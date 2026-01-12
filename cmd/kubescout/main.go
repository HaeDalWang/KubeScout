package main

import (
	"log"

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
	// TODO: Make port configurable
	log.Println("Server listening on :8080")
	if err := server.Start(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
