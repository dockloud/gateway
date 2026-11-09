FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/gateway ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway ./gateway

EXPOSE 8080

CMD ["./gateway"]