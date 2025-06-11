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

type ListPipelinesArgs struct {
	ProjectPath string `json:"project_path"`
	Status      string `json:"status"`
}

type GetPipelineArgs struct {
	ProjectPath string  `json:"project_path"`
	PipelineID  float64 `json:"pipeline_id"`
}

type TriggerPipelineArgs struct {
	ProjectPath string            `json:"project_path"`
	Ref         string            `json:"ref"`
	Variables   map[string]string `json:"variables,omitempty"`
}

func RegisterPipelineTools(s *server.MCPServer) {
	pipelineTool := mcp.NewTool("list_pipelines",
		mcp.WithDescription("List pipelines for a GitLab project"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("status", mcp.DefaultString("all"), mcp.Description("Pipeline status (running/pending/success/failed/canceled/skipped/all)")),
	)
	s.AddTool(pipelineTool, mcp.NewTypedToolHandler(listPipelinesHandler))
	
	getPipelineTool := mcp.NewTool("get_pipeline",
		mcp.WithDescription("Get details for a specific pipeline by ID"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithNumber("pipeline_id", mcp.Required(), mcp.Description("Pipeline ID")),
	)
	s.AddTool(getPipelineTool, mcp.NewTypedToolHandler(getPipelineHandler))

	triggerPipelineTool := mcp.NewTool("trigger_pipeline",
		mcp.WithDescription("Trigger a new pipeline on a specific branch with optional variables"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch, tag, or commit SHA to trigger pipeline on")),
		mcp.WithObject("variables", mcp.Description("Optional variables to pass to the pipeline (key-value pairs)")),
	)
	s.AddTool(triggerPipelineTool, mcp.NewTypedToolHandler(triggerPipelineHandler))
}

func listPipelinesHandler(ctx context.Context, request mcp.CallToolRequest, args ListPipelinesArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListProjectPipelinesOptions{}
	if args.Status != "all" {
		// Assuming gitlab.BuildStateValue is the correct type for status
		opt.Status = gitlab.Ptr(gitlab.BuildStateValue(args.Status))
	}

	pipelines, _, err := util.GitlabClient().Pipelines.ListProjectPipelines(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list pipelines: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pipelines for project %s:\n\n", args.ProjectPath))

	for _, pipeline := range pipelines {
		result.WriteString(fmt.Sprintf("Pipeline #%d\n", pipeline.ID))
		result.WriteString(fmt.Sprintf("Status: %s\n", pipeline.Status))
		result.WriteString(fmt.Sprintf("Ref: %s\n", pipeline.Ref))
		result.WriteString(fmt.Sprintf("SHA: %s\n", pipeline.SHA))
		result.WriteString(fmt.Sprintf("Created: %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("URL: %s\n\n", pipeline.WebURL))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getPipelineHandler(ctx context.Context, request mcp.CallToolRequest, args GetPipelineArgs) (*mcp.CallToolResult, error) {
	pipelineID := int(args.PipelineID)

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

func triggerPipelineHandler(ctx context.Context, request mcp.CallToolRequest, args TriggerPipelineArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.CreatePipelineOptions{
		Ref: gitlab.Ptr(args.Ref),
	}

	// Add variables if provided
	if len(args.Variables) > 0 {
		var variables []*gitlab.PipelineVariableOptions
		for key, value := range args.Variables {
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

	if len(args.Variables) > 0 {
		result.WriteString(fmt.Sprintf("\nVariables passed:\n"))
		for key, value := range args.Variables {
			result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
} 