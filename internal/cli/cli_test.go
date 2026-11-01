package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const testToken = "semaphore-test-tk16" // 16 chars // gitleaks:allow

func testEnv(url string) func(string) string {
	return func(k string) string {
		switch k {
		case "SEMAPHORE_URL":
			return url
		case "SEMAPHORE_API_TOKEN":
			return testToken
		}
		return ""
	}
}

func authHeader(r *http.Request) bool {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	return strings.HasPrefix(h, prefix) && strings.TrimPrefix(h, prefix) == testToken
}

func runCLI(t *testing.T, srv *httptest.Server, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Execute(args, strings.NewReader(""), &out, &errOut, testEnv(srv.URL))
	return code, out.String(), errOut.String()
}

func TestPing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		if !authHeader(r) {
			http.Error(w, "nope", http.StatusUnauthorized)
			return
		}
		w.Write([]byte(`{"response":"pong"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	code, out, _ := runCLI(t, srv, "ping")
	if code != 0 || !strings.Contains(out, "pong") {
		t.Fatalf("ping code=%d out=%s", code, out)
	}
}

func TestHealthAndVersion(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		if !authHeader(r) {
			http.Error(w, "nope", http.StatusUnauthorized)
			return
		}
		w.Write([]byte(`{"version":"2.16.14","update_available":false}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	code, out, _ := runCLI(t, srv, "health")
	if code != 0 || !strings.Contains(out, `"ok":true`) {
		t.Fatalf("health code=%d out=%s", code, out)
	}
	code, out, _ = runCLI(t, srv, "version")
	if code != 0 || !strings.Contains(out, "2.16.14") {
		t.Fatalf("version code=%d out=%s", code, out)
	}
}

func TestProjectsList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"demo"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	code, out, _ := runCLI(t, srv, "projects", "list")
	if code != 0 || !strings.Contains(out, "demo") {
		t.Fatalf("projects code=%d out=%s", code, out)
	}
}

func TestTemplatesList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/project/3/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("sort") != "name" {
			http.Error(w, "sort", http.StatusBadRequest)
			return
		}
		w.Write([]byte(`[{"id":10,"name":"deploy"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	code, out, _ := runCLI(t, srv, "templates", "list", "--project", "3")
	if code != 0 || !strings.Contains(out, "deploy") {
		t.Fatalf("templates code=%d out=%s", code, out)
	}
}

func TestTaskListGetOutput(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/project/1/tasks", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":5,"status":"success"}]`))
	})
	mux.HandleFunc("/api/project/1/tasks/5", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":5,"status":"success"}`))
	})
	mux.HandleFunc("/api/project/1/tasks/5/output", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"task_id":5,"output":"log line"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	if code, _, _ := runCLI(t, srv, "task", "list", "--project", "1"); code != 0 {
		t.Fatalf("task list exit %d", code)
	}
	if code, out, _ := runCLI(t, srv, "task", "get", "--project", "1", "--id", "5"); code != 0 || !strings.Contains(out, "success") {
		t.Fatalf("task get code=%d out=%s", code, out)
	}
	if code, out, _ := runCLI(t, srv, "task", "output", "--project", "1", "--id", "5"); code != 0 || !strings.Contains(out, "log line") {
		t.Fatalf("task output code=%d out=%s", code, out)
	}
}

func TestInventoriesEnvironments(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/project/2/inventory", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"prod"}]`))
	})
	mux.HandleFunc("/api/project/2/environment", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"env"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	if code, _, _ := runCLI(t, srv, "inventories", "list", "--project", "2"); code != 0 {
		t.Fatalf("inventories exit %d", code)
	}
	if code, _, _ := runCLI(t, srv, "environments", "list", "--project", "2"); code != 0 {
		t.Fatalf("environments exit %d", code)
	}
}

func TestKeysListStripsSecrets(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/project/1/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"ssh","type":"ssh","project_id":1,"body":"SECRETKEYMATERIAL"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	code, out, _ := runCLI(t, srv, "keys", "list", "--project", "1")
	if code != 0 {
		t.Fatalf("keys exit %d", code)
	}
	if strings.Contains(out, "SECRET") {
		t.Fatalf("secret leaked: %s", out)
	}
	if !strings.Contains(out, `"name":"ssh"`) {
		t.Fatalf("expected name in %s", out)
	}
}

func TestAPICallWithQuery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("foo") != "bar" {
			http.Error(w, "query", http.StatusBadRequest)
			return
		}
		w.Write([]byte(`[]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	code, _, _ := runCLI(t, srv, "api", "GET", "/api/projects", "--query", "foo=bar")
	if code != 0 {
		t.Fatalf("api exit %d", code)
	}
}

func TestAuthFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()
	code, _, errOut := runCLI(t, srv, "version")
	if code != 3 {
		t.Fatalf("want exit 3 got %d stderr=%s", code, errOut)
	}
}

func TestAuth401TokenNotInStderr(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		http.Error(w, `{"message":"invalid token `+key+`"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()
	code, _, errOut := runCLI(t, srv, "version")
	if code != 3 {
		t.Fatalf("want exit 3 got %d", code)
	}
	if strings.Contains(errOut, testToken) {
		t.Fatalf("token leaked to stderr: %s", errOut)
	}
	if !strings.Contains(errOut, "<API_TOKEN>") {
		t.Fatalf("expected redacted placeholder in stderr: %s", errOut)
	}
}

func TestDryRunOutputRedactsToken(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	code, out, _ := runCLI(t, srv, "health", "--dry-run")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(out, testToken) {
		t.Fatalf("dry-run leaked token: %s", out)
	}
	if !strings.Contains(out, "<API_TOKEN>") {
		t.Fatalf("expected dry-run placeholder: %s", out)
	}
}

func TestReportErrorRedactsToken(t *testing.T) {
	var errOut bytes.Buffer
	a := &app{token: testToken, errOut: &errOut}
	code := a.reportError(runtimeErr(fmt.Errorf("failure mentioning %s", testToken)))
	if code != 1 {
		t.Fatalf("code=%d", code)
	}
	out := errOut.String()
	if strings.Contains(out, testToken) {
		t.Fatalf("token leaked in reportError output: %s", out)
	}
}

func TestShortTokenRejected(t *testing.T) {
	short := "short-token-15c" // 15 chars // gitleaks:allow
	getenv := func(k string) string {
		switch k {
		case "SEMAPHORE_URL":
			return "http://127.0.0.1:1"
		case "SEMAPHORE_API_TOKEN":
			return short
		}
		return ""
	}
	var errOut bytes.Buffer
	code := Execute([]string{"version"}, strings.NewReader(""), io.Discard, &errOut, getenv)
	if code != 2 {
		t.Fatalf("want exit 2 got %d: %s", code, errOut.String())
	}
}

func TestSkillPrint(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	code, out, _ := runCLI(t, srv, "skill", "print")
	if code != 0 || !strings.Contains(out, "name: semaphore") {
		t.Fatalf("code=%d out=%s", code, out)
	}
}

func TestWriteRequiresWriteFlag(t *testing.T) {
	var errOut bytes.Buffer
	code := Execute([]string{"api", "POST", "/api/projects", "--body", "{}"},
		strings.NewReader(""), io.Discard, &errOut, testEnv("http://example"))
	if code != 3 {
		t.Fatalf("want 3 got %d: %s", code, errOut.String())
	}
}

func TestTaskRunRequiresWrite(t *testing.T) {
	var errOut bytes.Buffer
	code := Execute([]string{"task", "run", "--project", "1", "--template", "2"},
		strings.NewReader(""), io.Discard, &errOut, testEnv("http://example"))
	if code != 3 {
		t.Fatalf("want 3 got %d: %s", code, errOut.String())
	}
}

func TestProjectsListTableOutput(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"demo"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	code, out, _ := runCLI(t, srv, "projects", "list", "--output", "table")
	if code != 0 || !strings.Contains(out, "demo") {
		t.Fatalf("table code=%d out=%s", code, out)
	}
}

func TestInvalidBaseURLs(t *testing.T) {
	cases := []struct {
		url  string
		code int
	}{
		{"ftp://example.com", 2},
		{"https://", 2},
		{"https://user:pass@host.example", 2},
	}
	for _, tc := range cases {
		getenv := func(k string) string {
			switch k {
			case "SEMAPHORE_URL":
				return tc.url
			case "SEMAPHORE_API_TOKEN":
				return testToken
			}
			return ""
		}
		var errOut bytes.Buffer
		code := Execute([]string{"version"}, strings.NewReader(""), io.Discard, &errOut, getenv)
		if code != tc.code {
			t.Fatalf("url %q want exit %d got %d: %s", tc.url, tc.code, code, errOut.String())
		}
		errText := errOut.String()
		if strings.Contains(errText, "user:pass") || strings.Contains(errText, "pass@") {
			t.Fatalf("userinfo leaked in error for %q: %s", tc.url, errText)
		}
	}
}

func TestDryRunStripsURLUserinfo(t *testing.T) {
	getenv := func(k string) string {
		switch k {
		case "SEMAPHORE_URL":
			return "https://user:secret@host.example"
		case "SEMAPHORE_API_TOKEN":
			return testToken
		}
		return ""
	}
	var out bytes.Buffer
	code := Execute([]string{"health", "--dry-run"}, strings.NewReader(""), &out, io.Discard, getenv)
	if code != 2 {
		t.Fatalf("want config error for userinfo URL, got %d", code)
	}
}

func TestAPIPathValidation(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	for _, args := range [][]string{
		{"api", "GET", "/apix/foo"},
		{"api", "GET", "/api/../etc/passwd"},
		{"api", "GET", "/api/foo//bar"},
	} {
		code, _, errOut := runCLI(t, srv, args...)
		if code != 2 {
			t.Fatalf("args %v want exit 2 got %d stderr=%s", args, code, errOut)
		}
	}
}

func TestSkillMarkdownIdentical(t *testing.T) {
	a, err := os.ReadFile("SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile("../../skills/semaphore/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatalf("SKILL.md files differ; run scripts/sync-skill-md.sh")
	}
	if string(a) != skillMD {
		t.Fatalf("embedded SKILL.md out of sync with internal/cli/SKILL.md")
	}
}

func TestMissingToken(t *testing.T) {
	getenv := func(k string) string {
		if k == "SEMAPHORE_URL" {
			return "http://127.0.0.1:1"
		}
		return ""
	}
	var errOut bytes.Buffer
	code := Execute([]string{"ping"}, strings.NewReader(""), io.Discard, &errOut, getenv)
	if code != 3 {
		t.Fatalf("want exit 3 got %d", code)
	}
}

func TestInvalidIDFlags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP request during invalid ID validation: %s", r.URL.Path)
	}))
	defer srv.Close()
	cases := [][]string{
		{"templates", "list", "--project", "0"},
		{"templates", "list", "--project", "-1"},
		{"templates", "list", "--project", "1a"},
		{"templates", "list", "--project", "1/2"},
		{"templates", "list", "--project", "%2e"},
		{"templates", "list", "--project", "18446744073709551616"},
		{"task", "get", "--project", "1", "--id", "0"},
		{"task", "get", "--project", "1", "--id", "-1"},
		{"task", "get", "--project", "1", "--id", "1a"},
		{"task", "run", "--project", "1", "--template", "0"},
	}
	for _, args := range cases {
		code, _, errOut := runCLI(t, srv, args...)
		if code != 2 {
			t.Fatalf("args %v want exit 2 got %d stderr=%s", args, code, errOut)
		}
	}
}
