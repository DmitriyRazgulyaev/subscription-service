lint:
	golangci-lint run -v

run:
	ENV_PATH=./config/.env go run ./cmd/subscriptionService/main.go

migrate-up:
	ENV_PATH=./config/.env go run ./cmd/migrate/main.go ./migrations -command up

migrate-down:
	ENV_PATH=./config/.env go run ./cmd/migrate/main.go ./migrations -command down

migrate-version:
	ENV_PATH=./config/.env go run ./cmd/migrate/main.go ./migrations -command version

docker-build:
	docker-compose --env-file config/.env up --build -d

docker-up:
	docker-compose --env-file config/.env up -d

PROTOC = protoc
PROTO_PATH = ./proto/subscription
OUT_PATH = ./proto

generate-proto:
	$(PROTOC) --proto_path=$(PROTO_PATH) --go_out=$(OUT_PATH) --go-grpc_out=$(OUT_PATH) --grpc-gateway_out=$(OUT_PATH) --grpc-gateway_opt=generate_unbound_methods=true $(PROTO_PATH)/subscription.proto

