# -----------------------------------------------
# Stage 1: Install Python dependencies
# -----------------------------------------------
FROM python:3.12.1 AS python-deps
WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# -----------------------------------------------
# Stage 2: Build Go binary
# -----------------------------------------------
FROM golang:1.23.4 AS builder
WORKDIR /app

# Copy modules for go mod tidy
COPY go.mod .
COPY scheduler.go .
RUN go mod tidy

# Copy any additional files needed for building
COPY static ./static
COPY strategies ./strategies

# Build the binary
RUN go build -o /scheduler scheduler.go

# -----------------------------------------------
# Stage 3: Final image
# -----------------------------------------------
FROM python:3.12.1
WORKDIR /app

# Copy Python deps from python-deps stage
COPY --from=python-deps /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=python-deps /usr/local/bin /usr/local/bin

# Copy the Go binary and any required assets
COPY --from=builder /scheduler /scheduler
COPY --from=builder /app/static ./static
COPY --from=builder /app/strategies ./strategies

CMD ["/scheduler"]
