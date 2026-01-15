use crate::ad_client::ADClient;
use crate::ews_client::EWSClient;
use futures_util::{SinkExt, StreamExt};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::sync::Arc;
use tokio::sync::{broadcast, Mutex};
use tokio_tungstenite::{connect_async, tungstenite::Message};
use tracing::{error, info, warn};

#[derive(Debug, Clone, Serialize)]
pub struct ConnectorState {
    pub running: bool,
    pub connected: bool,
    pub ad_connected: bool,
    pub exchange_connected: bool,
    pub last_error: Option<String>,
}

#[derive(Debug, Deserialize)]
struct Command {
    command: String,
    request_id: String,
    params: Option<Value>,
}

pub struct Connector {
    backend_url: String,
    api_key: String,
    ad_client: ADClient,
    ews_client: EWSClient,
    state: Arc<Mutex<ConnectorState>>,
    log_tx: broadcast::Sender<String>,
}

impl Connector {
    pub fn new(
        backend_url: &str,
        api_key: &str,
        ad_client: ADClient,
        ews_client: EWSClient,
        log_tx: broadcast::Sender<String>,
    ) -> Self {
        Self {
            backend_url: backend_url.to_string(),
            api_key: api_key.to_string(),
            ad_client,
            ews_client,
            state: Arc::new(Mutex::new(ConnectorState {
                running: false,
                connected: false,
                ad_connected: false,
                exchange_connected: false,
                last_error: None,
            })),
            log_tx,
        }
    }

    fn log(&self, message: &str) {
        let timestamp = chrono::Local::now().format("%H:%M:%S");
        let log_line = format!("[{}] {}", timestamp, message);
        self.log_tx.send(log_line).ok();
    }

    pub async fn get_state(&self) -> ConnectorState {
        self.state.lock().await.clone()
    }

    pub async fn start(&self) -> Result<(), String> {
        {
            let mut state = self.state.lock().await;
            if state.running {
                return Err("Already running".to_string());
            }
            state.running = true;
        }

        self.log("Starting connector...");

        // Test AD connection
        self.log("Testing AD connection...");
        match self.ad_client.test_connection().await {
            Ok(_) => {
                self.log("AD connection OK");
                self.state.lock().await.ad_connected = true;
            }
            Err(e) => {
                self.log(&format!("AD connection failed: {}", e));
                self.state.lock().await.ad_connected = false;
            }
        }

        // Test EWS connection
        self.log("Testing Exchange connection...");
        match self.ews_client.test_connection().await {
            Ok(_) => {
                self.log("Exchange connection OK");
                self.state.lock().await.exchange_connected = true;
            }
            Err(e) => {
                self.log(&format!("Exchange connection failed: {}", e));
                self.state.lock().await.exchange_connected = false;
            }
        }

        // Connect to backend WebSocket
        self.log("Connecting to backend...");
        let url = format!("{}?token={}", self.backend_url, self.api_key);

        match connect_async(&url).await {
            Ok((ws_stream, _)) => {
                self.log("Connected to backend");
                self.state.lock().await.connected = true;

                let (mut write, mut read) = ws_stream.split();

                // Message handling loop
                while let Some(msg) = read.next().await {
                    match msg {
                        Ok(Message::Text(text)) => {
                            if let Ok(cmd) = serde_json::from_str::<Command>(&text) {
                                let response = self.handle_command(&cmd).await;
                                let response_json = serde_json::to_string(&response).unwrap();
                                if let Err(e) = write.send(Message::Text(response_json.into())).await {
                                    error!("Failed to send response: {}", e);
                                }
                            }
                        }
                        Ok(Message::Ping(data)) => {
                            write.send(Message::Pong(data)).await.ok();
                        }
                        Ok(Message::Close(_)) => {
                            self.log("Backend closed connection");
                            break;
                        }
                        Err(e) => {
                            error!("WebSocket error: {}", e);
                            break;
                        }
                        _ => {}
                    }
                }

                self.state.lock().await.connected = false;
            }
            Err(e) => {
                let err = format!("Failed to connect to backend: {}", e);
                self.log(&err);
                self.state.lock().await.last_error = Some(err.clone());
                return Err(err);
            }
        }

        Ok(())
    }

    pub async fn stop(&self) {
        self.log("Stopping connector...");
        let mut state = self.state.lock().await;
        state.running = false;
        state.connected = false;
    }

    async fn handle_command(&self, cmd: &Command) -> Value {
        info!("Received command: {} ({})", cmd.command, cmd.request_id);
        self.log(&format!("Command: {}", cmd.command));

        let result = match cmd.command.as_str() {
            "ping" => {
                json!({
                    "pong": true,
                    "timestamp": chrono::Utc::now().to_rfc3339()
                })
            }
            "sync_users" => {
                let params = cmd.params.as_ref();
                let require_dept = params
                    .and_then(|p| p.get("require_department"))
                    .and_then(|v| v.as_bool())
                    .unwrap_or(true);
                let include_photo = params
                    .and_then(|p| p.get("include_photo"))
                    .and_then(|v| v.as_bool())
                    .unwrap_or(true);

                match self.ad_client.get_all_users(require_dept, true, include_photo).await {
                    Ok((users, stats)) => {
                        self.log(&format!(
                            "Synced {} users ({} with dept)",
                            stats.returned, stats.with_department
                        ));
                        json!({
                            "users": users,
                            "total": stats.total_in_ad,
                            "stats": stats,
                            "has_more": false
                        })
                    }
                    Err(e) => {
                        self.log(&format!("Sync failed: {}", e));
                        json!({ "error": e })
                    }
                }
            }
            "authenticate" => {
                let params = cmd.params.as_ref().unwrap();
                let username = params.get("username").and_then(|v| v.as_str()).unwrap_or("");
                let password = params.get("password").and_then(|v| v.as_str()).unwrap_or("");

                match self.ad_client.authenticate(username, password).await {
                    Ok(Some(user)) => {
                        self.log(&format!("Auth success: {}", username));
                        json!({ "authenticated": true, "user": user })
                    }
                    Ok(None) => {
                        self.log(&format!("Auth failed: {}", username));
                        json!({ "authenticated": false, "error": "Invalid credentials" })
                    }
                    Err(e) => {
                        json!({ "authenticated": false, "error": e })
                    }
                }
            }
            "get_calendar" => {
                let params = cmd.params.as_ref().unwrap();
                let email = params.get("email").and_then(|v| v.as_str()).unwrap_or("");
                let days_back = params.get("days_back").and_then(|v| v.as_i64()).unwrap_or(7) as i32;
                let days_forward = params.get("days_forward").and_then(|v| v.as_i64()).unwrap_or(30) as i32;

                match self.ews_client.get_calendar_events(email, days_back, days_forward).await {
                    Ok(events) => {
                        self.log(&format!("Fetched {} events for {}", events.len(), email));
                        json!(events)
                    }
                    Err(e) => {
                        self.log(&format!("Calendar fetch failed: {}", e));
                        json!([])
                    }
                }
            }
            _ => {
                warn!("Unknown command: {}", cmd.command);
                json!({ "error": format!("Unknown command: {}", cmd.command) })
            }
        };

        json!({
            "type": "response",
            "request_id": cmd.request_id,
            "command": cmd.command,
            "success": !result.get("error").is_some(),
            "result": result,
            "timestamp": chrono::Utc::now().to_rfc3339()
        })
    }
}
