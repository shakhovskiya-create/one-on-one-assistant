package ews

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Client handles Exchange Web Services requests
type Client struct {
	URL    string
	Domain string
	client *http.Client
}

// CalendarEvent represents a calendar event from Exchange
type CalendarEvent struct {
	ID          string     `json:"id"`
	Subject     string     `json:"subject"`
	Start       string     `json:"start"`
	End         string     `json:"end"`
	Location    string     `json:"location,omitempty"`
	Organizer   *Person    `json:"organizer,omitempty"`
	Attendees   []Attendee `json:"attendees,omitempty"`
	IsRecurring bool       `json:"is_recurring"`
	IsCancelled bool       `json:"is_cancelled"`
}

// Person represents a person (organizer)
type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Attendee represents a meeting attendee
type Attendee struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Response string `json:"response,omitempty"`
	Optional bool   `json:"optional"`
}

// BusyTime represents a busy time slot
type BusyTime struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Status string `json:"status"`
}

// EmailMessage represents an email from Exchange
type EmailMessage struct {
	ID             string   `json:"id"`
	ChangeKey      string   `json:"change_key,omitempty"`
	ConversationID string   `json:"conversation_id,omitempty"`
	ItemClass      string   `json:"item_class,omitempty"`
	Subject        string   `json:"subject"`
	From           *Person  `json:"from,omitempty"`
	To             []Person `json:"to,omitempty"`
	CC             []Person `json:"cc,omitempty"`
	Body           string   `json:"body"`
	BodyPreview    string   `json:"body_preview,omitempty"`
	ReceivedAt     string   `json:"received_at"`
	IsRead         bool     `json:"is_read"`
	HasAttach      bool     `json:"has_attachments"`
	FolderID       string   `json:"folder_id,omitempty"`
}

// MailFolder represents a mail folder
type MailFolder struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	UnreadCount int    `json:"unread_count"`
	TotalCount  int    `json:"total_count"`
}

// Attachment represents an email attachment
type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	IsInline    bool   `json:"is_inline"`
	ContentID   string `json:"content_id,omitempty"`
}

// NewClient creates a new EWS client
// skipTLSVerify should only be true for development/internal certificates
func NewClient(url, domain string, skipTLSVerify bool) *Client {
	return &Client{
		URL:    url,
		Domain: domain,
		client: &http.Client{
			Timeout: 120 * time.Second, // Увеличен таймаут для медленных Exchange серверов
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: skipTLSVerify},
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
	}
}

// GetCalendarEvents fetches calendar events from Exchange
func (c *Client) GetCalendarEvents(email, username, password string, daysBack, daysForward int) ([]CalendarEvent, error) {
	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -daysBack).Format("2006-01-02T15:04:05Z")
	startDate = strings.Replace(startDate, now.Format("15:04:05"), "00:00:00", 1)
	endDate := now.AddDate(0, 0, daysForward).Format("2006-01-02T15:04:05Z")
	endDate = strings.Replace(endDate, now.AddDate(0, 0, daysForward).Format("15:04:05"), "23:59:59", 1)

	// Always specify the mailbox explicitly to avoid ambiguity
	var folderSpec string
	if email != "" {
		folderSpec = fmt.Sprintf(`<t:DistinguishedFolderId Id="calendar">
          <t:Mailbox>
            <t:EmailAddress>%s</t:EmailAddress>
          </t:Mailbox>
        </t:DistinguishedFolderId>`, email)
	} else {
		folderSpec = `<t:DistinguishedFolderId Id="calendar"/>`
	}

	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
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
      <m:CalendarView MaxEntriesReturned="200" StartDate="%s" EndDate="%s"/>
      <m:ParentFolderIds>
        %s
      </m:ParentFolderIds>
    </m:FindItem>
  </soap:Body>
</soap:Envelope>`, startDate, endDate, folderSpec)

	log.Printf("DEBUG SOAP Request for email=%s, username=%s:\n%s", email, username, soap)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG Exchange Response (first 2000 chars):\n%s", string(body[:min(2000, len(body))]))

	events := c.parseCalendarResponse(string(body))
	log.Printf("DEBUG Parsed %d events from Exchange response", len(events))

	return events, nil
}

// GetFreeBusy gets free/busy information for multiple users
func (c *Client) GetFreeBusy(emails []string, username, password, startDate, endDate string) (map[string][]BusyTime, error) {
	var mailboxes strings.Builder
	for _, email := range emails {
		mailboxes.WriteString(fmt.Sprintf(`<t:MailboxData>
			<t:Email><t:Address>%s</t:Address></t:Email>
			<t:AttendeeType>Required</t:AttendeeType>
		</t:MailboxData>`, email))
	}

	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:GetUserAvailabilityRequest>
      <t:TimeZone>
        <t:Bias>-180</t:Bias>
        <t:StandardTime><t:Bias>0</t:Bias><t:Time>00:00:00</t:Time><t:DayOrder>1</t:DayOrder><t:Month>1</t:Month><t:DayOfWeek>Sunday</t:DayOfWeek></t:StandardTime>
        <t:DaylightTime><t:Bias>0</t:Bias><t:Time>00:00:00</t:Time><t:DayOrder>1</t:DayOrder><t:Month>1</t:Month><t:DayOfWeek>Sunday</t:DayOfWeek></t:DaylightTime>
      </t:TimeZone>
      <m:MailboxDataArray>%s</m:MailboxDataArray>
      <t:FreeBusyViewOptions>
        <t:TimeWindow>
          <t:StartTime>%sT00:00:00</t:StartTime>
          <t:EndTime>%sT23:59:59</t:EndTime>
        </t:TimeWindow>
        <t:MergedFreeBusyIntervalInMinutes>30</t:MergedFreeBusyIntervalInMinutes>
        <t:RequestedView>FreeBusy</t:RequestedView>
      </t:FreeBusyViewOptions>
    </m:GetUserAvailabilityRequest>
  </soap:Body>
</soap:Envelope>`, mailboxes.String(), startDate, endDate)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	return c.parseFreeBusyResponse(string(body), emails), nil
}

func (c *Client) doRequest(soap, username, password string) ([]byte, error) {
	// Add domain to username if not present
	authUser := username
	if !strings.Contains(username, "\\") && !strings.Contains(username, "@") {
		authUser = c.Domain + "\\" + username
	}

	req, err := http.NewRequest("POST", c.URL, strings.NewReader(soap))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.SetBasicAuth(authUser, password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("EWS error %d: %s", resp.StatusCode, string(body)[:min(500, len(body))])
	}

	return body, nil
}

func (c *Client) parseCalendarResponse(xml string) []CalendarEvent {
	var events []CalendarEvent

	items := strings.Split(xml, "<t:CalendarItem>")
	log.Printf("DEBUG Parser: Split XML into %d parts by '<t:CalendarItem>'", len(items))

	for i, item := range items[1:] { // Skip first split
		endIdx := strings.Index(item, "</t:CalendarItem>")
		if endIdx == -1 {
			log.Printf("DEBUG Parser: Item %d has no closing tag </t:CalendarItem>", i)
			continue
		}
		itemXML := item[:endIdx]

		// Log first item for debugging
		if i == 0 {
			log.Printf("DEBUG Parser: First item XML (first 500 chars):\n%s", itemXML[:min(500, len(itemXML))])
		}

		event := CalendarEvent{
			ID:          extractValue(itemXML, `<t:ItemId Id="`, `"`),
			Subject:     extractValue(itemXML, "<t:Subject>", "</t:Subject>"),
			Start:       extractValue(itemXML, "<t:Start>", "</t:Start>"),
			End:         extractValue(itemXML, "<t:End>", "</t:End>"),
			Location:    extractValue(itemXML, "<t:Location>", "</t:Location>"),
			IsRecurring: strings.Contains(strings.ToLower(itemXML), "<t:isrecurring>true"),
			IsCancelled: strings.Contains(strings.ToLower(itemXML), "<t:iscancelled>true"),
		}

		if i == 0 {
			log.Printf("DEBUG Parser: First item - ID='%s', Subject='%s', Start='%s'", event.ID, event.Subject, event.Start)
		}

		if event.Subject == "" {
			event.Subject = "Без темы"
		}

		// Parse organizer
		if orgStart := strings.Index(itemXML, "<t:Organizer>"); orgStart != -1 {
			if orgEnd := strings.Index(itemXML[orgStart:], "</t:Organizer>"); orgEnd != -1 {
				orgXML := itemXML[orgStart : orgStart+orgEnd]
				event.Organizer = &Person{
					Name:  extractValue(orgXML, "<t:Name>", "</t:Name>"),
					Email: extractValue(orgXML, "<t:EmailAddress>", "</t:EmailAddress>"),
				}
			}
		}

		// Parse attendees
		event.Attendees = c.parseAttendees(itemXML)

		// Log attendees count for first few events
		if i < 3 {
			log.Printf("DEBUG Event '%s' has %d attendees", event.Subject, len(event.Attendees))
		}

		if event.Start != "" {
			events = append(events, event)
		}
	}

	return events
}

func (c *Client) parseAttendees(xml string) []Attendee {
	var attendees []Attendee

	// Debug: check if attendees sections exist
	hasRequired := strings.Contains(xml, "<t:RequiredAttendees>")
	hasOptional := strings.Contains(xml, "<t:OptionalAttendees>")
	log.Printf("DEBUG parseAttendees: hasRequiredAttendees=%v, hasOptionalAttendees=%v", hasRequired, hasOptional)

	for _, sectionTag := range []string{"<t:RequiredAttendees>", "<t:OptionalAttendees>"} {
		isOptional := strings.Contains(sectionTag, "Optional")

		sectionStart := strings.Index(xml, sectionTag)
		if sectionStart == -1 {
			continue
		}

		endTag := strings.Replace(sectionTag, "<", "</", 1)
		sectionEnd := strings.Index(xml[sectionStart:], endTag)
		if sectionEnd == -1 {
			continue
		}

		section := xml[sectionStart : sectionStart+sectionEnd]

		for _, attendeeXML := range strings.Split(section, "<t:Attendee>")[1:] {
			endIdx := strings.Index(attendeeXML, "</t:Attendee>")
			if endIdx == -1 {
				continue
			}

			email := extractValue(attendeeXML[:endIdx], "<t:EmailAddress>", "</t:EmailAddress>")
			if email != "" {
				attendees = append(attendees, Attendee{
					Name:     extractValue(attendeeXML[:endIdx], "<t:Name>", "</t:Name>"),
					Email:    email,
					Response: extractValue(attendeeXML[:endIdx], "<t:ResponseType>", "</t:ResponseType>"),
					Optional: isOptional,
				})
			}
		}
	}

	return attendees
}

func (c *Client) parseFreeBusyResponse(xml string, emails []string) map[string][]BusyTime {
	result := make(map[string][]BusyTime)

	responses := strings.Split(xml, "<FreeBusyResponse>")
	for i, resp := range responses[1:] {
		if i >= len(emails) {
			break
		}

		var busyTimes []BusyTime
		events := strings.Split(resp, "<CalendarEvent>")
		for _, event := range events[1:] {
			endIdx := strings.Index(event, "</CalendarEvent>")
			if endIdx == -1 {
				continue
			}

			start := extractValue(event[:endIdx], "<StartTime>", "</StartTime>")
			end := extractValue(event[:endIdx], "<EndTime>", "</EndTime>")
			status := extractValue(event[:endIdx], "<BusyType>", "</BusyType>")

			if start != "" && end != "" {
				busyTimes = append(busyTimes, BusyTime{
					Start:  start,
					End:    end,
					Status: status,
				})
			}
		}

		result[emails[i]] = busyTimes
	}

	return result
}

func extractValue(xml, startTag, endTag string) string {
	startIdx := strings.Index(xml, startTag)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startTag)

	endIdx := strings.Index(xml[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	value := xml[startIdx : startIdx+endIdx]
	// Unescape XML entities
	value = strings.ReplaceAll(value, "&amp;", "&")
	value = strings.ReplaceAll(value, "&lt;", "<")
	value = strings.ReplaceAll(value, "&gt;", ">")
	value = strings.ReplaceAll(value, "&quot;", "\"")
	value = strings.ReplaceAll(value, "&apos;", "'")

	// Remove HTML tags for body content
	if startTag == "<t:Body>" || strings.Contains(startTag, "Body") {
		re := regexp.MustCompile(`<[^>]*>`)
		value = re.ReplaceAllString(value, "")
	}

	return strings.TrimSpace(value)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetMailFolders fetches mail folders from Exchange
func (c *Client) GetMailFolders(email, username, password string) ([]MailFolder, error) {
	soap := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:FindFolder Traversal="Shallow">
      <m:FolderShape>
        <t:BaseShape>Default</t:BaseShape>
      </m:FolderShape>
      <m:ParentFolderIds>
        <t:DistinguishedFolderId Id="msgfolderroot"/>
      </m:ParentFolderIds>
    </m:FindFolder>
  </soap:Body>
</soap:Envelope>`

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	return c.parseMailFoldersResponse(string(body)), nil
}

// GetEmails fetches emails from a folder
func (c *Client) GetEmails(email, username, password, folderID string, limit int) ([]EmailMessage, error) {
	var folderSpec string
	if folderID != "" {
		folderSpec = fmt.Sprintf(`<t:FolderId Id="%s"/>`, folderID)
	} else {
		folderSpec = `<t:DistinguishedFolderId Id="inbox"/>`
	}

	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:FindItem Traversal="Shallow">
      <m:ItemShape>
        <t:BaseShape>Default</t:BaseShape>
        <t:AdditionalProperties>
          <t:FieldURI FieldURI="item:Subject"/>
          <t:FieldURI FieldURI="item:DateTimeReceived"/>
          <t:FieldURI FieldURI="item:ConversationId"/>
          <t:FieldURI FieldURI="item:ItemClass"/>
          <t:FieldURI FieldURI="message:From"/>
          <t:FieldURI FieldURI="message:ToRecipients"/>
          <t:FieldURI FieldURI="message:IsRead"/>
          <t:FieldURI FieldURI="item:HasAttachments"/>
        </t:AdditionalProperties>
      </m:ItemShape>
      <m:IndexedPageItemView MaxEntriesReturned="%d" Offset="0" BasePoint="Beginning"/>
      <m:SortOrder>
        <t:FieldOrder Order="Descending">
          <t:FieldURI FieldURI="item:DateTimeReceived"/>
        </t:FieldOrder>
      </m:SortOrder>
      <m:ParentFolderIds>
        %s
      </m:ParentFolderIds>
    </m:FindItem>
  </soap:Body>
</soap:Envelope>`, limit, folderSpec)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	return c.parseEmailsResponse(string(body)), nil
}

// GetEmailBody fetches the full body of an email
func (c *Client) GetEmailBody(username, password, itemID, changeKey string) (string, error) {
	// Build ItemId - ChangeKey is optional
	var itemIdElement string
	if changeKey != "" {
		itemIdElement = fmt.Sprintf(`<t:ItemId Id="%s" ChangeKey="%s"/>`, itemID, changeKey)
	} else {
		itemIdElement = fmt.Sprintf(`<t:ItemId Id="%s"/>`, itemID)
	}

	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:GetItem>
      <m:ItemShape>
        <t:BaseShape>Default</t:BaseShape>
        <t:BodyType>HTML</t:BodyType>
        <t:AdditionalProperties>
          <t:FieldURI FieldURI="item:Body"/>
        </t:AdditionalProperties>
      </m:ItemShape>
      <m:ItemIds>
        %s
      </m:ItemIds>
    </m:GetItem>
  </soap:Body>
</soap:Envelope>`, itemIdElement)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return "", err
	}

	// Extract body content - handle HTML body with attributes
	// The body tag looks like: <t:Body BodyType="HTML">content</t:Body>
	bodyStr := string(body)
	result := extractBodyContent(bodyStr)

	// Log for debugging
	if result == "" {
		log.Printf("GetEmailBody: empty body for item %s, response length: %d", itemID, len(bodyStr))
	}

	return result, nil
}

// extractBodyContent extracts the content from <t:Body ...>content</t:Body>
func extractBodyContent(xml string) string {
	// Find the start of Body tag
	startTag := "<t:Body"
	startIdx := strings.Index(xml, startTag)
	if startIdx == -1 {
		return ""
	}

	// Find the closing > of the opening tag
	closeTagIdx := strings.Index(xml[startIdx:], ">")
	if closeTagIdx == -1 {
		return ""
	}
	contentStart := startIdx + closeTagIdx + 1

	// Find the end tag
	endTag := "</t:Body>"
	endIdx := strings.Index(xml[contentStart:], endTag)
	if endIdx == -1 {
		return ""
	}

	content := xml[contentStart : contentStart+endIdx]

	// Unescape XML entities for the content
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&apos;", "'")

	return content
}

// GetAttachments fetches attachment list for an email
func (c *Client) GetAttachments(username, password, itemID, changeKey string) ([]Attachment, error) {
	var itemIdElement string
	if changeKey != "" {
		itemIdElement = fmt.Sprintf(`<t:ItemId Id="%s" ChangeKey="%s"/>`, itemID, changeKey)
	} else {
		itemIdElement = fmt.Sprintf(`<t:ItemId Id="%s"/>`, itemID)
	}

	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:GetItem>
      <m:ItemShape>
        <t:BaseShape>IdOnly</t:BaseShape>
        <t:AdditionalProperties>
          <t:FieldURI FieldURI="item:Attachments"/>
        </t:AdditionalProperties>
      </m:ItemShape>
      <m:ItemIds>
        %s
      </m:ItemIds>
    </m:GetItem>
  </soap:Body>
</soap:Envelope>`, itemIdElement)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	return c.parseAttachmentsResponse(string(body)), nil
}

// GetAttachmentContent fetches the content of a specific attachment
func (c *Client) GetAttachmentContent(username, password, attachmentID string) (string, string, []byte, error) {
	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:GetAttachment>
      <m:AttachmentIds>
        <t:AttachmentId Id="%s"/>
      </m:AttachmentIds>
    </m:GetAttachment>
  </soap:Body>
</soap:Envelope>`, attachmentID)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return "", "", nil, err
	}

	// Parse attachment response
	xml := string(body)
	name := extractValue(xml, "<t:Name>", "</t:Name>")
	contentType := extractValue(xml, "<t:ContentType>", "</t:ContentType>")
	contentB64 := extractValue(xml, "<t:Content>", "</t:Content>")

	// Decode base64 content
	var content []byte
	if contentB64 != "" {
		import_encoding_b64 := strings.ReplaceAll(contentB64, "\n", "")
		import_encoding_b64 = strings.ReplaceAll(import_encoding_b64, "\r", "")
		import_encoding_b64 = strings.ReplaceAll(import_encoding_b64, " ", "")
		// We'll return base64 string and let handler decode it
		content = []byte(import_encoding_b64)
	}

	return name, contentType, content, nil
}

func (c *Client) parseAttachmentsResponse(xml string) []Attachment {
	var attachments []Attachment

	// Parse FileAttachment
	for _, tag := range []string{"<t:FileAttachment>", "<t:ItemAttachment>"} {
		isItem := strings.Contains(tag, "ItemAttachment")
		items := strings.Split(xml, tag)
		for _, item := range items[1:] {
			var endTag string
			if isItem {
				endTag = "</t:ItemAttachment>"
			} else {
				endTag = "</t:FileAttachment>"
			}
			endIdx := strings.Index(item, endTag)
			if endIdx == -1 {
				continue
			}
			attXML := item[:endIdx]

			attachment := Attachment{
				ID:          extractValue(attXML, `<t:AttachmentId Id="`, `"`),
				Name:        extractValue(attXML, "<t:Name>", "</t:Name>"),
				ContentType: extractValue(attXML, "<t:ContentType>", "</t:ContentType>"),
				Size:        parseSize(extractValue(attXML, "<t:Size>", "</t:Size>")),
				IsInline:    strings.Contains(strings.ToLower(attXML), "<t:isinline>true"),
				ContentID:   extractValue(attXML, "<t:ContentId>", "</t:ContentId>"),
			}

			if attachment.ID != "" {
				attachments = append(attachments, attachment)
			}
		}
	}

	return attachments
}

func parseSize(s string) int {
	if s == "" {
		return 0
	}
	var size int
	fmt.Sscanf(s, "%d", &size)
	return size
}

// SendEmail sends an email via Exchange
func (c *Client) SendEmail(username, password, subject string, toEmails []string, body string, ccEmails []string) error {
	var toRecipients strings.Builder
	for _, email := range toEmails {
		toRecipients.WriteString(fmt.Sprintf(`<t:Mailbox><t:EmailAddress>%s</t:EmailAddress></t:Mailbox>`, email))
	}

	var ccRecipients strings.Builder
	for _, email := range ccEmails {
		ccRecipients.WriteString(fmt.Sprintf(`<t:Mailbox><t:EmailAddress>%s</t:EmailAddress></t:Mailbox>`, email))
	}

	ccSection := ""
	if len(ccEmails) > 0 {
		ccSection = fmt.Sprintf(`<t:CcRecipients>%s</t:CcRecipients>`, ccRecipients.String())
	}

	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:CreateItem MessageDisposition="SendAndSaveCopy">
      <m:SavedItemFolderId>
        <t:DistinguishedFolderId Id="sentitems"/>
      </m:SavedItemFolderId>
      <m:Items>
        <t:Message>
          <t:Subject>%s</t:Subject>
          <t:Body BodyType="HTML">%s</t:Body>
          <t:ToRecipients>%s</t:ToRecipients>
          %s
        </t:Message>
      </m:Items>
    </m:CreateItem>
  </soap:Body>
</soap:Envelope>`, escapeXML(subject), escapeXML(body), toRecipients.String(), ccSection)

	_, err := c.doRequest(soap, username, password)
	return err
}

// EmailAttachment represents an attachment to send
type EmailAttachment struct {
	Name    string
	Content []byte // Base64 will be encoded internally
}

// SendEmailWithAttachments sends an email with attachments via Exchange
func (c *Client) SendEmailWithAttachments(username, password, subject string, toEmails []string, body string, ccEmails []string, attachments []EmailAttachment) error {
	// If no attachments, use simple send
	if len(attachments) == 0 {
		return c.SendEmail(username, password, subject, toEmails, body, ccEmails)
	}

	// Step 1: Create the email as draft (SaveOnly)
	var toRecipients strings.Builder
	for _, email := range toEmails {
		toRecipients.WriteString(fmt.Sprintf(`<t:Mailbox><t:EmailAddress>%s</t:EmailAddress></t:Mailbox>`, email))
	}

	var ccRecipients strings.Builder
	for _, email := range ccEmails {
		ccRecipients.WriteString(fmt.Sprintf(`<t:Mailbox><t:EmailAddress>%s</t:EmailAddress></t:Mailbox>`, email))
	}

	ccSection := ""
	if len(ccEmails) > 0 {
		ccSection = fmt.Sprintf(`<t:CcRecipients>%s</t:CcRecipients>`, ccRecipients.String())
	}

	createSoap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:CreateItem MessageDisposition="SaveOnly">
      <m:SavedItemFolderId>
        <t:DistinguishedFolderId Id="drafts"/>
      </m:SavedItemFolderId>
      <m:Items>
        <t:Message>
          <t:Subject>%s</t:Subject>
          <t:Body BodyType="HTML">%s</t:Body>
          <t:ToRecipients>%s</t:ToRecipients>
          %s
        </t:Message>
      </m:Items>
    </m:CreateItem>
  </soap:Body>
</soap:Envelope>`, escapeXML(subject), escapeXML(body), toRecipients.String(), ccSection)

	createResp, err := c.doRequest(createSoap, username, password)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	// Extract ItemId and ChangeKey from response
	itemID := extractValue(string(createResp), `<t:ItemId Id="`, `"`)
	changeKey := extractValue(string(createResp), `ChangeKey="`, `"`)
	if itemID == "" {
		return fmt.Errorf("failed to get item ID from create response")
	}

	// Step 2: Add attachments
	for _, att := range attachments {
		attachSoap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:CreateAttachment>
      <m:ParentItemId Id="%s" ChangeKey="%s"/>
      <m:Attachments>
        <t:FileAttachment>
          <t:Name>%s</t:Name>
          <t:Content>%s</t:Content>
        </t:FileAttachment>
      </m:Attachments>
    </m:CreateAttachment>
  </soap:Body>
</soap:Envelope>`, itemID, changeKey, escapeXML(att.Name), base64.StdEncoding.EncodeToString(att.Content))

		attachResp, err := c.doRequest(attachSoap, username, password)
		if err != nil {
			return fmt.Errorf("failed to add attachment %s: %w", att.Name, err)
		}
		// Update changeKey for next attachment
		newChangeKey := extractValue(string(attachResp), `ChangeKey="`, `"`)
		if newChangeKey != "" {
			changeKey = newChangeKey
		}
	}

	// Step 3: Send the email
	sendSoap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:SendItem SaveItemToFolder="true">
      <m:ItemIds>
        <t:ItemId Id="%s" ChangeKey="%s"/>
      </m:ItemIds>
      <m:SavedItemFolderId>
        <t:DistinguishedFolderId Id="sentitems"/>
      </m:SavedItemFolderId>
    </m:SendItem>
  </soap:Body>
</soap:Envelope>`, itemID, changeKey)

	_, err = c.doRequest(sendSoap, username, password)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// MarkEmailAsRead marks an email as read
func (c *Client) MarkEmailAsRead(username, password, itemID, changeKey string) error {
	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:UpdateItem MessageDisposition="SaveOnly" ConflictResolution="AlwaysOverwrite">
      <m:ItemChanges>
        <t:ItemChange>
          <t:ItemId Id="%s" ChangeKey="%s"/>
          <t:Updates>
            <t:SetItemField>
              <t:FieldURI FieldURI="message:IsRead"/>
              <t:Message>
                <t:IsRead>true</t:IsRead>
              </t:Message>
            </t:SetItemField>
          </t:Updates>
        </t:ItemChange>
      </m:ItemChanges>
    </m:UpdateItem>
  </soap:Body>
</soap:Envelope>`, itemID, changeKey)

	_, err := c.doRequest(soap, username, password)
	return err
}

// DeleteEmail moves an email to deleted items
func (c *Client) DeleteEmail(username, password, itemID, changeKey string) error {
	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:DeleteItem DeleteType="MoveToDeletedItems">
      <m:ItemIds>
        <t:ItemId Id="%s" ChangeKey="%s"/>
      </m:ItemIds>
    </m:DeleteItem>
  </soap:Body>
</soap:Envelope>`, itemID, changeKey)

	_, err := c.doRequest(soap, username, password)
	return err
}

func (c *Client) parseMailFoldersResponse(xml string) []MailFolder {
	var folders []MailFolder

	items := strings.Split(xml, "<t:Folder>")
	for _, item := range items[1:] {
		endIdx := strings.Index(item, "</t:Folder>")
		if endIdx == -1 {
			continue
		}
		itemXML := item[:endIdx]

		unread := 0
		total := 0
		if v := extractValue(itemXML, "<t:UnreadCount>", "</t:UnreadCount>"); v != "" {
			fmt.Sscanf(v, "%d", &unread)
		}
		if v := extractValue(itemXML, "<t:TotalCount>", "</t:TotalCount>"); v != "" {
			fmt.Sscanf(v, "%d", &total)
		}

		folder := MailFolder{
			ID:          extractValue(itemXML, `<t:FolderId Id="`, `"`),
			DisplayName: extractValue(itemXML, "<t:DisplayName>", "</t:DisplayName>"),
			UnreadCount: unread,
			TotalCount:  total,
		}

		if folder.DisplayName != "" {
			folders = append(folders, folder)
		}
	}

	return folders
}

func (c *Client) parseEmailsResponse(xml string) []EmailMessage {
	var emails []EmailMessage

	items := strings.Split(xml, "<t:Message>")
	for _, item := range items[1:] {
		endIdx := strings.Index(item, "</t:Message>")
		if endIdx == -1 {
			continue
		}
		itemXML := item[:endIdx]

		email := EmailMessage{
			ID:             extractValue(itemXML, `<t:ItemId Id="`, `"`),
			ChangeKey:      extractChangeKey(itemXML),
			ConversationID: extractValue(itemXML, `<t:ConversationId Id="`, `"`),
			ItemClass:      extractValue(itemXML, "<t:ItemClass>", "</t:ItemClass>"),
			Subject:        extractValue(itemXML, "<t:Subject>", "</t:Subject>"),
			ReceivedAt:     extractValue(itemXML, "<t:DateTimeReceived>", "</t:DateTimeReceived>"),
			IsRead:         strings.Contains(strings.ToLower(itemXML), "<t:isread>true"),
			HasAttach:      strings.Contains(strings.ToLower(itemXML), "<t:hasattachments>true"),
		}

		// Parse From
		if fromStart := strings.Index(itemXML, "<t:From>"); fromStart != -1 {
			if fromEnd := strings.Index(itemXML[fromStart:], "</t:From>"); fromEnd != -1 {
				fromXML := itemXML[fromStart : fromStart+fromEnd]
				email.From = &Person{
					Name:  extractValue(fromXML, "<t:Name>", "</t:Name>"),
					Email: extractValue(fromXML, "<t:EmailAddress>", "</t:EmailAddress>"),
				}
			}
		}

		// Parse To recipients
		if toStart := strings.Index(itemXML, "<t:ToRecipients>"); toStart != -1 {
			if toEnd := strings.Index(itemXML[toStart:], "</t:ToRecipients>"); toEnd != -1 {
				toXML := itemXML[toStart : toStart+toEnd]
				for _, mailbox := range strings.Split(toXML, "<t:Mailbox>")[1:] {
					mbEnd := strings.Index(mailbox, "</t:Mailbox>")
					if mbEnd == -1 {
						continue
					}
					email.To = append(email.To, Person{
						Name:  extractValue(mailbox[:mbEnd], "<t:Name>", "</t:Name>"),
						Email: extractValue(mailbox[:mbEnd], "<t:EmailAddress>", "</t:EmailAddress>"),
					})
				}
			}
		}

		if email.Subject == "" {
			email.Subject = "(Без темы)"
		}

		emails = append(emails, email)
	}

	return emails
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func extractChangeKey(xml string) string {
	// ChangeKey is in <t:ItemId Id="..." ChangeKey="..."/>
	start := strings.Index(xml, `ChangeKey="`)
	if start == -1 {
		return ""
	}
	start += len(`ChangeKey="`)
	end := strings.Index(xml[start:], `"`)
	if end == -1 {
		return ""
	}
	return xml[start : start+end]
}

// MeetingRoom represents a meeting room from Exchange
type MeetingRoom struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Capacity int    `json:"capacity,omitempty"`
}

// RoomList represents a room list from Exchange
type RoomList struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetRoomLists fetches room lists from Exchange
func (c *Client) GetRoomLists(username, password string) ([]RoomList, error) {
	soap := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:GetRoomLists/>
  </soap:Body>
</soap:Envelope>`

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	return c.parseRoomListsResponse(string(body)), nil
}

// GetRooms fetches rooms from a specific room list
func (c *Client) GetRooms(roomListEmail, username, password string) ([]MeetingRoom, error) {
	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:GetRooms>
      <m:RoomList>
        <t:EmailAddress>%s</t:EmailAddress>
      </m:RoomList>
    </m:GetRooms>
  </soap:Body>
</soap:Envelope>`, roomListEmail)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	return c.parseRoomsResponse(string(body)), nil
}

// GetAllRooms fetches all rooms from all room lists
func (c *Client) GetAllRooms(username, password string) ([]MeetingRoom, error) {
	roomLists, err := c.GetRoomLists(username, password)
	if err != nil {
		return nil, err
	}

	var allRooms []MeetingRoom
	for _, rl := range roomLists {
		rooms, err := c.GetRooms(rl.Email, username, password)
		if err != nil {
			log.Printf("Failed to get rooms from list %s: %v", rl.Email, err)
			continue
		}
		allRooms = append(allRooms, rooms...)
	}

	return allRooms, nil
}

func (c *Client) parseRoomListsResponse(xml string) []RoomList {
	var roomLists []RoomList

	// Find all RoomList elements
	for _, roomXML := range strings.Split(xml, "<t:RoomList>")[1:] {
		endIdx := strings.Index(roomXML, "</t:RoomList>")
		if endIdx == -1 {
			continue
		}

		name := extractValue(roomXML[:endIdx], "<t:Name>", "</t:Name>")
		email := extractValue(roomXML[:endIdx], "<t:EmailAddress>", "</t:EmailAddress>")

		if email != "" {
			roomLists = append(roomLists, RoomList{
				Name:  name,
				Email: email,
			})
		}
	}

	return roomLists
}

func (c *Client) parseRoomsResponse(xml string) []MeetingRoom {
	var rooms []MeetingRoom

	// Find all Room elements
	for _, roomXML := range strings.Split(xml, "<t:Room>")[1:] {
		endIdx := strings.Index(roomXML, "</t:Room>")
		if endIdx == -1 {
			continue
		}

		// Room info is inside <t:Id>
		idXML := roomXML[:endIdx]
		name := extractValue(idXML, "<t:Name>", "</t:Name>")
		email := extractValue(idXML, "<t:EmailAddress>", "</t:EmailAddress>")

		if email != "" {
			rooms = append(rooms, MeetingRoom{
				Name:  name,
				Email: email,
			})
		}
	}

	return rooms
}

// RespondToMeetingRequest responds to a meeting invitation (Accept, Decline, Tentative)
func (c *Client) RespondToMeetingRequest(username, password, itemID, changeKey, response string) error {
	// Determine the response element based on response type
	var responseElement string
	switch response {
	case "Accept":
		responseElement = "t:AcceptItem"
	case "Decline":
		responseElement = "t:DeclineItem"
	case "Tentative":
		responseElement = "t:TentativelyAcceptItem"
	default:
		return fmt.Errorf("invalid response type: %s (must be Accept, Decline, or Tentative)", response)
	}

	var itemIdElement string
	if changeKey != "" {
		itemIdElement = fmt.Sprintf(`<t:ReferenceItemId Id="%s" ChangeKey="%s"/>`, itemID, changeKey)
	} else {
		itemIdElement = fmt.Sprintf(`<t:ReferenceItemId Id="%s"/>`, itemID)
	}

	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2013"/>
  </soap:Header>
  <soap:Body>
    <m:CreateItem MessageDisposition="SendAndSaveCopy">
      <m:Items>
        <%s>
          %s
        </%s>
      </m:Items>
    </m:CreateItem>
  </soap:Body>
</soap:Envelope>`, responseElement, itemIdElement, responseElement)

	_, err := c.doRequest(soap, username, password)
	return err
}
