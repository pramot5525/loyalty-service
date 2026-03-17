FROM golang:1.25-alpine AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 3000
CMD ["air"]

# ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o loyalty-service ./cmd/api

# ─────────────────────────────────────────────
FROM alpine:3.21 AS production

WORKDIR /app
COPY --from=builder /app/loyalty-service .

EXPOSE 3000
CMD ["./loyalty-service"]
