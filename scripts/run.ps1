# Запуск инфраструктуры
docker-compose -f infrastructure/docker-compose.yml up -d

# Ждем готовности инфраструктуры (порт 14268)
Write-Host "Waiting for infrastructure..."
while (-not (Test-NetConnection -ComputerName localhost -Port 14268).TcpTestSucceeded) {
    Start-Sleep -Seconds 1
}

# Запуск сервисов
docker-compose -f order/docker-compose.yml up -d
docker-compose -f spot_instrument/docker-compose.yml up -d

Write-Host "All services started!"
