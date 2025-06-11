package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type ListMergeRequestsArgs struct {
	ProjectPath string `json:"project_path"`
	State       string `json:"state"`
}

type GetMergeRequestArgs struct {
	ProjectPath string `json:"project_path"`
	MrIID       string `json:"mr_iid"`
}

type CreateMRNoteArgs struct {
	ProjectPath string `json:"project_path"`
	MrIID       string `json:"mr_iid"`
	Comment     string `json:"comment"`
}

type ListMRCommentsArgs struct {
	ProjectPath string `json:"project_path"`
	MrIID       string `json:"mr_iid"`
}

type CreateMergeRequestArgs struct {
	ProjectPath  string `json:"project_path"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Title        string `json:"title"`
	Description  string `json:"description"`
}

func RegisterMergeRequestTools(s *server.MCPServer) {
	mrListTool := mcp.NewTool("list_mrs",
		mcp.WithDescription("List merge requests"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("state", mcp.DefaultString("all"), mcp.Description("MR state (opened/closed/merged)")),
	)

	mrDetailsTool := mcp.NewTool("get_mr_details",
		mcp.WithDescription("Get merge request details"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
	)

	mrCommentTool := mcp.NewTool("create_mr_note",
		mcp.WithDescription("Create a note on a merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
		mcp.WithString("comment", mcp.Required(), mcp.Description("Comment text")),
	)

	listMRCommentsTool := mcp.NewTool("list_mr_comments",
		mcp.WithDescription("List all comments on a merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
	)

	createMRTool := mcp.NewTool("create_mr",
		mcp.WithDescription("Create a new merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("source_branch", mcp.Required(), mcp.Description("Source branch name")),
		mcp.WithString("target_branch", mcp.Required(), mcp.Description("Target branch name")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Merge request title")),
		mcp.WithString("description", mcp.Description("Merge request description")),
	)

	s.AddTool(mrListTool, mcp.NewTypedToolHandler(listMergeRequestsHandler))
	s.AddTool(mrDetailsTool, mcp.NewTypedToolHandler(getMergeRequestHandler))
	s.AddTool(mrCommentTool, mcp.NewTypedToolHandler(commentOnMergeRequestHandler))
	s.AddTool(listMRCommentsTool, mcp.NewTypedToolHandler(listMRCommentsHandler))
	s.AddTool(createMRTool, mcp.NewTypedToolHandler(createMergeRequestHandler))

}

func listMergeRequestsHandler(ctx context.Context, request mcp.CallToolRequest, args ListMergeRequestsArgs) (*mcp.CallToolResult, error) {
	state := args.State
	if state == "" {
		state = "all"
	}

	opt := &gitlab.ListProjectMergeRequestsOptions{
		State: &state,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	mrs, _, err := util.GitlabClient().MergeRequests.ListProjectMergeRequests(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list merge requests: %v", err)), nil
	}
	var result strings.Builder
	for _, mr := range mrs {
		result.WriteString(fmt.Sprintf("MR #%d: %s\nState: %s\nAuthor: %s\nURL: %s\nCreated: %s\n",
			mr.IID, mr.Title, mr.State, mr.Author.Username, mr.WebURL, mr.CreatedAt.Format("2006-01-02 15:04:05")))

		if mr.SourceBranch != "" {
			result.WriteString(fmt.Sprintf("Source Branch: %s\n", mr.SourceBranch))
		}
		if mr.TargetBranch != "" {
			result.WriteString(fmt.Sprintf("Target Branch: %s\n", mr.TargetBranch))
		}
		if mr.MergedAt != nil {
			result.WriteString(fmt.Sprintf("Merged At: %s\n", mr.MergedAt.Format("2006-01-02 15:04:05")))
		}
		if mr.ClosedAt != nil {
			result.WriteString(fmt.Sprintf("Closed At: %s\n", mr.ClosedAt.Format("2006-01-02 15:04:05")))
		}
		if mr.Description != "" {
			result.WriteString(fmt.Sprintf("Description: %s\n", mr.Description))
		}

		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getMergeRequestHandler(ctx context.Context, request mcp.CallToolRequest, args GetMergeRequestArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	// Get MR details
	mr, _, err := util.GitlabClient().MergeRequests.GetMergeRequest(args.ProjectPath, mrIID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get merge request: %v", err)), nil
	}

	// Get detailed changes
	changes, _, err := util.GitlabClient().MergeRequests.ListMergeRequestDiffs(args.ProjectPath, mrIID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get merge request changes: %v", err)), nil
	}

	var result strings.Builder

	// Write MR overview
	result.WriteString(fmt.Sprintf("Merge Request #%d: %s\n", mr.IID, mr.Title))
	result.WriteString(fmt.Sprintf("Author: %s\n", mr.Author.Username))
	result.WriteString(fmt.Sprintf("Source Branch: %s\n", mr.SourceBranch))
	result.WriteString(fmt.Sprintf("Target Branch: %s\n", mr.TargetBranch))
	result.WriteString(fmt.Sprintf("State: %s\n", mr.State))
	result.WriteString(fmt.Sprintf("Created: %s\n", mr.CreatedAt.Format("2006-01-02 15:04:05")))
	// Add SHAs information
	result.WriteString(fmt.Sprintf("Base SHA: %s\n", mr.DiffRefs.BaseSha))
	result.WriteString(fmt.Sprintf("Start SHA: %s\n", mr.DiffRefs.StartSha))
	result.WriteString(fmt.Sprintf("Head SHA: %s\n\n", mr.DiffRefs.HeadSha))

	if mr.Description != "" {
		result.WriteString("Description:\n")
		result.WriteString(mr.Description)
		result.WriteString("\n\n")
	}

	// Write changes overview
	result.WriteString(fmt.Sprintf("Changes Overview:\n"))
	result.WriteString(fmt.Sprintf("Total files changed: %d\n\n", len(changes)))

	// Write detailed changes for each file
	for _, change := range changes {
		result.WriteString(fmt.Sprintf("File: %s\n", change.NewPath))
		switch true {
		case change.NewFile:
			result.WriteString("Status: Added\n")
		case change.DeletedFile:
			result.WriteString("Status: Deleted\n")
		case change.RenamedFile:
			result.WriteString(fmt.Sprintf("Status: Renamed from %s\n", change.OldPath))
		default:
			result.WriteString("Status: Modified\n")
		}

		if change.Diff != "" {
			result.WriteString("Diff:\n")
			result.WriteString("```diff\n")
			result.WriteString(change.Diff)
			result.WriteString("\n```\n")
		}

		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func commentOnMergeRequestHandler(ctx context.Context, request mcp.CallToolRequest, args CreateMRNoteArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	opt := &gitlab.CreateMergeRequestNoteOptions{
		Body: &args.Comment,
	}

	note, _, err := util.GitlabClient().Notes.CreateMergeRequestNote(args.ProjectPath, mrIID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create comment: %v", err)), nil
	}

	result := fmt.Sprintf("Comment posted successfully!\nID: %d\nAuthor: %s\nCreated: %s\nContent: %s",
		note.ID, note.Author.Username, note.CreatedAt.Format("2006-01-02 15:04:05"), note.Body)

	return mcp.NewToolResultText(result), nil
}

func listMRCommentsHandler(ctx context.Context, request mcp.CallToolRequest, args ListMRCommentsArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	opt := &gitlab.ListMergeRequestNotesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
		OrderBy: gitlab.Ptr("created_at"),
		Sort:    gitlab.Ptr("desc"),
	}

	notes, _, err := util.GitlabClient().Notes.ListMergeRequestNotes(args.ProjectPath, mrIID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list merge request comments: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Comments for Merge Request !%d:\n\n", mrIID))

	for _, note := range notes {
		result.WriteString(fmt.Sprintf("ID: %d\n", note.ID))
		result.WriteString(fmt.Sprintf("Author: %s\n", note.Author.Username))
		result.WriteString(fmt.Sprintf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05")))
		if note.UpdatedAt != nil && !note.UpdatedAt.Equal(*note.CreatedAt) {
			result.WriteString(fmt.Sprintf("Updated: %s\n", note.UpdatedAt.Format("2006-01-02 15:04:05")))
		}
		result.WriteString(fmt.Sprintf("Content: %s\n", note.Body))

		if note.System {
			result.WriteString("Type: System Note\n")
		}

		if note.Position != nil {
			result.WriteString("Position Info:\n")
			result.WriteString(fmt.Sprintf("  Base SHA: %s\n", note.Position.BaseSHA))
			result.WriteString(fmt.Sprintf("  Start SHA: %s\n", note.Position.StartSHA))
			result.WriteString(fmt.Sprintf("  Head SHA: %s\n", note.Position.HeadSHA))
			result.WriteString(fmt.Sprintf("  Position Type: %s\n", note.Position.PositionType))

			if note.Position.NewPath != "" {
				result.WriteString(fmt.Sprintf("  New Path: %s\n", note.Position.NewPath))
			}
			if note.Position.NewLine != 0 {
				result.WriteString(fmt.Sprintf("  New Line: %d\n", note.Position.NewLine))
			}
			if note.Position.OldPath != "" {
				result.WriteString(fmt.Sprintf("  Old Path: %s\n", note.Position.OldPath))
			}
			if note.Position.OldLine != 0 {
				result.WriteString(fmt.Sprintf("  Old Line: %d\n", note.Position.OldLine))
			}

			if note.Position.LineRange != nil {
				result.WriteString("  Line Range:\n")
				if note.Position.LineRange.StartRange != nil {
					result.WriteString("    Start Range:\n")
					result.WriteString(fmt.Sprintf("      Line Code: %s\n", note.Position.LineRange.StartRange.LineCode))
					result.WriteString(fmt.Sprintf("      Type: %s\n", note.Position.LineRange.StartRange.Type))
					result.WriteString(fmt.Sprintf("      Old Line: %d\n", note.Position.LineRange.StartRange.OldLine))
					result.WriteString(fmt.Sprintf("      New Line: %d\n", note.Position.LineRange.StartRange.NewLine))
				}
				if note.Position.LineRange.EndRange != nil {
					result.WriteString("    End Range:\n")
					result.WriteString(fmt.Sprintf("      Line Code: %s\n", note.Position.LineRange.EndRange.LineCode))
					result.WriteString(fmt.Sprintf("      Type: %s\n", note.Position.LineRange.EndRange.Type))
					result.WriteString(fmt.Sprintf("      Old Line: %d\n", note.Position.LineRange.EndRange.OldLine))
					result.WriteString(fmt.Sprintf("      New Line: %d\n", note.Position.LineRange.EndRange.NewLine))
				}
			}
		}

		if note.Resolvable {
			result.WriteString("Resolvable: true\n")
			result.WriteString(fmt.Sprintf("Resolved: %v\n", note.Resolved))
			if note.Resolved {
				result.WriteString(fmt.Sprintf("Resolved By: %s\n", note.ResolvedBy.Username))
				if note.ResolvedAt != nil {
					result.WriteString(fmt.Sprintf("Resolved At: %s\n", note.ResolvedAt.Format("2006-01-02 15:04:05")))
				}
			}
		}

		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func createMergeRequestHandler(ctx context.Context, request mcp.CallToolRequest, args CreateMergeRequestArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.CreateMergeRequestOptions{
		Title:        &args.Title,
		SourceBranch: &args.SourceBranch,
		TargetBranch: &args.TargetBranch,
	}

	// Add description if provided
	if args.Description != "" {
		opt.Description = &args.Description
	}

	mr, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create merge request: %v", err)), nil
	}

	result := strings.Builder{}
	result.WriteString("Merge Request created successfully!\n\n")
	result.WriteString(fmt.Sprintf("MR #%d: %s\n", mr.IID, mr.Title))
	result.WriteString(fmt.Sprintf("State: %s\n", mr.State))
	result.WriteString(fmt.Sprintf("Source Branch: %s\n", mr.SourceBranch))
	result.WriteString(fmt.Sprintf("Target Branch: %s\n", mr.TargetBranch))
	result.WriteString(fmt.Sprintf("Author: %s\n", mr.Author.Username))
	result.WriteString(fmt.Sprintf("Created: %s\n", mr.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("URL: %s\n", mr.WebURL))

	if mr.Description != "" {
		result.WriteString("\nDescription:\n")
		result.WriteString(mr.Description)
	}

	return mcp.NewToolResultText(result.String()), nil
} 