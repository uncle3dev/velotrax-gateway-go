# AGENTS.md

## Mục tiêu repo

- Đây là API Gateway bằng Go cho hệ thống `velotrax`.
- Nhiệm vụ chính: nhận HTTP request, verify JWT, rồi forward sang auth/core service qua gRPC.

## File nguồn sự thật

- Entrypoint: `cmd/server/main.go`
- Route map: `internal/router/router.go`
- Config/env: `internal/config/config.go` và `.env.example`
- JWT/auth: `internal/middleware/auth.go`
- gRPC client/metadata: `internal/grpc/client/*`
- MongoDB/indexes: `internal/db/mongo.go`
- Contract gRPC: `proto/auth/auth.proto`, `proto/order/order.proto`
- Generated code: `internal/gen/`

## Quy ước làm việc

- Ưu tiên đọc code hiện có trước khi sửa.
- Không bịa thêm route/feature không có trong repo.
- Không sửa file generated trong `internal/gen/` nếu không thật sự cần.
- Khi đổi contract `.proto`, phải chạy lại `make proto`.
- Khi đổi route/auth/data flow, kiểm tra cả handler, middleware và gRPC client liên quan.

## Chỗ thường sửa

- Thêm hoặc chỉnh route trong `internal/router/router.go`
- Chỉnh logic auth/JWT trong `internal/middleware/auth.go`
- Chỉnh handler trong `internal/handler/auth/` và `internal/handler/order/`
- Chỉnh env/config trong `internal/config/config.go` và `.env.example`
- Chỉnh kết nối downstream trong `internal/grpc/client/`

## Không nên đụng

- `internal/gen/` nếu chỉ cần thay logic ứng dụng
- File sinh ra từ `make proto`
- `docker-compose.yml` chỉ để test luồng local, trừ khi task yêu cầu

## Lưu ý kỹ thuật

- JWT access token phải có `type = access`
- JWT refresh token phải có `type = refresh`
- Gateway đọc user id từ claim chuẩn `sub` qua `jwt.RegisteredClaims.Subject`
- Route `/v1/orders/*` forward nguyên `Authorization: Bearer <token>` sang gRPC metadata
- Nếu token thiếu `sub`, downstream có thể trả `Unauthenticated: missing subject`
- `CORS_ALLOWED_ORIGINS` đang có trong config nhưng middleware CORS hiện trả `*`

## Chạy kiểm tra

- Build: `go build ./...`
- Test: `go test ./...`
- Generate proto: `make proto`

## Khi cần cập nhật tài liệu

- Nếu đổi port, env, route, hoặc luồng auth, cập nhật luôn `README.md`
- Nếu đổi quy ước làm việc hoặc file nguồn sự thật, cập nhật luôn `AGENTS.md`
