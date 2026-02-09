# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Install ca-certificates for HTTPS requests (Gemini API, etc)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env.example ./.env.example

EXPOSE 8080

CMD ["./server"]
