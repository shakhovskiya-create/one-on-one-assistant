package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ekf/one-on-one-connector/internal/connector"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if exists
	godotenv.Load()

	// Parse flags
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	// Create connector
	conn, err := connector.New(*configPath)
	if err != nil {
		log.Fatalf("Failed to create connector: %v", err)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start connector in background
	go func() {
		if err := conn.Run(); err != nil {
			log.Printf("Connector error: %v", err)
		}
	}()

	log.Println("Connector started. Press Ctrl+C to stop.")

	// Wait for signal
	<-sigChan
	log.Println("Shutting down...")

	conn.Stop()
}
