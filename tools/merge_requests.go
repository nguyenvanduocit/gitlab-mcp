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

// Consolidated MR Management Args with action-based approach
type MergeRequestManagementArgs struct {
	Action      string `json:"action" validate:"required,oneof=list get create update accept rebase changes"`
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid,omitempty" validate:"omitempty,min=1"`
	
	// List action specific
	ListOptions struct {
		State string `json:"state" validate:"omitempty,oneof=opened closed merged all"`
	} `json:"list_options,omitempty"`
	
	// Create action specific
	CreateOptions struct {
		SourceBranch string `json:"source_branch" validate:"required_with=CreateOptions,min=1"`
		TargetBranch string `json:"target_branch" validate:"required_with=CreateOptions,min=1"`
		Title        string `json:"title" validate:"required_with=CreateOptions,min=1,max=255"`
		Description  string `json:"description" validate:"max=1000000"`
	} `json:"create_options,omitempty"`
	
	// Update action specific
	UpdateOptions struct {
		Title                string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
		Description          string `json:"description,omitempty" validate:"max=1000000"`
		TargetBranch         string `json:"target_branch,omitempty" validate:"omitempty,min=1"`
		StateEvent           string `json:"state_event,omitempty" validate:"omitempty,oneof=close reopen"`
		AssigneeID           int    `json:"assignee_id,omitempty" validate:"omitempty,min=1"`
		MilestoneID          int    `json:"milestone_id,omitempty" validate:"omitempty,min=1"`
		Labels               string `json:"labels,omitempty"`
		RemoveSourceBranch   bool   `json:"remove_source_branch,omitempty"`
		Squash               bool   `json:"squash,omitempty"`
		DiscussionLocked     bool   `json:"discussion_locked,omitempty"`
	} `json:"update_options,omitempty"`
	
	// Accept/Merge action specific
	AcceptOptions struct {
		MergeCommitMessage        string `json:"merge_commit_message,omitempty" validate:"max=1000"`
		SquashCommitMessage       string `json:"squash_commit_message,omitempty" validate:"max=1000"`
		Squash                    bool   `json:"squash,omitempty"`
		ShouldRemoveSourceBranch  bool   `json:"should_remove_source_branch,omitempty"`
		MergeWhenPipelineSucceeds bool   `json:"merge_when_pipeline_succeeds,omitempty"`
	} `json:"accept_options,omitempty"`
	
	// Rebase action specific
	RebaseOptions struct {
		SkipCI bool `json:"skip_ci,omitempty"`
	} `json:"rebase_options,omitempty"`
	
	// Changes action specific
	ChangesOptions struct {
		AccessRawDiffs bool `json:"access_raw_diffs,omitempty"`
		Unidiff        bool `json:"unidiff,omitempty"`
	} `json:"changes_options,omitempty"`
}

// Consolidated MR Comments Args with action-based approach
type MergeRequestCommentsArgs struct {
	Action      string `json:"action" validate:"required,oneof=list create"`
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
	
	// Create comment specific
	CommentOptions struct {
		Comment string `json:"comment" validate:"required_with=CommentOptions,min=1,max=1000000"`
	} `json:"comment_options,omitempty"`
}

// Consolidated MR Pipeline Args with action-based approach
type MergeRequestPipelineArgs struct {
	Action      string `json:"action" validate:"required,oneof=list create"`
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

// Legacy individual args for backward compatibility
type ListMergeRequestsArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	State       string `json:"state" validate:"omitempty,oneof=opened closed merged all"`
}

type GetMergeRequestArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

type CreateMRNoteArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
	Comment     string `json:"comment" validate:"required,min=1,max=1000000"`
}

type ListMRCommentsArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

type CreateMergeRequestArgs struct {
	ProjectPath  string `json:"project_path" validate:"required,min=1"`
	SourceBranch string `json:"source_branch" validate:"required,min=1"`
	TargetBranch string `json:"target_branch" validate:"required,min=1"`
	Title        string `json:"title" validate:"required,min=1,max=255"`
	Description  string `json:"description" validate:"max=1000000"`
}

type AcceptMergeRequestArgs struct {
	ProjectPath               string `json:"project_path" validate:"required,min=1"`
	MrIID                     string `json:"mr_iid" validate:"required,min=1"`
	MergeCommitMessage        string `json:"merge_commit_message,omitempty" validate:"max=1000"`
	SquashCommitMessage       string `json:"squash_commit_message,omitempty" validate:"max=1000"`
	Squash                    bool   `json:"squash,omitempty"`
	ShouldRemoveSourceBranch  bool   `json:"should_remove_source_branch,omitempty"`
	MergeWhenPipelineSucceeds bool   `json:"merge_when_pipeline_succeeds,omitempty"`
}

type UpdateMergeRequestArgs struct {
	ProjectPath        string `json:"project_path" validate:"required,min=1"`
	MrIID             string `json:"mr_iid" validate:"required,min=1"`
	Title             string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description       string `json:"description,omitempty" validate:"max=1000000"`
	TargetBranch      string `json:"target_branch,omitempty" validate:"omitempty,min=1"`
	StateEvent        string `json:"state_event,omitempty" validate:"omitempty,oneof=close reopen"`
	AssigneeID        int    `json:"assignee_id,omitempty" validate:"omitempty,min=1"`
	MilestoneID       int    `json:"milestone_id,omitempty" validate:"omitempty,min=1"`
	Labels            string `json:"labels,omitempty"`
	RemoveSourceBranch bool  `json:"remove_source_branch,omitempty"`
	Squash            bool   `json:"squash,omitempty"`
	DiscussionLocked  bool   `json:"discussion_locked,omitempty"`
}

type GetMRApprovalsArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

type GetMRParticipantsArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

type GetMRPipelinesArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

type GetMRCommitsArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

type CreateMRPipelineArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
}

type RebaseMRArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1"`
	MrIID       string `json:"mr_iid" validate:"required,min=1"`
	SkipCI      bool   `json:"skip_ci,omitempty"`
}

type GetMRChangesArgs struct {
	ProjectPath    string `json:"project_path" validate:"required,min=1"`
	MrIID          string `json:"mr_iid" validate:"required,min=1"`
	AccessRawDiffs bool   `json:"access_raw_diffs,omitempty"`
	Unidiff        bool   `json:"unidiff,omitempty"`
}

func RegisterMergeRequestTools(s *server.MCPServer) {
	// Consolidated MR Management Tool
	mrManagementTool := mcp.NewTool("manage_merge_request",
		mcp.WithDescription("Comprehensive merge request management with multiple actions: list, get, create, update, accept, rebase, changes"),
		mcp.WithString("action", 
			mcp.Required(), 
			mcp.Description("Action to perform: list, get, create, update, accept, rebase, changes")),
		mcp.WithString("project_path", 
			mcp.Required(), 
			mcp.Description("Project/repo path")),
		mcp.WithString("mr_iid", 
			mcp.Description("Merge request IID (required for get, update, accept, rebase, changes actions)")),
		
		// List options
		mcp.WithObject("list_options",
			mcp.Description("Options for list action"),
			mcp.Properties(map[string]any{
				"state": map[string]any{
					"type":        "string",
					"description": "MR state (opened/closed/merged/all)",
					"default":     "all",
				},
			}),
		),
		
		// Create options
		mcp.WithObject("create_options",
			mcp.Description("Options for create action"),
			mcp.Properties(map[string]any{
				"source_branch": map[string]any{
					"type":        "string",
					"description": "Source branch name",
				},
				"target_branch": map[string]any{
					"type":        "string", 
					"description": "Target branch name",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Merge request title",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Merge request description",
				},
			}),
		),
		
		// Update options
		mcp.WithObject("update_options",
			mcp.Description("Options for update action"),
			mcp.Properties(map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "New title for the issue",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "New description",
				},
				"target_branch": map[string]any{
					"type":        "string",
					"description": "New target branch",
				},
				"state_event": map[string]any{
					"type":        "string",
					"description": "State event (close, reopen)",
				},
				"assignee_id": map[string]any{
					"type":        "integer",
					"description": "Assignee user ID",
				},
				"milestone_id": map[string]any{
					"type":        "integer",
					"description": "Milestone ID",
				},
				"labels": map[string]any{
					"type":        "string",
					"description": "Comma-separated list of labels",
				},
				"remove_source_branch": map[string]any{
					"type":        "boolean",
					"description": "Remove source branch after merge",
				},
				"squash": map[string]any{
					"type":        "boolean",
					"description": "Squash commits when merging",
				},
				"discussion_locked": map[string]any{
					"type":        "boolean",
					"description": "Lock discussions",
				},
			}),
		),
		
		// Accept options
		mcp.WithObject("accept_options",
			mcp.Description("Options for accept action"),
			mcp.Properties(map[string]any{
				"merge_commit_message": map[string]any{
					"type":        "string",
					"description": "Custom merge commit message",
				},
				"squash_commit_message": map[string]any{
					"type":        "string",
					"description": "Custom squash commit message",
				},
				"squash": map[string]any{
					"type":        "boolean",
					"description": "Squash commits when merging",
				},
				"should_remove_source_branch": map[string]any{
					"type":        "boolean",
					"description": "Remove source branch after merge",
				},
				"merge_when_pipeline_succeeds": map[string]any{
					"type":        "boolean",
					"description": "Merge when pipeline succeeds",
				},
			}),
		),
		
		// Rebase options
		mcp.WithObject("rebase_options",
			mcp.Description("Options for rebase action"),
			mcp.Properties(map[string]any{
				"skip_ci": map[string]any{
					"type":        "boolean",
					"description": "Skip CI for rebase",
				},
			}),
		),
		
		// Changes options
		mcp.WithObject("changes_options",
			mcp.Description("Options for changes action"),
			mcp.Properties(map[string]any{
				"access_raw_diffs": map[string]any{
					"type":        "boolean",
					"description": "Access raw diffs",
				},
				"unidiff": map[string]any{
					"type":        "boolean",
					"description": "Show unified diff format",
				},
			}),
		),
	)

	// Consolidated MR Comments Tool
	mrCommentsTool := mcp.NewTool("manage_merge_request_comments",
		mcp.WithDescription("Manage merge request comments with actions: list, create"),
		mcp.WithString("action", 
			mcp.Required(), 
			mcp.Description("Action to perform: list, create")),
		mcp.WithString("project_path", 
			mcp.Required(), 
			mcp.Description("Project/repo path")),
		mcp.WithString("mr_iid", 
			mcp.Required(), 
			mcp.Description("Merge request IID")),
		
		// Comment options
		mcp.WithObject("comment_options",
			mcp.Description("Options for create action"),
			mcp.Properties(map[string]any{
				"comment": map[string]any{
					"type":        "string",
					"description": "Comment text",
				},
			}),
		),
	)

	// Consolidated MR Pipeline Tool
	mrPipelineTool := mcp.NewTool("manage_merge_request_pipeline",
		mcp.WithDescription("Manage merge request pipelines with actions: list, create"),
		mcp.WithString("action", 
			mcp.Required(), 
			mcp.Description("Action to perform: list, create")),
		mcp.WithString("project_path", 
			mcp.Required(), 
			mcp.Description("Project/repo path")),
		mcp.WithString("mr_iid", 
			mcp.Required(), 
			mcp.Description("Merge request IID")),
	)

	// MR Commits Tool (standalone as it's unique)
	getMRCommitsTool := mcp.NewTool("get_mr_commits",
		mcp.WithDescription("Get merge request commits"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
	)

	// Register consolidated tools
	s.AddTool(mrManagementTool, mcp.NewTypedToolHandler(mergeRequestManagementHandler))
	s.AddTool(mrCommentsTool, mcp.NewTypedToolHandler(mergeRequestCommentsHandler))
	s.AddTool(mrPipelineTool, mcp.NewTypedToolHandler(mergeRequestPipelineHandler))
	s.AddTool(getMRCommitsTool, mcp.NewTypedToolHandler(getMRCommitsHandler))
}

// Consolidated MR Management Handler
func mergeRequestManagementHandler(ctx context.Context, request mcp.CallToolRequest, args MergeRequestManagementArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "list":
		state := "all"
		if args.ListOptions.State != "" {
			state = args.ListOptions.State
		}
		return listMergeRequestsHandler(ctx, request, ListMergeRequestsArgs{
			ProjectPath: args.ProjectPath,
			State:       state,
		})
	
	case "get":
		if args.MrIID == "" {
			return mcp.NewToolResultError("mr_iid is required for get action"), nil
		}
		return getMergeRequestHandler(ctx, request, GetMergeRequestArgs{
			ProjectPath: args.ProjectPath,
			MrIID:       args.MrIID,
		})
	
	case "create":
		if args.CreateOptions.SourceBranch == "" || args.CreateOptions.TargetBranch == "" || args.CreateOptions.Title == "" {
			return mcp.NewToolResultError("source_branch, target_branch, and title are required for create action"), nil
		}
		return createMergeRequestHandler(ctx, request, CreateMergeRequestArgs{
			ProjectPath:  args.ProjectPath,
			SourceBranch: args.CreateOptions.SourceBranch,
			TargetBranch: args.CreateOptions.TargetBranch,
			Title:        args.CreateOptions.Title,
			Description:  args.CreateOptions.Description,
		})
	
	case "update":
		if args.MrIID == "" {
			return mcp.NewToolResultError("mr_iid is required for update action"), nil
		}
		return updateMergeRequestHandler(ctx, request, UpdateMergeRequestArgs{
			ProjectPath:        args.ProjectPath,
			MrIID:             args.MrIID,
			Title:             args.UpdateOptions.Title,
			Description:       args.UpdateOptions.Description,
			TargetBranch:      args.UpdateOptions.TargetBranch,
			StateEvent:        args.UpdateOptions.StateEvent,
			AssigneeID:        args.UpdateOptions.AssigneeID,
			MilestoneID:       args.UpdateOptions.MilestoneID,
			Labels:            args.UpdateOptions.Labels,
			RemoveSourceBranch: args.UpdateOptions.RemoveSourceBranch,
			Squash:            args.UpdateOptions.Squash,
			DiscussionLocked:  args.UpdateOptions.DiscussionLocked,
		})
	
	case "accept":
		if args.MrIID == "" {
			return mcp.NewToolResultError("mr_iid is required for accept action"), nil
		}
		return acceptMergeRequestHandler(ctx, request, AcceptMergeRequestArgs{
			ProjectPath:               args.ProjectPath,
			MrIID:                    args.MrIID,
			MergeCommitMessage:       args.AcceptOptions.MergeCommitMessage,
			SquashCommitMessage:      args.AcceptOptions.SquashCommitMessage,
			Squash:                   args.AcceptOptions.Squash,
			ShouldRemoveSourceBranch: args.AcceptOptions.ShouldRemoveSourceBranch,
			MergeWhenPipelineSucceeds: args.AcceptOptions.MergeWhenPipelineSucceeds,
		})
	
	case "rebase":
		if args.MrIID == "" {
			return mcp.NewToolResultError("mr_iid is required for rebase action"), nil
		}
		return rebaseMRHandler(ctx, request, RebaseMRArgs{
			ProjectPath: args.ProjectPath,
			MrIID:       args.MrIID,
			SkipCI:      args.RebaseOptions.SkipCI,
		})
	
	case "changes":
		if args.MrIID == "" {
			return mcp.NewToolResultError("mr_iid is required for changes action"), nil
		}
		return getMRChangesHandler(ctx, request, GetMRChangesArgs{
			ProjectPath:    args.ProjectPath,
			MrIID:          args.MrIID,
			AccessRawDiffs: args.ChangesOptions.AccessRawDiffs,
			Unidiff:        args.ChangesOptions.Unidiff,
		})
	
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported action: %s. Supported actions: list, get, create, update, accept, rebase, changes", args.Action)), nil
	}
}

// Consolidated MR Comments Handler
func mergeRequestCommentsHandler(ctx context.Context, request mcp.CallToolRequest, args MergeRequestCommentsArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "list":
		return listMRCommentsHandler(ctx, request, ListMRCommentsArgs{
			ProjectPath: args.ProjectPath,
			MrIID:       args.MrIID,
		})
	
	case "create":
		if args.CommentOptions.Comment == "" {
			return mcp.NewToolResultError("comment is required for create action"), nil
		}
		return commentOnMergeRequestHandler(ctx, request, CreateMRNoteArgs{
			ProjectPath: args.ProjectPath,
			MrIID:       args.MrIID,
			Comment:     args.CommentOptions.Comment,
		})
	
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported action: %s. Supported actions: list, create", args.Action)), nil
	}
}

// Consolidated MR Pipeline Handler
func mergeRequestPipelineHandler(ctx context.Context, request mcp.CallToolRequest, args MergeRequestPipelineArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "list":
		return getMRPipelinesHandler(ctx, request, GetMRPipelinesArgs{
			ProjectPath: args.ProjectPath,
			MrIID:       args.MrIID,
		})
	
	case "create":
		return createMRPipelineHandler(ctx, request, CreateMRPipelineArgs{
			ProjectPath: args.ProjectPath,
			MrIID:       args.MrIID,
		})
	
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported action: %s. Supported actions: list, create", args.Action)), nil
	}
}

// New handler for update MR
func updateMergeRequestHandler(ctx context.Context, request mcp.CallToolRequest, args UpdateMergeRequestArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	opt := &gitlab.UpdateMergeRequestOptions{}
	
	if args.Title != "" {
		opt.Title = &args.Title
	}
	if args.Description != "" {
		opt.Description = &args.Description
	}
	if args.TargetBranch != "" {
		opt.TargetBranch = &args.TargetBranch
	}
	if args.StateEvent != "" {
		opt.StateEvent = &args.StateEvent
	}
	if args.AssigneeID != 0 {
		opt.AssigneeID = &args.AssigneeID
	}
	if args.MilestoneID != 0 {
		opt.MilestoneID = &args.MilestoneID
	}
	// TODO: Fix Labels field assignment - requires proper LabelOptions type
	// if args.Labels != "" {
	//	opt.Labels = gitlab.Ptr(args.Labels)
	// }
	if args.RemoveSourceBranch {
		opt.RemoveSourceBranch = &args.RemoveSourceBranch
	}
	if args.Squash {
		opt.Squash = &args.Squash
	}
	if args.DiscussionLocked {
		opt.DiscussionLocked = &args.DiscussionLocked
	}

	mr, _, err := util.GitlabClient().MergeRequests.UpdateMergeRequest(args.ProjectPath, mrIID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update merge request: %v", err)), nil
	}

	result := strings.Builder{}
	result.WriteString("Merge Request updated successfully!\n\n")
	result.WriteString(fmt.Sprintf("MR #%d: %s\n", mr.IID, mr.Title))
	result.WriteString(fmt.Sprintf("State: %s\n", mr.State))
	result.WriteString(fmt.Sprintf("Source Branch: %s\n", mr.SourceBranch))
	result.WriteString(fmt.Sprintf("Target Branch: %s\n", mr.TargetBranch))
	result.WriteString(fmt.Sprintf("Author: %s\n", mr.Author.Username))
	result.WriteString(fmt.Sprintf("Updated: %s\n", mr.UpdatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("URL: %s\n", mr.WebURL))

	if mr.Description != "" {
		result.WriteString("\nDescription:\n")
		result.WriteString(mr.Description)
	}

	return mcp.NewToolResultText(result.String()), nil
}

// New handler for accept MR
func acceptMergeRequestHandler(ctx context.Context, request mcp.CallToolRequest, args AcceptMergeRequestArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	opt := &gitlab.AcceptMergeRequestOptions{}
	
	if args.MergeCommitMessage != "" {
		opt.MergeCommitMessage = &args.MergeCommitMessage
	}
	if args.SquashCommitMessage != "" {
		opt.SquashCommitMessage = &args.SquashCommitMessage
	}
	if args.Squash {
		opt.Squash = &args.Squash
	}
	if args.ShouldRemoveSourceBranch {
		opt.ShouldRemoveSourceBranch = &args.ShouldRemoveSourceBranch
	}
	if args.MergeWhenPipelineSucceeds {
		opt.MergeWhenPipelineSucceeds = &args.MergeWhenPipelineSucceeds
	}

	mr, _, err := util.GitlabClient().MergeRequests.AcceptMergeRequest(args.ProjectPath, mrIID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to accept merge request: %v", err)), nil
	}

	result := strings.Builder{}
	result.WriteString("Merge Request accepted successfully!\n\n")
	result.WriteString(fmt.Sprintf("MR #%d: %s\n", mr.IID, mr.Title))
	result.WriteString(fmt.Sprintf("State: %s\n", mr.State))
	result.WriteString(fmt.Sprintf("Source Branch: %s\n", mr.SourceBranch))
	result.WriteString(fmt.Sprintf("Target Branch: %s\n", mr.TargetBranch))
	if mr.MergedAt != nil {
		result.WriteString(fmt.Sprintf("Merged At: %s\n", mr.MergedAt.Format("2006-01-02 15:04:05")))
	}
	if mr.MergeCommitSHA != "" {
		result.WriteString(fmt.Sprintf("Merge Commit SHA: %s\n", mr.MergeCommitSHA))
	}
	result.WriteString(fmt.Sprintf("URL: %s\n", mr.WebURL))

	return mcp.NewToolResultText(result.String()), nil
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

func getMRPipelinesHandler(ctx context.Context, request mcp.CallToolRequest, args GetMRPipelinesArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	pipelines, _, err := util.GitlabClient().MergeRequests.ListMergeRequestPipelines(args.ProjectPath, mrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get merge request pipelines: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pipelines for Merge Request !%d:\n\n", mrIID))

	for _, pipeline := range pipelines {
		result.WriteString(fmt.Sprintf("Pipeline ID: %d\n", pipeline.ID))
		result.WriteString(fmt.Sprintf("Status: %s\n", pipeline.Status))
		result.WriteString(fmt.Sprintf("Ref: %s\n", pipeline.Ref))
		result.WriteString(fmt.Sprintf("SHA: %s\n", pipeline.SHA))
		if pipeline.CreatedAt != nil {
			result.WriteString(fmt.Sprintf("Created: %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05")))
		}
		if pipeline.UpdatedAt != nil {
			result.WriteString(fmt.Sprintf("Updated: %s\n", pipeline.UpdatedAt.Format("2006-01-02 15:04:05")))
		}
		if pipeline.WebURL != "" {
			result.WriteString(fmt.Sprintf("URL: %s\n", pipeline.WebURL))
		}
		result.WriteString("\n")
	}

	if len(pipelines) == 0 {
		result.WriteString("No pipelines found for this merge request.\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getMRCommitsHandler(ctx context.Context, request mcp.CallToolRequest, args GetMRCommitsArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	commits, _, err := util.GitlabClient().MergeRequests.GetMergeRequestCommits(args.ProjectPath, mrIID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get merge request commits: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commits for Merge Request !%d:\n\n", mrIID))

	for _, commit := range commits {
		result.WriteString(fmt.Sprintf("Commit: %s\n", commit.ID))
		result.WriteString(fmt.Sprintf("Short ID: %s\n", commit.ShortID))
		result.WriteString(fmt.Sprintf("Title: %s\n", commit.Title))
		if commit.AuthorName != "" {
			result.WriteString(fmt.Sprintf("Author: %s <%s>\n", commit.AuthorName, commit.AuthorEmail))
		}
		if commit.CommitterName != "" {
			result.WriteString(fmt.Sprintf("Committer: %s <%s>\n", commit.CommitterName, commit.CommitterEmail))
		}
		if commit.CreatedAt != nil {
			result.WriteString(fmt.Sprintf("Created: %s\n", commit.CreatedAt.Format("2006-01-02 15:04:05")))
		}
		if commit.CommittedDate != nil {
			result.WriteString(fmt.Sprintf("Committed: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
		}
		if commit.AuthoredDate != nil {
			result.WriteString(fmt.Sprintf("Authored: %s\n", commit.AuthoredDate.Format("2006-01-02 15:04:05")))
		}
		if commit.Message != "" {
			result.WriteString(fmt.Sprintf("Message:\n%s\n", commit.Message))
		}
		if commit.WebURL != "" {
			result.WriteString(fmt.Sprintf("URL: %s\n", commit.WebURL))
		}
		result.WriteString("\n")
	}

	if len(commits) == 0 {
		result.WriteString("No commits found for this merge request.\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func createMRPipelineHandler(ctx context.Context, request mcp.CallToolRequest, args CreateMRPipelineArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	pipeline, _, err := util.GitlabClient().MergeRequests.CreateMergeRequestPipeline(args.ProjectPath, mrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create merge request pipeline: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString("Pipeline created successfully!\n\n")
	result.WriteString(fmt.Sprintf("Pipeline ID: %d\n", pipeline.ID))
	result.WriteString(fmt.Sprintf("Status: %s\n", pipeline.Status))
	result.WriteString(fmt.Sprintf("Ref: %s\n", pipeline.Ref))
	result.WriteString(fmt.Sprintf("SHA: %s\n", pipeline.SHA))
	if pipeline.CreatedAt != nil {
		result.WriteString(fmt.Sprintf("Created: %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05")))
	}
	if pipeline.WebURL != "" {
		result.WriteString(fmt.Sprintf("URL: %s\n", pipeline.WebURL))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func rebaseMRHandler(ctx context.Context, request mcp.CallToolRequest, args RebaseMRArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	opt := &gitlab.RebaseMergeRequestOptions{
		SkipCI: &args.SkipCI,
	}

	_, err = util.GitlabClient().MergeRequests.RebaseMergeRequest(args.ProjectPath, mrIID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to rebase merge request: %v", err)), nil
	}

	result := fmt.Sprintf("Merge Request !%d has been successfully rebased.\n", mrIID)
	if args.SkipCI {
		result += "CI pipeline was skipped for this rebase.\n"
	}

	return mcp.NewToolResultText(result), nil
}

func getMRChangesHandler(ctx context.Context, request mcp.CallToolRequest, args GetMRChangesArgs) (*mcp.CallToolResult, error) {
	mrIID, err := strconv.Atoi(args.MrIID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mr_iid: %v", err)), nil
	}

	opt := &gitlab.GetMergeRequestChangesOptions{
		AccessRawDiffs: &args.AccessRawDiffs,
		Unidiff:        &args.Unidiff,
	}

	mr, _, err := util.GitlabClient().MergeRequests.GetMergeRequestChanges(args.ProjectPath, mrIID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get merge request changes: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Changes for Merge Request !%d: %s\n", mr.IID, mr.Title))
	result.WriteString(fmt.Sprintf("Author: %s\n", mr.Author.Username))
	result.WriteString(fmt.Sprintf("Source Branch: %s\n", mr.SourceBranch))
	result.WriteString(fmt.Sprintf("Target Branch: %s\n", mr.TargetBranch))
	result.WriteString(fmt.Sprintf("State: %s\n", mr.State))
	if mr.ChangesCount != "" {
		result.WriteString(fmt.Sprintf("Changes Count: %s\n", mr.ChangesCount))
	}
	result.WriteString("\n")

	if mr.Description != "" {
		result.WriteString("Description:\n")
		result.WriteString(mr.Description)
		result.WriteString("\n\n")
	}

	result.WriteString("Note: This endpoint is deprecated. Consider using 'get_mr_details' instead for detailed changes information.\n")

	return mcp.NewToolResultText(result.String()), nil
} 