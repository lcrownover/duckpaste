.PHONY: build install clean run container handler

build:
	@go build -o bin/github.com/lcrownover/duckpaste cmd/github.com/lcrownover/duckpaste/main.go

run: build
	@go run cmd/github.com/lcrownover/duckpaste/main.go

install: build
	@cp bin/github.com/lcrownover/duckpaste /usr/local/bin/github.com/lcrownover/duckpaste

container:
	@docker build -t github.com/lcrownover/duckpaste .

handler:
	@go build -o handler cmd/github.com/lcrownover/duckpaste/main.go

clean:
	@rm -f bin/github.com/lcrownover/duckpaste /usr/local/bin/github.com/lcrownover/duckpaste

