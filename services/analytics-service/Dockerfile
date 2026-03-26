# --- STAGE 1: Build Stage ---
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY services/analytics-service/go.mod services/analytics-service/go.sum* ./services/analytics-service/
COPY pkg ./pkg
COPY services/analytics-service/. ./services/analytics-service

WORKDIR /app/services/analytics-service
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/analytics-service ./cmd/analytics-api/main.go

# --- STAGE 2: Runtime Stage ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates mysql-client

WORKDIR /root/

COPY --from=builder /app/bin/analytics-service ./main
COPY --from=builder /app/services/analytics-service/migrations ./migrations
COPY --from=builder /app/services/analytics-service/migrate.sh ./migrate.sh

RUN chmod +x ./migrate.sh

EXPOSE 8081

CMD ["sh", "-c", "./migrate.sh && ./main"]
