package main

import (
	"log"
	"os"

	"github.com/haedalwang/kubescout/internal/api"
	"github.com/haedalwang/kubescout/internal/k8s"
	"github.com/haedalwang/kubescout/internal/upstream"
)

func main() {
	log.Println("🔭 KubeScout - Starting Server...")

	// 1. Initialize Clients
	helmClient := k8s.NewHelmClient()
	ahClient := upstream.NewArtifactHubClient()

	// 2. Initialize Server
	server := api.NewServer(helmClient, ahClient)

	// 3. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	addr := ":" + port
	log.Printf("Server listening on %s", addr)
	if err := server.Start(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
