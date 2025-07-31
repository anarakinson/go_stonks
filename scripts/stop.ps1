docker-compose -f client/docker-compose.yml down --remove-orphans
docker-compose -f spot_instrument/docker-compose.yml down --remove-orphans
docker-compose -f order/docker-compose.yml down --remove-orphans
docker-compose -f infrastructure/docker-compose.yml down --remove-orphans

