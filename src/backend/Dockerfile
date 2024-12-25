# Backend Dockerfile
FROM golang:1.23.1 AS builder

WORKDIR /app
COPY . .

# Build the Go app
RUN go mod tidy
RUN go build -o backend main.go

ENV CGO_ENABLED=0

# Runtime stage
FROM gcr.io/distroless/base-debian10
COPY --from=builder /app/backend /backend
CMD ["/backend"]