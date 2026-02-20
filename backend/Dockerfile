# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o drum-trainer-backend .

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
