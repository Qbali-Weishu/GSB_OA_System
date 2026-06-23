FROM golang:1.21-alpine AS builder

WORKDIR /build

# 先复制全部代码，让 go mod tidy 能解析 import
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o oa-server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata bash curl

WORKDIR /app

COPY --from=builder /build/oa-server ./
COPY migrations/ ./migrations/
COPY config.yaml ./

ENV OA_APP_ENV=production \
    OA_APP_PORT=8080 \
    OA_DATABASE_HOST=postgres \
    OA_DATABASE_PORT=5432 \
    OA_DATABASE_USER=oa \
    OA_DATABASE_PASSWORD=oa_pass \
    OA_DATABASE_NAME=oa_leave \
    OA_DATABASE_MIGRATIONSPATH=/app/migrations \
    OA_REDIS_ADDR=redis:6379 \
    OA_JWT_SECRET=change-me-in-production \
    OA_WORKFLOW_AUTOAPPROVE_MAXDAYS=1 \
    OA_WORKFLOW_MINIMUM_ONSITECOUNT=2

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./oa-server"]
