set PROJECT_NAME="github.com/anarakinson/go_stonks/stonks_pb"
set GEN_VERSION=v1

protoc -I ./proto --go_out=. --go_opt=module=%PROJECT_NAME% --go-grpc_out=. --go-grpc_opt=module=%PROJECT_NAME% ./proto/market/%GEN_VERSION%/*.proto
protoc -I ./proto --go_out=. --go_opt=module=%PROJECT_NAME% --go-grpc_out=. --go-grpc_opt=module=%PROJECT_NAME% ./proto/order/%GEN_VERSION%/*.proto
protoc -I ./proto --go_out=. --go_opt=module=%PROJECT_NAME% --go-grpc_out=. --go-grpc_opt=module=%PROJECT_NAME% ./proto/spot_instrument/%GEN_VERSION%/*.proto
