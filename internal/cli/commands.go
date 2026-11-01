package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swtormy-agents/semaphore-cli/internal/semaphore"
)

func (a *app) pingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "PING test (GET /api/ping)",
		Args:  cobra.NoArgs,
		RunE:  a.simpleGet("/api/ping", nil),
	}
}

func (a *app) healthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Verify SEMAPHORE_URL and API token (GET /api/info)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := a.client(ctx)
			if err != nil {
				return err
			}
			path := "/api/info"
			if a.dryRun {
				return a.emitValue(map[string]any{
					"request": semaphore.DryRun(http.MethodGet, c.BaseURL, path, nil, nil),
					"mode":    "read-only",
				})
			}
			raw, _, err := c.Get(ctx, path, nil)
			if err != nil {
				return a.wrapAPI(err)
			}
			var server json.RawMessage = raw
			return a.emitValue(map[string]any{"ok": true, "url": c.BaseURL, "server": server})
		},
	}
}

func (a *app) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"info"},
		Short:   "Semaphore server info (GET /api/info)",
		Args:    cobra.NoArgs,
		RunE:    a.simpleGet("/api/info", nil),
	}
}

func (a *app) simpleGet(path string, query map[string]string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		c, err := a.client(ctx)
		if err != nil {
			return err
		}
		if a.dryRun {
			return a.emitValue(semaphore.DryRun(http.MethodGet, c.BaseURL, path, urlValues(query), nil))
		}
		raw, _, err := c.Get(ctx, path, urlValues(query))
		if err != nil {
			return a.wrapAPI(err)
		}
		return a.emit(raw)
	}
}

func (a *app) projectsCmd() *cobra.Command {
	projects := &cobra.Command{Use: "projects", Short: "Projects"}
	list := &cobra.Command{
		Use:   "list",
		Short: "List projects (GET /api/projects)",
		Args:  cobra.NoArgs,
		RunE:  a.simpleGet("/api/projects", nil),
	}
	projects.AddCommand(list)
	return projects
}

func (a *app) templatesCmd() *cobra.Command {
	templates := &cobra.Command{Use: "templates", Short: "Job templates"}
	var project string
	list := &cobra.Command{
		Use:   "list",
		Short: "List templates in a project (GET /api/project/{id}/templates)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			path, err := projectPathFromID(projectID, "/templates")
			if err != nil {
				return err
			}
			return a.simpleGet(path, map[string]string{"sort": "name", "order": "asc"})(cmd, args)
		},
	}
	list.Flags().StringVar(&project, "project", "", "project id (required)")
	_ = list.MarkFlagRequired("project")
	templates.AddCommand(list)
	return templates
}

func (a *app) tasksCmd() *cobra.Command {
	tasks := &cobra.Command{Use: "task", Short: "Tasks"}
	var project, taskID, templateID string
	var writeFlag bool

	list := &cobra.Command{
		Use:   "list",
		Short: "List tasks in a project (GET /api/project/{id}/tasks)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			path, err := projectPathFromID(projectID, "/tasks")
			if err != nil {
				return err
			}
			return a.simpleGet(path, nil)(cmd, args)
		},
	}
	list.Flags().StringVar(&project, "project", "", "project id (required)")
	_ = list.MarkFlagRequired("project")

	get := &cobra.Command{
		Use:   "get",
		Short: "Get a task (GET /api/project/{id}/tasks/{task_id})",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			id, err := parsePositiveID("--id", taskID)
			if err != nil {
				return err
			}
			path, err := projectTaskPath(projectID, id, "")
			if err != nil {
				return err
			}
			return a.simpleGet(path, nil)(cmd, args)
		},
	}
	get.Flags().StringVar(&project, "project", "", "project id (required)")
	get.Flags().StringVar(&taskID, "id", "", "task id (required)")
	_ = get.MarkFlagRequired("project")
	_ = get.MarkFlagRequired("id")

	output := &cobra.Command{
		Use:   "output",
		Short: "Get task output (GET /api/project/{id}/tasks/{task_id}/output)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			id, err := parsePositiveID("--id", taskID)
			if err != nil {
				return err
			}
			path, err := projectTaskPath(projectID, id, "/output")
			if err != nil {
				return err
			}
			return a.simpleGet(path, nil)(cmd, args)
		},
	}
	output.Flags().StringVar(&project, "project", "", "project id (required)")
	output.Flags().StringVar(&taskID, "id", "", "task id (required)")
	_ = output.MarkFlagRequired("project")
	_ = output.MarkFlagRequired("id")

	run := &cobra.Command{
		Use:   "run",
		Short: "Start a task from a template (POST /api/project/{id}/tasks)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			tmplID, err := parsePositiveID("--template", templateID)
			if err != nil {
				return err
			}
			path, err := projectPathFromID(projectID, "/tasks")
			if err != nil {
				return err
			}
			if err := a.requireHumanWrite(cmd, "task run"); err != nil {
				return err
			}
			ctx := cmd.Context()
			c, err := a.client(ctx)
			if err != nil {
				return err
			}
			body := map[string]any{"template_id": tmplID}
			if a.dryRun {
				return a.emitValue(semaphore.DryRun(http.MethodPost, c.BaseURL, path, nil, body))
			}
			raw, _, err := c.JSON(ctx, http.MethodPost, path, body)
			if err != nil {
				return a.wrapAPI(err)
			}
			return a.emit(raw)
		},
	}
	run.Flags().StringVar(&project, "project", "", "project id (required)")
	run.Flags().StringVar(&templateID, "template", "", "template id (required)")
	bindWriteFlag(run, &writeFlag)
	_ = run.MarkFlagRequired("project")
	_ = run.MarkFlagRequired("template")

	tasks.AddCommand(list, get, output, run)
	return tasks
}

func (a *app) inventoriesCmd() *cobra.Command {
	inventories := &cobra.Command{Use: "inventories", Short: "Inventories"}
	var project string
	list := &cobra.Command{
		Use:   "list",
		Short: "List inventories (GET /api/project/{id}/inventory)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			path, err := projectPathFromID(projectID, "/inventory")
			if err != nil {
				return err
			}
			return a.simpleGet(path, map[string]string{"sort": "name", "order": "asc"})(cmd, args)
		},
	}
	list.Flags().StringVar(&project, "project", "", "project id (required)")
	_ = list.MarkFlagRequired("project")
	inventories.AddCommand(list)
	return inventories
}

func (a *app) environmentsCmd() *cobra.Command {
	environments := &cobra.Command{Use: "environments", Short: "Variable groups (environments)"}
	var project string
	list := &cobra.Command{
		Use:   "list",
		Short: "List environments (GET /api/project/{id}/environment)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			path, err := projectPathFromID(projectID, "/environment")
			if err != nil {
				return err
			}
			return a.simpleGet(path, map[string]string{"sort": "name", "order": "asc"})(cmd, args)
		},
	}
	list.Flags().StringVar(&project, "project", "", "project id (required)")
	_ = list.MarkFlagRequired("project")
	environments.AddCommand(list)
	return environments
}

func (a *app) keysCmd() *cobra.Command {
	keys := &cobra.Command{Use: "keys", Short: "Access keys (metadata only)"}
	var project string
	list := &cobra.Command{
		Use:   "list",
		Short: "List access key names and ids (GET /api/project/{id}/keys)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parsePositiveID("--project", project)
			if err != nil {
				return err
			}
			path, err := projectPathFromID(projectID, "/keys")
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			c, err := a.client(ctx)
			if err != nil {
				return err
			}
			q := urlValues(map[string]string{"sort": "name", "order": "asc"})
			if a.dryRun {
				return a.emitValue(semaphore.DryRun(http.MethodGet, c.BaseURL, path, q, nil))
			}
			raw, _, err := c.Get(ctx, path, q)
			if err != nil {
				return a.wrapAPI(err)
			}
			safe, err := semaphore.SanitizeAccessKeys(raw)
			if err != nil {
				return runtimeErr(err)
			}
			return a.emit(safe)
		},
	}
	list.Flags().StringVar(&project, "project", "", "project id (required)")
	_ = list.MarkFlagRequired("project")
	keys.AddCommand(list)
	return keys
}

func (a *app) apiCmd() *cobra.Command {
	var bodyJSON string
	var writeFlag bool
	var queryPairs []string
	cmd := &cobra.Command{
		Use:   "api <METHOD> <path>",
		Short: "Call a Semaphore API path (escape hatch)",
		Long:  "Paths must start with /api/. Query parameters use --query key=value. Mutating methods need --write and interactive confirmation.",
		Example: `  semaphore api GET /api/projects
  semaphore api GET /api/project/1/tasks --query limit=10`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			pathOnly, err := normalizeAPIPath(args[1])
			if err != nil {
				return err
			}
			q, err := parseQueryFlags(queryPairs)
			if err != nil {
				return err
			}
			if err := checkAPIAllowed(method, pathOnly); err != nil {
				return err
			}
			if apiNeedsWrite(method) {
				if err := a.requireHumanWrite(cmd, "api "+method+" "+pathOnly); err != nil {
					return err
				}
			}
			ctx := cmd.Context()
			c, err := a.client(ctx)
			if err != nil {
				return err
			}
			if a.dryRun {
				var body any
				if bodyJSON != "" {
					_ = json.Unmarshal([]byte(bodyJSON), &body)
				}
				return a.emitValue(semaphore.DryRun(method, c.BaseURL, pathOnly, q, body))
			}
			var raw json.RawMessage
			if bodyJSON != "" {
				var body any
				if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
					return usageErr("--body: %v", err)
				}
				raw, _, err = c.JSON(ctx, method, pathOnly, body)
			} else {
				raw, _, err = c.Do(ctx, method, pathOnly, q, nil)
			}
			if err != nil {
				return a.wrapAPI(err)
			}
			return a.emit(raw)
		},
	}
	cmd.Flags().StringVar(&bodyJSON, "body", "", "JSON body for POST/PUT/PATCH")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "query parameter (repeatable, key=value)")
	bindWriteFlag(cmd, &writeFlag)
	return cmd
}

func checkAPIAllowed(method, pathOnly string) error {
	p := strings.ToLower(pathOnly)
	if method == http.MethodDelete && strings.Contains(p, "/tasks/") {
		return blockedErr(fmt.Errorf("api %s %s is blocked in v0.1; confirm manually in the UI", method, pathOnly))
	}
	if method != http.MethodGet && method != http.MethodHead && strings.Contains(p, "/keys") {
		return blockedErr(fmt.Errorf("api %s %s is blocked in v0.1; keys mutations are not exposed", method, pathOnly))
	}
	return nil
}

func apiNeedsWrite(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

var isTerminal = func(r ioReader) bool {
	f, ok := r.(*os.File)
	return ok && isTerminalFile(f)
}

type ioReader interface {
	Read(p []byte) (n int, err error)
}
