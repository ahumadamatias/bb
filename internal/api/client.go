// Package api implements a minimal client for the Bitbucket Cloud REST
// API 2.0 (https://api.bitbucket.org/2.0), covering auth, pagination, and
// error handling shared by every endpoint-specific file in this package.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL is the Bitbucket Cloud API root used unless overridden
// (tests point this at an httptest.Server instead).
const DefaultBaseURL = "https://api.bitbucket.org/2.0"

// Client is a minimal Bitbucket Cloud API 2.0 client authenticated via
// HTTP Basic Auth with an Atlassian API token.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Email      string
	Token      string
	UserAgent  string
}

// NewClient builds a Client authenticated with email/token. version is
// embedded in the User-Agent header sent with every request.
func NewClient(email, token, version string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    DefaultBaseURL,
		Email:      email,
		Token:      token,
		UserAgent:  "bb-cli/" + version,
	}
}

// Error represents a non-2xx response from the Bitbucket API.
type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	switch e.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Sprintf("HTTP %d: %s (revisá tus credenciales con 'bb auth status' y los scopes del API token)", e.StatusCode, msg)
	default:
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, msg)
	}
}

// bitbucketErrorBody matches the {"error": {"message": "..."}} shape
// Bitbucket returns on failures.
type bitbucketErrorBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) resolveURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return c.BaseURL + path
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.resolveURL(path), body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Email, c.Token)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// doRequest executes req and returns the raw response body, translating
// non-2xx statuses into *Error.
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bitbucket request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &Error{StatusCode: resp.StatusCode}
		var parsed bitbucketErrorBody
		if json.Unmarshal(body, &parsed) == nil && parsed.Error.Message != "" {
			apiErr.Message = parsed.Error.Message
		} else if len(body) > 0 {
			apiErr.Message = strings.TrimSpace(string(body))
		}
		return nil, apiErr
	}

	return body, nil
}

// doJSON sends a JSON request (if reqBody is non-nil) and decodes a JSON
// response into respOut (if non-nil). path may be a relative API path
// (e.g. "/user") or an absolute URL (used for pagination's "next" links).
func (c *Client) doJSON(method, path string, reqBody, respOut interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		encoded, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := c.newRequest(method, path, bodyReader)
	if err != nil {
		return err
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if respOut == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, respOut); err != nil {
		return fmt.Errorf("decoding response body: %w", err)
	}
	return nil
}

// doRaw sends a request and returns the raw response body as a string,
// for endpoints that don't return JSON (e.g. the PR diff endpoint).
func (c *Client) doRaw(method, path string) (string, error) {
	req, err := c.newRequest(method, path, nil)
	if err != nil {
		return "", err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// page mirrors Bitbucket's paginated response envelope.
type page[T any] struct {
	Values []T    `json:"values"`
	Next   string `json:"next"`
}

// getPaginated fetches path and follows the "next" link until either the
// results are exhausted or limit items have been collected. limit <= 0
// means "no limit" (fetch every page).
func getPaginated[T any](c *Client, path string, limit int) ([]T, error) {
	var all []T
	next := path

	for next != "" {
		var p page[T]
		if err := c.doJSON(http.MethodGet, next, nil, &p); err != nil {
			return nil, err
		}
		all = append(all, p.Values...)

		if limit > 0 && len(all) >= limit {
			return all[:limit], nil
		}
		next = p.Next
	}

	return all, nil
}
