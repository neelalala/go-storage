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

migrate-up:
	docker compose run --rm migrate up $(step)

migrate-down:
	docker compose run --rm migrate down $(step)

migrate-version:
	docker compose run --rm migrate version

migrate-force:
	docker compose run --rm migrate force $(version)

prod-build:
	docker compose -f compose.yaml -f compose.prod.yaml build

prod-up:
	docker compose -f compose.yaml -f compose.prod.yaml up -d

prod-down:
	docker compose -f compose.yaml -f compose.prod.yaml down

prod-migrate-up:
	docker compose -f compose.yaml -f compose.prod.yaml run --rm migrate up $(step)

prod-migrate-down:
	docker compose -f compose.yaml -f compose.prod.yaml run --rm migrate down $(step)

prod-migrate-version:
	docker compose -f compose.yaml -f compose.prod.yaml run --rm migrate version

.PHONY: protobuf build up down clear migrate-up migrate-down migrate-version migrate-force prod-build prod-up prod-down prod-migrate-up prod-migrate-down

