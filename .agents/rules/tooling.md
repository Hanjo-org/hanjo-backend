# Tooling & Execution Environment Guidelines

## 1. Docker Management Policy
* **Prohibition**: Agents must **NEVER** autonomously execute Docker or Docker Compose commands (`docker ps`, `docker-compose up`, `docker-compose down`, `docker exec`, etc.).
* **User Control**: The user manages their own Docker environment, daemon state, and container lifecycles independently.
* **Database & Services**: When running database migrations, tests, or application servers, rely on standard Go tooling or provide the Makefile / CLI command for the user to execute if container startup is required.

## 2. Command Execution Guidelines
* **Safe Local Commands**: Prefer running pure Go commands:
  - `go test ./...`
  - `go build ./...`
  - `go vet ./...`
  - `git status`
* **Migrations**: To apply migrations, provide the `make migrate-up` or `migrate` command to the user or run it only if local database connectivity is already established and without launching Docker.
