// Package semaphore is a minimal Semaphore UI REST API client (community-built; not affiliated with Semaphore UI).
package semaphore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const MinTokenLength = 16

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

type APIError struct {
	Status  int
	Message string
	Body    json.RawMessage
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("semaphore HTTP %d: %s", e.Status, e.Message)
	}
	return fmt.Sprintf("semaphore HTTP %d", e.Status)
}

func (e *APIError) IsAuth() bool { return e.Status == 401 || e.Status == 403 }

func (c *Client) base() string {
	return strings.TrimRight(c.BaseURL, "/")
}

func (c *Client) http() *http.Client {
	if c.HTTP == nil {
		return http.DefaultClient
	}
	return c.HTTP
}

func (c *Client) Redact(s string) string {
	s = StripURLUserinfo(s)
	if c.Token == "" || len(c.Token) < MinTokenLength {
		return s
	}
	return strings.ReplaceAll(s, c.Token, "<API_TOKEN>")
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Request, error) {
	u := c.base() + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, errors.New(c.Redact(err.Error()))
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body io.Reader) (json.RawMessage, http.Header, error) {
	req, err := c.newRequest(ctx, method, path, query, body)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request %s %s failed: %s", method, path, c.Redact(err.Error()))
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %s", c.Redact(err.Error()))
	}
	if resp.StatusCode >= 400 {
		msg := c.Redact(parseMessage(data))
		return nil, resp.Header, &APIError{Status: resp.StatusCode, Message: msg, Body: data}
	}
	if len(data) == 0 {
		return json.RawMessage("null"), resp.Header, nil
	}
	if !json.Valid(data) {
		return json.RawMessage(data), resp.Header, nil
	}
	return json.RawMessage(data), resp.Header, nil
}

func parseMessage(data []byte) string {
	var m struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &m) == nil {
		if m.Message != "" {
			return m.Message
		}
		if m.Error != "" {
			return m.Error
		}
	}
	return strings.TrimSpace(string(data))
}

func (c *Client) Get(ctx context.Context, path string, query url.Values) (json.RawMessage, http.Header, error) {
	return c.Do(ctx, http.MethodGet, path, query, nil)
}

func (c *Client) JSON(ctx context.Context, method, path string, v any) (json.RawMessage, http.Header, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return c.Do(ctx, method, path, nil, bytes.NewReader(b))
}

func DryRun(method, baseURL, path string, query url.Values, body any) map[string]any {
	safeBase, _ := ValidateBaseURL(baseURL)
	if safeBase == "" {
		safeBase = StripURLUserinfo(strings.TrimSpace(baseURL))
	}
	u := strings.TrimRight(safeBase, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	out := map[string]any{
		"method": method,
		"url":    u,
		"header": map[string]string{"Authorization": "Bearer <API_TOKEN>"},
	}
	if body != nil {
		out["body"] = body
	}
	return out
}

// SanitizeAccessKeys returns id/name/type/project_id only — never secret material.
func SanitizeAccessKeys(raw json.RawMessage) (json.RawMessage, error) {
	var keys []map[string]any
	if err := json.Unmarshal(raw, &keys); err != nil {
		return raw, nil
	}
	out := make([]map[string]any, 0, len(keys))
	allowed := map[string]bool{"id": true, "name": true, "type": true, "project_id": true}
	for _, k := range keys {
		safe := make(map[string]any, 4)
		for key, val := range k {
			if allowed[key] {
				safe[key] = val
			}
		}
		out = append(out, safe)
	}
	return json.Marshal(out)
}
