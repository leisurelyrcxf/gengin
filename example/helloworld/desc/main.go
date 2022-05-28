package main

import (
	"fmt"
	"github.com/leisurelyrcxf/gengin/example/helloworld"
	"log"
)

func main() {
	srv := helloworld.NewServer(8080)
	if err := srv.RegisterServices(); err != nil {
		log.Fatalf("Register failed: %v", err)
	}
	desc := srv.ServiceDescription.GetDescription()
	fmt.Println(desc)
}
