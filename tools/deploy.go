package tools

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Deploy Tokens related structs
type ListProjectDeployTokensArgs struct {
	ProjectPath string `json:"project_path"`
}

type GetProjectDeployTokenArgs struct {
	ProjectPath     string `json:"project_path"`
	DeployTokenID   string `json:"deploy_token_id"`
}

type CreateProjectDeployTokenArgs struct {
	ProjectPath string   `json:"project_path"`
	Name        string   `json:"name"`
	ExpiresAt   string   `json:"expires_at,omitempty"` // ISO 8601 format
	Username    string   `json:"username,omitempty"`
	Scopes      []string `json:"scopes"`
}

type DeleteProjectDeployTokenArgs struct {
	ProjectPath   string `json:"project_path"`
	DeployTokenID string `json:"deploy_token_id"`
}

type ListGroupDeployTokensArgs struct {
	GroupID string `json:"group_id"`
}

type GetGroupDeployTokenArgs struct {
	GroupID       string `json:"group_id"`
	DeployTokenID string `json:"deploy_token_id"`
}

type CreateGroupDeployTokenArgs struct {
	GroupID   string   `json:"group_id"`
	Name      string   `json:"name"`
	ExpiresAt string   `json:"expires_at,omitempty"` // ISO 8601 format
	Username  string   `json:"username,omitempty"`
	Scopes    []string `json:"scopes"`
}

type DeleteGroupDeployTokenArgs struct {
	GroupID       string `json:"group_id"`
	DeployTokenID string `json:"deploy_token_id"`
}

func RegisterDeploymentTools(s *server.MCPServer) {
	// Deploy Tokens Tools
	listAllDeployTokensTool := mcp.NewTool("list_all_deploy_tokens",
		mcp.WithDescription("List all deploy tokens (requires administrator access)"),
	)

	listProjectDeployTokensTool := mcp.NewTool("list_project_deploy_tokens",
		mcp.WithDescription("List deploy tokens for a specific project"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project path (e.g., 'group/project')")),
	)

	getProjectDeployTokenTool := mcp.NewTool("get_project_deploy_token",
		mcp.WithDescription("Get details of a specific project deploy token"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project path (e.g., 'group/project')")),
		mcp.WithString("deploy_token_id", mcp.Required(), mcp.Description("Deploy token ID")),
	)

	createProjectDeployTokenTool := mcp.NewTool("create_project_deploy_token",
		mcp.WithDescription("Create a new deploy token for a project"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project path (e.g., 'group/project')")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name for the deploy token")),
		mcp.WithString("expires_at", mcp.Description("Expiration date in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)")),
		mcp.WithString("username", mcp.Description("Username for the deploy token")),
		mcp.WithArray("scopes", mcp.Required(), mcp.Description("Array of scopes (read_repository, read_registry, write_registry, read_package_registry, write_package_registry)")),
	)

	deleteProjectDeployTokenTool := mcp.NewTool("delete_project_deploy_token",
		mcp.WithDescription("Delete a deploy token from a project"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project path (e.g., 'group/project')")),
		mcp.WithString("deploy_token_id", mcp.Required(), mcp.Description("Deploy token ID to delete")),
	)

	listGroupDeployTokensTool := mcp.NewTool("list_group_deploy_tokens",
		mcp.WithDescription("List deploy tokens for a specific group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("Group ID or path")),
	)

	getGroupDeployTokenTool := mcp.NewTool("get_group_deploy_token",
		mcp.WithDescription("Get details of a specific group deploy token"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("Group ID or path")),
		mcp.WithString("deploy_token_id", mcp.Required(), mcp.Description("Deploy token ID")),
	)

	createGroupDeployTokenTool := mcp.NewTool("create_group_deploy_token",
		mcp.WithDescription("Create a new deploy token for a group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("Group ID or path")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name for the deploy token")),
		mcp.WithString("expires_at", mcp.Description("Expiration date in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)")),
		mcp.WithString("username", mcp.Description("Username for the deploy token")),
		mcp.WithArray("scopes", mcp.Required(), mcp.Description("Array of scopes (read_repository, read_registry, write_registry, read_package_registry, write_package_registry)")),
	)

	deleteGroupDeployTokenTool := mcp.NewTool("delete_group_deploy_token",
		mcp.WithDescription("Delete a deploy token from a group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("Group ID or path")),
		mcp.WithString("deploy_token_id", mcp.Required(), mcp.Description("Deploy token ID to delete")),
	)

	// Register Deploy Tokens handlers
	s.AddTool(listAllDeployTokensTool, mcp.NewTypedToolHandler(listAllDeployTokensHandler))
	s.AddTool(listProjectDeployTokensTool, mcp.NewTypedToolHandler(listProjectDeployTokensHandler))
	s.AddTool(getProjectDeployTokenTool, mcp.NewTypedToolHandler(getProjectDeployTokenHandler))
	s.AddTool(createProjectDeployTokenTool, mcp.NewTypedToolHandler(createProjectDeployTokenHandler))
	s.AddTool(deleteProjectDeployTokenTool, mcp.NewTypedToolHandler(deleteProjectDeployTokenHandler))
	s.AddTool(listGroupDeployTokensTool, mcp.NewTypedToolHandler(listGroupDeployTokensHandler))
	s.AddTool(getGroupDeployTokenTool, mcp.NewTypedToolHandler(getGroupDeployTokenHandler))
	s.AddTool(createGroupDeployTokenTool, mcp.NewTypedToolHandler(createGroupDeployTokenHandler))
	s.AddTool(deleteGroupDeployTokenTool, mcp.NewTypedToolHandler(deleteGroupDeployTokenHandler))
}

// Deploy Tokens Handlers

func listAllDeployTokensHandler(ctx context.Context, request mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, error) {
	tokens, _, err := util.GitlabClient().DeployTokens.ListAllDeployTokens()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list deploy tokens: %v", err)), nil
	}

	var result string
	result += fmt.Sprintf("Found %d deploy tokens:\n\n", len(tokens))
	
	for _, token := range tokens {
		result += fmt.Sprintf("ID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
			token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
		
		if token.ExpiresAt != nil {
			result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

func listProjectDeployTokensHandler(ctx context.Context, request mcp.CallToolRequest, args ListProjectDeployTokensArgs) (*mcp.CallToolResult, error) {
	tokens, _, err := util.GitlabClient().DeployTokens.ListProjectDeployTokens(args.ProjectPath, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list project deploy tokens: %v", err)), nil
	}

	var result string
	result += fmt.Sprintf("Deploy tokens for project '%s' (%d tokens):\n\n", args.ProjectPath, len(tokens))
	
	for _, token := range tokens {
		result += fmt.Sprintf("ID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
			token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
		
		if token.ExpiresAt != nil {
			result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

func getProjectDeployTokenHandler(ctx context.Context, request mcp.CallToolRequest, args GetProjectDeployTokenArgs) (*mcp.CallToolResult, error) {
	deployTokenID, err := strconv.Atoi(args.DeployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid deploy token ID: %v", err)), nil
	}

	token, _, err := util.GitlabClient().DeployTokens.GetProjectDeployToken(args.ProjectPath, deployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get deploy token: %v", err)), nil
	}

	result := fmt.Sprintf("Deploy Token Details:\n\nID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
		token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
	
	if token.ExpiresAt != nil {
		result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	return mcp.NewToolResultText(result), nil
}

func createProjectDeployTokenHandler(ctx context.Context, request mcp.CallToolRequest, args CreateProjectDeployTokenArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.CreateProjectDeployTokenOptions{
		Name:   gitlab.Ptr(args.Name),
		Scopes: &args.Scopes,
	}

	if args.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, args.ExpiresAt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid expires_at format: %v", err)), nil
		}
		opt.ExpiresAt = &expiresAt
	}

	if args.Username != "" {
		opt.Username = gitlab.Ptr(args.Username)
	}

	token, _, err := util.GitlabClient().DeployTokens.CreateProjectDeployToken(args.ProjectPath, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create deploy token: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Deploy token created successfully for project '%s'!\n\nID: %d\nName: %s\nUsername: %s\nToken: %s\nScopes: %v\n",
		args.ProjectPath, token.ID, token.Name, token.Username, token.Token, token.Scopes)
	
	if token.ExpiresAt != nil {
		result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	
	result += "\n⚠️  Important: Save the token value now. You won't be able to access it again!"

	return mcp.NewToolResultText(result), nil
}

func deleteProjectDeployTokenHandler(ctx context.Context, request mcp.CallToolRequest, args DeleteProjectDeployTokenArgs) (*mcp.CallToolResult, error) {
	deployTokenID, err := strconv.Atoi(args.DeployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid deploy token ID: %v", err)), nil
	}

	_, err = util.GitlabClient().DeployTokens.DeleteProjectDeployToken(args.ProjectPath, deployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete deploy token: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Deploy token %s deleted successfully from project '%s'", args.DeployTokenID, args.ProjectPath)
	return mcp.NewToolResultText(result), nil
}

func listGroupDeployTokensHandler(ctx context.Context, request mcp.CallToolRequest, args ListGroupDeployTokensArgs) (*mcp.CallToolResult, error) {
	tokens, _, err := util.GitlabClient().DeployTokens.ListGroupDeployTokens(args.GroupID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list group deploy tokens: %v", err)), nil
	}

	var result string
	result += fmt.Sprintf("Deploy tokens for group '%s' (%d tokens):\n\n", args.GroupID, len(tokens))
	
	for _, token := range tokens {
		result += fmt.Sprintf("ID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
			token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
		
		if token.ExpiresAt != nil {
			result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

func getGroupDeployTokenHandler(ctx context.Context, request mcp.CallToolRequest, args GetGroupDeployTokenArgs) (*mcp.CallToolResult, error) {
	deployTokenID, err := strconv.Atoi(args.DeployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid deploy token ID: %v", err)), nil
	}

	token, _, err := util.GitlabClient().DeployTokens.GetGroupDeployToken(args.GroupID, deployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get group deploy token: %v", err)), nil
	}

	result := fmt.Sprintf("Group Deploy Token Details:\n\nID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
		token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
	
	if token.ExpiresAt != nil {
		result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	return mcp.NewToolResultText(result), nil
}

func createGroupDeployTokenHandler(ctx context.Context, request mcp.CallToolRequest, args CreateGroupDeployTokenArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.CreateGroupDeployTokenOptions{
		Name:   gitlab.Ptr(args.Name),
		Scopes: &args.Scopes,
	}

	if args.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, args.ExpiresAt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid expires_at format: %v", err)), nil
		}
		opt.ExpiresAt = &expiresAt
	}

	if args.Username != "" {
		opt.Username = gitlab.Ptr(args.Username)
	}

	token, _, err := util.GitlabClient().DeployTokens.CreateGroupDeployToken(args.GroupID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create group deploy token: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Deploy token created successfully for group '%s'!\n\nID: %d\nName: %s\nUsername: %s\nToken: %s\nScopes: %v\n",
		args.GroupID, token.ID, token.Name, token.Username, token.Token, token.Scopes)
	
	if token.ExpiresAt != nil {
		result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	
	result += "\n⚠️  Important: Save the token value now. You won't be able to access it again!"

	return mcp.NewToolResultText(result), nil
}

func deleteGroupDeployTokenHandler(ctx context.Context, request mcp.CallToolRequest, args DeleteGroupDeployTokenArgs) (*mcp.CallToolResult, error) {
	deployTokenID, err := strconv.Atoi(args.DeployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid deploy token ID: %v", err)), nil
	}

	_, err = util.GitlabClient().DeployTokens.DeleteGroupDeployToken(args.GroupID, deployTokenID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete group deploy token: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Deploy token %s deleted successfully from group '%s'", args.DeployTokenID, args.GroupID)
	return mcp.NewToolResultText(result), nil
}
