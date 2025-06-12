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

// Complex typed structures for deploy tokens
type ListAllDeployTokensArgs struct {
	RandomString string `json:"random_string" validate:"required"` // Dummy parameter for no-parameter tools
}

// Nested structures for complex typed tools
type DeployTokenScope struct {
	Type        string `json:"type" validate:"required,oneof=project group"`        // project or group
	ProjectPath string `json:"project_path,omitempty" validate:"omitempty,min=1,max=255"` // Required for project scope
	GroupID     string `json:"group_id,omitempty" validate:"omitempty,min=1,max=255"`     // Required for group scope
}

type DeployTokenCreateOptions struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`                    // Token name
	Username  string   `json:"username,omitempty" validate:"omitempty,min=1,max=100"`    // Optional username
	ExpiresAt string   `json:"expires_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"` // Optional expiration
	Scopes    []string `json:"scopes" validate:"required,dive,oneof=read_repository read_registry write_registry read_package_registry write_package_registry"` // Required scopes
}

type DeployTokenIdentifier struct {
	ID string `json:"id" validate:"required,numeric"` // Deploy token ID
}

type ManageDeployTokensArgs struct {
	Action     string                     `json:"action" validate:"required,oneof=list get create delete"` // Action to perform
	Scope      DeployTokenScope          `json:"scope"`                                                    // Scope configuration
	TokenID    *DeployTokenIdentifier    `json:"token_id,omitempty"`                                      // For get/delete actions
	CreateOpts *DeployTokenCreateOptions `json:"create_options,omitempty"`                               // For create action
}

func RegisterDeploymentTools(s *server.MCPServer) {
	// List all deploy tokens (admin only)
	listAllDeployTokensTool := mcp.NewTool("list_all_deploy_tokens",
		mcp.WithDescription("List all deploy tokens (requires administrator access)"),
		mcp.WithString("random_string", 
			mcp.Required(), 
			mcp.Description("Dummy parameter for no-parameter tools")),
	)

	// Complex typed deploy tokens management tool
	manageDeployTokensTool := mcp.NewTool("manage_deploy_tokens",
		mcp.WithDescription("Manage deploy tokens for projects or groups. Supports list, get, create, and delete operations."),
		mcp.WithString("action", 
			mcp.Required(), 
			mcp.Description("Action to perform: list, get, create, delete")),
		mcp.WithObject("scope",
			mcp.Required(),
			mcp.Description("Scope configuration for the deploy token operation"),
			mcp.Properties(map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "Scope type: project or group",
					"enum":        []string{"project", "group"},
				},
				"project_path": map[string]any{
					"type":        "string",
					"description": "Project path (required for project scope, e.g., 'group/project')",
					"minLength":   1,
					"maxLength":   255,
				},
				"group_id": map[string]any{
					"type":        "string",
					"description": "Group ID or path (required for group scope)",
					"minLength":   1,
					"maxLength":   255,
				},
			})),
		mcp.WithObject("token_id",
			mcp.Description("Deploy token identifier (required for get/delete actions)"),
			mcp.Properties(map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "Deploy token ID",
					"pattern":     "^[0-9]+$",
				},
			})),
		mcp.WithObject("create_options",
			mcp.Description("Options for creating a new deploy token (required for create action)"),
			mcp.Properties(map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name for the deploy token",
					"minLength":   1,
					"maxLength":   100,
				},
				"username": map[string]any{
					"type":        "string",
					"description": "Username for the deploy token (optional)",
					"minLength":   1,
					"maxLength":   100,
				},
				"expires_at": map[string]any{
					"type":        "string",
					"description": "Expiration date in ISO 8601 format (optional)",
					"format":      "date-time",
				},
				"scopes": map[string]any{
					"type":        "array",
					"description": "Array of scopes for the deploy token",
					"items": map[string]any{
						"type": "string",
						"enum": []string{
							"read_repository",
							"read_registry", 
							"write_registry",
							"read_package_registry",
							"write_package_registry",
						},
					},
					"minItems": 1,
				},
			})),
	)

	// Register handlers
	s.AddTool(listAllDeployTokensTool, mcp.NewTypedToolHandler(listAllDeployTokensHandler))
	s.AddTool(manageDeployTokensTool, mcp.NewTypedToolHandler(manageDeployTokensHandler))
}

// Handlers

func listAllDeployTokensHandler(ctx context.Context, request mcp.CallToolRequest, args ListAllDeployTokensArgs) (*mcp.CallToolResult, error) {
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

func manageDeployTokensHandler(ctx context.Context, request mcp.CallToolRequest, args ManageDeployTokensArgs) (*mcp.CallToolResult, error) {
	// Validate scope configuration
	if args.Scope.Type != "project" && args.Scope.Type != "group" {
		return mcp.NewToolResultError("scope.type must be either 'project' or 'group'"), nil
	}

	// Validate scope-specific parameters
	if args.Scope.Type == "project" && args.Scope.ProjectPath == "" {
		return mcp.NewToolResultError("scope.project_path is required for project scope"), nil
	}
	if args.Scope.Type == "group" && args.Scope.GroupID == "" {
		return mcp.NewToolResultError("scope.group_id is required for group scope"), nil
	}

	// Validate action-specific parameters
	if (args.Action == "get" || args.Action == "delete") && args.TokenID == nil {
		return mcp.NewToolResultError("token_id is required for get/delete actions"), nil
	}
	if args.Action == "create" {
		if args.CreateOpts == nil {
			return mcp.NewToolResultError("create_options is required for create action"), nil
		}
		if args.CreateOpts.Name == "" {
			return mcp.NewToolResultError("create_options.name is required for create action"), nil
		}
		if len(args.CreateOpts.Scopes) == 0 {
			return mcp.NewToolResultError("create_options.scopes is required for create action"), nil
		}
	}

	// Route to appropriate handler based on action
	switch args.Action {
	case "list":
		return handleListDeployTokens(args)
	case "get":
		return handleGetDeployToken(args)
	case "create":
		return handleCreateDeployToken(args)
	case "delete":
		return handleDeleteDeployToken(args)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported action: %s", args.Action)), nil
	}
}

func handleListDeployTokens(args ManageDeployTokensArgs) (*mcp.CallToolResult, error) {
	var result string
	
	if args.Scope.Type == "project" {
		tokens, _, err := util.GitlabClient().DeployTokens.ListProjectDeployTokens(args.Scope.ProjectPath, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list project deploy tokens: %v", err)), nil
		}
		
		result += fmt.Sprintf("Deploy tokens for project '%s' (%d tokens):\n\n", args.Scope.ProjectPath, len(tokens))
		
		for _, token := range tokens {
			result += fmt.Sprintf("ID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
				token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
			
			if token.ExpiresAt != nil {
				result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
			}
			
			result += "\n"
		}
	} else { // group
		tokens, _, err := util.GitlabClient().DeployTokens.ListGroupDeployTokens(args.Scope.GroupID, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list group deploy tokens: %v", err)), nil
		}
		
		result += fmt.Sprintf("Deploy tokens for group '%s' (%d tokens):\n\n", args.Scope.GroupID, len(tokens))
		
		for _, token := range tokens {
			result += fmt.Sprintf("ID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
				token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
			
			if token.ExpiresAt != nil {
				result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
			}
			
			result += "\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

func handleGetDeployToken(args ManageDeployTokensArgs) (*mcp.CallToolResult, error) {
	deployTokenID, err := strconv.Atoi(args.TokenID.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid deploy token ID: %v", err)), nil
	}

	var result string
	
	if args.Scope.Type == "project" {
		token, _, err := util.GitlabClient().DeployTokens.GetProjectDeployToken(args.Scope.ProjectPath, deployTokenID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get project deploy token: %v", err)), nil
		}
		
		result = fmt.Sprintf("Project Deploy Token Details:\n\nID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
			token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
		
		if token.ExpiresAt != nil {
			result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	} else { // group
		token, _, err := util.GitlabClient().DeployTokens.GetGroupDeployToken(args.Scope.GroupID, deployTokenID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get group deploy token: %v", err)), nil
		}
		
		result = fmt.Sprintf("Group Deploy Token Details:\n\nID: %d\nName: %s\nUsername: %s\nRevoked: %t\nExpired: %t\nScopes: %v\n",
			token.ID, token.Name, token.Username, token.Revoked, token.Expired, token.Scopes)
		
		if token.ExpiresAt != nil {
			result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	}

	return mcp.NewToolResultText(result), nil
}

func handleCreateDeployToken(args ManageDeployTokensArgs) (*mcp.CallToolResult, error) {
	var result string
	
	if args.Scope.Type == "project" {
		opt := &gitlab.CreateProjectDeployTokenOptions{
			Name:   gitlab.Ptr(args.CreateOpts.Name),
			Scopes: &args.CreateOpts.Scopes,
		}

		if args.CreateOpts.ExpiresAt != "" {
			expiresAt, err := time.Parse(time.RFC3339, args.CreateOpts.ExpiresAt)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid expires_at format: %v", err)), nil
			}
			opt.ExpiresAt = &expiresAt
		}

		if args.CreateOpts.Username != "" {
			opt.Username = gitlab.Ptr(args.CreateOpts.Username)
		}

		token, _, err := util.GitlabClient().DeployTokens.CreateProjectDeployToken(args.Scope.ProjectPath, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create project deploy token: %v", err)), nil
		}

		result = fmt.Sprintf("✅ Deploy token created successfully for project '%s'!\n\nID: %d\nName: %s\nUsername: %s\nToken: %s\nScopes: %v\n",
			args.Scope.ProjectPath, token.ID, token.Name, token.Username, token.Token, token.Scopes)
		
		if token.ExpiresAt != nil {
			result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	} else { // group
		opt := &gitlab.CreateGroupDeployTokenOptions{
			Name:   gitlab.Ptr(args.CreateOpts.Name),
			Scopes: &args.CreateOpts.Scopes,
		}

		if args.CreateOpts.ExpiresAt != "" {
			expiresAt, err := time.Parse(time.RFC3339, args.CreateOpts.ExpiresAt)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid expires_at format: %v", err)), nil
			}
			opt.ExpiresAt = &expiresAt
		}

		if args.CreateOpts.Username != "" {
			opt.Username = gitlab.Ptr(args.CreateOpts.Username)
		}

		token, _, err := util.GitlabClient().DeployTokens.CreateGroupDeployToken(args.Scope.GroupID, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create group deploy token: %v", err)), nil
		}

		result = fmt.Sprintf("✅ Deploy token created successfully for group '%s'!\n\nID: %d\nName: %s\nUsername: %s\nToken: %s\nScopes: %v\n",
			args.Scope.GroupID, token.ID, token.Name, token.Username, token.Token, token.Scopes)
		
		if token.ExpiresAt != nil {
			result += fmt.Sprintf("Expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	}
	
	result += "\n⚠️  Important: Save the token value now. You won't be able to access it again!"
	return mcp.NewToolResultText(result), nil
}

func handleDeleteDeployToken(args ManageDeployTokensArgs) (*mcp.CallToolResult, error) {
	deployTokenID, err := strconv.Atoi(args.TokenID.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid deploy token ID: %v", err)), nil
	}

	var result string
	
	if args.Scope.Type == "project" {
		_, err = util.GitlabClient().DeployTokens.DeleteProjectDeployToken(args.Scope.ProjectPath, deployTokenID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to delete project deploy token: %v", err)), nil
		}
		
		result = fmt.Sprintf("✅ Deploy token %s deleted successfully from project '%s'", args.TokenID.ID, args.Scope.ProjectPath)
	} else { // group
		_, err = util.GitlabClient().DeployTokens.DeleteGroupDeployToken(args.Scope.GroupID, deployTokenID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to delete group deploy token: %v", err)), nil
		}
		
		result = fmt.Sprintf("✅ Deploy token %s deleted successfully from group '%s'", args.TokenID.ID, args.Scope.GroupID)
	}

	return mcp.NewToolResultText(result), nil
}
