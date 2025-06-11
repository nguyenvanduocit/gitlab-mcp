package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type ListProjectsArgs struct {
	GroupID string `json:"group_id"`
	Search  string `json:"search"`
}

type GetProjectArgs struct {
	ProjectPath string `json:"project_path"`
}

func RegisterProjectTools(s *server.MCPServer) {
	listProjectsTool := mcp.NewTool("list_projects",
		mcp.WithDescription("List GitLab projects"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("gitlab group ID")),
		mcp.WithString("search", mcp.Description("Multiple terms can be provided, separated by an escaped space, either + or %20, and will be ANDed together. Example: one+two will match substrings one and two (in any order).")),
	)

	projectTool := mcp.NewTool("get_project",
		mcp.WithDescription("Get GitLab project details"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
	)

	s.AddTool(listProjectsTool, mcp.NewTypedToolHandler(listProjectsHandler))
	s.AddTool(projectTool, mcp.NewTypedToolHandler(getProjectHandler))
}

func listProjectsHandler(ctx context.Context, request mcp.CallToolRequest, args ListProjectsArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListGroupProjectsOptions{
		Archived: gitlab.Ptr(false),
		OrderBy:  gitlab.Ptr("last_activity_at"),
		Sort:     gitlab.Ptr("desc"),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	if args.Search != "" {
		opt.Search = gitlab.Ptr(args.Search)
	}

	projects, _, err := util.GitlabClient().Groups.ListGroupProjects(args.GroupID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to search projects: %v", err)), nil
	}

	var result string
	for _, project := range projects {
		result += fmt.Sprintf("ID: %d\nName: %s\nPath: %s\nDescription: %s\nLast Activity: %s\n\n",
			project.ID, project.Name, project.PathWithNamespace, project.Description, project.LastActivityAt.Format("2006-01-02 15:04:05"))
	}

	return mcp.NewToolResultText(result), nil
}

func getProjectHandler(ctx context.Context, request mcp.CallToolRequest, args GetProjectArgs) (*mcp.CallToolResult, error) {
	// Get project details
	project, _, err := util.GitlabClient().Projects.GetProject(args.ProjectPath, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %v", err)), nil
	}

	// Get branches
	branches, _, err := util.GitlabClient().Branches.ListBranches(args.ProjectPath, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list branches: %v", err)), nil
	}

	// Get tags
	tags, _, err := util.GitlabClient().Tags.ListTags(args.ProjectPath, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list tags: %v", err)), nil
	}

	// Build basic project info
	result := fmt.Sprintf("Project Details:\nID: %d\nName: %s\nPath: %s\nDescription: %s\nURL: %s\nDefault Branch: %s\n\n",
		project.ID, project.Name, project.PathWithNamespace, project.Description, project.WebURL,
		project.DefaultBranch)

	// Add branches
	result += "Branches:\n"
	for _, branch := range branches {
		result += fmt.Sprintf("- %s\n", branch.Name)
	}

	// Add tags
	result += "\nTags:\n"
	for _, tag := range tags {
		result += fmt.Sprintf("- %s\n", tag.Name)
	}

	return mcp.NewToolResultText(result), nil
} 