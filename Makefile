dbConnStr = "postgresql://postgres:postgres@localhost:5432/crypto_gateway_test"
migrationsDir = ./sql/migrations
sqlcFile = ./sql/sqlc.yaml

protoPathDir = ./proto
protoGoOutDir = ./internal/pb
protoGoGrpcOutDir = $(protoGoOutDir)

cmdServerDir = ./cmd/server

clean-pb:
	rm -rf $(protoGoOutDir)/*

gen-pb:
	./script/generate_pb.sh -i $(protoPathDir) -o $(protoGoOutDir) v1

run-migrations:
	goose -dir $(migrationsDir) postgres $(dbConnStr) up

reset-migrations:
	goose -dir $(migrationsDir) postgres $(dbConnStr) reset

run-sqlc-gen:
	sqlc -f $(sqlcFile) generate

build:
	go build -o ./bin/server $(cmdServerDir)/main.go

build-debug:
	go build -gcflags=all="-N -l" -o ./bin/server $(cmdServerDir)/main.go