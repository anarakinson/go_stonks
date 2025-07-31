FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go mod download

# собираем основное приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o entrypoint ./cmd/main.go

EXPOSE 8080

CMD ["./entrypoint"]
