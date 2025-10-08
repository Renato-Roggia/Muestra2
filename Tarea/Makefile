MODULE=Tarea
PROTO_DIR=proto
PROTO_FILE=$(PROTO_DIR)/heist.proto

.PHONY: proto build run-lester run-michael run-franklin run-trevor clean

all: init proto build

init:
	sudo service rabbitmq-server start
	go mod init Tarea
	go mod tidy

proto:
	protoc --go_out=. --go-grpc_out=. $(PROTO_FILE)

build:
	go build -o bin/lester ./lester
	go build -o bin/michael ./michael
	go build -o bin/franklin ./franklin
	go build -o bin/trevor ./trevor

run-lester:
	go run ./lester

run-michael:
	go run ./michael

run-franklin:
	go run ./franklin

run-trevor:
	go run ./trevor

clean:
	rm -rf bin
	rm -f proto/*.pb.go