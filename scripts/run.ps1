# Запуск инфраструктуры
docker-compose -f infrastructure/docker-compose.yml up -d --build

# Ждем готовности инфраструктуры (порт 14268)
Write-Host "Waiting for infrastructure..."
while (-not (Test-NetConnection -ComputerName localhost -Port 14268).TcpTestSucceeded) {
    Start-Sleep -Seconds 1
}

# Запуск сервисов
docker-compose -f spot_instrument/docker-compose.yml up -d --build
docker-compose -f order/docker-compose.yml up -d --build
# Собираем и запускаем клиент в фоне
docker-compose -f client/docker-compose.yml up -d --build

Write-Host "All services started!"

# Ждем, пока клиент начнет слушать порт 8080 (или другой признак готовности)
Write-Host "Waiting for client to be ready (port 8080)..."
while (-not (Test-NetConnection -ComputerName localhost -Port 8080).TcpTestSucceeded) {
    Start-Sleep -Seconds 1
}

Write-Host "All services started! Launching interactive client..."

# Запускаем интерактивную сессию
docker-compose -f client/docker-compose.yml run client
