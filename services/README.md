# Services

This directory contains the Go services used by the Docker Compose stack.

| Service | Container host | Gateway route | Local port |
| --- | --- | --- | --- |
| Auth | `auth:8484` | `/auth` | `8484` |
| Task tracker | `task-tracker:8485` | `/tasks` | `8485` |
| Expense tracker | `expense-tracker:8486` | `/expenses` | `8486` |

Each service reads configuration from its own local `.env` file through `docker-compose.yml`.
