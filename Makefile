protobuf:
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               pkg/proto/storage/storage.proto
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               pkg/proto/metadata/metadata.proto
	protoc --go_out=. --go_opt=paths=source_relative \
                   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
                   pkg/proto/users/users.proto
build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

clear:
	docker compose down -v --remove-orphans
	rm -rf ./pgdata
	mkdir -p ./pgdata
	rm -rf ./uploads/node1 ./uploads/node2 ./uploads/node3
	mkdir -p ./uploads/node1 ./uploads/node2 ./uploads/node3

migrate-up:
	docker compose run --rm migrate up $(step)

migrate-down:
	docker compose run --rm migrate down $(step)

migrate-version:
	docker compose run --rm migrate version

migrate-force:
	docker compose run --rm migrate force $(version)

.PHONY: protobuf build up down clear migrate-up migrate-down migrate-version migrate-force

