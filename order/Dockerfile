FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

# # Установка protoc
# RUN apt-get update && apt-get install -y protobuf-compiler

# # Установка Go плагинов
# RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
# RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
# RUN export PATH="$PATH:$(go env GOPATH)/bin"

# # Генерируем код
# RUN protoc -I ./proto \
    # --go_out=. \
    # --go_opt=module=stonks \
    # --go-grpc_out=. \
    # --go-grpc_opt=module=stonks  \
    # ./proto/*.proto

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
# COPY ./config ./config
COPY .env .env
COPY --from=builder /entrypoint ./
