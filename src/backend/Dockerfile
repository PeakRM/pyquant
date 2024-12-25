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
COPY --from=builder /app/positions.json /positions.json

CMD ["/backend"]