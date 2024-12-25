# Scheduler Dockerfile
FROM golang:1.23.4 AS builder

WORKDIR /app
COPY . .

# Build the Go app
RUN go mod tidy
RUN go build -o scheduler scheduler.go

# Python Runtime stage
FROM python:3.12.1

WORKDIR /app
COPY --from=builder /app/scheduler /scheduler
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .
CMD ["/scheduler"]