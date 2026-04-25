# velotrax-gateway-go

API Gateway bằng Go cho hệ thống logistics `velotrax`. Dự án này nhận request HTTP từ client, xác thực JWT, rồi gọi sang các downstream service qua gRPC.

## Repo này làm gì

- Expose REST-style API dưới prefix `/v1`
- Xử lý xác thực JWT bằng middleware của gateway
- Forward request sang:
  - `velotrax-auth-go` qua gRPC cho auth/user flow
  - `velotrax-core-go` qua gRPC cho order flow
- Kết nối MongoDB và tạo indexes khi khởi động
- Ghi log request theo format JSON bằng Zap

## Tech Stack

- Go 1.25
- Gin
- gRPC
- JWT: `github.com/golang-jwt/jwt/v5`
- Config: Viper + biến môi trường
- MongoDB driver v2
- Logging: Zap

## Cấu trúc chính

```text
cmd/server/main.go        # entrypoint
internal/config/          # load và validate env
internal/db/              # kết nối MongoDB + tạo indexes
internal/router/          # khai báo routes Gin
internal/middleware/      # logger, recovery, CORS, JWT auth
internal/handler/auth/    # /v1/auth/*
internal/handler/order/   # /v1/orders/*
internal/grpc/client/     # gRPC clients + metadata
internal/gen/             # code sinh từ proto
proto/                    # source .proto
scripts/gen_proto.sh      # generate protobuf
docker-compose.yml        # gateway + mongodb
```

## Luồng chạy

1. `cmd/server/main.go` load config từ env.
2. Khởi tạo logger, MongoDB, rồi kết nối tới 2 gRPC service.
3. Register routes trong `internal/router/router.go`.
4. Request đi qua middleware theo thứ tự:
   - logger
   - recovery
   - CORS
   - auth middleware nếu route cần auth
5. Handler gọi downstream qua gRPC.

## Routes hiện có

### Auth

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/logout`
- `POST /v1/auth/refresh`
- `GET /v1/auth/profile`
- `PUT /v1/auth/profile`

`POST /v1/auth/refresh` cần header `Authorization: Bearer <refresh-token>` và body:

```json
{
  "refresh_token": "<refresh-token>"
}
```

`PUT /v1/auth/profile` nhận body tùy chọn:

```json
{
  "email": "new@email.com",
  "userName": "new-name",
  "roles": ["ADMIN"]
}
```

### Orders

- `GET /v1/orders`
- `POST /v1/orders`
- `GET /v1/orders/:id`
- `GET /v1/orders/:id/tracking`

`GET /v1/orders` dùng query params `page`, `pageSize`, `status`.

`POST /v1/orders` nhận body tùy chọn với các field tương tự:

```json
{
  "page": 1,
  "pageSize": 20,
  "status": "pending"
}
```

### Health

- `GET /health`

## JWT / Auth

- Gateway đọc header `Authorization: Bearer <token>`.
- Middleware parse JWT và kiểm tra `type`:
  - `access` cho route cần access token
  - `refresh` cho route refresh token
- Gateway lấy user id từ claim chuẩn `sub` thông qua `jwt.RegisteredClaims.Subject`.
- `/v1/auth/profile` forward access token xuống auth service trong field `access_token`.
- Với route orders, gateway forward nguyên header `Authorization` sang gRPC metadata.

Lưu ý:
- Nếu token không có `sub`, downstream có thể báo `missing subject`.
- Gateway không tự ký JWT, token đến từ auth service.

## MongoDB

Gateway kết nối MongoDB lúc startup và tạo indexes cho collection:

- `users`
- `orders`

Database name cố định là `velotrax`.

## Chạy local

### Cần chuẩn bị

- Go 1.25
- MongoDB chạy local hoặc trong Docker
- `velotrax-auth-go` và `velotrax-core-go` đang chạy
- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` nếu muốn regenerate proto

### Cách chạy

```bash
cp .env.example .env
go run ./cmd/server
```

Hoặc dùng Makefile:

```bash
make run
```

## Biến môi trường

File tham chiếu: [.env.example](/Users/uncle3/Projects/velotrax-gateway-go/.env.example)

| Biến | Ý nghĩa | Bắt buộc |
|---|---|---|
| `APP_PORT` | Port HTTP của gateway | Không |
| `APP_ENV` | `development` hoặc `production` | Không |
| `JWT_SECRET` | Secret để verify JWT | Có |
| `JWT_EXPIRY` | Thời gian sống access token | Không |
| `JWT_REFRESH_EXPIRY` | Thời gian sống refresh token | Không |
| `GRPC_AUTH_ADDR` | Địa chỉ gRPC của auth service | Có |
| `GRPC_CORE_ADDR` | Địa chỉ gRPC của core/order service | Có |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` | Không |
| `CORS_ALLOWED_ORIGINS` | Origin được phép gọi từ browser; hiện middleware CORS vẫn đang trả `*` nên biến này chưa được áp dụng | Không |
| `MONGO_URI` | Mongo connection string | Có |

## Scripts / lệnh hữu ích

Từ [Makefile](/Users/uncle3/Projects/velotrax-gateway-go/Makefile):

- `make run` - chạy server
- `make build` - build binary vào `./bin/server`
- `make proto` - regenerate protobuf code
- `make lint` - chạy `golangci-lint`
- `make test` - chạy `go test -v ./...`
- `make docker-build` - build image Docker
- `make docker-run` - `docker compose up -d`
- `make docker-down` - `docker compose down`

## Docker

`Dockerfile` build 2 stage:

- stage build dùng `golang:1.25-alpine`
- stage runtime dùng `alpine:3.21`

`docker-compose.yml` hiện có 2 service:

- `gateway`
- `mongodb`

## Proto / codegen

- File nguồn là `proto/auth/auth.proto` và `proto/order/order.proto`
- Code sinh ra nằm trong `internal/gen/`
- Script `scripts/gen_proto.sh` gọi `protoc`, sau đó move file sinh ra vào `internal/gen/`
- `UserDetail` hiện có `id`, `email`, `user_name`, `roles`
- `PUT /v1/auth/profile` có thể cập nhật `email`, `userName`, `roles`

Không sửa tay các file trong `internal/gen/` nếu có thể tránh được.
