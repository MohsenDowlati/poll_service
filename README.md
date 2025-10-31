# Poll Service (Go Clean Architecture)

Poll Service is a reference implementation of a poll management backend written in Go. It follows a clean architecture layout on top of Gin and MongoDB, provides JWT-powered authentication, and exposes interactive API docs via Swagger.

## Features
- Multistage authentication flow with signup, login, refresh tokens, and cookie helpers
- Admin-focused poll management (create, edit, delete, list) mapped to sheets
- Client poll participation endpoints with vote tracking and participation limits
- Sheet orchestration and notification workflows for onboarding new participants
- MongoDB data access through repository interfaces and domain-focused use cases
- Auto seeding for a super administrator account when env variables are provided

## Tech Stack
- Go 1.19
- Gin Web Framework
- MongoDB 6.x
- JWT (github.com/golang-jwt/jwt/v4)
- Swagger via swaggo
- Docker + Docker Compose (optional)

## Project Layout
```
cmd/             Entry point (`cmd/main.go`) and HTTP server bootstrap
bootstrap/       Environment loading, Mongo client, and seed helpers
api/controller/  HTTP handlers mapped to use cases
api/middleware/  Authorization middleware and supporting filters
api/route/       Route registration grouped by feature
domain/          Entities, DTOs, contracts, and mocks used for testing
repository/      Mongo-backed repository implementations
usecase/         Business logic orchestrating repositories and domain rules
internal/        Helper utilities (e.g., token helpers, fake data)
docs/            Generated Swagger documentation
```

## Getting Started
1. Install [Go 1.19+](https://go.dev/dl/) and ensure `go` is on your `PATH`.
2. Clone the repository and move into the project directory.
3. Copy `.env.example` to `.env` and update the values for your setup:
   - `SERVER_ADDRESS` controls the Gin listen address (default `:8080`).
   - `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` configure MongoDB connection.
   - `ACCESS_TOKEN_SECRET` and `REFRESH_TOKEN_SECRET` secure JWT generation.
   - Super admin fields seed an initial admin user when the service starts.
4. Start MongoDB locally or run the stack with Docker (see below).

## Running the Service
```bash
# Run against your local environment
go run ./cmd/main.go

# Rebuild Swagger docs after changing annotations
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/main.go -o docs
```
The Swagger UI is available at `http://localhost:8080/swagger/index.html` once the service is running (replace the host if you change `SERVER_ADDRESS`).

## Docker Compose
The provided `docker-compose.yaml` spins up the API and MongoDB together. Ensure `.env` holds the same port values exposed by Docker.
```bash
docker compose up --build
```
MongoDB data persists in the named `dbdata` volume. Stop with `Ctrl+C` and remove resources via `docker compose down` when you are finished.

## Testing
Run the full test suite:
```bash
go test ./...
```
Mocks for interfaces live under `domain/mocks` to simplify unit testing of use cases and repositories.

## Authentication Notes
- Access tokens embed user claims and are expected in the `Authorization: Bearer <token>` header for protected routes.
- Refresh tokens support session renewal through the refresh endpoint.
- `bootstrap/seeder.go` automatically provisions a super admin user if the required env variables are present when the service boots.
- Cookie metadata (domain, secure flag, same-site) is configurable through environment variables to support multiple deployment targets.

## Contributing
1. Fork the project and create a feature branch.
2. Make your changes with clear commits and accompanying tests when applicable.
3. Run `go test ./...` before opening a pull request.
4. Describe the motivation, approach, and testing in the PR template.

## License
This project is licensed under the [MIT License](LICENSE).
