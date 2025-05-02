FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o person-enrichment-service ./cmd/service

FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/person-enrichment-service .
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/docs/swagger /app/docs/swagger

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://${HTTP_HOST:-0.0.0.0}:${HTTP_PORT:-8080}/health || exit 1

ENTRYPOINT ["./person-enrichment-service"]