package main

import (
	"log"
	"xm_test/cmd/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
