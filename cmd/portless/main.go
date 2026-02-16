package main

import (
	"fmt"
	"os"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "proxy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: portless proxy [start|stop|status]")
			os.Exit(1)
		}
		handleProxy(os.Args[2])
	case "ngrok":
		if len(os.Args) < 3 {
			fmt.Println("Usage: portless ngrok [start|stop|status]")
			os.Exit(1)
		}
		handleNgrok(os.Args[2:])
	case "dashboard":
		startDashboard()
	case "run":
		if len(os.Args) < 4 {
			fmt.Println("Usage: portless run <name> <command> [args...]")
			os.Exit(1)
		}
		handleRun(os.Args[2], os.Args[3:])
	case "version":
		fmt.Printf("portless version %s\n", version)
	case "help":
		printUsage()
	default:
		// If not a known command, treat as: portless <name> <command> [args...]
		if len(os.Args) >= 3 {
			handleRun(os.Args[1], os.Args[2:])
		} else {
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println(`portless - Replace port numbers with stable, named .localhost URLs

Usage:
  portless proxy [start|stop|status]     Manage the portless proxy server
  portless ngrok [start|stop|status]     Expose proxy via ngrok tunnel
  portless dashboard                     Open web dashboard
  portless run <name> <command> [args]   Run a command with a named .localhost URL
  portless <name> <command> [args]       Short form of 'run' command
  portless version                       Show version information
  portless help                          Show this help message

Examples:
  # Start the proxy
  portless proxy start

  # Run your apps
  portless myapp npm start               # Access at http://myapp.localhost:1355
  portless api go run main.go            # Access at http://api.localhost:1355

  # Expose to internet (optional)
  portless ngrok start                   # All services accessible via ngrok URL
  portless ngrok start --region eu       # Start with EU region
  portless ngrok status                  # Check public URL

  # View dashboard (optional)
  portless dashboard                     # Opens web UI at http://localhost:1355

Note: The proxy server must be running before using named URLs.`)
}

func handleProxy(action string) {
	switch action {
	case "start":
		startProxy()
	case "stop":
		stopProxy()
	case "status":
		proxyStatus()
	default:
		fmt.Printf("Unknown proxy action: %s\n", action)
		fmt.Println("Usage: portless proxy [start|stop|status]")
		os.Exit(1)
	}
}

func handleRun(name string, command []string) {
	runWithProxy(name, command)
}

func handleNgrok(args []string) {
	action := args[0]
	
	switch action {
	case "start":
		region := ""
		subdomain := ""
		
		// Parse flags
		for i := 1; i < len(args); i++ {
			if args[i] == "--region" && i+1 < len(args) {
				region = args[i+1]
				i++
			} else if args[i] == "--subdomain" && i+1 < len(args) {
				subdomain = args[i+1]
				i++
			}
		}
		
		startNgrok(region, subdomain)
	case "stop":
		stopNgrok()
	case "status":
		ngrokStatus()
	default:
		fmt.Printf("Unknown ngrok action: %s\n", action)
		fmt.Println("Usage: portless ngrok [start|stop|status]")
		os.Exit(1)
	}
}
