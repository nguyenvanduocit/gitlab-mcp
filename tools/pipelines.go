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