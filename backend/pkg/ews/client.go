package ews

import (
	"crypto/tls"
	"fmt"
	"io"
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

// NewClient creates a new EWS client
func NewClient(url, domain string) *Client {
	return &Client{
		URL:    url,
		Domain: domain,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// GetCalendarEvents fetches calendar events from Exchange
func (c *Client) GetCalendarEvents(email, username, password string, daysBack, daysForward int) ([]CalendarEvent, error) {
	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -daysBack).Format("2006-01-02T00:00:00Z")
	endDate := now.AddDate(0, 0, daysForward).Format("2006-01-02T23:59:59Z")

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
        <t:DistinguishedFolderId Id="calendar"/>
      </m:ParentFolderIds>
    </m:FindItem>
  </soap:Body>
</soap:Envelope>`, startDate, endDate)

	body, err := c.doRequest(soap, username, password)
	if err != nil {
		return nil, err
	}

	return c.parseCalendarResponse(string(body)), nil
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

	items := strings.Split(xml, "<t:CalendarItem")
	for _, item := range items[1:] { // Skip first split
		endIdx := strings.Index(item, "</t:CalendarItem>")
		if endIdx == -1 {
			continue
		}
		itemXML := item[:endIdx]

		event := CalendarEvent{
			ID:          extractValue(itemXML, `ItemId Id="`, `"`),
			Subject:     extractValue(itemXML, "<t:Subject>", "</t:Subject>"),
			Start:       extractValue(itemXML, "<t:Start>", "</t:Start>"),
			End:         extractValue(itemXML, "<t:End>", "</t:End>"),
			Location:    extractValue(itemXML, "<t:Location>", "</t:Location>"),
			IsRecurring: strings.Contains(strings.ToLower(itemXML), "<t:isrecurring>true"),
			IsCancelled: strings.Contains(strings.ToLower(itemXML), "<t:iscancelled>true"),
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

		if event.Start != "" {
			events = append(events, event)
		}
	}

	return events
}

func (c *Client) parseAttendees(xml string) []Attendee {
	var attendees []Attendee

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
