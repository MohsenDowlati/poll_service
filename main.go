package main

import (
	"log"

	_ "github.com/amitshekhariitbhu/go-backend-clean-architecture/docs"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
