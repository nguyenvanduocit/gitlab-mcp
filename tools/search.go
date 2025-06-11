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

// Search arguments structures
type GlobalSearchArgs struct {
	Query string `json:"query"`
	Scope string `json:"scope"`
	Ref   string `json:"ref,omitempty"`
}

type GroupSearchArgs struct {
	GroupID string `json:"group_id"`
	Query   string `json:"query"`
	Scope   string `json:"scope"`
	Ref     string `json:"ref,omitempty"`
}

type ProjectSearchArgs struct {
	ProjectID string `json:"project_id"`
	Query     string `json:"query"`
	Scope     string `json:"scope"`
	Ref       string `json:"ref,omitempty"`
}

// RegisterSearchTools registers all search-related tools
func RegisterSearchTools(s *server.MCPServer) {
	// Global search tool
	globalSearchTool := mcp.NewTool("search_global",
		mcp.WithDescription("Search across all GitLab content globally. Supports searching projects, merge requests, commits, blobs, and users."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query string")),
		mcp.WithString("scope", mcp.Required(), mcp.Description("Search scope: projects, merge_requests, commits, blobs, users")),
		mcp.WithString("ref", mcp.Description("Repository branch or tag to search in (optional)")),
	)

	// Group search tool
	groupSearchTool := mcp.NewTool("search_group",
		mcp.WithDescription("Search within a specific GitLab group. Supports searching projects, merge requests, commits, blobs, and users within the group."),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("Group ID or path")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query string")),
		mcp.WithString("scope", mcp.Required(), mcp.Description("Search scope: projects, merge_requests, commits, blobs, users")),
		mcp.WithString("ref", mcp.Description("Repository branch or tag to search in (optional)")),
	)

	// Project search tool
	projectSearchTool := mcp.NewTool("search_project",
		mcp.WithDescription("Search within a specific GitLab project. Supports searching merge requests, commits, blobs, and users within the project."),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query string")),
		mcp.WithString("scope", mcp.Required(), mcp.Description("Search scope: merge_requests, commits, blobs, users")),
		mcp.WithString("ref", mcp.Description("Repository branch or tag to search in (optional)")),
	)

	searchMergeRequestsGlobalTool := mcp.NewTool("search_merge_requests_global",
		mcp.WithDescription("Search for merge requests across all GitLab projects"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query for merge requests")),
		mcp.WithString("ref", mcp.Description("Repository branch or tag to search in (optional)")),
	)

	searchCommitsGlobalTool := mcp.NewTool("search_commits_global",
		mcp.WithDescription("Search for commits across all GitLab projects"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query for commits")),
		mcp.WithString("ref", mcp.Description("Repository branch or tag to search in (optional)")),
	)

	searchCodeGlobalTool := mcp.NewTool("search_code_global",
		mcp.WithDescription("Search for code across all GitLab projects"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query for code content")),
		mcp.WithString("ref", mcp.Description("Repository branch or tag to search in (optional)")),
	)

	// Register all tools
	s.AddTool(globalSearchTool, mcp.NewTypedToolHandler(globalSearchHandler))
	s.AddTool(groupSearchTool, mcp.NewTypedToolHandler(groupSearchHandler))
	s.AddTool(projectSearchTool, mcp.NewTypedToolHandler(projectSearchHandler))
	s.AddTool(searchMergeRequestsGlobalTool, mcp.NewTypedToolHandler(searchMergeRequestsGlobalHandler))
	s.AddTool(searchCommitsGlobalTool, mcp.NewTypedToolHandler(searchCommitsGlobalHandler))
	s.AddTool(searchCodeGlobalTool, mcp.NewTypedToolHandler(searchCodeGlobalHandler))
}

// Global search handler
func globalSearchHandler(ctx context.Context, request mcp.CallToolRequest, args GlobalSearchArgs) (*mcp.CallToolResult, error) {
	client := util.GitlabClient()
	
	opt := &gitlab.SearchOptions{}
	if args.Ref != "" {
		opt.Ref = &args.Ref
	}

	var result string

	switch args.Scope {
	case "projects":
		projects, _, err := client.Search.Projects(args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search projects: %v", err)), nil
		}
		result = formatProjectsResult(projects)

	case "merge_requests":
		mrs, _, err := client.Search.MergeRequests(args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search merge requests: %v", err)), nil
		}
		result = formatMergeRequestsResult(mrs)

	case "commits":
		commits, _, err := client.Search.Commits(args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search commits: %v", err)), nil
		}
		result = formatCommitsResult(commits)

	case "blobs":
		blobs, _, err := client.Search.Blobs(args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search blobs: %v", err)), nil
		}
		result = formatBlobsResult(blobs)

	case "users":
		users, _, err := client.Search.Users(args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search users: %v", err)), nil
		}
		result = formatUsersResult(users)

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported scope: %s. Supported scopes: projects, issues, merge_requests, milestones, snippet_titles, wiki_blobs, commits, blobs, users", args.Scope)), nil
	}

	if result == "" {
		result = fmt.Sprintf("No results found for query '%s' in scope '%s'", args.Query, args.Scope)
	}

	return mcp.NewToolResultText(result), nil
}

// Group search handler
func groupSearchHandler(ctx context.Context, request mcp.CallToolRequest, args GroupSearchArgs) (*mcp.CallToolResult, error) {
	client := util.GitlabClient()
	
	opt := &gitlab.SearchOptions{}
	if args.Ref != "" {
		opt.Ref = &args.Ref
	}

	var result string

	switch args.Scope {
	case "blobs":
		blobs, _, err := client.Search.BlobsByGroup(args.GroupID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search blobs in group: %v", err)), nil
		}
		result = formatBlobsResult(blobs)

	case "projects":
		projects, _, err := client.Search.ProjectsByGroup(args.GroupID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search projects in group: %v", err)), nil
		}
		result = formatProjectsResult(projects)

	case "merge_requests":
		mrs, _, err := client.Search.MergeRequestsByGroup(args.GroupID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search merge requests in group: %v", err)), nil
		}
		result = formatMergeRequestsResult(mrs)

	case "commits":
		commits, _, err := client.Search.CommitsByGroup(args.GroupID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search commits in group: %v", err)), nil
		}
		result = formatCommitsResult(commits)

	case "users":
		users, _, err := client.Search.UsersByGroup(args.GroupID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search users in group: %v", err)), nil
		}
		result = formatUsersResult(users)

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported scope for group search: %s. Supported scopes: projects, issues, merge_requests, milestones, wiki_blobs, commits, blobs, users", args.Scope)), nil
	}

	if result == "" {
		result = fmt.Sprintf("No results found for query '%s' in scope '%s' within group '%s'", args.Query, args.Scope, args.GroupID)
	}

	return mcp.NewToolResultText(result), nil
}

// Project search handler
func projectSearchHandler(ctx context.Context, request mcp.CallToolRequest, args ProjectSearchArgs) (*mcp.CallToolResult, error) {
	client := util.GitlabClient()
	
	opt := &gitlab.SearchOptions{}
	if args.Ref != "" {
		opt.Ref = &args.Ref
	}

	var result string

	switch args.Scope {
	case "blobs":
		blobs, _, err := client.Search.BlobsByProject(args.ProjectID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search blobs in project: %v", err)), nil
		}
		result = formatBlobsResult(blobs)

	case "merge_requests":
		mrs, _, err := client.Search.MergeRequestsByProject(args.ProjectID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search merge requests in project: %v", err)), nil
		}
		result = formatMergeRequestsResult(mrs)

	case "commits":
		commits, _, err := client.Search.CommitsByProject(args.ProjectID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search commits in project: %v", err)), nil
		}
		result = formatCommitsResult(commits)

	case "users":
		users, _, err := client.Search.UsersByProject(args.ProjectID, args.Query, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search users in project: %v", err)), nil
		}
		result = formatUsersResult(users)

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported scope for project search: %s. Supported scopes: issues, merge_requests, milestones, notes, wiki_blobs, commits, blobs, users", args.Scope)), nil
	}

	if result == "" {
		result = fmt.Sprintf("No results found for query '%s' in scope '%s' within project '%s'", args.Query, args.Scope, args.ProjectID)
	}

	return mcp.NewToolResultText(result), nil
}

// Specialized search handlers
func searchIssuesGlobalHandler(ctx context.Context, request mcp.CallToolRequest, args GlobalSearchArgs) (*mcp.CallToolResult, error) {
	args.Scope = "issues"
	return globalSearchHandler(ctx, request, args)
}

func searchMergeRequestsGlobalHandler(ctx context.Context, request mcp.CallToolRequest, args GlobalSearchArgs) (*mcp.CallToolResult, error) {
	args.Scope = "merge_requests"
	return globalSearchHandler(ctx, request, args)
}

func searchCommitsGlobalHandler(ctx context.Context, request mcp.CallToolRequest, args GlobalSearchArgs) (*mcp.CallToolResult, error) {
	args.Scope = "commits"
	return globalSearchHandler(ctx, request, args)
}

func searchCodeGlobalHandler(ctx context.Context, request mcp.CallToolRequest, args GlobalSearchArgs) (*mcp.CallToolResult, error) {
	args.Scope = "blobs"
	return globalSearchHandler(ctx, request, args)
}

// Formatting functions
func formatProjectsResult(projects []*gitlab.Project) string {
	if len(projects) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d project(s):\n\n", len(projects)))

	for i, project := range projects {
		result.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", i+1, project.Name, project.PathWithNamespace))
		result.WriteString(fmt.Sprintf("   ID: %d\n", project.ID))
		if project.Description != "" {
			result.WriteString(fmt.Sprintf("   Description: %s\n", project.Description))
		}
		result.WriteString(fmt.Sprintf("   URL: %s\n", project.WebURL))
		if project.LastActivityAt != nil {
			result.WriteString(fmt.Sprintf("   Last Activity: %s\n", project.LastActivityAt.Format("2006-01-02 15:04:05")))
		}
		result.WriteString("\n")
	}

	return result.String()
}

func formatIssuesResult(issues []*gitlab.Issue) string {
	if len(issues) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d issue(s):\n\n", len(issues)))

	for i, issue := range issues {
		result.WriteString(fmt.Sprintf("%d. **#%d: %s**\n", i+1, issue.IID, issue.Title))
		result.WriteString(fmt.Sprintf("   Project: %s\n", issue.ProjectID))
		result.WriteString(fmt.Sprintf("   State: %s\n", issue.State))
		result.WriteString(fmt.Sprintf("   Author: %s\n", issue.Author.Name))
		if issue.Assignee != nil {
			result.WriteString(fmt.Sprintf("   Assignee: %s\n", issue.Assignee.Name))
		}
		result.WriteString(fmt.Sprintf("   Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("   URL: %s\n", issue.WebURL))
		result.WriteString("\n")
	}

	return result.String()
}

func formatMergeRequestsResult(mrs []*gitlab.MergeRequest) string {
	if len(mrs) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d merge request(s):\n\n", len(mrs)))

	for i, mr := range mrs {
		result.WriteString(fmt.Sprintf("%d. **!%d: %s**\n", i+1, mr.IID, mr.Title))
		result.WriteString(fmt.Sprintf("   Project: %s\n", mr.ProjectID))
		result.WriteString(fmt.Sprintf("   State: %s\n", mr.State))
		result.WriteString(fmt.Sprintf("   Author: %s\n", mr.Author.Name))
		if mr.Assignee != nil {
			result.WriteString(fmt.Sprintf("   Assignee: %s\n", mr.Assignee.Name))
		}
		result.WriteString(fmt.Sprintf("   Source Branch: %s\n", mr.SourceBranch))
		result.WriteString(fmt.Sprintf("   Target Branch: %s\n", mr.TargetBranch))
		result.WriteString(fmt.Sprintf("   Created: %s\n", mr.CreatedAt.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("   URL: %s\n", mr.WebURL))
		result.WriteString("\n")
	}

	return result.String()
}

func formatMilestonesResult(milestones []*gitlab.Milestone) string {
	if len(milestones) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d milestone(s):\n\n", len(milestones)))

	for i, milestone := range milestones {
		result.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, milestone.Title))
		result.WriteString(fmt.Sprintf("   ID: %d\n", milestone.ID))
		if milestone.Description != "" {
			result.WriteString(fmt.Sprintf("   Description: %s\n", milestone.Description))
		}
		result.WriteString(fmt.Sprintf("   State: %s\n", milestone.State))
		if milestone.DueDate != nil {
			result.WriteString(fmt.Sprintf("   Due Date: %s\n", milestone.DueDate.String()))
		}
		result.WriteString(fmt.Sprintf("   URL: %s\n", milestone.WebURL))
		result.WriteString("\n")
	}

	return result.String()
}

func formatSnippetsResult(snippets []*gitlab.Snippet) string {
	if len(snippets) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d snippet(s):\n\n", len(snippets)))

	for i, snippet := range snippets {
		result.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, snippet.Title))
		result.WriteString(fmt.Sprintf("   ID: %d\n", snippet.ID))
		result.WriteString(fmt.Sprintf("   Author: %s\n", snippet.Author.Name))
		result.WriteString(fmt.Sprintf("   Filename: %s\n", snippet.FileName))
		result.WriteString(fmt.Sprintf("   Created: %s\n", snippet.CreatedAt.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("   URL: %s\n", snippet.WebURL))
		result.WriteString("\n")
	}

	return result.String()
}

func formatWikisResult(wikis []*gitlab.Wiki) string {
	if len(wikis) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d wiki page(s):\n\n", len(wikis)))

	for i, wiki := range wikis {
		result.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, wiki.Title))
		if wiki.Content != "" {
			// Truncate content for display
			content := wiki.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			result.WriteString(fmt.Sprintf("   Content: %s\n", content))
		}
		result.WriteString("\n")
	}

	return result.String()
}

func formatCommitsResult(commits []*gitlab.Commit) string {
	if len(commits) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d commit(s):\n\n", len(commits)))

	for i, commit := range commits {
		result.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, commit.Title))
		result.WriteString(fmt.Sprintf("   SHA: %s\n", commit.ID))
		result.WriteString(fmt.Sprintf("   Author: %s <%s>\n", commit.AuthorName, commit.AuthorEmail))
		result.WriteString(fmt.Sprintf("   Date: %s\n", commit.CreatedAt.Format("2006-01-02 15:04:05")))
		if commit.Message != commit.Title && commit.Message != "" {
			// Show first few lines of commit message
			lines := strings.Split(commit.Message, "\n")
			if len(lines) > 3 {
				result.WriteString(fmt.Sprintf("   Message: %s...\n", strings.Join(lines[:3], " ")))
			} else {
				result.WriteString(fmt.Sprintf("   Message: %s\n", commit.Message))
			}
		}
		result.WriteString(fmt.Sprintf("   URL: %s\n", commit.WebURL))
		result.WriteString("\n")
	}

	return result.String()
}

func formatBlobsResult(blobs []*gitlab.Blob) string {
	if len(blobs) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d code file(s):\n\n", len(blobs)))

	for i, blob := range blobs {
		result.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, blob.Filename))
		result.WriteString(fmt.Sprintf("   Path: %s\n", blob.Path))
		result.WriteString(fmt.Sprintf("   Project ID: %d\n", blob.ProjectID))
		result.WriteString(fmt.Sprintf("   Ref: %s\n", blob.Ref))
		if blob.Startline > 0 {
			result.WriteString(fmt.Sprintf("   Start Line: %d\n", blob.Startline))
		}
		if blob.Data != "" {
			// Show first few lines of the blob data
			lines := strings.Split(blob.Data, "\n")
			if len(lines) > 5 {
				result.WriteString(fmt.Sprintf("   Preview:\n   %s\n   ...\n", strings.Join(lines[:5], "\n   ")))
			} else {
				result.WriteString(fmt.Sprintf("   Content:\n   %s\n", strings.Join(lines, "\n   ")))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

func formatUsersResult(users []*gitlab.User) string {
	if len(users) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d user(s):\n\n", len(users)))

	for i, user := range users {
		result.WriteString(fmt.Sprintf("%d. **%s** (@%s)\n", i+1, user.Name, user.Username))
		result.WriteString(fmt.Sprintf("   ID: %d\n", user.ID))
		if user.Email != "" {
			result.WriteString(fmt.Sprintf("   Email: %s\n", user.Email))
		}
		result.WriteString(fmt.Sprintf("   State: %s\n", user.State))
		if user.WebURL != "" {
			result.WriteString(fmt.Sprintf("   URL: %s\n", user.WebURL))
		}
		result.WriteString("\n")
	}

	return result.String()
}

func formatNotesResult(notes []*gitlab.Note) string {
	if len(notes) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d note(s):\n\n", len(notes)))

	for i, note := range notes {
		result.WriteString(fmt.Sprintf("%d. **Note by %s**\n", i+1, note.Author.Name))
		result.WriteString(fmt.Sprintf("   ID: %d\n", note.ID))
		result.WriteString(fmt.Sprintf("   Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05")))
		
		// Truncate note body for display
		body := note.Body
		if len(body) > 300 {
			body = body[:300] + "..."
		}
		result.WriteString(fmt.Sprintf("   Content: %s\n", body))
		result.WriteString("\n")
	}

	return result.String()
}
