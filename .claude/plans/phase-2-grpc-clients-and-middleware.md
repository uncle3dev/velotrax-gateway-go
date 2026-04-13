# Phase 2: gRPC Clients & Middleware Setup

## Objective
Implement gRPC client connections to downstream services (auth, core) and middleware pipeline (logger, recovery, JWT auth).

---

## Tasks

### 2.1 gRPC Client Implementation

#### 2.1.1 Auth gRPC Client
**File:** `internal/grpc/client/auth_client.go`

- [ ] Create `AuthClient` interface (abstract from implementation)
- [ ] Implement connection logic:
  - Dial with `grpc.WithBlock()` + timeout (fail-fast at startup)
  - Wrap with auth interceptor để attach JWT metadata
  - Singleton pattern — khởi tạo lúc main.go startup
- [ ] Implement methods to forward calls từ HTTP handlers → gRPC:
  - `Register(ctx, req) → response`
  - `Login(ctx, req) → response`
  - `Logout(ctx, req) → response`
  - `Refresh(ctx, req) → response`
- [ ] Error handling: map gRPC status codes → internal errors

#### 2.1.2 Core gRPC Client
**File:** `internal/grpc/client/core_client.go`

- [ ] Create `CoreClient` interface
- [ ] Similar connection logic as auth client
- [ ] Implement methods:
  - `ListOrders(ctx, req) → response`
  - `GetOrder(ctx, req) → response`
  - `TrackOrder(ctx, req) → response`

#### 2.1.3 Auth Interceptor
**File:** `internal/grpc/interceptor/auth.go`

- [ ] Create unary interceptor that:
  - Reads JWT token from `c.Get("token")` in Gin context
  - Injects token vào gRPC metadata: `metadata.Set("authorization", "Bearer <token>")`
- [ ] Attach to both auth_client và core_client tại creation

### 2.2 Middleware Pipeline

#### 2.2.1 Logger Middleware
**File:** `internal/middleware/logger.go`

- [ ] Create Gin middleware that:
  - Logs request: method, path, query params
  - Captures response: status code, latency, response size
  - Output format: Zap structured JSON
  - Unique request ID (trace ID) cho mỗi request

#### 2.2.2 Recovery Middleware
**File:** `internal/middleware/recovery.go`

- [ ] Create Gin middleware that:
  - Catches panics
  - Logs panic stack trace (error level)
  - Returns `500 Internal Server Error` với generic message
  - **Important:** Không leak stack trace ra client ở production

#### 2.2.3 JWT Auth Middleware
**File:** `internal/middleware/auth.go`

- [ ] Create Gin middleware that:
  - Reads `Authorization: Bearer <token>` header
  - Validates JWT signature & expiry locally (không gọi gRPC)
  - **Success case:**
    - Extract userID, roles từ token claims
    - Inject vào Gin context: `c.Set("userID", userID)`, `c.Set("roles", roles)`
    - Inject token vào context: `c.Set("token", token)` để gRPC interceptor dùng
    - Call next()
  - **Failure cases:**
    - Missing token → `401 Unauthorized`
    - Invalid signature/expired → `401 Unauthorized`
    - Insufficient roles (check route permissions) → `403 Forbidden`

### 2.3 Router Setup with Middleware

#### 2.3.1 Router Implementation
**File:** `internal/router/router.go`

- [ ] Create `Setup(engine *gin.Engine, authClient, coreClient, jwtSecret) error` function that:
  - Registers middleware in order: Logger → Recovery → JWT Auth
  - Creates route groups: `/v1/auth`, `/v1/permissions`, `/v1/users`, `/v1/orders`
  - **Auth routes** (no middleware): `/v1/auth/register`, `/v1/auth/login`
  - **Protected routes** (JWT required): `/v1/users/*`, `/v1/permissions/*`
  - **Order routes** (JWT required): `/v1/orders/*`
  - Returns error if setup fails (e.g., invalid JWT secret length)

### 2.4 Integration & Testing

#### 2.4.1 Update main.go
**File:** `cmd/server/main.go`

- [ ] Pass gRPC clients to router.Setup()
- [ ] Verify middleware pipeline execution order

#### 2.4.2 Manual Testing Checklist
- [ ] `docker-compose up`: Server starts without errors
- [ ] Verify middleware logs appear (structured JSON)
- [ ] Test JWT validation:
  - Request without token → 401
  - Request with invalid token → 401
  - Request with valid token → allowed (or handled by handler)

---

## Acceptance Criteria

- [x] gRPC clients connect at startup (fail-fast if services unavailable)
- [x] Middleware pipeline executes in correct order
- [x] JWT middleware validates token locally
- [x] All errors are mapped & logged correctly
- [x] `go build ./...` passes, no warnings
- [x] `golangci-lint` passes (if configured)

---

## Notes

- gRPC proto files assumed available in `proto/auth/` & `proto/core/`
- JWT secret validation already in phase 1
- No handlers implemented in this phase — pure infrastructure
