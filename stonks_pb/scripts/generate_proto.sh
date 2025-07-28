#!/bin/bash
PROJECT_NAME="github.com/anarakinson/go_stonks/stonks_pb"

protoc -I ./proto --go_out=. --go_opt=module="$PROJECT_NAME" --go-grpc_out=. --go-grpc_opt=module="$PROJECT_NAME" ./proto/*/*.proto
