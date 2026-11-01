# Semaphore API scope (semaphore-cli v0.1)

This document lists REST endpoints used by **semaphore-cli v0.1**. The CLI is unofficial and community-built; it is not affiliated with Semaphore UI.

Authentication for all calls below: header `Authorization: Bearer <token>` (from env `SEMAPHORE_API_TOKEN` only). Base URL from env `SEMAPHORE_URL` or flag `--url`. OpenAPI source: [semaphoreui/semaphore `api-docs.yml`](https://github.com/semaphoreui/semaphore/blob/develop/api-docs.yml) (`basePath: /api`).

## Covered in v0.1

| Command | Method | Path | OpenAPI (`api-docs.yml`) |
|---------|--------|------|--------------------------|
| `semaphore ping` | GET | `/api/ping` | line 2080 (`/ping`) |
| `semaphore health`, `semaphore version` / `info` | GET | `/api/info` | line 2108 (`/info`) |
| `semaphore projects list` | GET | `/api/projects` | line 2699 (`/projects`) |
| `semaphore templates list --project` | GET | `/api/project/{project_id}/templates` | line 3711 |
| `semaphore task list --project` | GET | `/api/project/{project_id}/tasks` | line 4154 |
| `semaphore task get --project --id` | GET | `/api/project/{project_id}/tasks/{task_id}` | line 4250 |
| `semaphore task output --project --id` | GET | `/api/project/{project_id}/tasks/{task_id}/output` | line 4271 |
| `semaphore inventories list --project` | GET | `/api/project/{project_id}/inventory` | line 3555 |
| `semaphore environments list --project` | GET | `/api/project/{project_id}/environment` | line 3632 |
| `semaphore keys list --project` | GET | `/api/project/{project_id}/keys` | line 3320 (response stripped to id/name/type/project_id) |
| `semaphore task run --template --project --write` | POST | `/api/project/{project_id}/tasks` | line 4168 (`post` under `/project/{project_id}/tasks`) |
| `semaphore api` | * | `/api/...` | Escape hatch; query via `--query key=value` |

List endpoints that require `sort` and `order` in OpenAPI use defaults `sort=name` and `order=asc` (or `sort=type` where applicable for keys via `name`).

## Planned (not in v0.1 yet)

- `GET /api/project/{project_id}/tasks/last`
- `GET /api/project/{project_id}/tasks/{task_id}/raw_output`
- Project create/update/delete
- Template and inventory mutations

## Out of scope

- Routes outside `/api/` without explicit confirmation
- Returning secret key material from `keys list` (CLI never exposes secrets)
- Non-interactive writes (agents must not pass `--write`)
