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

// New args structs for advanced commit tools
type SearchCommitsArgs struct {
	ProjectPath string `json:"project_path"`
	Author      string `json:"author"`
	Path        string `json:"path"`
	Since       string `json:"since"`
	Until       string `json:"until"`
	Ref         string `json:"ref"`
}

type GetCommitCommentsArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
}

type PostCommitCommentArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
	Note        string `json:"note"`
	Path        string `json:"path"`
	Line        int    `json:"line"`
	LineType    string `json:"line_type"`
}

type GetCommitStatusesArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
	Ref         string `json:"ref"`
}

type GetCommitMergeRequestsArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
}

type GetCommitGPGSignatureArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
}

type CherryPickCommitArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
	Branch      string `json:"branch"`
	DryRun      bool   `json:"dry_run"`
	Message     string `json:"message"`
}

type RevertCommitArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
	Branch      string `json:"branch"`
}

type GetCommitRefsArgs struct {
	ProjectPath string `json:"project_path"`
	CommitSHA   string `json:"commit_sha"`
	Type        string `json:"type"`
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

	// Advanced commit tools
	searchCommitsTool := mcp.NewTool("search_commits",
		mcp.WithDescription("Search commits by author, file path, and date range"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("author", mcp.Description("Filter by author name or email")),
		mcp.WithString("path", mcp.Description("Filter by file path")),
		mcp.WithString("since", mcp.Description("Start date (YYYY-MM-DD)")),
		mcp.WithString("until", mcp.Description("End date (YYYY-MM-DD)")),
		mcp.WithString("ref", mcp.Description("Branch name, tag, or commit SHA")),
	)

	commitCommentsTool := mcp.NewTool("get_commit_comments",
		mcp.WithDescription("Get comments of a commit"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA")),
	)

	postCommitCommentTool := mcp.NewTool("post_commit_comment",
		mcp.WithDescription("Add a comment to a commit"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA")),
		mcp.WithString("note", mcp.Required(), mcp.Description("Comment text")),
		mcp.WithString("path", mcp.Description("File path for line-specific comment")),
		mcp.WithNumber("line", mcp.Description("Line number for line-specific comment")),
		mcp.WithString("line_type", mcp.Description("Line type: 'new' or 'old'")),
	)

	commitMergeRequestsTool := mcp.NewTool("get_commit_merge_requests",
		mcp.WithDescription("Get merge requests associated with a commit"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA")),
	)

	cherryPickCommitTool := mcp.NewTool("cherry_pick_commit",
		mcp.WithDescription("Cherry-pick a commit to another branch"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA to cherry-pick")),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Target branch")),
		mcp.WithBoolean("dry_run", mcp.Description("Perform a dry run without making changes")),
		mcp.WithString("message", mcp.Description("Custom commit message")),
	)

	revertCommitTool := mcp.NewTool("revert_commit",
		mcp.WithDescription("Revert a commit"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA to revert")),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Target branch")),
	)

	// Register existing tools
	s.AddTool(fileContentTool, mcp.NewTypedToolHandler(getFileContentHandler))
	s.AddTool(commitsTool, mcp.NewTypedToolHandler(listCommitsHandler))
	s.AddTool(commitDetailsTool, mcp.NewTypedToolHandler(getCommitDetailsHandler))

	// Register new advanced commit tools
	s.AddTool(searchCommitsTool, mcp.NewTypedToolHandler(searchCommitsHandler))
	s.AddTool(commitCommentsTool, mcp.NewTypedToolHandler(getCommitCommentsHandler))
	s.AddTool(postCommitCommentTool, mcp.NewTypedToolHandler(postCommitCommentHandler))
	s.AddTool(commitMergeRequestsTool, mcp.NewTypedToolHandler(getCommitMergeRequestsHandler))
	s.AddTool(cherryPickCommitTool, mcp.NewTypedToolHandler(cherryPickCommitHandler))
	s.AddTool(revertCommitTool, mcp.NewTypedToolHandler(revertCommitHandler))
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

// Advanced commit handlers
func searchCommitsHandler(ctx context.Context, request mcp.CallToolRequest, args SearchCommitsArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListCommitsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	if args.Author != "" {
		opt.Author = gitlab.Ptr(args.Author)
	}
	if args.Path != "" {
		opt.Path = gitlab.Ptr(args.Path)
	}
	if args.Ref != "" {
		opt.RefName = gitlab.Ptr(args.Ref)
	} else {
		opt.RefName = gitlab.Ptr("develop")
	}

	if args.Since != "" {
		sinceTime, err := time.Parse("2006-01-02", args.Since)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid since date: %v", err)), nil
		}
		opt.Since = gitlab.Ptr(sinceTime)
	}

	if args.Until != "" {
		untilTime, err := time.Parse("2006-01-02 15:04:05", args.Until+" 23:00:00")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid until date: %v", err)), nil
		}
		opt.Until = gitlab.Ptr(untilTime)
	}

	commits, _, err := util.GitlabClient().Commits.ListCommits(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to search commits: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search results for project %s:\n", args.ProjectPath))
	if args.Author != "" {
		result.WriteString(fmt.Sprintf("Author: %s\n", args.Author))
	}
	if args.Path != "" {
		result.WriteString(fmt.Sprintf("Path: %s\n", args.Path))
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

func getCommitCommentsHandler(ctx context.Context, request mcp.CallToolRequest, args GetCommitCommentsArgs) (*mcp.CallToolResult, error) {
	comments, _, err := util.GitlabClient().Commits.GetCommitComments(args.ProjectPath, args.CommitSHA, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit comments: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Comments for commit %s:\n\n", args.CommitSHA))

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

func postCommitCommentHandler(ctx context.Context, request mcp.CallToolRequest, args PostCommitCommentArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.PostCommitCommentOptions{
		Note: gitlab.Ptr(args.Note),
	}

	if args.Path != "" {
		opt.Path = gitlab.Ptr(args.Path)
	}
	if args.Line > 0 {
		opt.Line = gitlab.Ptr(args.Line)
	}
	if args.LineType != "" {
		opt.LineType = gitlab.Ptr(args.LineType)
	}

	comment, _, err := util.GitlabClient().Commits.PostCommitComment(args.ProjectPath, args.CommitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to post commit comment: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Comment posted successfully to commit %s:\n\n", args.CommitSHA))
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

func getCommitStatusesHandler(ctx context.Context, request mcp.CallToolRequest, args GetCommitStatusesArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.GetCommitStatusesOptions{}
	if args.Ref != "" {
		opt.Ref = gitlab.Ptr(args.Ref)
	}

	statuses, _, err := util.GitlabClient().Commits.GetCommitStatuses(args.ProjectPath, args.CommitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit statuses: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pipeline statuses for commit %s:\n\n", args.CommitSHA))

	if len(statuses) == 0 {
		result.WriteString("No pipeline statuses found.\n")
	} else {
		for _, status := range statuses {
			result.WriteString(fmt.Sprintf("Status: %s\n", status.Status))
			result.WriteString(fmt.Sprintf("Name: %s\n", status.Name))
			result.WriteString(fmt.Sprintf("Ref: %s\n", status.Ref))
			result.WriteString(fmt.Sprintf("Description: %s\n", status.Description))
			result.WriteString(fmt.Sprintf("Created: %s\n", status.CreatedAt.Format("2006-01-02 15:04:05")))
			if status.StartedAt != nil {
				result.WriteString(fmt.Sprintf("Started: %s\n", status.StartedAt.Format("2006-01-02 15:04:05")))
			}
			if status.FinishedAt != nil {
				result.WriteString(fmt.Sprintf("Finished: %s\n", status.FinishedAt.Format("2006-01-02 15:04:05")))
			}
			if status.TargetURL != "" {
				result.WriteString(fmt.Sprintf("URL: %s\n", status.TargetURL))
			}
			result.WriteString("\n")
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getCommitMergeRequestsHandler(ctx context.Context, request mcp.CallToolRequest, args GetCommitMergeRequestsArgs) (*mcp.CallToolResult, error) {
	mrs, _, err := util.GitlabClient().Commits.ListMergeRequestsByCommit(args.ProjectPath, args.CommitSHA)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit merge requests: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Merge requests associated with commit %s:\n\n", args.CommitSHA))

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

func getCommitGPGSignatureHandler(ctx context.Context, request mcp.CallToolRequest, args GetCommitGPGSignatureArgs) (*mcp.CallToolResult, error) {
	signature, _, err := util.GitlabClient().Commits.GetGPGSignature(args.ProjectPath, args.CommitSHA)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get GPG signature: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("GPG signature for commit %s:\n\n", args.CommitSHA))
	result.WriteString(fmt.Sprintf("Key ID: %d\n", signature.KeyID))
	result.WriteString(fmt.Sprintf("Primary Key ID: %s\n", signature.KeyPrimaryKeyID))
	result.WriteString(fmt.Sprintf("User Name: %s\n", signature.KeyUserName))
	result.WriteString(fmt.Sprintf("User Email: %s\n", signature.KeyUserEmail))
	result.WriteString(fmt.Sprintf("Verification Status: %s\n", signature.VerificationStatus))
	if signature.KeySubkeyID > 0 {
		result.WriteString(fmt.Sprintf("Subkey ID: %d\n", signature.KeySubkeyID))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func cherryPickCommitHandler(ctx context.Context, request mcp.CallToolRequest, args CherryPickCommitArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.CherryPickCommitOptions{
		Branch: gitlab.Ptr(args.Branch),
	}

	if args.DryRun {
		opt.DryRun = gitlab.Ptr(true)
	}
	if args.Message != "" {
		opt.Message = gitlab.Ptr(args.Message)
	}

	commit, _, err := util.GitlabClient().Commits.CherryPickCommit(args.ProjectPath, args.CommitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to cherry-pick commit: %v", err)), nil
	}

	var result strings.Builder
	if args.DryRun {
		result.WriteString(fmt.Sprintf("Dry run: Cherry-pick of commit %s to branch %s would succeed.\n\n", args.CommitSHA, args.Branch))
	} else {
		result.WriteString(fmt.Sprintf("Successfully cherry-picked commit %s to branch %s:\n\n", args.CommitSHA, args.Branch))
	}

	result.WriteString(fmt.Sprintf("New Commit: %s\n", commit.ID))
	result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
	result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
	result.WriteString(fmt.Sprintf("URL: %s\n", commit.WebURL))

	return mcp.NewToolResultText(result.String()), nil
}

func revertCommitHandler(ctx context.Context, request mcp.CallToolRequest, args RevertCommitArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.RevertCommitOptions{
		Branch: gitlab.Ptr(args.Branch),
	}

	commit, _, err := util.GitlabClient().Commits.RevertCommit(args.ProjectPath, args.CommitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to revert commit: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Successfully reverted commit %s on branch %s:\n\n", args.CommitSHA, args.Branch))
	result.WriteString(fmt.Sprintf("Revert Commit: %s\n", commit.ID))
	result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
	result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
	result.WriteString(fmt.Sprintf("URL: %s\n", commit.WebURL))

	return mcp.NewToolResultText(result.String()), nil
}

func getCommitRefsHandler(ctx context.Context, request mcp.CallToolRequest, args GetCommitRefsArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.GetCommitRefsOptions{}
	if args.Type != "" {
		opt.Type = gitlab.Ptr(args.Type)
	}

	refs, _, err := util.GitlabClient().Commits.GetCommitRefs(args.ProjectPath, args.CommitSHA, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit refs: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("References containing commit %s:\n\n", args.CommitSHA))

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