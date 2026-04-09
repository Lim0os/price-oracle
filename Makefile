APP_NAME=price-oracle
CMD_DIR=./cmd/price-oracle

# Default DB connection (localhost via docker compose)
DB_DSN ?= "postgres://postgres:postgres@localhost:5432/price_oracle?sslmode=disable"

.PHONY: build run test lint docker-build proto migrate-up migrate-down migrate-force clean

build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(APP_NAME):local .

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

proto:
	protoc \
		-I. \
		-I$(shell go env GOPATH)/pkg/mod/github.com/bufbuild/protocompile@v0.14.1/wellknownimports \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		proto/rates/v1/rates.proto

migrate-up:
	migrate -path migrations -database $(DB_DSN) up

migrate-down:
	migrate -path migrations -database $(DB_DSN) down

migrate-force:
	migrate -path migrations -database $(DB_DSN) force $(VERSION)

clean:
	rm -rf bin/
	rm -f $(APP_NAME)
