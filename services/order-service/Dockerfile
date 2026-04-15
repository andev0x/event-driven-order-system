# --- STAGE 1: Build Stage ---
# Use golang:alpine for a small build footprint
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go.mod/go.sum first for caching
COPY services/order-service/go.mod services/order-service/go.sum* ./services/order-service/

# Copy shared pkg module before running go mod download to satisfy replace directive
COPY pkg ./pkg
COPY services/order-service/. ./services/order-service

WORKDIR /app/services/order-service
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/order-service ./cmd/order-api/main.go

# --- STAGE 2: Runtime Stage ---
# Use a minimal alpine image for the final executable
FROM alpine:latest

RUN apk --no-cache add ca-certificates mysql-client mariadb-connector-c

WORKDIR /root/

COPY --from=builder /app/bin/order-service ./main
COPY --from=builder /app/services/order-service/migrations ./migrations
COPY --from=builder /app/services/order-service/migrate.sh ./migrate.sh

RUN chmod +x ./migrate.sh

EXPOSE 8080

CMD ["sh", "-c", "./migrate.sh && ./main"]
