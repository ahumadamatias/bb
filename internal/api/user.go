package api

import "net/http"

// User is the response shape of GET /user.
type User struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	DisplayName string `json:"display_name"`
	AccountID   string `json:"account_id"`
}

// CurrentUser fetches the authenticated user, doubling as a credential
// check: a 401/403 here means the configured email/token are invalid.
func (c *Client) CurrentUser() (*User, error) {
	var u User
	if err := c.doJSON(http.MethodGet, "/user", nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
