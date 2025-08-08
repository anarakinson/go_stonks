#!/bin/bash

# Определяем пути к файлам
infra="infrastructure/docker-compose.yml"
services=(
    "order/docker-compose.yml"
    "spot_instrument/docker-compose.yml"
    "client/docker-compose.yml"
)

# Функция для очистки
cleanup() {
    echo "Cleaning up..."
    # Останавливаем сервисы
    for service in "${services[@]}"; do
        docker-compose -f "$service" down -v
    done
    # Останавливаем инфраструктуру
    docker-compose -f "$infra" down -v
}

# Запуск инфраструктуры
echo "Starting infrastructure..."
docker-compose -f "$infra" up -d --build
if [ $? -ne 0 ]; then
    echo -e "Infrastructure failed! Cleaning up..."
    cleanup
    exit 1
fi

# Запуск сервисов
for service in "${services[@]}"; do
    echo "Starting $service..."
    docker-compose -f "$service" up -d --build
    if [ $? -ne 0 ]; then
        echo -e "$service failed! Cleaning up..."
        cleanup
        exit 1
    fi
done

# Запуск интерактивного клиента
echo "Starting client session..."
docker-compose -f "client/docker-compose.yml" run client
