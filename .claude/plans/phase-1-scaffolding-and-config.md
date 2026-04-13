# Phase 1: Scaffolding & Config (COMPLETED)

## Objective
Initialize project structure, setup config loading, environment variables, and Docker configuration.

---

## Completed Tasks

### 1.1 Config Loading & Validation ✅

**File:** `internal/config/config.go`

- ✅ Implemented `Config` struct with all required fields:
  - App: `AppPort`, `AppEnv`
  - JWT: `JWTSecret`, `JWTExpiry`, `JWTRefreshExpiry`
  - gRPC: `GRPCAuthAddr`, `GRPCCoreAddr`
  - Logging: `LogLevel`
  - CORS: `CORSAllowedOrigins`

- ✅ Implemented `Load()` function:
  - Load from environment variables using Viper
  - Set defaults for each field
  - Unmarshal into Config struct
  - Call validate() for required field checks

- ✅ Implemented `validate()` function:
  - Check JWT_SECRET >= 32 characters (required for signing)
  - Check GRPC_AUTH_ADDR not empty
  - Check GRPC_CORE_ADDR not empty
  - Print config object for debugging: `fmt.Printf("Config: %+v\n", cfg)`

### 1.2 Environment Variables & .env Setup ✅

**Files:** `.env`, `.env.example`

- ✅ Created `.env.example` with all required variables:
  ```env
  APP_PORT=8080
  APP_ENV=development
  JWT_SECRET=change_me_to_a_strong_secret_at_least_32_chars
  JWT_EXPIRY=15m
  JWT_REFRESH_EXPIRY=168h
  GRPC_AUTH_ADDR=localhost:50051
  GRPC_CORE_ADDR=localhost:50052
  LOG_LEVEL=info
  CORS_ALLOWED_ORIGINS=http://localhost:3000
  ```

- ✅ Created `.env` with actual values (git-ignored)
  - JWT_SECRET set to valid 32+ char string: `mysupersecretkey1234567890abcdef`

### 1.3 Docker & Docker Compose Setup ✅

**Files:** `Dockerfile`, `docker-compose.yml`

#### Dockerfile
- ✅ Multi-stage build:
  - **Stage 1 (Builder):** golang:1.25-alpine
    - Copy go.mod/go.sum
    - Run `go mod download`
    - Copy source code
    - Build binary: `CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server`
  
  - **Stage 2 (Runtime):** alpine:3.21
    - Install ca-certificates & tzdata for gRPC TLS
    - Copy binary from builder
    - Expose port 8080
    - Entrypoint: `./server`

#### docker-compose.yml
- ✅ Service definition for `gateway`:
  - Build from Dockerfile
  - Container name: `velotrax-gateway`
  - Port mapping: `8080:8080`
  - Load `.env` file: `env_file: .env`
  - Explicitly pass env vars to container via `environment` section:
    ```yaml
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - GRPC_AUTH_ADDR=${GRPC_AUTH_ADDR}
      - GRPC_CORE_ADDR=${GRPC_CORE_ADDR}
    ```
  - Restart policy: `unless-stopped`

### 1.4 Main Entry Point ✅

**File:** `cmd/server/main.go`

- ✅ Implemented `main()` function:
  1. Load config via `config.Load()`
  2. Print debug config object: `fmt.Printf("DEBUG: Loaded Config: %+v\n", cfg)`
  3. Initialize Zap logger based on `cfg.LogLevel`
  4. Log startup info (env, port)
  5. Connect to auth gRPC service (placeholder client creation)
  6. Set Gin mode (debug if development, release if production)
  7. Create Gin engine
  8. Setup routes (placeholder)
  9. Start HTTP server on `APP_PORT` in goroutine
  10. Setup graceful shutdown:
      - Wait for SIGINT/SIGTERM
      - Shutdown server with 5s timeout
      - Log server stopped

- ✅ Implemented `initLogger()` helper:
  - Return Zap logger based on level (debug/production config)
  - Parse log level string → Zap level

- ✅ Implemented `parseLogLevel()` helper:
  - Map string (debug, info, warn, error) → zapcore.Level

### 1.5 Go Module Setup ✅

**File:** `go.mod`, `go.sum`

- ✅ Initialized Go module: `module github.com/uncle3dev/velotrax-gateway-go`
- ✅ Set Go version: 1.24+ (specified in Dockerfile: 1.25)
- ✅ Added main dependencies:
  - `github.com/gin-gonic/gin` (HTTP framework)
  - `go.uber.org/zap` (structured logging)
  - `github.com/spf13/viper` (config management)
  - `google.golang.org/grpc` (gRPC client, will be used in phase 2)

---

## Testing & Verification ✅

- ✅ `docker-compose up --build` successfully builds image
- ✅ Container starts and runs server
- ✅ Config loads from `.env` file
- ✅ Debug print shows correct config values:
  ```
  Config: &{AppPort:8080 AppEnv:development JWTSecret:mysupersecretkey1234567890abcdef ...}
  ```
- ✅ Logger initializes successfully
- ✅ Server listens on port 8080

---

## Current Status

**Phase 1 COMPLETE** — Scaffolding & config layer ready.

### Ready for Phase 2:
- gRPC client setup (auth, core)
- Middleware pipeline (logger, recovery, JWT auth)
- Router configuration with all endpoints

### Known Limitations:
- gRPC clients not yet connected (will fail until auth/core services available)
- No HTTP handlers implemented
- No routes registered
- Router.Setup() placeholder only

---

## Notes

- All environment variables required by config validator must be present in `.env`
- JWT_SECRET is validated to be >= 32 chars for secure signing
- Docker-compose mounts `.env` file and passes via environment section for Viper AutomaticEnv()
- No dependencies on downstream services in phase 1 — pure local config & logging setup
