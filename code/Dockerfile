FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o blog-engine ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl
WORKDIR /app
COPY --from=builder /app/blog-engine .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
CMD ["./blog-engine"]
