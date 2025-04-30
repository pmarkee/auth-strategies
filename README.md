# Authentication Strategies

This project showcases a few basic authentication methods in Go. It is not meant to be production ready or used as a
dependency in a real project, but rather as educational material or simply a template to get started with.

## Features

- 🔒 Basic Authentication
- 🔒 Email + password login with argon2 hashing and server side sessions
- 🪪 JWT access token based sessions
- 🔑 API key authentication
- 📄 OpenAPI 2.0 docs via `swaggo`
- 💾 SQL-first approach to persistence via `sqlc` and `golang-migrate`
- 🧭 Routing via `chi`
- 🏗️ Pooled database connections to PostgreSQL via `pgx` and `pgxpool`
- 🐳 Dockerized for easy setup
- 📁 Following the [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

## How to use

### Just running

1. install Docker and Docker Compose
2. clone the repo
3. in the repo root, `docker compose up -d --build`
4. in your browser, open `localhost:8080`

### Development

#### Setup

1. install dependencies
    - `go`, `make`, `docker`, `docker-compose`
    - [swag](https://github.com/swaggo/swag)
    - [migrate](https://github.com/golang-migrate/migrate)
    - [sqlc](https://github.com/sqlc-dev/sqlc)
2. `make deps`
3. setup postgres: `docker compose up -d database`
4. `go run ./cmd/server`

#### Project layout

The project mostly follows the [Standard Project Layout](https://github.com/golang-standards/project-layout).

- Any executables (in our case `server` and `migrate`) live in the `/cmd` directory in their own packages.
- Our `config.yaml` is located in `/configs` - note that this is embedded into the binary.
- `/internal` is where all of our own logic is located.
  - `/internal/auth` has the actual authentication endpoints and logic (in `handler.go` and `service.go` respectively)
  and our middlewares (in the `*_auth.go` files)
  - `/internal/user` hosts the protected routes - which in this case is all the same route replicated for each
  authentication method.

#### Common tasks

#### DB schema changes

To modify DB schema:

1. write a pair of up and down-migrations in `internal/db/migrations` (following naming schema).
2. apply the migrations via `make migrate`, or run the full server and the migrations will be applied on startup.

#### DB interaction in code

To add new DB interactions to the application:
1. write your query in `internal/db/queries/queries.sql` (or another SQL file in the same directory). Make sure to
follow the format required by SQLC!
2. run `make generate` to generate boilerplate data access layer code
3. call the newly generated boilerplate function from your code (preferably in the service layer)

#### Adding API docs

To add OpenAPI documentation to your endpoint:

1. write a doc comment - see the `swaggo/swag` repo for specs, or one of the `handler.go` files of this project for
examples. Related request and response structs can also be annotated!
2. run `make docs`.

## Q&A

### Where are the unit tests?
This project is mostly glue code with minimal business logic. Unit tests would end up being a tower of mocks and would
mean little as to the correctness of the system. Integration tests would be more suitable, but their setup would be
overkill for a small, non-critical showcase like this.

### Most of the world uses OAuth now, why is it missing?
Including OAuth would require a frontend, extra configuration, and because it's quite difficult to get working on a
local setup, a proper deployment with a domain name. That’s all out of scope for this demo.

### You have secrets checked into version control!
Indeed, and that's something that should never be done with a production application. However, secrets management is
outside the scope of this project.

### Why are you embedding your config.yaml and migration scripts into the binary?
For a production application I would avoid this, as a change to the scripts would require a rebuild. However, for the
purposes of this showcase it's easier than dealing with the filesystem or an external service, and the impact on binary
size is negligible.
