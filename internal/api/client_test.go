package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL
	return c, server
}

func TestDoJSONSuccess(t *testing.T) {
	type resp struct {
		Name string `json:"name"`
	}

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/thing" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		email, token, ok := r.BasicAuth()
		if !ok || email != "user@example.com" || token != "token" {
			t.Errorf("missing/incorrect basic auth: %s %s %v", email, token, ok)
		}
		if ua := r.Header.Get("User-Agent"); ua != "bb-cli/test" {
			t.Errorf("User-Agent = %q, want bb-cli/test", ua)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp{Name: "hello"})
	})

	var out resp
	if err := c.doJSON(http.MethodGet, "/thing", nil, &out); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if out.Name != "hello" {
		t.Errorf("out.Name = %q, want hello", out.Name)
	}
}

func TestDoJSONErrorMapsBitbucketErrorBody(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "Bad title"},
		})
	})

	err := c.doJSON(http.MethodPost, "/thing", map[string]string{"a": "b"}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *Error", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Message != "Bad title" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Bad title")
	}
}

func TestDoJSONUnauthorizedSuggestsCredentials(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "Invalid credentials"},
		})
	})

	err := c.doJSON(http.MethodGet, "/user", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "credenciales") {
		t.Errorf("error message = %q, want it to mention credenciales", err.Error())
	}
}

func TestGetPaginatedFollowsNextAndRespectsLimit(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageParam := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch pageParam {
		case "", "1":
			fmt.Fprintf(w, `{"values":[{"id":1},{"id":2}],"next":"%s/things?page=2"}`, server.URL)
		case "2":
			fmt.Fprintf(w, `{"values":[{"id":3},{"id":4}],"next":"%s/things?page=3"}`, server.URL)
		case "3":
			fmt.Fprint(w, `{"values":[{"id":5}]}`)
		}
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	// No limit: should follow all pages.
	all, err := getPaginated[item](c, "/things", 0)
	if err != nil {
		t.Fatalf("getPaginated() error = %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("len(all) = %d, want 5", len(all))
	}

	// Limit: should stop early.
	limited, err := getPaginated[item](c, "/things", 3)
	if err != nil {
		t.Fatalf("getPaginated() error = %v", err)
	}
	if len(limited) != 3 {
		t.Fatalf("len(limited) = %d, want 3", len(limited))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
