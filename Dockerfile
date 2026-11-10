# syntax=docker/dockerfile:1.7

FROM golang:1.27-alpine AS builder

RUN apk add --no-cache git

ENV GOPRIVATE=github.com/dockloud/*

WORKDIR /app

COPY go.mod go.sum ./

COPY . .

RUN CGO_ENABLED=0 go build -o /app/gateway ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway ./gateway

EXPOSE 8080

CMD ["./gateway"]