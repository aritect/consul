FROM golang:1.24-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

COPY . .

WORKDIR /workspace

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o target/consul-telegram-bot ./cmd/consul-telegram-bot
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o target/db-migrate ./cmd/db-migrate
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o target/db-dump ./cmd/db-dump

FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates

WORKDIR /workspace

RUN addgroup -g 1000 consul && \
    adduser -D -u 1000 -G consul consul

RUN mkdir -p /workspace/data && \
    chown -R consul:consul /workspace/data

USER consul

COPY --from=builder /workspace/target/consul-telegram-bot ./consul-telegram-bot
COPY --from=builder /workspace/target/db-migrate ./db-migrate
COPY --from=builder /workspace/target/db-dump ./db-dump
COPY --from=builder /workspace/assets ./assets

EXPOSE 8080

CMD ["./consul-telegram-bot"]
