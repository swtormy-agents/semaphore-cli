# semaphore-cli

> **Unofficial, community-built, not affiliated with Semaphore UI**

`semaphore` is a small Go CLI for the [Semaphore UI](https://github.com/semaphoreui/semaphore) REST API. It is **unofficial, community-built, not affiliated with Semaphore UI**. It targets humans and AI agents that want structured JSON (or tables) instead of wiring an MCP server — see [semaphoreui/semaphore#3126](https://github.com/semaphoreui/semaphore/issues/3126).

MIT licensed. Third-party notices in [NOTICE](NOTICE); release archives also ship `THIRD_PARTY_NOTICES` and `third_party_licenses/`.

## Install

Download a release archive from [GitHub Releases](https://github.com/swtormy-agents/semaphore-cli/releases) (Linux, macOS, and Windows — amd64/arm64). Extract the binary and add it to your `PATH`.

**macOS** (example: Apple Silicon):

```bash
curl -fsSL -o semaphore_darwin_arm64.tar.gz \
  https://github.com/swtormy-agents/semaphore-cli/releases/latest/download/semaphore_darwin_arm64.tar.gz
tar -xzf semaphore_darwin_arm64.tar.gz
sudo install semaphore_darwin_arm64/semaphore /usr/local/bin/semaphore
```

Use `semaphore_darwin_amd64.tar.gz` on Intel Macs.

Or install from source with Go **1.26.10** (matches `go.mod`):

```bash
go install github.com/swtormy-agents/semaphore-cli/cmd/semaphore@latest
```

## Quick start

```bash
export SEMAPHORE_URL=https://semaphore.example.com
export SEMAPHORE_API_TOKEN=...    # never pass on the command line

semaphore health
semaphore projects list
semaphore templates list --project 1
semaphore task list --project 1
semaphore keys list --project 1
```

## Output

- Default: JSON on stdout; one JSON error object on stderr
- `--output table` for simple tabular views
- `--jq '...'` to filter JSON (gojq)
- `--pretty` to indent JSON

## Safety

- Never log or print the API token; `--dry-run` and errors redact it as `<API_TOKEN>`
- Agents stay read-only. Mutating `semaphore api` calls and `task run` need `--write`, a TTY, and typing `write` when prompted
- Exit codes: `0` ok, `1` runtime/API, `2` usage or config, `3` auth, `4` blocked

See [docs/api-scope.md](docs/api-scope.md) for endpoints covered in v0.1.

## Agent skills

Bundled skill (also in [skills/semaphore/SKILL.md](skills/semaphore/SKILL.md)):

```bash
semaphore skill install --agent cursor
semaphore skill print
```

## License

MIT — see [LICENSE](LICENSE).
