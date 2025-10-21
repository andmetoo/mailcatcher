package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andmetoo/mailcatcher"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	smtpPort := flag.Int("smtp-port", 1025, "SMTP server port")
	httpPort := flag.Int("http-port", 8025, "HTTP API server port")
	showVersion := flag.Bool("version", false, "Show version information")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	if *showVersion {
		fmt.Printf("mailcatcher %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	// Check environment variables (they override defaults but not flags)
	if !isFlagPassed("smtp-port") {
		if port := os.Getenv("MAILCATCHER_SMTP_PORT"); port != "" {
			if _, err := fmt.Sscanf(port, "%d", smtpPort); err != nil {
				log.Printf("Warning: invalid MAILCATCHER_SMTP_PORT value '%s', using default", port)
			}
		}
	}
	if !isFlagPassed("http-port") {
		if port := os.Getenv("MAILCATCHER_HTTP_PORT"); port != "" {
			if _, err := fmt.Sscanf(port, "%d", httpPort); err != nil {
				log.Printf("Warning: invalid MAILCATCHER_HTTP_PORT value '%s', using default", port)
			}
		}
	}

	logger := log.New(os.Stdout, "[mailcatcher] ", log.LstdFlags)

	logger.Printf("Starting mailcatcher %s", version)
	logger.Printf("SMTP server will listen on port %d", *smtpPort)
	logger.Printf("HTTP API will listen on port %d", *httpPort)

	// Create server
	server := mailcatcher.New(*smtpPort, *httpPort)

	// Set logger if verbose
	if *verbose {
		server.SetLogger(logger)
	}

	// Start server
	if err := server.Start(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}

	logger.Printf("SMTP server started on :%d", *smtpPort)
	logger.Printf("HTTP API started on :%d", *httpPort)
	logger.Printf("Web interface: http://localhost:%d/api/v1/emails", *httpPort)
	logger.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		logger.Printf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	logger.Println("Server stopped")
}

// isFlagPassed checks if a flag was explicitly passed
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
