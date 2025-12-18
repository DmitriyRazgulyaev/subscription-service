FROM golang:1.25.4-alpine as builder
LABEL authors="dmitriyrazgulyaev"

RUN apk add --no-cache git make
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build \
    -o /app/bin/subscriptionService \
    -ldflags="-w -s" \
    ./cmd/subscriptionService/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build \
        -o /app/bin/migrate \
        -ldflags="-w -s" \
        ./cmd/migrate/main.go

FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app

COPY --from=builder /app/bin/subscriptionService /app/subscriptionService
COPY --from=builder /app/bin/migrate /app/migrate

COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

RUN chown -R appuser:appuser /app
USER appuser

FROM runtime AS subscriptionService
EXPOSE 50051
ENTRYPOINT ["/app/subscriptionService"]

FROM runtime AS migrate
ENTRYPOINT ["/app/migrate"]