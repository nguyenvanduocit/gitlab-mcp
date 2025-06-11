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

type GetFileContentArgs struct {
	ProjectPath string `json:"project_path"`
	FilePath    string `json:"file_path"`
	Ref         string `json:"ref"`
}

type ListCommitsArgs struct {
	ProjectPath string `json:"project_path"`
	Since       string `json:"since"`
	Until       string `json:"until"`
	Ref         string `json:"ref"`
}

type GetCommitDetailsArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
}

func RegisterRepositoryTools(s *server.MCPServer) {
	fileContentTool := mcp.NewTool("get_file_content",
		mcp.WithDescription("Get file content from a GitLab repository"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file in the repository")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch name, tag, or commit SHA")),
	)

	commitsTool := mcp.NewTool("list_commits",
		mcp.WithDescription("List commits in a GitLab project within a date range"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("since", mcp.Required(), mcp.Description("Start date (YYYY-MM-DD)")),
		mcp.WithString("until", mcp.Description("End date (YYYY-MM-DD). If not provided, defaults to current date")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch name, tag, or commit SHA")),
	)

	commitDetailsTool := mcp.NewTool("get_commit_details",
		mcp.WithDescription("Get details of a commit"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA")),
	)

	s.AddTool(fileContentTool, mcp.NewTypedToolHandler(getFileContentHandler))
	s.AddTool(commitsTool, mcp.NewTypedToolHandler(listCommitsHandler))
	s.AddTool(commitDetailsTool, mcp.NewTypedToolHandler(getCommitDetailsHandler))
}

func getFileContentHandler(ctx context.Context, request mcp.CallToolRequest, args GetFileContentArgs) (*mcp.CallToolResult, error) {
	ref := args.Ref
	if ref == "" {
		ref = "develop" // Default ref if not provided
	}

	// Get raw file content
	fileContent, _, err := util.GitlabClient().RepositoryFiles.GetRawFile(args.ProjectPath, args.FilePath, &gitlab.GetRawFileOptions{
		Ref: gitlab.Ptr(ref),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get file content: %v; maybe wrong ref?", err)), nil
	}

	var result strings.Builder

	// Write file information
	result.WriteString(fmt.Sprintf("File: %s\n", args.FilePath))
	result.WriteString(fmt.Sprintf("Ref: %s\n", ref))
	result.WriteString("Content:\n")
	result.WriteString(string(fileContent))

	return mcp.NewToolResultText(result.String()), nil
}

func listCommitsHandler(ctx context.Context, request mcp.CallToolRequest, args ListCommitsArgs) (*mcp.CallToolResult, error) {
	until := args.Until
	if until == "" {
		until = time.Now().Format("2006-01-02")
	}

	ref := args.Ref
	if ref == "" {
		ref = "develop" // Default ref if not provided
	}

	sinceTime, err := time.Parse("2006-01-02", args.Since)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid since date: %v", err)), nil
	}

	untilTime, err := time.Parse("2006-01-02 15:04:05", until+" 23:00:00")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid until date: %v", err)), nil
	}

	opt := &gitlab.ListCommitsOptions{
		Since:   gitlab.Ptr(sinceTime),
		Until:   gitlab.Ptr(untilTime),
		RefName: gitlab.Ptr(ref),
	}

	commits, _, err := util.GitlabClient().Commits.ListCommits(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list commits: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commits for project %s between %s and %s (ref: %s):\n\n",
		args.ProjectPath, args.Since, until, ref))

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

func getCommitDetailsHandler(ctx context.Context, request mcp.CallToolRequest, args GetCommitDetailsArgs) (*mcp.CallToolResult, error) {
	commit, _, err := util.GitlabClient().Commits.GetCommit(args.ProjectPath, args.CommitSHA, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit details: %v", err)), nil
	}

	opt := &gitlab.GetCommitDiffOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	diffs, _, err := util.GitlabClient().Commits.GetCommitDiff(args.ProjectPath, args.CommitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit diffs: %v", err)), nil
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