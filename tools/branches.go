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

// Branch Protection Management
type BranchProtectionArgs struct {
	Action      string `json:"action" validate:"required,oneof=protect unprotect list get_protection"`
	ProjectPath string `json:"project_path" validate:"required,min=1,max=255"`
	BranchName  string `json:"branch_name" validate:"omitempty,min=1,max=255"`
	Confirmed   bool   `json:"confirmed,omitempty"`

	// Protection options
	ProtectionOptions ProtectionOptions `json:"protection_options"`
}

func RegisterBranchTools(s *server.MCPServer) {
	// Branch Protection Management Tool
	branchProtectionTool := mcp.NewTool("manage_branch_protection",
		mcp.WithDescription("Manage branch protection for GitLab projects: protect, unprotect, list, get_protection"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: protect, unprotect, list, get_protection")),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path (1-255 characters)")),
		mcp.WithString("branch_name", mcp.Description("Branch name (1-255 characters, required for: protect, unprotect, get_protection)")),
		mcp.WithBoolean("confirmed", mcp.Description("Confirmation required for protect and unprotect actions")),

		// Protection options
		mcp.WithObject("protection_options",
			mcp.Description("Branch protection configuration options"),
			mcp.Properties(map[string]any{
				"push_access_level": map[string]any{
					"type":        "string",
					"description": "Push access level: 0 (No access), 30 (Developer), 40 (Master)",
					"enum":        []string{"0", "30", "40"},
				},
				"merge_access_level": map[string]any{
					"type":        "string",
					"description": "Merge access level: 0 (No access), 30 (Developer), 40 (Master)",
					"enum":        []string{"0", "30", "40"},
				},
				"unprotect_access_level": map[string]any{
					"type":        "string",
					"description": "Unprotect access level: 30 (Developer), 40 (Master)",
					"enum":        []string{"30", "40"},
				},
				"allowed_to_push": map[string]any{
					"type":        "array",
					"description": "List of user IDs allowed to push",
					"items":       map[string]any{"type": "string"},
				},
				"allowed_to_merge": map[string]any{
					"type":        "array",
					"description": "List of user IDs allowed to merge",
					"items":       map[string]any{"type": "string"},
				},
				"allowed_to_unprotect": map[string]any{
					"type":        "array",
					"description": "List of user IDs allowed to unprotect",
					"items":       map[string]any{"type": "string"},
				},
				"code_owner_approval_required": map[string]any{
					"type":        "boolean",
					"description": "Require code owner approval for merges",
					"default":     false,
				},
			}),
		),
	)

	// Register tool
	s.AddTool(branchProtectionTool, mcp.NewTypedToolHandler(branchProtectionHandler))
}

func branchProtectionHandler(ctx context.Context, request mcp.CallToolRequest, args BranchProtectionArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "protect":
		if args.BranchName == "" {
			return mcp.NewToolResultError("branch_name is required for protect action"), nil
		}
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with protecting the branch."), nil
		}
		return protectBranch(ctx, args.ProjectPath, args.BranchName, args.ProtectionOptions)

	case "unprotect":
		if args.BranchName == "" {
			return mcp.NewToolResultError("branch_name is required for unprotect action"), nil
		}
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with unprotecting the branch."), nil
		}
		return unprotectBranch(ctx, args.ProjectPath, args.BranchName)

	case "list":
		return listProtectedBranches(ctx, args.ProjectPath)

	case "get_protection":
		if args.BranchName == "" {
			return mcp.NewToolResultError("branch_name is required for get_protection action"), nil
		}
		return getBranchProtection(ctx, args.ProjectPath, args.BranchName)

	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action: %s. Valid actions are: protect, unprotect, list, get_protection", args.Action)), nil
	}
}

type ProtectionOptions struct {
	PushAccessLevel            string   `json:"push_access_level,omitempty"`
	MergeAccessLevel           string   `json:"merge_access_level,omitempty"`
	UnprotectAccessLevel       string   `json:"unprotect_access_level,omitempty"`
	AllowedToPush              []string `json:"allowed_to_push,omitempty"`
	AllowedToMerge             []string `json:"allowed_to_merge,omitempty"`
	AllowedToUnprotect         []string `json:"allowed_to_unprotect,omitempty"`
	CodeOwnerApprovalRequired  bool     `json:"code_owner_approval_required,omitempty"`
}

func protectBranch(ctx context.Context, projectPath, branchName string, options ProtectionOptions) (*mcp.CallToolResult, error) {
	opt := &gitlab.ProtectRepositoryBranchesOptions{
		Name: gitlab.Ptr(branchName),
	}

	// Set access levels with defaults if not specified
	if options.PushAccessLevel != "" {
		if level := parseAccessLevel(options.PushAccessLevel); level != nil {
			opt.PushAccessLevel = level
		}
	} else {
		opt.PushAccessLevel = gitlab.Ptr(gitlab.MaintainerPermissions) // Default to maintainer
	}

	if options.MergeAccessLevel != "" {
		if level := parseAccessLevel(options.MergeAccessLevel); level != nil {
			opt.MergeAccessLevel = level
		}
	} else {
		opt.MergeAccessLevel = gitlab.Ptr(gitlab.MaintainerPermissions) // Default to maintainer
	}

	if options.UnprotectAccessLevel != "" {
		if level := parseAccessLevel(options.UnprotectAccessLevel); level != nil {
			opt.UnprotectAccessLevel = level
		}
	}

	// Set specific user permissions
	if len(options.AllowedToPush) > 0 {
		allowedToPush := make([]*gitlab.BranchPermissionOptions, len(options.AllowedToPush))
		for i, userID := range options.AllowedToPush {
			allowedToPush[i] = &gitlab.BranchPermissionOptions{
				UserID: gitlab.Ptr(parseUserID(userID)),
			}
		}
		opt.AllowedToPush = &allowedToPush
	}

	if len(options.AllowedToMerge) > 0 {
		allowedToMerge := make([]*gitlab.BranchPermissionOptions, len(options.AllowedToMerge))
		for i, userID := range options.AllowedToMerge {
			allowedToMerge[i] = &gitlab.BranchPermissionOptions{
				UserID: gitlab.Ptr(parseUserID(userID)),
			}
		}
		opt.AllowedToMerge = &allowedToMerge
	}

	if len(options.AllowedToUnprotect) > 0 {
		allowedToUnprotect := make([]*gitlab.BranchPermissionOptions, len(options.AllowedToUnprotect))
		for i, userID := range options.AllowedToUnprotect {
			allowedToUnprotect[i] = &gitlab.BranchPermissionOptions{
				UserID: gitlab.Ptr(parseUserID(userID)),
			}
		}
		opt.AllowedToUnprotect = &allowedToUnprotect
	}

	// Set code owner settings
	if options.CodeOwnerApprovalRequired {
		opt.CodeOwnerApprovalRequired = gitlab.Ptr(true)
	}

	branch, _, err := util.GitlabClient().ProtectedBranches.ProtectRepositoryBranches(projectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to protect branch: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Successfully protected branch '%s' in project %s:\n\n", branchName, projectPath))
	result.WriteString(fmt.Sprintf("Branch: %s\n", branch.Name))
	result.WriteString(fmt.Sprintf("Push Access Level: %s\n", formatAccessLevel(branch.PushAccessLevels)))
	result.WriteString(fmt.Sprintf("Merge Access Level: %s\n", formatAccessLevel(branch.MergeAccessLevels)))
	result.WriteString(fmt.Sprintf("Unprotect Access Level: %s\n", formatAccessLevel(branch.UnprotectAccessLevels)))

	if branch.CodeOwnerApprovalRequired {
		result.WriteString("Code Owner Approval: Required\n")
	}

	if len(branch.PushAccessLevels) > 0 {
		result.WriteString("\nPush Access Details:\n")
		for _, level := range branch.PushAccessLevels {
			result.WriteString(fmt.Sprintf("- Access Level: %d\n", level.AccessLevel))
		}
	}

	if len(branch.MergeAccessLevels) > 0 {
		result.WriteString("\nMerge Access Details:\n")
		for _, level := range branch.MergeAccessLevels {
			result.WriteString(fmt.Sprintf("- Access Level: %d\n", level.AccessLevel))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func unprotectBranch(ctx context.Context, projectPath, branchName string) (*mcp.CallToolResult, error) {
	_, err := util.GitlabClient().ProtectedBranches.UnprotectRepositoryBranches(projectPath, branchName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to unprotect branch: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Successfully unprotected branch '%s' in project %s\n", branchName, projectPath))
	result.WriteString("The branch is now unprotected and can be pushed to by any user with push access to the repository.\n")

	return mcp.NewToolResultText(result.String()), nil
}

func listProtectedBranches(ctx context.Context, projectPath string) (*mcp.CallToolResult, error) {
	branches, _, err := util.GitlabClient().ProtectedBranches.ListProtectedBranches(projectPath, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list protected branches: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Protected branches for project %s:\n\n", projectPath))

	if len(branches) == 0 {
		result.WriteString("No protected branches found.\n")
	} else {
		for i, branch := range branches {
			result.WriteString(fmt.Sprintf("%d. Branch: %s\n", i+1, branch.Name))
			result.WriteString(fmt.Sprintf("   Push Access: %s\n", formatAccessLevel(branch.PushAccessLevels)))
			result.WriteString(fmt.Sprintf("   Merge Access: %s\n", formatAccessLevel(branch.MergeAccessLevels)))
			result.WriteString(fmt.Sprintf("   Unprotect Access: %s\n", formatAccessLevel(branch.UnprotectAccessLevels)))
			
			if branch.CodeOwnerApprovalRequired {
				result.WriteString("   Code Owner Approval: Required\n")
			}
			
			result.WriteString("\n")
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getBranchProtection(ctx context.Context, projectPath, branchName string) (*mcp.CallToolResult, error) {
	branch, _, err := util.GitlabClient().ProtectedBranches.GetProtectedBranch(projectPath, branchName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get branch protection: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Protection details for branch '%s' in project %s:\n\n", branchName, projectPath))
	result.WriteString(fmt.Sprintf("Branch: %s\n", branch.Name))
	result.WriteString(fmt.Sprintf("Push Access Level: %s\n", formatAccessLevel(branch.PushAccessLevels)))
	result.WriteString(fmt.Sprintf("Merge Access Level: %s\n", formatAccessLevel(branch.MergeAccessLevels)))
	result.WriteString(fmt.Sprintf("Unprotect Access Level: %s\n", formatAccessLevel(branch.UnprotectAccessLevels)))

	if branch.CodeOwnerApprovalRequired {
		result.WriteString("Code Owner Approval: Required\n")
	} else {
		result.WriteString("Code Owner Approval: Not required\n")
	}

	if len(branch.PushAccessLevels) > 0 {
		result.WriteString("\nDetailed Push Access:\n")
		for _, level := range branch.PushAccessLevels {
			result.WriteString(fmt.Sprintf("- Access Level: %d\n", level.AccessLevel))
		}
	}

	if len(branch.MergeAccessLevels) > 0 {
		result.WriteString("\nDetailed Merge Access:\n")
		for _, level := range branch.MergeAccessLevels {
			result.WriteString(fmt.Sprintf("- Access Level: %d\n", level.AccessLevel))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

// Helper functions
func parseAccessLevel(level string) *gitlab.AccessLevelValue {
	switch level {
	case "0":
		return gitlab.Ptr(gitlab.NoPermissions)
	case "30":
		return gitlab.Ptr(gitlab.DeveloperPermissions)
	case "40":
		return gitlab.Ptr(gitlab.MaintainerPermissions)
	default:
		return nil
	}
}

func parseUserID(userIDStr string) int {
	// Simple conversion - in a real implementation you might want more robust parsing
	if userIDStr == "" {
		return 0
	}
	// For simplicity, we'll assume the string is a valid integer
	// In practice, you'd want proper error handling here
	var userID int
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return 0
	}
	return userID
}

func formatAccessLevel(levels []*gitlab.BranchAccessDescription) string {
	if len(levels) == 0 {
		return "None"
	}

	var parts []string
	for _, level := range levels {
		switch level.AccessLevel {
		case 0:
			parts = append(parts, "No access")
		case 30:
			parts = append(parts, "Developer")
		case 40:
			parts = append(parts, "Maintainer")
		default:
			parts = append(parts, fmt.Sprintf("Level %d", level.AccessLevel))
		}
	}

	return strings.Join(parts, ", ")
}