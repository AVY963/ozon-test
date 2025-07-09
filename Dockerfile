# Используем официальный образ Golang для сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Используем минимальный образ для запуска
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS-запросов
RUN apk --no-cache add ca-certificates

# Создаем пользователя app
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/migrations ./migrations

# Меняем владельца файлов
RUN chown -R appuser:appgroup /root

# Переключаемся на пользователя app
USER appuser

EXPOSE 8080

CMD ["./main"] 