package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Portless Dashboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            theme: {
                extend: {
                    colors: {
                        primary: '#3b82f6',
                    }
                }
            }
        }
    </script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <!-- Header -->
        <div class="bg-white rounded-lg shadow-md p-6 mb-6">
            <div class="flex items-center justify-between">
                <div>
                    <h1 class="text-3xl font-bold text-gray-800">Portless Dashboard</h1>
                    <p class="text-gray-600 mt-1">Named localhost URLs for your development servers</p>
                </div>
                <div class="text-right">
                    <div class="text-sm text-gray-500">Proxy Port</div>
                    <div class="text-2xl font-bold text-primary">{{.ProxyPort}}</div>
                </div>
            </div>
        </div>

        <!-- Status Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <!-- Proxy Status -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center justify-between mb-2">
                    <h3 class="text-lg font-semibold text-gray-700">Proxy Server</h3>
                    {{if .ProxyRunning}}
                    <span class="px-3 py-1 bg-green-100 text-green-800 text-sm font-medium rounded-full">Running</span>
                    {{else}}
                    <span class="px-3 py-1 bg-red-100 text-red-800 text-sm font-medium rounded-full">Stopped</span>
                    {{end}}
                </div>
                <p class="text-gray-600 text-sm">Local proxy server handling service routing</p>
            </div>

            <!-- ngrok Status -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center justify-between mb-2">
                    <h3 class="text-lg font-semibold text-gray-700">ngrok Tunnel</h3>
                    {{if .NgrokRunning}}
                    <span class="px-3 py-1 bg-green-100 text-green-800 text-sm font-medium rounded-full">Active</span>
                    {{else}}
                    <span class="px-3 py-1 bg-gray-100 text-gray-800 text-sm font-medium rounded-full">Inactive</span>
                    {{end}}
                </div>
                <p class="text-gray-600 text-sm">Public internet access to your services</p>
            </div>

            <!-- Services Count -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center justify-between mb-2">
                    <h3 class="text-lg font-semibold text-gray-700">Active Services</h3>
                    <span class="text-3xl font-bold text-primary">{{.ServiceCount}}</span>
                </div>
                <p class="text-gray-600 text-sm">Currently registered services</p>
            </div>
        </div>

        <!-- ngrok Public URL -->
        {{if .NgrokPublicURL}}
        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
            <div class="flex items-start">
                <svg class="w-5 h-5 text-blue-500 mt-0.5 mr-3" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path>
                </svg>
                <div class="flex-1">
                    <h4 class="text-sm font-medium text-blue-800 mb-1">Public ngrok URL</h4>
                    <a href="{{.NgrokPublicURL}}" target="_blank" class="text-blue-600 hover:text-blue-800 break-all">{{.NgrokPublicURL}}</a>
                </div>
            </div>
        </div>
        {{end}}

        <!-- Services Table -->
        <div class="bg-white rounded-lg shadow-md overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-200">
                <h2 class="text-xl font-semibold text-gray-800">Registered Services</h2>
            </div>
            
            {{if .Services}}
            <div class="overflow-x-auto">
                <table class="w-full">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Service Name</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Local Port</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Local URL</th>
                            {{if $.NgrokPublicURL}}
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Public URL</th>
                            {{end}}
                        </tr>
                    </thead>
                    <tbody class="bg-white divide-y divide-gray-200">
                        {{range .Services}}
                        <tr class="hover:bg-gray-50">
                            <td class="px-6 py-4 whitespace-nowrap">
                                <div class="flex items-center">
                                    <div class="w-2 h-2 bg-green-500 rounded-full mr-3"></div>
                                    <span class="text-sm font-medium text-gray-900">{{.Name}}</span>
                                </div>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                {{.Port}}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                <a href="{{.LocalURL}}" target="_blank" class="text-sm text-blue-600 hover:text-blue-800">{{.LocalURL}}</a>
                            </td>
                            {{if $.NgrokPublicURL}}
                            <td class="px-6 py-4 whitespace-nowrap">
                                <a href="{{.PublicURL}}" target="_blank" class="text-sm text-blue-600 hover:text-blue-800 break-all">{{.PublicURL}}</a>
                            </td>
                            {{end}}
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            {{else}}
            <div class="px-6 py-12 text-center">
                <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"></path>
                </svg>
                <h3 class="mt-2 text-sm font-medium text-gray-900">No services registered</h3>
                <p class="mt-1 text-sm text-gray-500">Start a service using: portless myapp &lt;command&gt;</p>
            </div>
            {{end}}
        </div>

        <!-- Quick Commands -->
        <div class="mt-6 bg-white rounded-lg shadow-md p-6">
            <h2 class="text-lg font-semibold text-gray-800 mb-4">Quick Commands</h2>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="bg-gray-50 rounded p-4">
                    <h3 class="text-sm font-medium text-gray-700 mb-2">Start a service</h3>
                    <code class="text-xs bg-gray-800 text-green-400 px-3 py-1 rounded block overflow-x-auto">portless myapp npm start</code>
                </div>
                <div class="bg-gray-50 rounded p-4">
                    <h3 class="text-sm font-medium text-gray-700 mb-2">Check proxy status</h3>
                    <code class="text-xs bg-gray-800 text-green-400 px-3 py-1 rounded block overflow-x-auto">portless proxy status</code>
                </div>
                {{if not .NgrokRunning}}
                <div class="bg-gray-50 rounded p-4">
                    <h3 class="text-sm font-medium text-gray-700 mb-2">Expose to internet</h3>
                    <code class="text-xs bg-gray-800 text-green-400 px-3 py-1 rounded block overflow-x-auto">portless ngrok start</code>
                </div>
                {{else}}
                <div class="bg-gray-50 rounded p-4">
                    <h3 class="text-sm font-medium text-gray-700 mb-2">Stop ngrok tunnel</h3>
                    <code class="text-xs bg-gray-800 text-green-400 px-3 py-1 rounded block overflow-x-auto">portless ngrok stop</code>
                </div>
                {{end}}
                <div class="bg-gray-50 rounded p-4">
                    <h3 class="text-sm font-medium text-gray-700 mb-2">View help</h3>
                    <code class="text-xs bg-gray-800 text-green-400 px-3 py-1 rounded block overflow-x-auto">portless help</code>
                </div>
            </div>
        </div>

        <!-- Footer -->
        <div class="mt-8 text-center text-sm text-gray-500">
            <p>Portless v1.0.0 - Named localhost URLs for development</p>
            <p class="mt-1">Auto-refreshes every 5 seconds</p>
        </div>
    </div>

    <script>
        // Auto-refresh data every 5 seconds without full page reload
        async function refreshDashboard() {
            try {
                const response = await fetch('/__dashboard/api');
                if (response.ok) {
                    window.location.reload();
                }
            } catch (error) {
                console.error('Failed to refresh dashboard:', error);
            }
        }

        // Refresh every 5 seconds
        setInterval(refreshDashboard, 5000);
    </script>
</body>
</html>`

type ServiceInfo struct {
	Name      string
	Port      int
	LocalURL  string
	PublicURL string
}

type DashboardData struct {
	ProxyPort      int
	ProxyRunning   bool
	NgrokRunning   bool
	NgrokPublicURL string
	ServiceCount   int
	Services       []ServiceInfo
}

func startDashboard() {
	if !isProxyRunning() {
		fmt.Println("Error: Proxy is not running")
		fmt.Println("Please start the proxy first with: portless proxy start")
		return
	}

	dashboardURL := fmt.Sprintf("http://localhost:%d/__dashboard", proxyPort)
	fmt.Printf("Opening dashboard at %s\n", dashboardURL)
	
	// Try to open browser
	if err := openBrowser(dashboardURL); err != nil {
		fmt.Printf("Dashboard URL: %s\n", dashboardURL)
		fmt.Println("Please open this URL in your browser")
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Load registry
	if err := loadRegistry(); err != nil {
		log.Printf("Failed to load registry: %v", err)
	}

	// Load ngrok config
	ngrokConfig, _ := loadNgrokConfig()
	
	// Build dashboard data
	data := DashboardData{
		ProxyPort:    proxyPort,
		ProxyRunning: isProxyRunning(),
		NgrokRunning: isNgrokRunning(),
	}

	if ngrokConfig != nil {
		data.NgrokPublicURL = ngrokConfig.PublicURL
	}

	// Build services list
	registry.RLock()
	data.ServiceCount = len(registry.Services)
	for name, port := range registry.Services {
		service := ServiceInfo{
			Name:     name,
			Port:     port,
			LocalURL: fmt.Sprintf("http://%s.localhost:%d", name, proxyPort),
		}
		
		if data.NgrokPublicURL != "" {
			service.PublicURL = buildPublicURL(data.NgrokPublicURL, name)
		}
		
		data.Services = append(data.Services, service)
	}
	registry.RUnlock()

	// Parse and execute template
	tmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

func handleDashboardAPI(w http.ResponseWriter, r *http.Request) {
	// Load registry
	if err := loadRegistry(); err != nil {
		log.Printf("Failed to load registry: %v", err)
	}

	// Load ngrok config
	ngrokConfig, _ := loadNgrokConfig()

	data := DashboardData{
		ProxyPort:    proxyPort,
		ProxyRunning: isProxyRunning(),
		NgrokRunning: isNgrokRunning(),
	}

	if ngrokConfig != nil {
		data.NgrokPublicURL = ngrokConfig.PublicURL
	}

	registry.RLock()
	data.ServiceCount = len(registry.Services)
	for name, port := range registry.Services {
		service := ServiceInfo{
			Name:     name,
			Port:     port,
			LocalURL: fmt.Sprintf("http://%s.localhost:%d", name, proxyPort),
		}
		
		if data.NgrokPublicURL != "" {
			service.PublicURL = buildPublicURL(data.NgrokPublicURL, name)
		}
		
		data.Services = append(data.Services, service)
	}
	registry.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// buildPublicURL creates a public URL for a service given the ngrok base URL
func buildPublicURL(ngrokURL, serviceName string) string {
	if ngrokURL == "" {
		return ""
	}
	// Extract the domain from ngrok URL (remove protocol)
	publicDomain := strings.TrimPrefix(ngrokURL, "https://")
	publicDomain = strings.TrimPrefix(publicDomain, "http://")
	return fmt.Sprintf("https://%s.%s", serviceName, publicDomain)
}
