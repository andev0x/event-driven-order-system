# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY services/notification-worker/go.mod services/notification-worker/go.sum* ./services/notification-worker/
COPY pkg ./pkg
COPY services/notification-worker/. ./services/notification-worker

WORKDIR /app/services/notification-worker
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/notification-worker ./cmd/notification-worker/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/notification-worker ./main

CMD ["./main"]
