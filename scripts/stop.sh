#!/bin/bash

docker-compose -f spot_instrument/docker-compose.yml down
docker-compose -f order/docker-compose.yml down
docker-compose -f infrastructure/docker-compose.yml down
