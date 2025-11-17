FROM golang:1.24-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

COPY cmd/consul-telegram-bot/go.mod cmd/consul-telegram-bot/go.sum ./cmd/consul-telegram-bot/

WORKDIR /workspace/cmd/consul-telegram-bot

RUN go mod download

COPY cmd/consul-telegram-bot/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o target/consul-telegram-bot ./cmd/consul-telegram-bot

FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates

WORKDIR /workspace

RUN addgroup -g 1000 aritect && \
    adduser -D -u 1000 -G aritect aritect

RUN mkdir -p /workspace/data && \
    chown -R aritect:aritect /workspace/data

USER aritect

COPY --from=builder /workspace/cmd/consul-telegram-bot/target/consul-telegram-bot ./consul-telegram-bot

EXPOSE 8080

CMD ["./consul-telegram-bot"]
