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
	@./scripts/release.sh build $(version)

prod-release:
	@./scripts/release.sh release $(version) "$(desc)"

prod-run:
	@./scripts/release.sh run

prod-down:
	@./scripts/release.sh down

prod-rollback:
	@./scripts/release.sh rollback "$(release)"

prod-releases:
	@./scripts/release.sh list

prod-current:
	@./scripts/release.sh current

prod-migrate-up:
	@./scripts/release.sh migrate up "$(step)"

prod-migrate-down:
	@./scripts/release.sh migrate down "$(step)"

prod-migrate-version:
	@./scripts/release.sh migrate version

prod-migrate-force:
	@./scripts/release.sh migrate force "$(version)"

.PHONY: protobuf build up down clear migrate-up migrate-down migrate-version migrate-force prod-build prod-release prod-run prod-down prod-rollback prod-releases prod-current prod-migrate-up prod-migrate-down prod-migrate-version prod-migrate-force

