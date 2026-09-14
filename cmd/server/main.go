package main

import (
	"fmt"
	"log"
	"makarov-chains/api"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler, err := api.NewHandler()
	if err != nil {
		log.Fatal("Failed to initialize:", err)
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	fmt.Printf("Server starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed:", err)
	}
}
