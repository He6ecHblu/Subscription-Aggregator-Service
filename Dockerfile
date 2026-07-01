FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/subscription-aggregator-service ./cmd/app

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/bin/subscription-aggregator-service /app/subscription-aggregator-service

USER app

EXPOSE 8080

ENTRYPOINT ["/app/subscription-aggregator-service"]
