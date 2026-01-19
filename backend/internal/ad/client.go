package ad

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type Client struct {
	URL          string
	BaseDN       string
	BindUser     string
	BindPassword string
	SkipVerify   bool
	conn         *ldap.Conn
}

type User struct {
	DN              string   `json:"dn"`
	Username        string   `json:"username"`
	Login           string   `json:"login"` // Same as Username, for backend compatibility
	Email           string   `json:"email"`
	Name            string   `json:"name"`
	DisplayName     string   `json:"display_name"`
	GivenName       string   `json:"given_name"`
	Surname         string   `json:"surname"`
	Department      string   `json:"department"`
	Title           string   `json:"title"`
	Phone           string   `json:"phone"`
	Mobile          string   `json:"mobile"`
	Manager         string   `json:"manager"`
	ManagerDN       string   `json:"manager_dn"`
	PhotoBase64     string   `json:"photo_base64,omitempty"`
	MemberOf        []string `json:"member_of"`
	Enabled         bool     `json:"enabled"`
	PasswordExpired bool     `json:"password_expired"`
}

func NewClient(url, baseDN, bindUser, bindPassword string, skipVerify bool) *Client {
	return &Client{
		URL:          url,
		BaseDN:       baseDN,
		BindUser:     bindUser,
		BindPassword: bindPassword,
		SkipVerify:   skipVerify,
	}
}

func (c *Client) Connect() error {
	var err error

	if strings.HasPrefix(c.URL, "ldaps://") {
		c.conn, err = ldap.DialURL(c.URL, ldap.DialWithTLSConfig(&tls.Config{
			InsecureSkipVerify: c.SkipVerify,
		}))
	} else {
		c.conn, err = ldap.DialURL(c.URL)
	}

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Bind with service account
	if c.BindUser != "" && c.BindPassword != "" {
		err = c.conn.Bind(c.BindUser, c.BindPassword)
		if err != nil {
			return fmt.Errorf("failed to bind: %w", err)
		}
	}

	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) Authenticate(username, password string) (*User, error) {
	// First, try to authenticate with user credentials directly
	// Try different username formats: domain\username, username@domain, or just username
	var userConn *ldap.Conn
	var err error

	// Connect to LDAP
	if strings.HasPrefix(c.URL, "ldaps://") {
		userConn, err = ldap.DialURL(c.URL, ldap.DialWithTLSConfig(&tls.Config{
			InsecureSkipVerify: c.SkipVerify,
		}))
	} else {
		userConn, err = ldap.DialURL(c.URL)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer userConn.Close()

	// Try to bind with user credentials in different formats
	bindFormats := []string{
		username,                                                       // plain username
		fmt.Sprintf("%s@%s", username, extractDomainFromDN(c.BaseDN)), // username@domain
		fmt.Sprintf("CN=%s,%s", username, c.BaseDN),                    // CN=username,OU=...
	}

	var bindErr error
	authenticated := false
	for _, bindUser := range bindFormats {
		bindErr = userConn.Bind(bindUser, password)
		if bindErr == nil {
			authenticated = true
			break
		}
	}

	if !authenticated {
		return nil, fmt.Errorf("invalid credentials: %w", bindErr)
	}

	// Now search for user info using the authenticated connection
	searchFilter := fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))", ldap.EscapeFilter(username))
	searchRequest := ldap.NewSearchRequest(
		c.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn", "sAMAccountName", "userPrincipalName", "mail", "displayName", "givenName", "sn", "department", "title", "telephoneNumber", "mobile", "manager", "memberOf", "userAccountControl"},
		nil,
	)

	sr, err := userConn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(sr.Entries) == 0 {
		// User authenticated but not found in search - create minimal user info
		return &User{
			Username: username,
			Login:    username,
			Email:    fmt.Sprintf("%s@%s", username, extractDomainFromDN(c.BaseDN)),
			Name:     username,
			Enabled:  true,
		}, nil
	}

	// Parse full user info
	entry := sr.Entries[0]
	user := c.parseUser(entry, false)
	return user, nil
}

// extractDomainFromDN extracts domain name from DN (e.g., DC=ekfgroup,DC=ru -> ekfgroup.ru)
func extractDomainFromDN(dn string) string {
	parts := strings.Split(dn, ",")
	var domainParts []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToUpper(part), "DC=") {
			domainParts = append(domainParts, part[3:])
		}
	}
	if len(domainParts) > 0 {
		return strings.Join(domainParts, ".")
	}
	return "local"
}

func (c *Client) GetAllUsers(includePhotos bool) ([]*User, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	attributes := []string{
		"dn", "sAMAccountName", "mail", "displayName", "givenName", "sn",
		"department", "title", "telephoneNumber", "mobile", "manager",
		"memberOf", "userAccountControl",
	}

	if includePhotos {
		attributes = append(attributes, "thumbnailPhoto")
	}

	// Filter: user object, has department, account enabled
	searchFilter := "(&(objectClass=user)(objectCategory=person)(department=*)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))"

	searchRequest := ldap.NewSearchRequest(
		c.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		attributes,
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	users := make([]*User, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		user := c.parseUser(entry, includePhotos)
		if user.Department != "" { // Double check department exists
			users = append(users, user)
		}
	}

	return users, nil
}

func (c *Client) GetUserByEmail(email string) (*User, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	searchFilter := fmt.Sprintf("(&(objectClass=user)(mail=%s))", ldap.EscapeFilter(email))
	searchRequest := ldap.NewSearchRequest(
		c.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn", "sAMAccountName", "mail", "displayName", "givenName", "sn", "department", "title", "telephoneNumber", "mobile", "manager", "memberOf", "userAccountControl"},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return c.parseUser(sr.Entries[0], false), nil
}

func (c *Client) GetSubordinates(managerDN string) ([]*User, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	searchFilter := fmt.Sprintf("(&(objectClass=user)(manager=%s))", ldap.EscapeFilter(managerDN))
	searchRequest := ldap.NewSearchRequest(
		c.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn", "sAMAccountName", "mail", "displayName", "department", "title"},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	users := make([]*User, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		users = append(users, c.parseUser(entry, false))
	}

	return users, nil
}

func (c *Client) parseUser(entry *ldap.Entry, includePhoto bool) *User {
	username := entry.GetAttributeValue("sAMAccountName")
	user := &User{
		DN:          entry.DN,
		Username:    username,
		Login:       username, // Same as Username for backend compatibility
		Email:       entry.GetAttributeValue("mail"),
		DisplayName: entry.GetAttributeValue("displayName"),
		GivenName:   entry.GetAttributeValue("givenName"),
		Surname:     entry.GetAttributeValue("sn"),
		Department:  entry.GetAttributeValue("department"),
		Title:       entry.GetAttributeValue("title"),
		Phone:       entry.GetAttributeValue("telephoneNumber"),
		Mobile:      entry.GetAttributeValue("mobile"),
		ManagerDN:   entry.GetAttributeValue("manager"),
		MemberOf:    entry.GetAttributeValues("memberOf"),
	}

	// Parse name
	if user.DisplayName != "" {
		user.Name = user.DisplayName
	} else if user.GivenName != "" && user.Surname != "" {
		user.Name = user.GivenName + " " + user.Surname
	} else {
		user.Name = user.Username
	}

	// Parse manager name from DN
	if user.ManagerDN != "" {
		parts := strings.Split(user.ManagerDN, ",")
		if len(parts) > 0 {
			cnPart := parts[0]
			if strings.HasPrefix(cnPart, "CN=") {
				user.Manager = cnPart[3:]
			}
		}
	}

	// Check if account is enabled (userAccountControl bit 2 = disabled)
	uac := entry.GetAttributeValue("userAccountControl")
	user.Enabled = !strings.Contains(uac, "2") // Simple check

	// Parse photo
	if includePhoto {
		photoBytes := entry.GetRawAttributeValue("thumbnailPhoto")
		if len(photoBytes) > 0 {
			user.PhotoBase64 = base64.StdEncoding.EncodeToString(photoBytes)
		}
	}

	return user
}
