package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/nguyenvanduocit/gitlab-mcp/tools"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
)	

func main() {
	envFile := flag.String("env", ".env", "Path to environment file")
	ssePort := flag.String("sse_port", "", "Port for SSE server. If not provided, will use stdio")
	flag.Parse()

	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			fmt.Printf("Warning: Error loading env file %s: %v\n", *envFile, err)
		}
	}

	mcpServer := server.NewMCPServer(
		"GitLab Tool",
		"1.0.0",
		server.WithLogging(),
	)

	tools.RegisterProjectTools(mcpServer)
	tools.RegisterMergeRequestTools(mcpServer)
	tools.RegisterRepositoryTools(mcpServer)
	tools.RegisterPipelineTools(mcpServer)
	tools.RegisterUserTools(mcpServer)
	tools.RegisterGroupTools(mcpServer)

	if *ssePort != "" {
		sseServer := server.NewSSEServer(mcpServer)
		if err := sseServer.Start(fmt.Sprintf(":%s", *ssePort)); err != nil && !util.IsContextCanceled(err) {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil && !util.IsContextCanceled(err) {
			panic(fmt.Sprintf("Server error: %v", err))
		}
	}
}
