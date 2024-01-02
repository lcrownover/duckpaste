.PHONY: build install clean run container 

all: build

build:
	@go build -o bin/duckpaste cmd/duckpaste/main.go

run: build
	@go run cmd/duckpaste/main.go

install: 
	@cp bin/duckpaste /usr/local/bin/duckpaste

container:
	@docker build -t duckpaste .

clean:
	@rm -f bin/duckpaste /usr/local/bin/duckpaste

