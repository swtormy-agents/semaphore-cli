package cli

import (
	"testing"
)

func TestParsePositiveID(t *testing.T) {
	ok, err := parsePositiveID("--project", "42")
	if err != nil || ok != 42 {
		t.Fatalf("42 => %d err=%v", ok, err)
	}
	bad := []string{"0", "-1", "1a", "1/2", "%2e", "18446744073709551616", "99999999999999999999999999999"}
	for _, raw := range bad {
		if _, err := parsePositiveID("--project", raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestProjectPathFromIDUsesFormattedInteger(t *testing.T) {
	p, err := projectPathFromID(7, "/tasks")
	if err != nil || p != "/api/project/7/tasks" {
		t.Fatalf("path=%q err=%v", p, err)
	}
}
