package main

import (
	"log"

	"github.com/ruial/busca/internal/api"
	"github.com/ruial/busca/internal/repository"
)

func main() {
	log.Println("Starting busca")
	log.Println(api.Server(":8080", repository.NewInMemoryIndexRepo()))
}
