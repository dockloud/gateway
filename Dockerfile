# syntax=docker/dockerfile:1.7

FROM golang:1.27-alpine AS builder

ENV GOPRIVATE=github.com/dockloud/*

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=secret,id=github_token \
    set -eu; \
    TOKEN="$(cat /run/secrets/github_token)"; \
    git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"; \
    go mod download; \
    git config --global --unset-all url."https://x-access-token:${TOKEN}@github.com/".insteadOf

COPY . .

RUN CGO_ENABLED=0 go build -o /app/gateway ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway ./gateway

EXPOSE 8080

CMD ["./gateway"]