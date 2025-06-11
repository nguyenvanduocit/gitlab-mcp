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

type ListGroupUsersArgs struct {
	GroupID string `json:"group_id"`
}

type ListGroupsArgs struct {
	Search     string `json:"search"`
	Owned      bool   `json:"owned"`
	MinAccess  string `json:"min_access_level"`
}

func RegisterGroupTools(s *server.MCPServer) {
	listGroupUsersTool := mcp.NewTool("list_group_users",
		mcp.WithDescription("List all users in a GitLab group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("GitLab group ID")),
	)
	s.AddTool(listGroupUsersTool, mcp.NewTypedToolHandler(listGroupUsersHandler))

	listGroupsTool := mcp.NewTool("list_groups",
		mcp.WithDescription("List GitLab groups accessible to the user"),
		mcp.WithString("search", mcp.Description("Search for groups by name or path")),
		mcp.WithBoolean("owned", mcp.Description("List only groups owned by the authenticated user")),
		mcp.WithString("min_access_level", mcp.Description("Minimum access level (guest, reporter, developer, maintainer, owner)")),
	)
	s.AddTool(listGroupsTool, mcp.NewTypedToolHandler(listGroupsHandler))
}

func listGroupUsersHandler(ctx context.Context, request mcp.CallToolRequest, args ListGroupUsersArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListGroupMembersOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	members, _, err := util.GitlabClient().Groups.ListGroupMembers(args.GroupID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list group members: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Users in group %s:\n\n", args.GroupID))

	for _, member := range members {
		result.WriteString(fmt.Sprintf("User: %s\n", member.Username))
		result.WriteString(fmt.Sprintf("Name: %s\n", member.Name))
		result.WriteString(fmt.Sprintf("ID: %d\n", member.ID))
		result.WriteString(fmt.Sprintf("State: %s\n", member.State))
		result.WriteString(fmt.Sprintf("Access Level: %s\n", getAccessLevelString(member.AccessLevel)))
		if member.ExpiresAt != nil {
			result.WriteString(fmt.Sprintf("Expires At: %s\n", member.ExpiresAt.String()))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

// Helper function to convert access level to string
func getAccessLevelString(level gitlab.AccessLevelValue) string {
	switch level {
	case gitlab.GuestPermissions:
		return "Guest"
	case gitlab.ReporterPermissions:
		return "Reporter"
	case gitlab.DeveloperPermissions:
		return "Developer"
	case gitlab.MaintainerPermissions:
		return "Maintainer"
	case gitlab.OwnerPermissions:
		return "Owner"
	default:
		return fmt.Sprintf("Unknown (%d)", level)
	}
}

func listGroupsHandler(ctx context.Context, request mcp.CallToolRequest, args ListGroupsArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
		OrderBy: gitlab.Ptr("name"),
		Sort:    gitlab.Ptr("asc"),
	}

	// Apply search filter if provided
	if args.Search != "" {
		opt.Search = gitlab.Ptr(args.Search)
	}

	// Apply owned filter if provided
	if args.Owned {
		opt.Owned = gitlab.Ptr(true)
	}

	// Apply minimum access level filter if provided
	if args.MinAccess != "" {
		switch strings.ToLower(args.MinAccess) {
		case "guest":
			opt.MinAccessLevel = gitlab.Ptr(gitlab.GuestPermissions)
		case "reporter":
			opt.MinAccessLevel = gitlab.Ptr(gitlab.ReporterPermissions)
		case "developer":
			opt.MinAccessLevel = gitlab.Ptr(gitlab.DeveloperPermissions)
		case "maintainer":
			opt.MinAccessLevel = gitlab.Ptr(gitlab.MaintainerPermissions)
		case "owner":
			opt.MinAccessLevel = gitlab.Ptr(gitlab.OwnerPermissions)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("invalid min_access_level: %s. Valid values: guest, reporter, developer, maintainer, owner", args.MinAccess)), nil
		}
	}

	groups, _, err := util.GitlabClient().Groups.ListGroups(opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list groups: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString("GitLab Groups:\n\n")

	for _, group := range groups {
		result.WriteString(fmt.Sprintf("Group: %s\n", group.Name))
		result.WriteString(fmt.Sprintf("Path: %s\n", group.Path))
		result.WriteString(fmt.Sprintf("Full Path: %s\n", group.FullPath))
		result.WriteString(fmt.Sprintf("ID: %d\n", group.ID))
		result.WriteString(fmt.Sprintf("Visibility: %s\n", group.Visibility))
		result.WriteString(fmt.Sprintf("Web URL: %s\n", group.WebURL))
		
		if group.Description != "" {
			result.WriteString(fmt.Sprintf("Description: %s\n", group.Description))
		}
		
		if group.AvatarURL != "" {
			result.WriteString(fmt.Sprintf("Avatar: %s\n", group.AvatarURL))
		}
		
		result.WriteString(fmt.Sprintf("Created: %s\n", group.CreatedAt.Format("2006-01-02 15:04:05")))
		
		// Show parent group if available
		if group.ParentID != 0 {
			result.WriteString(fmt.Sprintf("Parent ID: %d\n", group.ParentID))
		}
		
		// Show statistics if available
		if group.Statistics != nil {
			result.WriteString(fmt.Sprintf("Repository Size: %d bytes\n", group.Statistics.RepositorySize))
		}
		
		result.WriteString("\n")
	}

	if len(groups) == 0 {
		result.WriteString("No groups found matching the criteria.\n")
	}

	return mcp.NewToolResultText(result.String()), nil
} 