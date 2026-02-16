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
  portless run <name> <command> [args]   Run a command with a named .localhost URL
  portless <name> <command> [args]       Short form of 'run' command
  portless version                       Show version information
  portless help                          Show this help message

Examples:
  # First, start the proxy (run this in one terminal)
  portless proxy start

  # Then run your apps (in separate terminals)
  portless myapp npm start               # Access at http://myapp.localhost:1355
  portless run api go run main.go        # Access at http://api.localhost:1355

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
