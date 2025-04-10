package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func RegisterProjectTools(s *server.MCPServer) {
	listProjectsTool := mcp.NewTool("gitlab_list_projects",
		mcp.WithDescription("List GitLab projects"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("gitlab group ID")),
		mcp.WithString("search", mcp.Description("Multiple terms can be provided, separated by an escaped space, either + or %20, and will be ANDed together. Example: one+two will match substrings one and two (in any order).")),
	)

	projectTool := mcp.NewTool("gitlab_get_project",
		mcp.WithDescription("Get GitLab project details"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
	)

	s.AddTool(listProjectsTool, util.ErrorGuard(listProjectsHandler))
	s.AddTool(projectTool, util.ErrorGuard(getProjectHandler))
}

func listProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupID := request.Params.Arguments["group_id"].(string)

	opt := &gitlab.ListGroupProjectsOptions{
		Archived: gitlab.Ptr(false),
		OrderBy:  gitlab.Ptr("last_activity_at"),
		Sort:     gitlab.Ptr("desc"),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	if search, ok := request.Params.Arguments["search"]; ok {
		opt.Search = gitlab.Ptr(search.(string))
	}

	projects, _, err := util.GitlabClient().Groups.ListGroupProjects(groupID, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %v", err)
	}

	var result string
	for _, project := range projects {
		result += fmt.Sprintf("ID: %d\nName: %s\nPath: %s\nDescription: %s\nLast Activity: %s\n\n",
			project.ID, project.Name, project.PathWithNamespace, project.Description, project.LastActivityAt.Format("2006-01-02 15:04:05"))
	}

	return mcp.NewToolResultText(result), nil
}

func getProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.Params.Arguments["project_path"].(string)

	// Get project details
	project, _, err := util.GitlabClient().Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	// Get branches
	branches, _, err := util.GitlabClient().Branches.ListBranches(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %v", err)
	}

	// Get tags
	tags, _, err := util.GitlabClient().Tags.ListTags(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %v", err)
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