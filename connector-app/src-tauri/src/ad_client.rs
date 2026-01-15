use ldap3::{Ldap, LdapConnAsync, LdapConnSettings, Scope, SearchEntry, controls::{MakeCritical, SimplePagedResults}};
use serde::{Deserialize, Serialize};
use std::collections::HashSet;
use tracing::{error, info};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ADUser {
    pub dn: String,
    pub name: Option<String>,
    pub email: Option<String>,
    pub login: Option<String>,
    pub title: Option<String>,
    pub department: Option<String>,
    pub manager_dn: Option<String>,
    pub phone: Option<String>,
    pub mobile: Option<String>,
    pub photo_base64: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
pub struct SyncStats {
    pub total_in_ad: usize,
    pub with_department: usize,
    pub without_department: usize,
    pub with_email: usize,
    pub without_email: usize,
    pub filtered_out: usize,
    pub returned: usize,
}

pub struct ADClient {
    server: String,
    port: u16,
    use_ssl: bool,
    bind_user: String,
    bind_password: String,
    base_dn: String,
    users_ou: String,
}

impl ADClient {
    pub fn new(
        server: &str,
        port: u16,
        use_ssl: bool,
        bind_user: &str,
        bind_password: &str,
        base_dn: &str,
    ) -> Self {
        // Derive users OU from base_dn
        let users_ou = format!("OU=EKF-USERS,{}", base_dn);

        Self {
            server: server.to_string(),
            port,
            use_ssl,
            bind_user: bind_user.to_string(),
            bind_password: bind_password.to_string(),
            base_dn: base_dn.to_string(),
            users_ou,
        }
    }

    async fn connect(&self) -> Result<Ldap, String> {
        let url = if self.use_ssl {
            format!("ldaps://{}:{}", self.server, self.port)
        } else {
            format!("ldap://{}:{}", self.server, self.port)
        };

        info!("Connecting to AD: {}", url);

        let settings = LdapConnSettings::new();
        let (conn, mut ldap) = LdapConnAsync::with_settings(settings, &url)
            .await
            .map_err(|e| format!("Failed to connect to AD: {}", e))?;

        // Spawn connection handler
        tokio::spawn(async move {
            if let Err(e) = conn.drive().await {
                error!("LDAP connection error: {}", e);
            }
        });

        // Bind with credentials
        ldap.simple_bind(&self.bind_user, &self.bind_password)
            .await
            .map_err(|e| format!("Failed to bind to AD: {}", e))?
            .success()
            .map_err(|e| format!("Bind failed: {}", e))?;

        info!("Connected to AD successfully");
        Ok(ldap)
    }

    pub async fn test_connection(&self) -> Result<(), String> {
        let mut ldap = self.connect().await?;
        ldap.unbind().await.ok();
        Ok(())
    }

    pub async fn get_all_users(
        &self,
        require_department: bool,
        require_email: bool,
        include_photo: bool,
    ) -> Result<(Vec<ADUser>, SyncStats), String> {
        let mut ldap = self.connect().await?;

        // Search for active users
        let filter = "(&(objectClass=user)(objectCategory=person)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))";

        let mut attrs = vec![
            "cn",
            "mail",
            "sAMAccountName",
            "userPrincipalName",
            "title",
            "department",
            "manager",
            "telephoneNumber",
            "mobile",
        ];

        if include_photo {
            attrs.push("thumbnailPhoto");
        }

        let mut stats = SyncStats {
            total_in_ad: 0,
            with_department: 0,
            without_department: 0,
            with_email: 0,
            without_email: 0,
            filtered_out: 0,
            returned: 0,
        };

        let mut users = Vec::new();

        // Use paged search to handle large directories (page size 500)
        let mut page_ctrl = SimplePagedResults::new(500, false);

        loop {
            let controls = vec![page_ctrl.clone().critical().into()];

            let search_result = ldap
                .with_controls(controls)
                .search(&self.base_dn, Scope::Subtree, filter, attrs.clone())
                .await
                .map_err(|e| format!("Search failed: {}", e))?;

            let (rs, res) = search_result
                .success()
                .map_err(|e| format!("Search result error: {}", e))?;

            stats.total_in_ad += rs.len();

            for entry in rs {
                let entry = SearchEntry::construct(entry);

                let department = entry.attrs.get("department").and_then(|v| v.first().cloned());
                let email = entry.attrs.get("mail").and_then(|v| v.first().cloned());

                // Update stats
                if department.is_some() {
                    stats.with_department += 1;
                } else {
                    stats.without_department += 1;
                }

                if email.is_some() {
                    stats.with_email += 1;
                } else {
                    stats.without_email += 1;
                }

                // Apply filters
                if require_department && department.is_none() {
                    stats.filtered_out += 1;
                    continue;
                }

                if require_email && email.is_none() {
                    stats.filtered_out += 1;
                    continue;
                }

                // Parse photo
                let photo_base64 = if include_photo {
                    entry
                        .bin_attrs
                        .get("thumbnailPhoto")
                        .and_then(|photos| photos.first())
                        .map(|photo| base64::Engine::encode(&base64::engine::general_purpose::STANDARD, photo))
                } else {
                    None
                };

                let user = ADUser {
                    dn: entry.dn,
                    name: entry.attrs.get("cn").and_then(|v| v.first().cloned()),
                    email,
                    login: entry.attrs.get("sAMAccountName").and_then(|v| v.first().cloned()),
                    title: entry.attrs.get("title").and_then(|v| v.first().cloned()),
                    department,
                    manager_dn: entry.attrs.get("manager").and_then(|v| v.first().cloned()),
                    phone: entry.attrs.get("telephoneNumber").and_then(|v| v.first().cloned()),
                    mobile: entry.attrs.get("mobile").and_then(|v| v.first().cloned()),
                    photo_base64,
                };

                users.push(user);
            }

            // Check if there are more pages
            if let Some(cookie) = SimplePagedResults::from(&res) {
                if cookie.is_empty() {
                    break;
                }
                page_ctrl.set_cookie(cookie);
            } else {
                break;
            }
        }

        stats.returned = users.len();

        ldap.unbind().await.ok();

        info!(
            "AD sync: {} total, {} with dept, {} filtered, {} returned",
            stats.total_in_ad, stats.with_department, stats.filtered_out, stats.returned
        );

        Ok((users, stats))
    }

    pub async fn authenticate(&self, username: &str, password: &str) -> Result<Option<ADUser>, String> {
        // Try to bind with user credentials
        let url = if self.use_ssl {
            format!("ldaps://{}:{}", self.server, self.port)
        } else {
            format!("ldap://{}:{}", self.server, self.port)
        };

        let settings = LdapConnSettings::new();
        let (conn, mut ldap) = LdapConnAsync::with_settings(settings, &url)
            .await
            .map_err(|e| format!("Failed to connect: {}", e))?;

        tokio::spawn(async move {
            conn.drive().await.ok();
        });

        // Try NTLM-style bind
        let result = ldap.simple_bind(username, password).await;

        match result {
            Ok(res) => {
                if res.rc == 0 {
                    // Success - now fetch user info
                    let sam = username.split('\\').last().unwrap_or(username);
                    let filter = format!("(sAMAccountName={})", sam);

                    let (rs, _) = ldap
                        .search(
                            &self.base_dn,
                            Scope::Subtree,
                            &filter,
                            vec!["cn", "mail", "title", "department", "manager"],
                        )
                        .await
                        .map_err(|e| format!("Search failed: {}", e))?
                        .success()
                        .map_err(|e| format!("Search error: {}", e))?;

                    ldap.unbind().await.ok();

                    if let Some(entry) = rs.into_iter().next() {
                        let entry = SearchEntry::construct(entry);
                        return Ok(Some(ADUser {
                            dn: entry.dn,
                            name: entry.attrs.get("cn").and_then(|v| v.first().cloned()),
                            email: entry.attrs.get("mail").and_then(|v| v.first().cloned()),
                            login: Some(sam.to_string()),
                            title: entry.attrs.get("title").and_then(|v| v.first().cloned()),
                            department: entry.attrs.get("department").and_then(|v| v.first().cloned()),
                            manager_dn: entry.attrs.get("manager").and_then(|v| v.first().cloned()),
                            phone: None,
                            mobile: None,
                            photo_base64: None,
                        }));
                    }
                }
                ldap.unbind().await.ok();
                Ok(None)
            }
            Err(_) => {
                ldap.unbind().await.ok();
                Ok(None)
            }
        }
    }
}
