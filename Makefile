
run:
	ENV_PATH=./config/.env go run ./cmd/subscriptionService/main.go



PROTOC = protoc
PROTO_PATH = ./proto/subscription
OUT_PATH = ./proto

generate-proto:
	$(PROTOC) --proto_path=$(PROTO_PATH) --go_out=$(OUT_PATH) --go-grpc_out=$(OUT_PATH) --grpc-gateway_out=$(OUT_PATH) --grpc-gateway_opt=generate_unbound_methods=true $(PROTO_PATH)/subscription.proto

