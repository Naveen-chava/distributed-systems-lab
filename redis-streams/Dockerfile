FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o producer ./cmd/producer/main.go
RUN go build -o consumer ./cmd/consumer/main.go

# Use a minimal runtime image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/producer .
COPY --from=builder /app/consumer .

# Default command (can be overridden in docker-compose)
CMD ["./producer"]
