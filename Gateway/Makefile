.PHONY: build run test clean docker-build docker-run bench

BINARY_NAME=gateway
MAIN_PATH=cmd/gateway/main.go

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

bench:
	go test -bench=. ./...
