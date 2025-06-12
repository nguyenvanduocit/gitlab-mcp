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

// Consolidated args structures
type JobListArgs struct {
	ProjectPath    string   `json:"project_path" validate:"required,min=1"`
	PipelineID     *float64 `json:"pipeline_id,omitempty" validate:"omitempty,min=1"` // Optional - if provided, list pipeline jobs; if not, list project jobs
	Scope          []string `json:"scope,omitempty" validate:"omitempty,dive,oneof=created pending running failed success canceled skipped"`
	IncludeRetried bool     `json:"include_retried,omitempty"`
}

type JobManageArgs struct {
	ProjectPath string  `json:"project_path" validate:"required,min=1"`
	JobID       float64 `json:"job_id" validate:"required,min=1"`
	Action      string  `json:"action" validate:"required,oneof=get cancel retry"` // "get", "cancel", "retry"
}

func RegisterJobTools(s *server.MCPServer) {
	// Consolidated job listing tool
	jobListTool := mcp.NewTool("manage_jobs_list",
		mcp.WithDescription("List jobs for a GitLab project or specific pipeline"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithNumber("pipeline_id", mcp.Description("Pipeline ID (optional - if provided, lists pipeline jobs; if not, lists project jobs)")),
		mcp.WithArray("scope", mcp.Description("Job scope filter (created, pending, running, failed, success, canceled, skipped)")),
		mcp.WithBoolean("include_retried", mcp.DefaultBool(false), mcp.Description("Include retried jobs")),
	)
	s.AddTool(jobListTool, mcp.NewTypedToolHandler(jobListHandler))

	// Consolidated job management tool
	jobManageTool := mcp.NewTool("manage_job_actions",
		mcp.WithDescription("Perform actions on a specific job (get details, cancel, or retry)"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithNumber("job_id", mcp.Required(), mcp.Description("Job ID")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: 'get' (get details), 'cancel' (cancel job), 'retry' (retry job)")),
	)
	s.AddTool(jobManageTool, mcp.NewTypedToolHandler(jobManageHandler))
}

// Consolidated job listing handler
func jobListHandler(ctx context.Context, request mcp.CallToolRequest, args JobListArgs) (*mcp.CallToolResult, error) {
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

	var jobs []*gitlab.Job
	var err error
	var result strings.Builder

	// Check if pipeline_id is provided to determine which API to call
	if args.PipelineID != nil {
		pipelineID := int(*args.PipelineID)
		jobs, _, err = util.GitlabClient().Jobs.ListPipelineJobs(args.ProjectPath, pipelineID, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list pipeline jobs: %v", err)), nil
		}
		result.WriteString(fmt.Sprintf("Jobs for pipeline #%d in project %s:\n\n", pipelineID, args.ProjectPath))
	} else {
		jobs, _, err = util.GitlabClient().Jobs.ListProjectJobs(args.ProjectPath, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list project jobs: %v", err)), nil
		}
		result.WriteString(fmt.Sprintf("Jobs for project %s:\n\n", args.ProjectPath))
	}

	for _, job := range jobs {
		result.WriteString(formatJobInfo(job))
		result.WriteString("\n")
	}

	if len(jobs) == 0 {
		if args.PipelineID != nil {
			result.WriteString("No jobs found for the specified pipeline.\n")
		} else {
			result.WriteString("No jobs found for the specified criteria.\n")
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

// Consolidated job management handler
func jobManageHandler(ctx context.Context, request mcp.CallToolRequest, args JobManageArgs) (*mcp.CallToolResult, error) {
	jobID := int(args.JobID)

	switch strings.ToLower(args.Action) {
	case "get":
		return getJobDetails(args.ProjectPath, jobID)
	case "cancel":
		return cancelJobAction(args.ProjectPath, jobID)
	case "retry":
		return retryJobAction(args.ProjectPath, jobID)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action '%s'. Valid actions are: get, cancel, retry", args.Action)), nil
	}
}

// Helper functions for job management actions
func getJobDetails(projectPath string, jobID int) (*mcp.CallToolResult, error) {
	job, _, err := util.GitlabClient().Jobs.GetJob(projectPath, jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get job: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Job #%d Details:\n\n", job.ID))
	result.WriteString(formatJobDetailedInfo(job))

	return mcp.NewToolResultText(result.String()), nil
}

func cancelJobAction(projectPath string, jobID int) (*mcp.CallToolResult, error) {
	job, _, err := util.GitlabClient().Jobs.CancelJob(projectPath, jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to cancel job: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Job #%d has been canceled successfully!\n\n", job.ID))
	result.WriteString(formatJobInfo(job))

	return mcp.NewToolResultText(result.String()), nil
}

func retryJobAction(projectPath string, jobID int) (*mcp.CallToolResult, error) {
	job, _, err := util.GitlabClient().Jobs.RetryJob(projectPath, jobID)
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
