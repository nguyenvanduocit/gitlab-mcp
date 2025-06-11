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

type ListProjectJobsArgs struct {
	ProjectPath    string   `json:"project_path"`
	Scope          []string `json:"scope,omitempty"`
	IncludeRetried bool     `json:"include_retried,omitempty"`
}

type ListPipelineJobsArgs struct {
	ProjectPath    string   `json:"project_path"`
	PipelineID     float64  `json:"pipeline_id"`
	Scope          []string `json:"scope,omitempty"`
	IncludeRetried bool     `json:"include_retried,omitempty"`
}

type GetJobArgs struct {
	ProjectPath string  `json:"project_path"`
	JobID       float64 `json:"job_id"`
}

type CancelJobArgs struct {
	ProjectPath string  `json:"project_path"`
	JobID       float64 `json:"job_id"`
}

type RetryJobArgs struct {
	ProjectPath string  `json:"project_path"`
	JobID       float64 `json:"job_id"`
}

func RegisterJobTools(s *server.MCPServer) {
	listProjectJobsTool := mcp.NewTool("list_project_jobs",
		mcp.WithDescription("List jobs for a GitLab project"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithArray("scope", mcp.Description("Job scope filter (created, pending, running, failed, success, canceled, skipped)")),
		mcp.WithBoolean("include_retried", mcp.DefaultBool(false), mcp.Description("Include retried jobs")),
	)
	s.AddTool(listProjectJobsTool, mcp.NewTypedToolHandler(listProjectJobsHandler))

	listPipelineJobsTool := mcp.NewTool("list_pipeline_jobs",
		mcp.WithDescription("List jobs for a specific pipeline"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithNumber("pipeline_id", mcp.Required(), mcp.Description("Pipeline ID")),
		mcp.WithArray("scope", mcp.Description("Job scope filter (created, pending, running, failed, success, canceled, skipped)")),
		mcp.WithBoolean("include_retried", mcp.DefaultBool(false), mcp.Description("Include retried jobs")),
	)
	s.AddTool(listPipelineJobsTool, mcp.NewTypedToolHandler(listPipelineJobsHandler))

	getJobTool := mcp.NewTool("get_job",
		mcp.WithDescription("Get details for a specific job by ID"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithNumber("job_id", mcp.Required(), mcp.Description("Job ID")),
	)
	s.AddTool(getJobTool, mcp.NewTypedToolHandler(getJobHandler))

	cancelJobTool := mcp.NewTool("cancel_job",
		mcp.WithDescription("Cancel a specific job"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithNumber("job_id", mcp.Required(), mcp.Description("Job ID to cancel")),
	)
	s.AddTool(cancelJobTool, mcp.NewTypedToolHandler(cancelJobHandler))

	retryJobTool := mcp.NewTool("retry_job",
		mcp.WithDescription("Retry a specific job"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithNumber("job_id", mcp.Required(), mcp.Description("Job ID to retry")),
	)
	s.AddTool(retryJobTool, mcp.NewTypedToolHandler(retryJobHandler))
}

func listProjectJobsHandler(ctx context.Context, request mcp.CallToolRequest, args ListProjectJobsArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListJobsOptions{}

	// Convert scope strings to BuildStateValue
	if len(args.Scope) > 0 {
		var scopes []gitlab.BuildStateValue
		for _, s := range args.Scope {
			scopes = append(scopes, gitlab.BuildStateValue(s))
		}
		opt.Scope = &scopes
	}

	if args.IncludeRetried {
		opt.IncludeRetried = gitlab.Ptr(args.IncludeRetried)
	}

	jobs, _, err := util.GitlabClient().Jobs.ListProjectJobs(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list project jobs: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Jobs for project %s:\n\n", args.ProjectPath))

	for _, job := range jobs {
		result.WriteString(formatJobInfo(job))
		result.WriteString("\n")
	}

	if len(jobs) == 0 {
		result.WriteString("No jobs found for the specified criteria.\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func listPipelineJobsHandler(ctx context.Context, request mcp.CallToolRequest, args ListPipelineJobsArgs) (*mcp.CallToolResult, error) {
	pipelineID := int(args.PipelineID)
	opt := &gitlab.ListJobsOptions{}

	// Convert scope strings to BuildStateValue
	if len(args.Scope) > 0 {
		var scopes []gitlab.BuildStateValue
		for _, s := range args.Scope {
			scopes = append(scopes, gitlab.BuildStateValue(s))
		}
		opt.Scope = &scopes
	}

	if args.IncludeRetried {
		opt.IncludeRetried = gitlab.Ptr(args.IncludeRetried)
	}

	jobs, _, err := util.GitlabClient().Jobs.ListPipelineJobs(args.ProjectPath, pipelineID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list pipeline jobs: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Jobs for pipeline #%d in project %s:\n\n", pipelineID, args.ProjectPath))

	for _, job := range jobs {
		result.WriteString(formatJobInfo(job))
		result.WriteString("\n")
	}

	if len(jobs) == 0 {
		result.WriteString("No jobs found for the specified pipeline.\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getJobHandler(ctx context.Context, request mcp.CallToolRequest, args GetJobArgs) (*mcp.CallToolResult, error) {
	jobID := int(args.JobID)

	job, _, err := util.GitlabClient().Jobs.GetJob(args.ProjectPath, jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get job: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Job #%d Details:\n\n", job.ID))
	result.WriteString(formatJobDetailedInfo(job))

	return mcp.NewToolResultText(result.String()), nil
}

func cancelJobHandler(ctx context.Context, request mcp.CallToolRequest, args CancelJobArgs) (*mcp.CallToolResult, error) {
	jobID := int(args.JobID)

	job, _, err := util.GitlabClient().Jobs.CancelJob(args.ProjectPath, jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to cancel job: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Job #%d has been canceled successfully!\n\n", job.ID))
	result.WriteString(formatJobInfo(job))

	return mcp.NewToolResultText(result.String()), nil
}

func retryJobHandler(ctx context.Context, request mcp.CallToolRequest, args RetryJobArgs) (*mcp.CallToolResult, error) {
	jobID := int(args.JobID)

	job, _, err := util.GitlabClient().Jobs.RetryJob(args.ProjectPath, jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to retry job: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Job #%d has been retried successfully!\n\n", job.ID))
	result.WriteString(formatJobInfo(job))

	return mcp.NewToolResultText(result.String()), nil
}

// Helper function to format basic job information
func formatJobInfo(job *gitlab.Job) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Job #%d - %s\n", job.ID, job.Name))
	result.WriteString(fmt.Sprintf("Status: %s\n", job.Status))
	result.WriteString(fmt.Sprintf("Stage: %s\n", job.Stage))
	result.WriteString(fmt.Sprintf("Ref: %s\n", job.Ref))
	
	if job.CreatedAt != nil {
		result.WriteString(fmt.Sprintf("Created: %s\n", job.CreatedAt.Format("2006-01-02 15:04:05")))
	}
	
	if job.StartedAt != nil {
		result.WriteString(fmt.Sprintf("Started: %s\n", job.StartedAt.Format("2006-01-02 15:04:05")))
	}
	
	if job.FinishedAt != nil {
		result.WriteString(fmt.Sprintf("Finished: %s\n", job.FinishedAt.Format("2006-01-02 15:04:05")))
	}
	
	result.WriteString(fmt.Sprintf("Duration: %.2f seconds\n", job.Duration))
	result.WriteString(fmt.Sprintf("URL: %s\n", job.WebURL))
	
	return result.String()
}

// Helper function to format detailed job information
func formatJobDetailedInfo(job *gitlab.Job) string {
	var result strings.Builder
	
	// Basic info
	result.WriteString(fmt.Sprintf("Name: %s\n", job.Name))
	result.WriteString(fmt.Sprintf("Status: %s\n", job.Status))
	result.WriteString(fmt.Sprintf("Stage: %s\n", job.Stage))
	result.WriteString(fmt.Sprintf("Ref: %s\n", job.Ref))
	result.WriteString(fmt.Sprintf("Allow Failure: %t\n", job.AllowFailure))
	result.WriteString(fmt.Sprintf("Tag: %t\n", job.Tag))
	
	// Timing information
	if job.CreatedAt != nil {
		result.WriteString(fmt.Sprintf("Created: %s\n", job.CreatedAt.Format("2006-01-02 15:04:05")))
	}
	
	if job.StartedAt != nil {
		result.WriteString(fmt.Sprintf("Started: %s\n", job.StartedAt.Format("2006-01-02 15:04:05")))
	}
	
	if job.FinishedAt != nil {
		result.WriteString(fmt.Sprintf("Finished: %s\n", job.FinishedAt.Format("2006-01-02 15:04:05")))
	}
	
	if job.ErasedAt != nil {
		result.WriteString(fmt.Sprintf("Erased: %s\n", job.ErasedAt.Format("2006-01-02 15:04:05")))
	}
	
	result.WriteString(fmt.Sprintf("Duration: %.2f seconds\n", job.Duration))
	result.WriteString(fmt.Sprintf("Queued Duration: %.2f seconds\n", job.QueuedDuration))
	
	// Coverage and failure reason
	if job.Coverage > 0 {
		result.WriteString(fmt.Sprintf("Coverage: %.2f%%\n", job.Coverage))
	}
	
	if job.FailureReason != "" {
		result.WriteString(fmt.Sprintf("Failure Reason: %s\n", job.FailureReason))
	}
	
	// Pipeline information
	result.WriteString(fmt.Sprintf("\nPipeline Information:\n"))
	result.WriteString(fmt.Sprintf("Pipeline ID: %d\n", job.Pipeline.ID))
	result.WriteString(fmt.Sprintf("Pipeline Status: %s\n", job.Pipeline.Status))
	result.WriteString(fmt.Sprintf("Pipeline Ref: %s\n", job.Pipeline.Ref))
	result.WriteString(fmt.Sprintf("Pipeline SHA: %s\n", job.Pipeline.Sha))
	
	// Runner information
	if job.Runner.ID > 0 {
		result.WriteString(fmt.Sprintf("\nRunner Information:\n"))
		result.WriteString(fmt.Sprintf("Runner ID: %d\n", job.Runner.ID))
		result.WriteString(fmt.Sprintf("Runner Name: %s\n", job.Runner.Name))
		result.WriteString(fmt.Sprintf("Runner Description: %s\n", job.Runner.Description))
		result.WriteString(fmt.Sprintf("Runner Active: %t\n", job.Runner.Active))
		result.WriteString(fmt.Sprintf("Runner Shared: %t\n", job.Runner.IsShared))
	}
	
	// Artifacts information
	if len(job.Artifacts) > 0 {
		result.WriteString(fmt.Sprintf("\nArtifacts:\n"))
		for _, artifact := range job.Artifacts {
			result.WriteString(fmt.Sprintf("- %s (%s, %d bytes)\n", artifact.Filename, artifact.FileType, artifact.Size))
		}
	}
	
	if job.ArtifactsFile.Filename != "" {
		result.WriteString(fmt.Sprintf("Artifacts File: %s (%d bytes)\n", job.ArtifactsFile.Filename, job.ArtifactsFile.Size))
	}
	
	if job.ArtifactsExpireAt != nil {
		result.WriteString(fmt.Sprintf("Artifacts Expire: %s\n", job.ArtifactsExpireAt.Format("2006-01-02 15:04:05")))
	}
	
	// Tags
	if len(job.TagList) > 0 {
		result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(job.TagList, ", ")))
	}
	
	// User information
	if job.User != nil {
		result.WriteString(fmt.Sprintf("\nTriggered by: %s (%s)\n", job.User.Name, job.User.Username))
	}
	
	// Commit information
	if job.Commit != nil {
		result.WriteString(fmt.Sprintf("\nCommit Information:\n"))
		result.WriteString(fmt.Sprintf("Commit SHA: %s\n", job.Commit.ID))
		result.WriteString(fmt.Sprintf("Commit Title: %s\n", job.Commit.Title))
		result.WriteString(fmt.Sprintf("Commit Message: %s\n", job.Commit.Message))
		if job.Commit.AuthorName != "" {
			result.WriteString(fmt.Sprintf("Author: %s <%s>\n", job.Commit.AuthorName, job.Commit.AuthorEmail))
		}
	}
	
	result.WriteString(fmt.Sprintf("\nWeb URL: %s\n", job.WebURL))
	
	return result.String()
}
