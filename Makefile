.PHONY: help run build proto-gen

# Application
APP_NAME=kfc-notification
BINARY_NAME=notification

help:
	@echo "Available targets:"
	@echo "  run		Run the server"
	@echo "  build		Build binary to bin/notification"
	@echo "  proto-gen	Generate protobuf files"

# Run the application
run:
	@go run cmd/notification/main.go

# Build the application
build:
	@mkdir -p bin
	@go build -o bin/$(BINARY_NAME) cmd/notification/main.go
	@echo "built bin/notification"

# Generate the protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	@find proto -name '*.proto' -print0 | xargs -0 -n1 \
		protoc \
			--go_out=. \
			--go_opt=module=github.com/ramisoul84/kfc-notification \
			--go-grpc_out=. \
			--go-grpc_opt=module=github.com/ramisoul84/kfc-notification \
			--proto_path=.
	@echo "✅ Proto files generated"