package main

import (
	"github.com/lcrownover/duckpaste/internal/example"
	"github.com/lcrownover/duckpaste/internal/web"
)

func main() {
	example.HelloWorld()
	web.StartServer()
}
