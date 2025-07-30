#!/bin/bash

# Запуск инфраструктуры
docker-compose -f infrastructure/docker-compose.yml up -d

# Ждем готовности инфраструктуры 
echo "Waiting for infrastructure..."
while ! nc -z localhost 14268; do
  sleep 1
done

# Запуск сервисов
docker-compose -f order/docker-compose.yml up -d
docker-compose -f spot_instrument/docker-compose.yml up -d

echo "All services started!"
