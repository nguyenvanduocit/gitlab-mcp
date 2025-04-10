package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func RegisterRepositoryTools(s *server.MCPServer) {
	fileContentTool := mcp.NewTool("gitlab_get_file_content",
		mcp.WithDescription("Get file content from a GitLab repository"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file in the repository")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch name, tag, or commit SHA")),
	)

	commitsTool := mcp.NewTool("gitlab_list_commits",
		mcp.WithDescription("List commits in a GitLab project within a date range"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("since", mcp.Required(), mcp.Description("Start date (YYYY-MM-DD)")),
		mcp.WithString("until", mcp.Description("End date (YYYY-MM-DD). If not provided, defaults to current date")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch name, tag, or commit SHA")),
	)

	commitDetailsTool := mcp.NewTool("gitlab_get_commit_details",
		mcp.WithDescription("Get details of a commit"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA")),
	)

	s.AddTool(fileContentTool, util.ErrorGuard(getFileContentHandler))
	s.AddTool(commitsTool, util.ErrorGuard(listCommitsHandler))
	s.AddTool(commitDetailsTool, util.ErrorGuard(getCommitDetailsHandler))
}

func getFileContentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.Params.Arguments["project_path"].(string)
	filePath := request.Params.Arguments["file_path"].(string)

	ref := "develop" // Default ref if not provided
	if value, ok := request.Params.Arguments["ref"]; ok {
		ref = value.(string)
	}

	// Get raw file content
	fileContent, _, err := util.GitlabClient().RepositoryFiles.GetRawFile(projectID, filePath, &gitlab.GetRawFileOptions{
		Ref: gitlab.Ptr(ref),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %v; maybe wrong ref?", err)
	}

	var result strings.Builder

	// Write file information
	result.WriteString(fmt.Sprintf("File: %s\n", filePath))
	result.WriteString(fmt.Sprintf("Ref: %s\n", ref))
	result.WriteString("Content:\n")
	result.WriteString(string(fileContent))

	return mcp.NewToolResultText(result.String()), nil
}

func listCommitsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.Params.Arguments["project_path"].(string)
	since, ok := request.Params.Arguments["since"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: since")
	}

	until := time.Now().Format("2006-01-02")
	if value, ok := request.Params.Arguments["until"]; ok {
		until = value.(string)
	}

	ref := "develop" // Default ref if not provided
	if value, ok := request.Params.Arguments["ref"]; ok {
		ref = value.(string)
	}

	sinceTime, err := time.Parse("2006-01-02", since)
	if err != nil {
		return nil, fmt.Errorf("invalid since date: %v", err)
	}

	untilTime, err := time.Parse("2006-01-02 15:04:05", until+" 23:00:00")
	if err != nil {
		return nil, fmt.Errorf("invalid until date: %v", err)
	}

	opt := &gitlab.ListCommitsOptions{
		Since:   gitlab.Ptr(sinceTime),
		Until:   gitlab.Ptr(untilTime),
		RefName: gitlab.Ptr(ref),
	}

	commits, _, err := util.GitlabClient().Commits.ListCommits(projectID, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commits for project %s between %s and %s (ref: %s):\n\n",
		projectID, since, until, ref))

	for _, commit := range commits {
		result.WriteString(fmt.Sprintf("Commit: %s\n", commit.ID))
		result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
		result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
		if commit.LastPipeline != nil {
			result.WriteString("Last Pipeline: \n")
			result.WriteString(fmt.Sprintf("  Status: %s\n", commit.LastPipeline.Status))
			result.WriteString(fmt.Sprintf("  Ref: %s\n", commit.LastPipeline.Ref))
			result.WriteString(fmt.Sprintf("  SHA: %s\n", commit.LastPipeline.SHA))
			result.WriteString(fmt.Sprintf("  Created: %s\n", commit.LastPipeline.CreatedAt.Format("2006-01-02 15:04:05")))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getCommitDetailsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.Params.Arguments["project_path"].(string)
	commitSHA := request.Params.Arguments["commit_sha"].(string)

	commit, _, err := util.GitlabClient().Commits.GetCommit(projectID, commitSHA, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit details: %v", err)
	}

	opt := &gitlab.GetCommitDiffOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	diffs, _, err := util.GitlabClient().Commits.GetCommitDiff(projectID, commitSHA, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit diffs: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commit: %s\n", commit.ShortID))
	result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
	result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
	result.WriteString(fmt.Sprintf("URL: %s\n\n", commit.WebURL))

	if commit.ParentIDs != nil {
		result.WriteString("Parents:\n")
		for _, parentID := range commit.ParentIDs {
			result.WriteString(fmt.Sprintf("- %s\n", parentID))
		}
		result.WriteString("\n")
	}

	result.WriteString("Diffs:\n")
	for _, diff := range diffs {
		result.WriteString(fmt.Sprintf("File: %s\n", diff.NewPath))
		result.WriteString(fmt.Sprintf("Status: %s\n", getDiffStatus(diff)))

		if diff.Diff != "" {
			result.WriteString("```diff\n")
			result.WriteString(diff.Diff)
			result.WriteString("\n```\n")
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getDiffStatus(diff *gitlab.Diff) string {
	if diff.NewFile {
		return "Added"
	}
	if diff.DeletedFile {
		return "Deleted"
	}
	if diff.RenamedFile {
		return fmt.Sprintf("Renamed from %s", diff.OldPath)
	}
	return "Modified"
} 