# Uni Backend (Go)

Lightweight backend using Go’s stdlib `http.ServeMux`, MongoDB, and JWT-based auth. Swagger (OpenAPI) docs are generated with `swaggo`.

## Prerequisites

- Go 1.22+
- MongoDB running (local or remote)
- `swag` CLI:

  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

## Quick start

### 1) Configure environment

There’s a sample file at `.env.sample`. Copy it and fill values:

```bash
cp .env.sample .env
```

### 2) Run the server

```bash
go run ./cmd/server
```

Server boots at `http://<HOST>:<PORT>`.

### 3) API docs (Swagger UI)

After generating docs (see below), open:

```
http://<HOST>:<PORT>/swagger/index.html
```

## Generate Swagger docs

```bash
cd cmd/server
```

```bash
swag init --generalInfo cmd/server/main.go --dir cmd/server,internal/handlers,internal/routes --output internal/docs --parseInternal --parseDependency
```

## Dev tips

- Protected routes use Bearer tokens (Authorization: `Bearer <jwt>`). Your middleware also accepts the `access_token` cookie.
