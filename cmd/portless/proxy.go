package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultProxyPort = 1355
	configDir        = ".portless"
	registryFile     = "registry.json"
	pidFile          = "proxy.pid"
)

var proxyPort int

func init() {
	// Allow proxy port to be configured via environment variable
	if portStr := os.Getenv("PORTLESS_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			proxyPort = port
		} else {
			log.Printf("Warning: Invalid PORTLESS_PORT value '%s', using default %d", portStr, defaultProxyPort)
			proxyPort = defaultProxyPort
		}
	} else {
		proxyPort = defaultProxyPort
	}
}

type Registry struct {
	sync.RWMutex
	Services map[string]int `json:"services"` // service name -> port
}

var registry = &Registry{
	Services: make(map[string]int),
}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(home, configDir)
}

func ensureConfigDir() error {
	dir := getConfigDir()
	return os.MkdirAll(dir, 0755)
}

func getRegistryPath() string {
	return filepath.Join(getConfigDir(), registryFile)
}

func getPidPath() string {
	return filepath.Join(getConfigDir(), pidFile)
}

func loadRegistry() error {
	path := getRegistryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No registry yet, that's ok
		}
		return err
	}

	registry.Lock()
	defer registry.Unlock()
	return json.Unmarshal(data, &registry.Services)
}

func saveRegistry() error {
	registry.RLock()
	data, err := json.MarshalIndent(registry.Services, "", "  ")
	registry.RUnlock()
	if err != nil {
		return err
	}

	return os.WriteFile(getRegistryPath(), data, 0644)
}

func startProxy() {
	if err := ensureConfigDir(); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	// Check if already running
	if isProxyRunning() {
		fmt.Println("Proxy is already running")
		os.Exit(0)
	}

	// Load existing registry
	if err := loadRegistry(); err != nil {
		log.Printf("Warning: Failed to load registry: %v", err)
	}

	// Write PID file
	pid := os.Getpid()
	if err := os.WriteFile(getPidPath(), []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create reverse proxy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleProxyRequest(w, r)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", proxyPort),
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("Portless proxy server started on :%d\n", proxyPort)
		fmt.Printf("Access your apps at http://<name>.localhost:%d\n", proxyPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down proxy server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Clean up PID file
	os.Remove(getPidPath())
	fmt.Println("Proxy server stopped")
}

func handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Extract service name from host
	host := r.Host
	// Remove port if present
	hostWithoutPort := strings.Split(host, ":")[0]
	parts := strings.Split(hostWithoutPort, ".")
	
	if len(parts) < 2 || parts[len(parts)-1] != "localhost" {
		http.Error(w, fmt.Sprintf("Invalid host format. Use <name>.localhost:%d", proxyPort), http.StatusBadRequest)
		return
	}

	serviceName := parts[0]

	// Reload registry to get latest service registrations
	if err := loadRegistry(); err != nil {
		log.Printf("Warning: Failed to reload registry: %v", err)
	}

	// Look up service port
	registry.RLock()
	port, exists := registry.Services[serviceName]
	registry.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf("Service '%s' not found. Make sure it's running with: portless %s <command>", serviceName, serviceName), http.StatusNotFound)
		return
	}

	// Create reverse proxy to the actual service
	target, err := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		http.Error(w, "Failed to parse target URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Modify the request to forward to the actual service
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	// Handle proxy errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for service '%s': %v", serviceName, err)
		http.Error(w, fmt.Sprintf("Failed to connect to service '%s' on port %d. Is it running?", serviceName, port), http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}

func stopProxy() {
	pidPath := getPidPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Proxy is not running")
			return
		}
		log.Fatalf("Failed to read PID file: %v", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		log.Fatalf("Invalid PID in file: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("Proxy is not running")
		os.Remove(pidPath)
		return
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		fmt.Printf("Failed to stop proxy: %v\n", err)
		os.Remove(pidPath)
		return
	}

	// Wait for process to exit with timeout
	maxWait := 5 * time.Second
	start := time.Now()
	for time.Since(start) < maxWait {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			// Process no longer exists
			os.Remove(pidPath)
			fmt.Println("Proxy stopped")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	// If still running after timeout, force kill
	fmt.Println("Proxy did not stop gracefully, forcing...")
	process.Kill()
	os.Remove(pidPath)
	fmt.Println("Proxy stopped")
}

func isProxyRunning() bool {
	pidPath := getPidPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process is alive
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func proxyStatus() {
	if isProxyRunning() {
		fmt.Printf("Proxy is running on port %d\n", proxyPort)
		
		// Show registered services
		if err := loadRegistry(); err == nil {
			registry.RLock()
			defer registry.RUnlock()
			
			if len(registry.Services) > 0 {
				fmt.Println("\nRegistered services:")
				for name, port := range registry.Services {
					fmt.Printf("  - %s.localhost:%d -> localhost:%d\n", name, proxyPort, port)
				}
			} else {
				fmt.Println("\nNo services registered yet")
			}
		}
	} else {
		fmt.Println("Proxy is not running")
		fmt.Println("Start it with: portless proxy start")
	}
}

// Helper function to register a service
func registerService(name string, port int) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	if err := loadRegistry(); err != nil {
		return err
	}

	registry.Lock()
	registry.Services[name] = port
	registry.Unlock()

	return saveRegistry()
}

// Helper function to unregister a service
func unregisterService(name string) error {
	if err := loadRegistry(); err != nil {
		return err
	}

	registry.Lock()
	delete(registry.Services, name)
	registry.Unlock()

	return saveRegistry()
}

// Find a free port
func findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
