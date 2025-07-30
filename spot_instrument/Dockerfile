FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go mod download

# # генерация документации 
# RUN go install github.com/swaggo/swag/cmd/swag@latest
# # RUN go generate ./...
# RUN swag init -g ./cmd/app/main.go

# собираем основное приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint ./cmd/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app
COPY ./config ./config
COPY .env .env
COPY --from=builder /entrypoint ./
