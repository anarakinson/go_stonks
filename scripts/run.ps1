
# Определяем пути к файлам
$infra = "infrastructure/docker-compose.yml"
$services = @(
    "order/docker-compose.yml",
    "spot_instrument/docker-compose.yml",
    "client/docker-compose.yml"
)

# Запуск инфраструктуры
docker-compose -f $infra up -d --build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Infrastructure failed! Cleaning up..." -ForegroundColor Red
    docker-compose -f $infra down -v
    exit 1
}

# Запуск сервисов
foreach ($service in $services) {
    docker-compose -f $service up -d --build
    if ($LASTEXITCODE -ne 0) {
        Write-Host "$service failed! Cleaning up..." -ForegroundColor Red
        
        # Остановка всех сервисов
        foreach ($s in $services) {
            docker-compose -f $s down -v
        }
        
        # Остановка инфраструктуры
        docker-compose -f $infra down -v
        exit 1
    }
}

# Запуск интерактивного клиента
docker-compose -f client/docker-compose.yml run client

