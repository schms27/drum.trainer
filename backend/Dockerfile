# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

ARG APP_VERSION=dev

WORKDIR /build

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath       -ldflags="-s -w -X github.com/simu/drum-trainer/backend/api.Version=${APP_VERSION}"       -o drum-trainer-backend .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/drum-trainer-backend .

# Migrations are embedded in the binary; only a writable data volume is needed.
VOLUME ["/data"]

ENV PORT=8080 \
    DB_PATH=/data/drum_trainer.db \
    UPLOAD_DIR=/data/uploads

EXPOSE 8080
CMD ["./drum-trainer-backend"]
