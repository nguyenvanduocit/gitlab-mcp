package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nguyenvanduocit/gitlab-mcp/tools"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	envFile := flag.String("env", "", "Path to environment file (optional when environment variables are set directly)")
	httpPort := flag.String("http_port", "", "Port for HTTP server. If not provided, will use stdio")
	flag.Parse()

	// Load environment file if specified
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			fmt.Printf("⚠️  Warning: Error loading env file %s: %v\n", *envFile, err)
		} else {
			fmt.Printf("✅ Loaded environment variables from %s\n", *envFile)
		}
	}

	// Check required environment variables
	requiredEnvs := []string{"GITLAB_TOKEN", "GITLAB_URL"}
	missingEnvs := []string{}
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			missingEnvs = append(missingEnvs, env)
		}
	}

	if len(missingEnvs) > 0 {
		fmt.Println("❌ Configuration Error: Missing required environment variables")
		fmt.Println()
		fmt.Println("Missing variables:")
		for _, env := range missingEnvs {
			fmt.Printf("  - %s\n", env)
		}
		fmt.Println()
		fmt.Println("📋 Setup Instructions:")
		fmt.Println("1. Get your GitLab access token from: https://gitlab.com/-/profile/personal_access_tokens")
		fmt.Println("2. Set the environment variables:")
		fmt.Println()
		fmt.Println("   Option A - Using .env file:")
		fmt.Println("   Create a .env file with:")
		fmt.Println("   GITLAB_URL=https://gitlab.com")
		fmt.Println("   GITLAB_TOKEN=your-access-token")
		fmt.Println()
		fmt.Println("   Option B - Using environment variables:")
		fmt.Println("   export GITLAB_URL=https://gitlab.com")
		fmt.Println("   export GITLAB_TOKEN=your-access-token")
		fmt.Println()
		fmt.Println("   Option C - Using Docker:")
		fmt.Printf("   docker run -e GITLAB_URL=https://gitlab.com \\\n")
		fmt.Printf("              -e GITLAB_TOKEN=your-access-token \\\n")
		fmt.Printf("              ghcr.io/nguyenvanduocit/gitlab-mcp:latest\n")
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println("✅ All required environment variables are set")
	fmt.Printf("🔗 Connected to: %s\n", os.Getenv("GITLAB_URL"))

	mcpServer := server.NewMCPServer(
		"GitLab Tool",
		"1.0.0",
		server.WithLogging(),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)

	tools.RegisterProjectTools(mcpServer)
	tools.RegisterMergeRequestTools(mcpServer)
	tools.RegisterRepositoryTools(mcpServer)
	tools.RegisterPipelineTools(mcpServer)
	tools.RegisterUserTools(mcpServer)
	tools.RegisterGroupTools(mcpServer)
	tools.RegisterFlowTools(mcpServer)

	if *httpPort != "" {
		fmt.Println()
		fmt.Println("🚀 Starting GitLab MCP Server in HTTP mode...")
		fmt.Printf("📡 Server will be available at: http://localhost:%s/mcp\n", *httpPort)
		fmt.Println()
		fmt.Println("📋 Cursor Configuration:")
		fmt.Println("Add the following to your Cursor MCP settings (.cursor/mcp.json):")
		fmt.Println()
		fmt.Println("```json")
		fmt.Println("{")
		fmt.Println("  \"mcpServers\": {")
		fmt.Println("    \"gitlab\": {")
		fmt.Printf("      \"url\": \"http://localhost:%s/mcp\"\n", *httpPort)
		fmt.Println("    }")
		fmt.Println("  }")
		fmt.Println("}")
		fmt.Println("```")
		fmt.Println()
		fmt.Println("💡 Tips:")
		fmt.Println("- Restart Cursor after adding the configuration")
		fmt.Println("- Test the connection by asking Claude: 'List my GitLab projects'")
		fmt.Println("- Use '@gitlab' in Cursor to reference GitLab-related context")
		fmt.Println()
		fmt.Println("🔄 Server starting...")
		
		httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath("/mcp"))
		if err := httpServer.Start(fmt.Sprintf(":%s", *httpPort)); err != nil && !isContextCanceled(err) {
			log.Fatalf("❌ Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil && !isContextCanceled(err) {
			log.Fatalf("❌ Server error: %v", err)
		}
	}
}

// IsContextCanceled checks if the error is related to context cancellation
func isContextCanceled(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if it's directly context.Canceled
	if errors.Is(err, context.Canceled) {
		return true
	}
	
	// Check if the error message contains context canceled
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "context canceled") || 
	       strings.Contains(errMsg, "operation was canceled") ||
	       strings.Contains(errMsg, "context deadline exceeded")
}
