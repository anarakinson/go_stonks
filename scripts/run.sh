#!/bin/bash

# Запуск инфраструктуры
docker-compose -f infrastructure/docker-compose.yml up -d --build

# Ждем готовности инфраструктуры 
echo "Waiting for infrastructure..."
while ! nc -z localhost 14268; do
  sleep 1
done

# Запуск сервисов
docker-compose -f spot_instrument/docker-compose.yml up -d --build
docker-compose -f order/docker-compose.yml up -d --build
# Собираем и запускаем клиент в фоне
docker-compose -f client/docker-compose.yml up -d --build

echo "All services started!"

while ! nc -z localhost 8080; do
  sleep 1
done

# Запускаем интерактивную сессию
docker-compose -f client/docker-compose.yml run client

