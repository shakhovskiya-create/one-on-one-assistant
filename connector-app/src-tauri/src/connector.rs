use crate::ad_client::ADClient;
use crate::LogBuffer;
use futures_util::{SinkExt, StreamExt};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::sync::Arc;
use tokio::sync::Mutex;
use tokio_tungstenite::{connect_async, tungstenite::Message};
use tracing::{error, info, warn};

#[derive(Debug, Clone, Serialize)]
pub struct ConnectorState {
    pub running: bool,
    pub connected: bool,
    pub ad_connected: bool,
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
    state: Arc<Mutex<ConnectorState>>,
    logs: LogBuffer,
}

impl Connector {
    pub fn new(
        backend_url: &str,
        api_key: &str,
        ad_client: ADClient,
        logs: LogBuffer,
    ) -> Self {
        Self {
            backend_url: backend_url.to_string(),
            api_key: api_key.to_string(),
            ad_client,
            state: Arc::new(Mutex::new(ConnectorState {
                running: false,
                connected: false,
                ad_connected: false,
                last_error: None,
            })),
            logs,
        }
    }

    async fn log(&self, message: &str) {
        let timestamp = chrono::Local::now().format("%H:%M:%S");
        let log_line = format!("[{}] {}", timestamp, message);
        let mut logs = self.logs.lock().await;
        logs.push(log_line);
        // Keep only last 100 logs
        if logs.len() > 100 {
            logs.remove(0);
        }
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

        self.log("Starting connector...").await;

        // Test AD connection
        self.log("Testing AD connection...").await;
        match self.ad_client.test_connection().await {
            Ok(_) => {
                self.log("AD connection OK").await;
                self.state.lock().await.ad_connected = true;
            }
            Err(e) => {
                self.log(&format!("AD connection failed: {}", e)).await;
                self.state.lock().await.ad_connected = false;
            }
        }

        // Connect to backend WebSocket
        self.log("Connecting to backend...").await;
        let url = format!("{}?token={}", self.backend_url, self.api_key);

        match connect_async(&url).await {
            Ok((ws_stream, _)) => {
                self.log("Connected to backend").await;
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
                            self.log("Backend closed connection").await;
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
                self.log(&err).await;
                self.state.lock().await.last_error = Some(err.clone());
                return Err(err);
            }
        }

        Ok(())
    }

    pub async fn stop(&self) {
        self.log("Stopping connector...").await;
        let mut state = self.state.lock().await;
        state.running = false;
        state.connected = false;
    }

    async fn handle_command(&self, cmd: &Command) -> Value {
        info!("Received command: {} ({})", cmd.command, cmd.request_id);
        self.log(&format!("Command: {}", cmd.command)).await;

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

                self.log("Starting AD sync...").await;

                let result = tokio::time::timeout(
                    std::time::Duration::from_secs(300), // 5 minute timeout
                    self.ad_client.get_all_users(require_dept, true, include_photo)
                ).await;

                match result {
                    Ok(Ok((users, stats))) => {
                        self.log(&format!(
                            "Synced {} users ({} with dept)",
                            stats.returned, stats.with_department
                        )).await;
                        json!({
                            "users": users,
                            "total": stats.total_in_ad,
                            "stats": stats,
                            "has_more": false
                        })
                    }
                    Ok(Err(e)) => {
                        self.log(&format!("Sync failed: {}", e)).await;
                        json!({ "error": e })
                    }
                    Err(_) => {
                        self.log("Sync timed out after 5 minutes").await;
                        json!({ "error": "Sync timed out" })
                    }
                }
            }
            "authenticate" => {
                let params = cmd.params.as_ref().unwrap();
                let username = params.get("username").and_then(|v| v.as_str()).unwrap_or("");
                let password = params.get("password").and_then(|v| v.as_str()).unwrap_or("");

                match self.ad_client.authenticate(username, password).await {
                    Ok(Some(user)) => {
                        self.log(&format!("Auth success: {}", username)).await;
                        json!({ "authenticated": true, "user": user })
                    }
                    Ok(None) => {
                        self.log(&format!("Auth failed: {}", username)).await;
                        json!({ "authenticated": false, "error": "Invalid credentials" })
                    }
                    Err(e) => {
                        json!({ "authenticated": false, "error": e })
                    }
                }
            }
            "get_calendar" => {
                // Calendar is now handled directly by the backend
                json!({ "error": "Calendar integration moved to cloud backend" })
            }
            _ => {
                warn!("Unknown command: {}", cmd.command);
                json!({ "error": format!("Unknown command: {}", cmd.command) })
            }
        };

        let success = result.get("error").is_none();
        let error = result.get("error").and_then(|e| e.as_str()).map(|s| s.to_string());

        json!({
            "type": "response",
            "request_id": cmd.request_id,
            "command": cmd.command,
            "success": success,
            "error": error,
            "result": result,
            "timestamp": chrono::Utc::now().to_rfc3339()
        })
    }
}
