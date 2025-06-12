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

// Consolidated Repository Files Management
type RepositoryFilesArgs struct {
	Action      string `json:"action" validate:"required,oneof=get_content"`
	ProjectPath string `json:"project_path" validate:"required,min=1,max=255"`
	FilePath    string `json:"file_path" validate:"required,min=1,max=500"`
	Ref         string `json:"ref" validate:"required,min=1,max=255"`
}

// Consolidated Commits Management
type CommitsManagementArgs struct {
	Action      string `json:"action" validate:"required,oneof=list search get_details get_comments post_comment get_merge_requests get_refs"`
	ProjectPath string `json:"project_path" validate:"required,min=1,max=255"`
	
	// Common commit parameters
	CommitSHA string `json:"commit_sha,omitempty" validate:"omitempty,min=7,max=40,alphanum"`
	Ref       string `json:"ref,omitempty" validate:"omitempty,min=1,max=255"`
	
	// List/Search specific parameters
	ListOptions struct {
		Since string `json:"since,omitempty" validate:"omitempty,datetime=2006-01-02"`
		Until string `json:"until,omitempty" validate:"omitempty,datetime=2006-01-02"`
	} `json:"list_options"`
	
	SearchOptions struct {
		Author string `json:"author,omitempty" validate:"omitempty,min=1,max=100"`
		Path   string `json:"path,omitempty" validate:"omitempty,min=1,max=500"`
		Since  string `json:"since,omitempty" validate:"omitempty,datetime=2006-01-02"`
		Until  string `json:"until,omitempty" validate:"omitempty,datetime=2006-01-02"`
	} `json:"search_options"`
	
	// Comment specific parameters
	CommentOptions struct {
		Note     string `json:"note,omitempty" validate:"omitempty,min=1,max=1000"`
		Path     string `json:"path,omitempty" validate:"omitempty,min=1,max=500"`
		Line     int    `json:"line,omitempty" validate:"omitempty,min=1"`
		LineType string `json:"line_type,omitempty" validate:"omitempty,oneof=new old"`
	} `json:"comment_options"`
	
	// Refs specific parameters
	RefsOptions struct {
		Type string `json:"type,omitempty" validate:"omitempty,oneof=branch tag"`
	} `json:"refs_options"`
}

// Consolidated Commit Operations
type CommitOperationsArgs struct {
	Action      string `json:"action" validate:"required,oneof=cherry_pick revert"`
	ProjectPath string `json:"project_path" validate:"required,min=1,max=255"`
	CommitSHA   string `json:"commit_sha" validate:"required,min=7,max=40,alphanum"`
	Branch      string `json:"branch" validate:"required,min=1,max=255"`
	
	// Cherry-pick specific options
	CherryPickOptions struct {
		DryRun  bool   `json:"dry_run"`
		Message string `json:"message,omitempty" validate:"omitempty,min=1,max=500"`
	} `json:"cherry_pick_options"`
}

func RegisterRepositoryTools(s *server.MCPServer) {
	// Consolidated Repository Files Tool
	repositoryFilesTool := mcp.NewTool("manage_repository_files",
		mcp.WithDescription("Manage repository files with various actions: get_content"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: get_content")),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path (1-255 characters)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file in the repository (1-500 characters)")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch name, tag, or commit SHA (1-255 characters)")),
	)

	// Consolidated Commits Management Tool
	commitsManagementTool := mcp.NewTool("manage_commits",
		mcp.WithDescription("Comprehensive commits management with multiple actions: list, search, get_details, get_comments, post_comment, get_merge_requests, get_refs"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: list, search, get_details, get_comments, post_comment, get_merge_requests, get_refs")),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path (1-255 characters)")),
		mcp.WithString("commit_sha", mcp.Description("Commit SHA (7-40 alphanumeric characters, required for: get_details, get_comments, post_comment, get_merge_requests, get_refs)")),
		mcp.WithString("ref", mcp.Description("Branch name, tag, or commit SHA (1-255 characters, required for list action)")),
		
		// List options
		mcp.WithObject("list_options",
			mcp.Description("Options for list action"),
			mcp.Properties(map[string]any{
				"since": map[string]any{
					"type":        "string",
					"description": "Start date (YYYY-MM-DD, required for list action)",
					"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
				},
				"until": map[string]any{
					"type":        "string",
					"description": "End date (YYYY-MM-DD, optional - defaults to current date)",
					"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
				},
			}),
		),
		
		// Search options
		mcp.WithObject("search_options",
			mcp.Description("Options for search action"),
			mcp.Properties(map[string]any{
				"author": map[string]any{
					"type":        "string",
					"description": "Filter by author name or email (1-100 characters)",
					"minLength":   1,
					"maxLength":   100,
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Filter by file path (1-500 characters)",
					"minLength":   1,
					"maxLength":   500,
				},
				"since": map[string]any{
					"type":        "string",
					"description": "Start date (YYYY-MM-DD)",
					"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
				},
				"until": map[string]any{
					"type":        "string",
					"description": "End date (YYYY-MM-DD)",
					"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
				},
			}),
		),
		
		// Comment options
		mcp.WithObject("comment_options",
			mcp.Description("Options for post_comment action"),
			mcp.Properties(map[string]any{
				"note": map[string]any{
					"type":        "string",
					"description": "Comment text (1-1000 characters, required for post_comment)",
					"minLength":   1,
					"maxLength":   1000,
				},
				"path": map[string]any{
					"type":        "string",
					"description": "File path for line-specific comment (1-500 characters)",
					"minLength":   1,
					"maxLength":   500,
				},
				"line": map[string]any{
					"type":        "number",
					"description": "Line number for line-specific comment (minimum: 1)",
					"minimum":     1,
				},
				"line_type": map[string]any{
					"type":        "string",
					"description": "Line type for line-specific comment",
					"enum":        []string{"new", "old"},
				},
			}),
		),
		
		// Refs options
		mcp.WithObject("refs_options",
			mcp.Description("Options for get_refs action"),
			mcp.Properties(map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "Reference type filter",
					"enum":        []string{"branch", "tag"},
				},
			}),
		),
	)

	// Consolidated Commit Operations Tool
	commitOperationsTool := mcp.NewTool("commit_operations",
		mcp.WithDescription("Perform operations on commits: cherry_pick, revert"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Operation to perform: cherry_pick, revert")),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path (1-255 characters)")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA to operate on (7-40 alphanumeric characters)")),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Target branch (1-255 characters)")),
		
		// Cherry-pick options
		mcp.WithObject("cherry_pick_options",
			mcp.Description("Options for cherry_pick action"),
			mcp.Properties(map[string]any{
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "Perform a dry run without making changes",
					"default":     false,
				},
				"message": map[string]any{
					"type":        "string",
					"description": "Custom commit message (1-500 characters)",
					"minLength":   1,
					"maxLength":   500,
				},
			}),
		),
	)

	// Register consolidated tools
	s.AddTool(repositoryFilesTool, mcp.NewTypedToolHandler(repositoryFilesHandler))
	s.AddTool(commitsManagementTool, mcp.NewTypedToolHandler(commitsManagementHandler))
	s.AddTool(commitOperationsTool, mcp.NewTypedToolHandler(commitOperationsHandler))
}

// Consolidated handlers
func repositoryFilesHandler(ctx context.Context, request mcp.CallToolRequest, args RepositoryFilesArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "get_content":
		return getFileContent(ctx, args.ProjectPath, args.FilePath, args.Ref)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action: %s. Valid actions are: get_content", args.Action)), nil
	}
}

func commitsManagementHandler(ctx context.Context, request mcp.CallToolRequest, args CommitsManagementArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "list":
		if args.ListOptions.Since == "" {
			return mcp.NewToolResultError("since date is required for list action"), nil
		}
		if args.Ref == "" {
			return mcp.NewToolResultError("ref is required for list action"), nil
		}
		return listCommits(ctx, args.ProjectPath, args.ListOptions.Since, args.ListOptions.Until, args.Ref)
		
	case "search":
		return searchCommits(ctx, args.ProjectPath, args.SearchOptions.Author, args.SearchOptions.Path, 
			args.SearchOptions.Since, args.SearchOptions.Until, args.Ref)
		
	case "get_details":
		if args.CommitSHA == "" {
			return mcp.NewToolResultError("commit_sha is required for get_details action"), nil
		}
		return getCommitDetails(ctx, args.ProjectPath, args.CommitSHA)
		
	case "get_comments":
		if args.CommitSHA == "" {
			return mcp.NewToolResultError("commit_sha is required for get_comments action"), nil
		}
		return getCommitComments(ctx, args.ProjectPath, args.CommitSHA)
		
	case "post_comment":
		if args.CommitSHA == "" {
			return mcp.NewToolResultError("commit_sha is required for post_comment action"), nil
		}
		if args.CommentOptions.Note == "" {
			return mcp.NewToolResultError("note is required for post_comment action"), nil
		}
		return postCommitComment(ctx, args.ProjectPath, args.CommitSHA, args.CommentOptions.Note,
			args.CommentOptions.Path, args.CommentOptions.Line, args.CommentOptions.LineType)
		
	case "get_merge_requests":
		if args.CommitSHA == "" {
			return mcp.NewToolResultError("commit_sha is required for get_merge_requests action"), nil
		}
		return getCommitMergeRequests(ctx, args.ProjectPath, args.CommitSHA)
		
	case "get_refs":
		if args.CommitSHA == "" {
			return mcp.NewToolResultError("commit_sha is required for get_refs action"), nil
		}
		return getCommitRefs(ctx, args.ProjectPath, args.CommitSHA, args.RefsOptions.Type)
		
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action: %s. Valid actions are: list, search, get_details, get_comments, post_comment, get_merge_requests, get_refs", args.Action)), nil
	}
}

func commitOperationsHandler(ctx context.Context, request mcp.CallToolRequest, args CommitOperationsArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "cherry_pick":
		return cherryPickCommit(ctx, args.ProjectPath, args.CommitSHA, args.Branch,
			args.CherryPickOptions.DryRun, args.CherryPickOptions.Message)
		
	case "revert":
		return revertCommit(ctx, args.ProjectPath, args.CommitSHA, args.Branch)
		
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action: %s. Valid actions are: cherry_pick, revert", args.Action)), nil
	}
}

// Direct implementation functions (no more legacy handlers)
func getFileContent(ctx context.Context, projectPath, filePath, ref string) (*mcp.CallToolResult, error) {
	if ref == "" {
		ref = "develop" // Default ref if not provided
	}

	// Get raw file content
	fileContent, _, err := util.GitlabClient().RepositoryFiles.GetRawFile(projectPath, filePath, &gitlab.GetRawFileOptions{
		Ref: gitlab.Ptr(ref),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get file content: %v; maybe wrong ref?", err)), nil
	}

	var result strings.Builder

	// Write file information
	result.WriteString(fmt.Sprintf("File: %s\n", filePath))
	result.WriteString(fmt.Sprintf("Ref: %s\n", ref))
	result.WriteString("Content:\n")
	result.WriteString(string(fileContent))

	return mcp.NewToolResultText(result.String()), nil
}

func listCommits(ctx context.Context, projectPath, since, until, ref string) (*mcp.CallToolResult, error) {
	if until == "" {
		until = time.Now().Format("2006-01-02")
	}

	if ref == "" {
		ref = "develop" // Default ref if not provided
	}

	sinceTime, err := time.Parse("2006-01-02", since)
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

	commits, _, err := util.GitlabClient().Commits.ListCommits(projectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list commits: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commits for project %s between %s and %s (ref: %s):\n\n",
		projectPath, since, until, ref))

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

func getCommitDetails(ctx context.Context, projectPath, commitSHA string) (*mcp.CallToolResult, error) {
	commit, _, err := util.GitlabClient().Commits.GetCommit(projectPath, commitSHA, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit details: %v", err)), nil
	}

	opt := &gitlab.GetCommitDiffOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	diffs, _, err := util.GitlabClient().Commits.GetCommitDiff(projectPath, commitSHA, opt)
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

func searchCommits(ctx context.Context, projectPath, author, path, since, until, ref string) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListCommitsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	if author != "" {
		opt.Author = gitlab.Ptr(author)
	}
	if path != "" {
		opt.Path = gitlab.Ptr(path)
	}
	if ref != "" {
		opt.RefName = gitlab.Ptr(ref)
	} else {
		opt.RefName = gitlab.Ptr("develop")
	}

	if since != "" {
		sinceTime, err := time.Parse("2006-01-02", since)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid since date: %v", err)), nil
		}
		opt.Since = gitlab.Ptr(sinceTime)
	}

	if until != "" {
		untilTime, err := time.Parse("2006-01-02 15:04:05", until+" 23:00:00")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid until date: %v", err)), nil
		}
		opt.Until = gitlab.Ptr(untilTime)
	}

	commits, _, err := util.GitlabClient().Commits.ListCommits(projectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to search commits: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search results for project %s:\n", projectPath))
	if author != "" {
		result.WriteString(fmt.Sprintf("Author: %s\n", author))
	}
	if path != "" {
		result.WriteString(fmt.Sprintf("Path: %s\n", path))
	}
	result.WriteString(fmt.Sprintf("Found %d commits:\n\n", len(commits)))

	for _, commit := range commits {
		result.WriteString(fmt.Sprintf("Commit: %s\n", commit.ID))
		result.WriteString(fmt.Sprintf("Author: %s <%s>\n", commit.AuthorName, commit.AuthorEmail))
		result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
		result.WriteString(fmt.Sprintf("URL: %s\n\n", commit.WebURL))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getCommitComments(ctx context.Context, projectPath, commitSHA string) (*mcp.CallToolResult, error) {
	comments, _, err := util.GitlabClient().Commits.GetCommitComments(projectPath, commitSHA, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit comments: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Comments for commit %s:\n\n", commitSHA))

	if len(comments) == 0 {
		result.WriteString("No comments found.\n")
	} else {
		for i, comment := range comments {
			result.WriteString(fmt.Sprintf("Comment #%d:\n", i+1))
			result.WriteString(fmt.Sprintf("Author: %s <%s>\n", comment.Author.Name, comment.Author.Email))
			result.WriteString(fmt.Sprintf("Note: %s\n", comment.Note))
			if comment.Path != "" {
				result.WriteString(fmt.Sprintf("File: %s", comment.Path))
				if comment.Line > 0 {
					result.WriteString(fmt.Sprintf(" (line %d, %s)", comment.Line, comment.LineType))
				}
				result.WriteString("\n")
			}
			result.WriteString("\n")
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func postCommitComment(ctx context.Context, projectPath, commitSHA, note, path string, line int, lineType string) (*mcp.CallToolResult, error) {
	opt := &gitlab.PostCommitCommentOptions{
		Note: gitlab.Ptr(note),
	}

	if path != "" {
		opt.Path = gitlab.Ptr(path)
	}
	if line > 0 {
		opt.Line = gitlab.Ptr(line)
	}
	if lineType != "" {
		opt.LineType = gitlab.Ptr(lineType)
	}

	comment, _, err := util.GitlabClient().Commits.PostCommitComment(projectPath, commitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to post commit comment: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Comment posted successfully to commit %s:\n\n", commitSHA))
	result.WriteString(fmt.Sprintf("Author: %s <%s>\n", comment.Author.Name, comment.Author.Email))
	result.WriteString(fmt.Sprintf("Note: %s\n", comment.Note))
	if comment.Path != "" {
		result.WriteString(fmt.Sprintf("File: %s", comment.Path))
		if comment.Line > 0 {
			result.WriteString(fmt.Sprintf(" (line %d, %s)", comment.Line, comment.LineType))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getCommitMergeRequests(ctx context.Context, projectPath, commitSHA string) (*mcp.CallToolResult, error) {
	mrs, _, err := util.GitlabClient().Commits.ListMergeRequestsByCommit(projectPath, commitSHA)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit merge requests: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Merge requests associated with commit %s:\n\n", commitSHA))

	if len(mrs) == 0 {
		result.WriteString("No merge requests found.\n")
	} else {
		for _, mr := range mrs {
			result.WriteString(fmt.Sprintf("MR !%d: %s\n", mr.IID, mr.Title))
			result.WriteString(fmt.Sprintf("State: %s\n", mr.State))
			result.WriteString(fmt.Sprintf("Author: %s\n", mr.Author.Name))
			result.WriteString(fmt.Sprintf("Source: %s -> %s\n", mr.SourceBranch, mr.TargetBranch))
			result.WriteString(fmt.Sprintf("URL: %s\n\n", mr.WebURL))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func cherryPickCommit(ctx context.Context, projectPath, commitSHA, branch string, dryRun bool, message string) (*mcp.CallToolResult, error) {
	opt := &gitlab.CherryPickCommitOptions{
		Branch: gitlab.Ptr(branch),
	}

	if dryRun {
		opt.DryRun = gitlab.Ptr(true)
	}
	if message != "" {
		opt.Message = gitlab.Ptr(message)
	}

	commit, _, err := util.GitlabClient().Commits.CherryPickCommit(projectPath, commitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to cherry-pick commit: %v", err)), nil
	}

	var result strings.Builder
	if dryRun {
		result.WriteString(fmt.Sprintf("Dry run: Cherry-pick of commit %s to branch %s would succeed.\n\n", commitSHA, branch))
	} else {
		result.WriteString(fmt.Sprintf("Successfully cherry-picked commit %s to branch %s:\n\n", commitSHA, branch))
	}

	result.WriteString(fmt.Sprintf("New Commit: %s\n", commit.ID))
	result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
	result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
	result.WriteString(fmt.Sprintf("URL: %s\n", commit.WebURL))

	return mcp.NewToolResultText(result.String()), nil
}

func revertCommit(ctx context.Context, projectPath, commitSHA, branch string) (*mcp.CallToolResult, error) {
	opt := &gitlab.RevertCommitOptions{
		Branch: gitlab.Ptr(branch),
	}

	commit, _, err := util.GitlabClient().Commits.RevertCommit(projectPath, commitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to revert commit: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Successfully reverted commit %s on branch %s:\n\n", commitSHA, branch))
	result.WriteString(fmt.Sprintf("Revert Commit: %s\n", commit.ID))
	result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
	result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
	result.WriteString(fmt.Sprintf("URL: %s\n", commit.WebURL))

	return mcp.NewToolResultText(result.String()), nil
}

func getCommitRefs(ctx context.Context, projectPath, commitSHA, refType string) (*mcp.CallToolResult, error) {
	opt := &gitlab.GetCommitRefsOptions{}
	if refType != "" {
		opt.Type = gitlab.Ptr(refType)
	}

	refs, _, err := util.GitlabClient().Commits.GetCommitRefs(projectPath, commitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit refs: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("References containing commit %s:\n\n", commitSHA))

	if len(refs) == 0 {
		result.WriteString("No references found.\n")
	} else {
		branches := make([]string, 0)
		tags := make([]string, 0)

		for _, ref := range refs {
			if ref.Type == "branch" {
				branches = append(branches, ref.Name)
			} else if ref.Type == "tag" {
				tags = append(tags, ref.Name)
			}
		}

		if len(branches) > 0 {
			result.WriteString("Branches:\n")
			for _, branch := range branches {
				result.WriteString(fmt.Sprintf("- %s\n", branch))
			}
			result.WriteString("\n")
		}

		if len(tags) > 0 {
			result.WriteString("Tags:\n")
			for _, tag := range tags {
				result.WriteString(fmt.Sprintf("- %s\n", tag))
			}
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}