# Uni Backend (Go)

Lightweight backend using Go’s stdlib `http.ServeMux`, MongoDB, and JWT-based auth. Swagger (OpenAPI) docs are generated with `swaggo`.

## Prerequisites

- Go 1.22+
- MongoDB running (local or remote)

## Quick start

# Configure environment

There’s a sample file at `.env.sample`. Copy it and fill values:

```bash
cp .env.sample .env
```

# Install prerequisites (CLI)

**Install dev CLIs:**

```cmd
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/air-verse/air@latest
```

**Fetch project Go modules:**

```cmd
go mod download
```

# Using `make`

Your Makefile targets:

- **Run server (no reload)**

  ```cmd
  make run
  ```

- **Generate Swagger docs**

  ```cmd
  make swag
  ```

- **Run with hot reload (Air)**

  ```cmd
  make watch
  ```

- **Clean generated folders**

  ```cmd
  make clean
  ```

# Without `make` (raw commands)

- **Run server (no reload)**

  ```cmd
  go run ./cmd/server
  ```

- **Generate Swagger docs**

  ```cmd
  swag init -g ./cmd/server/main.go -o ./internal/docs --parseInternal
  ```

- **Run with hot reload (Air)**

  ```cmd
  air -c .air.toml
  ```

- **Clean generated folders**

  ```cmd
  if exist internal\docs rmdir /s /q internal\docs
  if exist tmp rmdir /s /q tmp
  ```
