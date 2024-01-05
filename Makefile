.PHONY: build install clean run container

all: build

build:
	@go build -o bin/duckpaste cmd/duckpaste/main.go

run: build
	@go run cmd/duckpaste/main.go

debug: build
	@go run cmd/duckpaste/main.go -debug

install:
	@cp bin/duckpaste /usr/local/bin/duckpaste

container:
	@docker build -t duckpaste .

run_container: container
	@docker run -it --env-file env -p 8080:8080 duckpaste

clean:
	@rm -f bin/duckpaste /usr/local/bin/duckpaste
