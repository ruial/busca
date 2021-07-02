package main

import (
	"log"

	"github.com/ruial/busca/internal/api"
)

func main() {
	log.Println("Starting busca")
	log.Println(api.Server(":8080"))
}
