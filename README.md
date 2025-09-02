# Uni Backend (Go)

Lightweight backend using Go's stdlib `http.ServeMux`, MongoDB, and JWT-based auth. Swagger (OpenAPI) docs are generated with `swaggo`.

## Features

- **Authentication**: JWT-based auth with refresh tokens
- **Profile Management**: Complete user profiles with document uploads
- **File Storage**: Appwrite integration for secure document storage
- **Document Upload**: Support for SSC certificates and marksheets (PDF only)
- **API Documentation**: Auto-generated Swagger/OpenAPI docs

## Prerequisites

- Go 1.22+
- MongoDB running (local or remote)
- Appwrite instance for file storage

## Quick start

# Configure environment

There's a sample file at `.env.sample`. Copy it and fill values:

```bash
cp .env.sample .env
```

**Required environment variables:**

- `MONGO_URI`: MongoDB connection string
- `MONGO_DB_NAME`: Database name
- `JWT_SECRET`: Secret key for JWT tokens
- `APPWRITE_ENDPOINT`: Appwrite instance URL
- `APPWRITE_PROJECT_ID`: Appwrite project ID
- `APPWRITE_API_KEY`: Appwrite API key
- `APPWRITE_BUCKET_ID`: Appwrite storage bucket ID

See `APPWRITE_SETUP.md` for detailed Appwrite configuration instructions.

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
