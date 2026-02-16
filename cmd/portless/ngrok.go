package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	ngrokConfigFile = "ngrok.json"
	ngrokPidFile    = "ngrok.pid"
)

type NgrokConfig struct {
	URL       string    `json:"url"`
	PublicURL string    `json:"public_url"`
	StartedAt time.Time `json:"started_at"`
	Region    string    `json:"region,omitempty"`
}

func getNgrokConfigPath() string {
	return filepath.Join(getConfigDir(), ngrokConfigFile)
}

func getNgrokPidPath() string {
	return filepath.Join(getConfigDir(), ngrokPidFile)
}

func loadNgrokConfig() (*NgrokConfig, error) {
	path := getNgrokConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var config NgrokConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveNgrokConfig(config *NgrokConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getNgrokConfigPath(), data, 0644)
}

func startNgrok(region string, subdomain string) {
	// Check if ngrok is installed
	if _, err := exec.LookPath("ngrok"); err != nil {
		fmt.Println("Error: ngrok is not installed")
		fmt.Println("Please install ngrok from https://ngrok.com/download")
		os.Exit(1)
	}

	// Check if proxy is running
	if !isProxyRunning() {
		fmt.Println("Error: Proxy is not running")
		fmt.Println("Please start the proxy first with: portless proxy start")
		os.Exit(1)
	}

	// Check if ngrok is already running
	if isNgrokRunning() {
		fmt.Println("ngrok tunnel is already running")
		ngrokStatus()
		os.Exit(0)
	}

	// Build ngrok command
	args := []string{"http", fmt.Sprintf("%d", proxyPort)}
	
	if region != "" {
		args = append(args, "--region", region)
	}
	
	if subdomain != "" {
		args = append(args, "--subdomain", subdomain)
	}

	// Start ngrok
	cmd := exec.Command("ngrok", args...)
	
	// Create log file
	logPath := filepath.Join(getConfigDir(), "ngrok.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start ngrok: %v", err)
	}

	// Save PID
	pid := cmd.Process.Pid
	if err := os.WriteFile(getNgrokPidPath(), []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		log.Fatalf("Failed to write ngrok PID file: %v", err)
	}

	fmt.Printf("Starting ngrok tunnel on port %d...\n", proxyPort)
	
	// Wait for ngrok to be ready and get the public URL
	time.Sleep(2 * time.Second)
	
	publicURL, err := getNgrokPublicURL()
	if err != nil {
		fmt.Printf("Warning: Could not get ngrok public URL: %v\n", err)
		fmt.Println("ngrok is starting, use 'portless ngrok status' to check the URL")
		return
	}

	// Save ngrok config
	config := &NgrokConfig{
		URL:       fmt.Sprintf("http://localhost:%d", proxyPort),
		PublicURL: publicURL,
		StartedAt: time.Now(),
		Region:    region,
	}

	if err := saveNgrokConfig(config); err != nil {
		log.Printf("Warning: Failed to save ngrok config: %v", err)
	}

	fmt.Println("\n✓ ngrok tunnel started successfully!")
	fmt.Printf("✓ Public URL: %s\n", publicURL)
	fmt.Println("\nYour services are now accessible from the internet:")
	
	// Show registered services
	if err := loadRegistry(); err == nil {
		registry.RLock()
		for name := range registry.Services {
			// Replace localhost with ngrok domain
			publicServiceURL := strings.Replace(publicURL, "https://", fmt.Sprintf("https://%s.", name), 1)
			if !strings.HasPrefix(publicURL, "https://") {
				publicServiceURL = strings.Replace(publicURL, "http://", fmt.Sprintf("http://%s.", name), 1)
			}
			fmt.Printf("  - %s: %s\n", name, publicServiceURL)
		}
		registry.RUnlock()
	}
	
	fmt.Println("\nNote: Keep this terminal open or run in background")
}

func stopNgrok() {
	pidPath := getNgrokPidPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("ngrok is not running")
			return
		}
		log.Fatalf("Failed to read ngrok PID file: %v", err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		log.Fatalf("Invalid PID in file: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("ngrok is not running")
		os.Remove(pidPath)
		os.Remove(getNgrokConfigPath())
		return
	}

	// Send SIGTERM
	if err := process.Kill(); err != nil {
		fmt.Printf("Failed to stop ngrok: %v\n", err)
		return
	}

	// Clean up
	os.Remove(pidPath)
	os.Remove(getNgrokConfigPath())
	
	fmt.Println("ngrok tunnel stopped")
}

func isNgrokRunning() bool {
	pidPath := getNgrokPidPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process is alive (Unix-like systems)
	err = process.Signal(os.Signal(nil))
	return err == nil
}

func ngrokStatus() {
	if !isNgrokRunning() {
		fmt.Println("ngrok is not running")
		fmt.Println("Start it with: portless ngrok start")
		return
	}

	config, err := loadNgrokConfig()
	if err != nil {
		fmt.Printf("ngrok is running but config could not be loaded: %v\n", err)
		return
	}

	if config == nil {
		fmt.Println("ngrok is running but config is not available")
		fmt.Println("Trying to fetch current status...")
		
		publicURL, err := getNgrokPublicURL()
		if err != nil {
			fmt.Printf("Could not get ngrok public URL: %v\n", err)
			return
		}
		
		config = &NgrokConfig{
			PublicURL: publicURL,
		}
	}

	fmt.Println("ngrok tunnel is active")
	fmt.Printf("Public URL: %s\n", config.PublicURL)
	
	if !config.StartedAt.IsZero() {
		fmt.Printf("Started: %s (uptime: %s)\n", 
			config.StartedAt.Format("2006-01-02 15:04:05"),
			time.Since(config.StartedAt).Round(time.Second))
	}
	
	if config.Region != "" {
		fmt.Printf("Region: %s\n", config.Region)
	}

	// Show registered services
	if err := loadRegistry(); err == nil {
		registry.RLock()
		if len(registry.Services) > 0 {
			fmt.Println("\nPublic service URLs:")
			for name := range registry.Services {
				// Extract domain from public URL
				publicServiceURL := strings.Replace(config.PublicURL, "://", fmt.Sprintf("://%s.", name), 1)
				fmt.Printf("  - %s: %s\n", name, publicServiceURL)
			}
		}
		registry.RUnlock()
	}
}

// getNgrokPublicURL fetches the public URL from ngrok's local API
func getNgrokPublicURL() (string, error) {
	// ngrok exposes a local API at http://localhost:4040/api/tunnels
	resp, err := http.Get("http://localhost:4040/api/tunnels")
	if err != nil {
		return "", fmt.Errorf("failed to connect to ngrok API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read ngrok API response: %v", err)
	}

	var result struct {
		Tunnels []struct {
			PublicURL string `json:"public_url"`
			Proto     string `json:"proto"`
		} `json:"tunnels"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse ngrok API response: %v", err)
	}

	// Find the HTTP/HTTPS tunnel
	for _, tunnel := range result.Tunnels {
		if tunnel.Proto == "https" || tunnel.Proto == "http" {
			return tunnel.PublicURL, nil
		}
	}

	return "", fmt.Errorf("no HTTP/HTTPS tunnel found")
}
