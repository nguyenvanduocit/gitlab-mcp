package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Consolidated pipeline management arguments with action-based routing
type PipelineManagementArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	Action      string `json:"action" validate:"required,oneof=list get trigger"`
	
	// List action options
	ListOptions struct {
		Status string `json:"status,omitempty" validate:"omitempty,oneof=running pending success failed canceled skipped all"`
	} `json:"list_options,omitempty"`
	
	// Get action options
	GetOptions struct {
		PipelineID float64 `json:"pipeline_id" validate:"required,min=1"`
	} `json:"get_options,omitempty"`
	
	// Trigger action options
	TriggerOptions struct {
		Ref       string            `json:"ref" validate:"required,min=1"`
		Variables map[string]string `json:"variables,omitempty" validate:"omitempty,dive,keys,min=1,endkeys,min=1"`
		Metadata  struct {
			Description string `json:"description,omitempty" validate:"omitempty,max=500"`
			Source      string `json:"source,omitempty" validate:"omitempty,max=100"`
		} `json:"metadata,omitempty"`
	} `json:"trigger_options,omitempty"`
}

func RegisterPipelineTools(s *server.MCPServer) {
	// Consolidated pipeline management tool
	pipelineManagementTool := mcp.NewTool("manage_pipelines",
		mcp.WithDescription("Comprehensive pipeline management for GitLab projects. Supports list, get details, and trigger operations."),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: 'list' (list pipelines), 'get' (get pipeline details), 'trigger' (create new pipeline)")),
		
		// List options
		mcp.WithObject("list_options", 
			mcp.Description("Options for list action"),
			mcp.Properties(map[string]any{
				"status": map[string]any{
					"type":        "string",
					"description": "Pipeline status filter (running/pending/success/failed/canceled/skipped/all)",
					"default":     "all",
				},
			}),
		),
		
		// Get options
		mcp.WithObject("get_options",
			mcp.Description("Options for get action"),
			mcp.Properties(map[string]any{
				"pipeline_id": map[string]any{
					"type":        "number",
					"description": "Pipeline ID to retrieve details for",
				},
			}),
		),
		
		// Trigger options
		mcp.WithObject("trigger_options",
			mcp.Description("Options for trigger action"),
			mcp.Properties(map[string]any{
				"ref": map[string]any{
					"type":        "string",
					"description": "Branch, tag, or commit SHA to trigger pipeline on",
				},
				"variables": map[string]any{
					"type":        "object",
					"description": "Optional variables to pass to the pipeline (key-value pairs)",
				},
				"metadata": map[string]any{
					"type": "object",
					"description": "Additional pipeline metadata",
					"properties": map[string]any{
						"description": map[string]any{
							"type":        "string",
							"description": "Pipeline description",
						},
						"source": map[string]any{
							"type":        "string", 
							"description": "Pipeline source identifier",
						},
					},
				},
			}),
		),
	)
	
	s.AddTool(pipelineManagementTool, mcp.NewTypedToolHandler(pipelineManagementHandler))
}

// Consolidated pipeline management handler
func pipelineManagementHandler(ctx context.Context, request mcp.CallToolRequest, args PipelineManagementArgs) (*mcp.CallToolResult, error) {
	switch strings.ToLower(args.Action) {
	case "list":
		return handleListPipelines(args)
	case "get":
		if args.GetOptions.PipelineID == 0 {
			return mcp.NewToolResultError("pipeline_id is required in get_options for get action"), nil
		}
		return handleGetPipeline(args)
	case "trigger":
		if args.TriggerOptions.Ref == "" {
			return mcp.NewToolResultError("ref is required in trigger_options for trigger action"), nil
		}
		return handleTriggerPipeline(args)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported action: %s. Supported actions: list, get, trigger", args.Action)), nil
	}
}

// Handle list pipelines action
func handleListPipelines(args PipelineManagementArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListProjectPipelinesOptions{}
	
	status := "all"
	if args.ListOptions.Status != "" {
		status = args.ListOptions.Status
	}
	
	if status != "all" {
		opt.Status = gitlab.Ptr(gitlab.BuildStateValue(status))
	}

	pipelines, _, err := util.GitlabClient().Pipelines.ListProjectPipelines(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list pipelines: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pipelines for project %s (status: %s):\n\n", args.ProjectPath, status))

	if len(pipelines) == 0 {
		result.WriteString("No pipelines found matching the criteria.\n")
	} else {
		for _, pipeline := range pipelines {
			result.WriteString(fmt.Sprintf("Pipeline #%d\n", pipeline.ID))
			result.WriteString(fmt.Sprintf("Status: %s\n", pipeline.Status))
			result.WriteString(fmt.Sprintf("Ref: %s\n", pipeline.Ref))
			result.WriteString(fmt.Sprintf("SHA: %s\n", pipeline.SHA))
			result.WriteString(fmt.Sprintf("Created: %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05")))
			result.WriteString(fmt.Sprintf("URL: %s\n\n", pipeline.WebURL))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

// Handle get pipeline details action
func handleGetPipeline(args PipelineManagementArgs) (*mcp.CallToolResult, error) {
	pipelineID := int(args.GetOptions.PipelineID)

	pipeline, _, err := util.GitlabClient().Pipelines.GetPipeline(args.ProjectPath, pipelineID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get pipeline: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pipeline #%d Details:\n\n", pipeline.ID))
	result.WriteString(fmt.Sprintf("Status: %s\n", pipeline.Status))
	result.WriteString(fmt.Sprintf("Ref: %s\n", pipeline.Ref))
	result.WriteString(fmt.Sprintf("SHA: %s\n", pipeline.SHA))
	result.WriteString(fmt.Sprintf("Created: %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Updated: %s\n", pipeline.UpdatedAt.Format("2006-01-02 15:04:05")))
	
	if pipeline.StartedAt != nil {
		result.WriteString(fmt.Sprintf("Started: %s\n", pipeline.StartedAt.Format("2006-01-02 15:04:05")))
	}
	
	if pipeline.FinishedAt != nil {
		result.WriteString(fmt.Sprintf("Finished: %s\n", pipeline.FinishedAt.Format("2006-01-02 15:04:05")))
	}
	
	result.WriteString(fmt.Sprintf("Duration: %d seconds\n", pipeline.Duration))
	result.WriteString(fmt.Sprintf("Coverage: %s\n", pipeline.Coverage))
	result.WriteString(fmt.Sprintf("URL: %s\n", pipeline.WebURL))

	return mcp.NewToolResultText(result.String()), nil
}

// Handle trigger pipeline action
func handleTriggerPipeline(args PipelineManagementArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.CreatePipelineOptions{
		Ref: gitlab.Ptr(args.TriggerOptions.Ref),
	}

	// Add variables if provided
	if len(args.TriggerOptions.Variables) > 0 {
		var variables []*gitlab.PipelineVariableOptions
		for key, value := range args.TriggerOptions.Variables {
			variables = append(variables, &gitlab.PipelineVariableOptions{
				Key:   gitlab.Ptr(key),
				Value: gitlab.Ptr(value),
			})
		}
		opt.Variables = &variables
	}

	pipeline, _, err := util.GitlabClient().Pipelines.CreatePipeline(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to trigger pipeline: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pipeline triggered successfully!\n\n"))
	result.WriteString(fmt.Sprintf("Pipeline #%d\n", pipeline.ID))
	result.WriteString(fmt.Sprintf("Status: %s\n", pipeline.Status))
	result.WriteString(fmt.Sprintf("Ref: %s\n", pipeline.Ref))
	result.WriteString(fmt.Sprintf("SHA: %s\n", pipeline.SHA))
	result.WriteString(fmt.Sprintf("Created: %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("URL: %s\n", pipeline.WebURL))

	if len(args.TriggerOptions.Variables) > 0 {
		result.WriteString(fmt.Sprintf("\nVariables passed:\n"))
		for key, value := range args.TriggerOptions.Variables {
			result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}
	
	if args.TriggerOptions.Metadata.Description != "" {
		result.WriteString(fmt.Sprintf("\nDescription: %s\n", args.TriggerOptions.Metadata.Description))
	}
	
	if args.TriggerOptions.Metadata.Source != "" {
		result.WriteString(fmt.Sprintf("Source: %s\n", args.TriggerOptions.Metadata.Source))
	}

	return mcp.NewToolResultText(result.String()), nil
} 