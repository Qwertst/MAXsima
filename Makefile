BINARY := maxsima
PROTO_DIR := proto
GEN_DIR := proto/gen

.PHONY: all build proto clean

all: build

proto:
	protoc \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/chat.proto

build:
	go build -o $(BINARY) ./cmd/chat

clean:
	rm -f $(BINARY)
