# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o admin ./cmd/admin/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/api /app/admin /app/
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080 8081

CMD ["./api"]
