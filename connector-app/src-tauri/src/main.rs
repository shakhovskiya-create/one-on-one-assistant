#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod ad_client;
mod connector;
mod ews_client;

use ad_client::ADClient;
use connector::Connector;
use ews_client::EWSClient;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tauri::State;
use tokio::sync::{broadcast, Mutex};
use tracing_subscriber::EnvFilter;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectorConfig {
    pub backend_url: String,
    pub api_key: String,
    pub ad_server: String,
    pub ad_port: u16,
    pub ad_use_ssl: bool,
    pub ad_bind_user: String,
    pub ad_bind_password: String,
    pub ad_base_dn: String,
    pub ews_url: String,
    pub ews_username: String,
    pub ews_password: String,
}

impl Default for ConnectorConfig {
    fn default() -> Self {
        Self {
            backend_url: "wss://one-on-one-back.up.railway.app/ws/connector".to_string(),
            api_key: String::new(),
            ad_server: "172.20.0.33".to_string(),
            ad_port: 389,
            ad_use_ssl: false,
            ad_bind_user: String::new(),
            ad_bind_password: String::new(),
            ad_base_dn: "DC=ekfgroup,DC=ru".to_string(),
            ews_url: "https://post.ekf.su/EWS/Exchange.asmx".to_string(),
            ews_username: String::new(),
            ews_password: String::new(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Default)]
pub struct ConnectorStatus {
    pub running: bool,
    pub connected: bool,
    pub ad_connected: bool,
    pub exchange_connected: bool,
    pub last_error: Option<String>,
    pub logs: Vec<String>,
}

struct AppState {
    config: Mutex<ConnectorConfig>,
    status: Mutex<ConnectorStatus>,
    connector: Mutex<Option<Arc<Connector>>>,
    log_tx: broadcast::Sender<String>,
}

#[tauri::command]
async fn get_config(state: State<'_, AppState>) -> Result<ConnectorConfig, String> {
    Ok(state.config.lock().await.clone())
}

#[tauri::command]
async fn save_config(config: ConnectorConfig, state: State<'_, AppState>) -> Result<(), String> {
    let mut current = state.config.lock().await;
    *current = config;
    Ok(())
}

#[tauri::command]
async fn get_status(state: State<'_, AppState>) -> Result<ConnectorStatus, String> {
    let mut status = state.status.lock().await.clone();

    // Get recent logs
    let mut log_rx = state.log_tx.subscribe();
    while let Ok(log) = log_rx.try_recv() {
        status.logs.push(log);
        if status.logs.len() > 100 {
            status.logs.remove(0);
        }
    }

    // Update from connector state if running
    if let Some(connector) = state.connector.lock().await.as_ref() {
        let conn_state = connector.get_state().await;
        status.running = conn_state.running;
        status.connected = conn_state.connected;
        status.ad_connected = conn_state.ad_connected;
        status.exchange_connected = conn_state.exchange_connected;
        status.last_error = conn_state.last_error;
    }

    Ok(status)
}

#[tauri::command]
async fn start_connector(state: State<'_, AppState>) -> Result<(), String> {
    let config = state.config.lock().await.clone();

    // Validate config
    if config.api_key.is_empty() {
        return Err("API Key не задан".to_string());
    }
    if config.ad_bind_user.is_empty() || config.ad_bind_password.is_empty() {
        return Err("Учетные данные AD не заданы".to_string());
    }

    // Create clients
    let ad_client = ADClient::new(
        &config.ad_server,
        config.ad_port,
        config.ad_use_ssl,
        &config.ad_bind_user,
        &config.ad_bind_password,
        &config.ad_base_dn,
    );

    let ews_client = EWSClient::new(&config.ews_url, &config.ews_username, &config.ews_password);

    // Create connector
    let connector = Arc::new(Connector::new(
        &config.backend_url,
        &config.api_key,
        ad_client,
        ews_client,
        state.log_tx.clone(),
    ));

    *state.connector.lock().await = Some(connector.clone());

    // Start in background
    tokio::spawn(async move {
        if let Err(e) = connector.start().await {
            tracing::error!("Connector error: {}", e);
        }
    });

    Ok(())
}

#[tauri::command]
async fn stop_connector(state: State<'_, AppState>) -> Result<(), String> {
    if let Some(connector) = state.connector.lock().await.as_ref() {
        connector.stop().await;
    }
    *state.connector.lock().await = None;
    Ok(())
}

#[tauri::command]
async fn clear_logs(state: State<'_, AppState>) -> Result<(), String> {
    let mut status = state.status.lock().await;
    status.logs.clear();
    Ok(())
}

#[tauri::command]
async fn test_ad(state: State<'_, AppState>) -> Result<String, String> {
    let config = state.config.lock().await.clone();

    let ad_client = ADClient::new(
        &config.ad_server,
        config.ad_port,
        config.ad_use_ssl,
        &config.ad_bind_user,
        &config.ad_bind_password,
        &config.ad_base_dn,
    );

    ad_client.test_connection().await.map(|_| "OK".to_string())
}

#[tauri::command]
async fn test_exchange(state: State<'_, AppState>) -> Result<String, String> {
    let config = state.config.lock().await.clone();

    let ews_client = EWSClient::new(&config.ews_url, &config.ews_username, &config.ews_password);

    ews_client.test_connection().await.map(|_| "OK".to_string())
}

fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            EnvFilter::from_default_env()
                .add_directive("ekf_connector=info".parse().unwrap()),
        )
        .init();

    let (log_tx, _) = broadcast::channel::<String>(100);

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(AppState {
            config: Mutex::new(ConnectorConfig::default()),
            status: Mutex::new(ConnectorStatus::default()),
            connector: Mutex::new(None),
            log_tx,
        })
        .invoke_handler(tauri::generate_handler![
            get_config,
            save_config,
            get_status,
            start_connector,
            stop_connector,
            clear_logs,
            test_ad,
            test_exchange
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
