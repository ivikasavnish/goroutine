package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

func runWithProxy(name string, command []string) {
	// Validate service name
	if strings.Contains(name, ".") {
		log.Fatal("Service name cannot contain dots")
	}

	// Check if proxy is running
	if !isProxyRunning() {
		fmt.Println("Error: Proxy is not running")
		fmt.Println("Please start the proxy first with: portless proxy start")
		fmt.Println("Then run your command in a new terminal")
		os.Exit(1)
	}

	// Find a free port
	port, err := findFreePort()
	if err != nil {
		log.Fatalf("Failed to find free port: %v", err)
	}

	// Register the service
	if err := registerService(name, port); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}
	defer func() {
		if err := unregisterService(name); err != nil {
			log.Printf("Warning: Failed to unregister service: %v", err)
		}
	}()

	// Set up environment variables
	env := append(os.Environ(), 
		fmt.Sprintf("PORT=%d", port),
		fmt.Sprintf("SERVICE_NAME=%s", name),
		fmt.Sprintf("PORTLESS_PORT=%d", proxyPort),
	)

	// Build the command
	var cmd *exec.Cmd
	if len(command) == 1 {
		cmd = exec.Command(command[0])
	} else {
		cmd = exec.Command(command[0], command[1:]...)
	}

	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	fmt.Printf("\n✓ Service '%s' started on port %d\n", name, port)
	fmt.Printf("✓ Access it at: http://%s.localhost:%d\n\n", name, proxyPort)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for command to exit or signal
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case err := <-cmdDone:
		if err != nil {
			fmt.Printf("\nCommand exited with error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\nCommand exited successfully")
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal %v, stopping service...\n", sig)
		
		// Send signal to the command
		if err := cmd.Process.Signal(sig); err != nil {
			log.Printf("Failed to send signal to process: %v", err)
			cmd.Process.Kill()
		}
		
		// Wait for it to exit
		cmd.Wait()
		fmt.Println("Service stopped")
	}
}
