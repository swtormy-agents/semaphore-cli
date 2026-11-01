package semaphore

import (
	"strings"
	"testing"
)

func TestValidateBaseURL(t *testing.T) {
	ok, err := ValidateBaseURL("https://semaphore.example.com/")
	if err != nil || ok != "https://semaphore.example.com" {
		t.Fatalf("ok=%q err=%v", ok, err)
	}
	for _, raw := range []string{
		"ftp://example.com",
		"https://",
		"https:///path",
		"https://user:pass@host.example",
		"://no-scheme",
	} {
		if _, err := ValidateBaseURL(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestStripURLUserinfo(t *testing.T) {
	in := "https://user:pass@host.example/api"
	out := StripURLUserinfo(in)
	if strings.Contains(out, "user") || strings.Contains(out, "pass") {
		t.Fatalf("userinfo leaked: %q", out)
	}
	if !strings.Contains(out, "host.example") {
		t.Fatalf("host missing: %q", out)
	}
}

func TestSanitizeAccessKeys(t *testing.T) {
	raw := []byte(`[{"id":1,"name":"k","type":"ssh","project_id":2,"secret":"PRIVATE","private_key":"-----BEGIN"}]`)
	out, err := SanitizeAccessKeys(raw)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "PRIVATE") || strings.Contains(s, "BEGIN") {
		t.Fatalf("secrets leaked: %s", s)
	}
	if !strings.Contains(s, `"name":"k"`) {
		t.Fatalf("expected name: %s", s)
	}
}
