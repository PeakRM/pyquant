# Backend Dockerfile
FROM golang:1.23.1 AS builder

WORKDIR /app
COPY . .

# Build the Go app
RUN go mod tidy
RUN go build -o backend main.go

# Use Debian for runtime
FROM debian:bookworm-slim
WORKDIR /root/
COPY --from=builder /app/backend /backend

# Install PostgreSQL client libraries
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq5 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

CMD ["/backend"]