package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Portless automatically sets the PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		fmt.Println("Warning: PORT not set, using default 3000")
		fmt.Println("Tip: Run with 'portless myapp go run main.go' for named URL access")
	} else {
		fmt.Printf("Using PORT=%s from portless\n", port)
	}

	// Simple HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Portless Demo!\n")
		fmt.Fprintf(w, "Server: %s\n", r.Host)
		fmt.Fprintf(w, "Time: %s\n", time.Now().Format(time.RFC3339))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	addr := ":" + port
	fmt.Printf("\nDemo server starting on port %s\n", port)
	
	if os.Getenv("PORT") != "" {
		serviceName := os.Getenv("SERVICE_NAME")
		if serviceName == "" {
			serviceName = "myapp"
		}
		proxyPort := os.Getenv("PORTLESS_PORT")
		if proxyPort == "" {
			proxyPort = "1355"
		}
		fmt.Printf("Access via: http://%s.localhost:%s\n", serviceName, proxyPort)
	} else {
		fmt.Printf("Access via: http://localhost:%s\n", port)
	}
	
	fmt.Println("\nEndpoints:")
	fmt.Println("  /        - Main page")
	fmt.Println("  /health  - Health check")
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
