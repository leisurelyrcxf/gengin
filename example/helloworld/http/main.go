package main

import (
	"log"

	"github.com/leisurelyrcxf/gengin/example/helloworld"
)

func main() {
	srv := helloworld.NewServer(8080)
	if err := srv.Serve(); err != nil {
		log.Fatalf("Serve failed: '%v'", err)
	}
}
