# Используем официальный образ Golang как базовый
FROM golang:1.18 as builder

# Устанавливаем рабочий каталог внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка Go-приложения
RUN go build -o /streamer

# Используем минимальный образ для финального контейнера
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ffmpeg

# Копируем собранное приложение из образа сборки
COPY --from=builder /streamer /streamer

# Устанавливаем рабочий каталог внутри контейнера
WORKDIR /

# Указываем команду запуска контейнера
CMD ["/streamer"]

# Открываем порт 8080 для входящих запросов
EXPOSE 8080
