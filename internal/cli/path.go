package cli

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// normalizeAPIPath validates and normalizes an API path for the api passthrough command.
// Query strings must be passed via --query, not embedded in the path.
func normalizeAPIPath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", usageErr("path is required")
	}
	if raw[0] != '/' {
		raw = "/" + raw
	}
	if i := strings.IndexAny(raw, "?#"); i >= 0 {
		return "", usageErr("path must not contain ? or #; use --query for query parameters")
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", usageErr("invalid path encoding: %v", err)
	}
	if strings.ContainsAny(decoded, "?#") {
		return "", usageErr("path must not contain ? or # after decoding")
	}
	if strings.Contains(decoded, "..") {
		return "", usageErr("path must not contain ..")
	}
	if strings.Contains(decoded, "//") {
		return "", usageErr("path must not contain //")
	}
	cleaned := path.Clean(decoded)
	if cleaned == "/api" {
		return cleaned, nil
	}
	if !strings.HasPrefix(cleaned, "/api/") {
		return "", usageErr("path must start with /api/")
	}
	return cleaned, nil
}

func projectPath(projectID int, suffix string) (string, error) {
	if projectID <= 0 {
		return "", usageErr("--project must be a positive project id")
	}
	p := fmt.Sprintf("/api/project/%d%s", projectID, suffix)
	return normalizeAPIPath(p)
}

func parseQueryFlags(pairs []string) (url.Values, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	q := make(url.Values)
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok || k == "" {
			return nil, usageErr("--query must be key=value")
		}
		q.Add(k, v)
	}
	return q, nil
}
