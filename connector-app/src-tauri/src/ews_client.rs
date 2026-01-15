use reqwest::Client;
use serde::{Deserialize, Serialize};
use tracing::{error, info};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CalendarEvent {
    pub id: String,
    pub subject: String,
    pub body: Option<String>,
    pub start: String,
    pub end: String,
    pub location: Option<String>,
    pub organizer: Option<String>,
    pub attendees: Vec<Attendee>,
    pub is_recurring: bool,
    pub is_cancelled: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attendee {
    pub email: String,
    pub name: Option<String>,
    pub response: Option<String>,
    pub optional: bool,
}

pub struct EWSClient {
    url: String,
    username: String,
    password: String,
    client: Client,
}

impl EWSClient {
    pub fn new(url: &str, username: &str, password: &str) -> Self {
        let client = Client::builder()
            .danger_accept_invalid_certs(true)  // For self-signed certs
            .build()
            .expect("Failed to create HTTP client");

        Self {
            url: url.to_string(),
            username: username.to_string(),
            password: password.to_string(),
            client,
        }
    }

    pub async fn test_connection(&self) -> Result<(), String> {
        // Try to get folder list as a simple test
        let soap = r#"<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:GetFolder>
      <m:FolderShape>
        <t:BaseShape>IdOnly</t:BaseShape>
      </m:FolderShape>
      <m:FolderIds>
        <t:DistinguishedFolderId Id="calendar"/>
      </m:FolderIds>
    </m:GetFolder>
  </soap:Body>
</soap:Envelope>"#;

        let response = self
            .client
            .post(&self.url)
            .basic_auth(&self.username, Some(&self.password))
            .header("Content-Type", "text/xml; charset=utf-8")
            .body(soap)
            .send()
            .await
            .map_err(|e| format!("EWS request failed: {}", e))?;

        if response.status().is_success() {
            info!("EWS connection successful");
            Ok(())
        } else {
            let status = response.status();
            let body = response.text().await.unwrap_or_default();
            error!("EWS connection failed: {} - {}", status, body);
            Err(format!("EWS returned {}", status))
        }
    }

    pub async fn get_calendar_events(
        &self,
        email: &str,
        days_back: i32,
        days_forward: i32,
    ) -> Result<Vec<CalendarEvent>, String> {
        let now = chrono::Utc::now();
        let start = now - chrono::Duration::days(days_back as i64);
        let end = now + chrono::Duration::days(days_forward as i64);

        let soap = format!(
            r#"<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
    <t:ExchangeImpersonation>
      <t:ConnectingSID>
        <t:SmtpAddress>{}</t:SmtpAddress>
      </t:ConnectingSID>
    </t:ExchangeImpersonation>
  </soap:Header>
  <soap:Body>
    <m:FindItem Traversal="Shallow">
      <m:ItemShape>
        <t:BaseShape>Default</t:BaseShape>
        <t:AdditionalProperties>
          <t:FieldURI FieldURI="item:Subject"/>
          <t:FieldURI FieldURI="item:Body"/>
          <t:FieldURI FieldURI="calendar:Start"/>
          <t:FieldURI FieldURI="calendar:End"/>
          <t:FieldURI FieldURI="calendar:Location"/>
          <t:FieldURI FieldURI="calendar:Organizer"/>
          <t:FieldURI FieldURI="calendar:RequiredAttendees"/>
          <t:FieldURI FieldURI="calendar:OptionalAttendees"/>
          <t:FieldURI FieldURI="calendar:IsRecurring"/>
          <t:FieldURI FieldURI="calendar:IsCancelled"/>
        </t:AdditionalProperties>
      </m:ItemShape>
      <m:CalendarView MaxEntriesReturned="1000" StartDate="{}" EndDate="{}"/>
      <m:ParentFolderIds>
        <t:DistinguishedFolderId Id="calendar">
          <t:Mailbox>
            <t:EmailAddress>{}</t:EmailAddress>
          </t:Mailbox>
        </t:DistinguishedFolderId>
      </m:ParentFolderIds>
    </m:FindItem>
  </soap:Body>
</soap:Envelope>"#,
            email,
            start.format("%Y-%m-%dT%H:%M:%SZ"),
            end.format("%Y-%m-%dT%H:%M:%SZ"),
            email
        );

        let response = self
            .client
            .post(&self.url)
            .basic_auth(&self.username, Some(&self.password))
            .header("Content-Type", "text/xml; charset=utf-8")
            .body(soap)
            .send()
            .await
            .map_err(|e| format!("EWS request failed: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("EWS returned {}", response.status()));
        }

        let body = response.text().await.map_err(|e| format!("Failed to read response: {}", e))?;

        // Parse XML response
        let events = self.parse_calendar_response(&body)?;

        info!("Fetched {} calendar events for {}", events.len(), email);
        Ok(events)
    }

    fn parse_calendar_response(&self, xml: &str) -> Result<Vec<CalendarEvent>, String> {
        // Simple XML parsing - extract calendar items
        // In production, use a proper XML parser like quick-xml with serde
        let mut events = Vec::new();

        // Basic regex-style extraction (simplified for demo)
        // In real implementation, use quick-xml properly
        for item in xml.split("<t:CalendarItem").skip(1) {
            if let Some(end_idx) = item.find("</t:CalendarItem>") {
                let item_xml = &item[..end_idx];

                let event = CalendarEvent {
                    id: extract_xml_value(item_xml, "ItemId Id=\"", "\"").unwrap_or_default(),
                    subject: extract_xml_value(item_xml, "<t:Subject>", "</t:Subject>")
                        .unwrap_or_else(|| "Untitled".to_string()),
                    body: extract_xml_value(item_xml, "<t:Body", "</t:Body>")
                        .map(|b| b.split('>').last().unwrap_or(&b).to_string()),
                    start: extract_xml_value(item_xml, "<t:Start>", "</t:Start>").unwrap_or_default(),
                    end: extract_xml_value(item_xml, "<t:End>", "</t:End>").unwrap_or_default(),
                    location: extract_xml_value(item_xml, "<t:Location>", "</t:Location>"),
                    organizer: extract_xml_value(item_xml, "<t:EmailAddress>", "</t:EmailAddress>"),
                    attendees: vec![], // TODO: Parse attendees properly
                    is_recurring: item_xml.contains("<t:IsRecurring>true"),
                    is_cancelled: item_xml.contains("<t:IsCancelled>true"),
                };

                events.push(event);
            }
        }

        Ok(events)
    }
}

fn extract_xml_value(xml: &str, start_tag: &str, end_tag: &str) -> Option<String> {
    let start_idx = xml.find(start_tag)?;
    let value_start = start_idx + start_tag.len();
    let end_idx = xml[value_start..].find(end_tag)?;
    Some(xml[value_start..value_start + end_idx].to_string())
}
