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

func RegisterGroupTools(s *server.MCPServer) {
	listGroupUsersTool := mcp.NewTool("list_group_users",
		mcp.WithDescription("List all users in a GitLab group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("GitLab group ID")),
	)
	s.AddTool(listGroupUsersTool, util.ErrorGuard(listGroupUsersHandler))
}

func listGroupUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupID := request.Params.Arguments["group_id"].(string)

	opt := &gitlab.ListGroupMembersOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	members, _, err := util.GitlabClient().Groups.ListGroupMembers(groupID, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Users in group %s:\n\n", groupID))

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
	case gitlab.OwnerPermission:
		return "Owner"
	default:
		return fmt.Sprintf("Unknown (%d)", level)
	}
} 