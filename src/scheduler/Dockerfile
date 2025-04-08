# Stage 1: Install Python dependencies
FROM python:3.12.1 AS python-deps
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Stage 2: Build React app
FROM node:18 AS react-builder
WORKDIR /app
COPY static/react-app/package*.json ./
RUN npm install
COPY static/react-app/ ./
RUN npm run build

# Stage 3: Build Go binary
FROM golang:1.23.4 AS go-builder
WORKDIR /app
COPY go.mod .
COPY scheduler.go .
RUN go mod tidy
COPY static ./static
COPY strategies ./strategies
COPY tradepb ./tradepb
RUN go build -o /scheduler scheduler.go

# Stage 4: Final image
FROM python:3.12.1
WORKDIR /app
COPY --from=python-deps /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=python-deps /usr/local/bin /usr/local/bin
COPY --from=go-builder /scheduler /scheduler
COPY --from=go-builder /app/strategies ./strategies
COPY --from=go-builder /app/static ./static
# Copy built React app to static directory
COPY --from=react-builder /app/build ./static/react-app/build

CMD ["/scheduler"]