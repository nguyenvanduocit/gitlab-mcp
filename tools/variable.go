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

// GroupVariableArgs defines the consolidated arguments for all group variable operations
type GroupVariableArgs struct {
	Action            string            `json:"action" validate:"required,oneof=list get create update remove"`
	GroupID           string            `json:"group_id" validate:"required"`
	Key               string            `json:"key" validate:"required_unless=Action list"`
	Value             string            `json:"value" validate:"required_if=Action create"`
	VariableType      string            `json:"variable_type" validate:"omitempty,oneof=env_var file"`
	Protected         *bool             `json:"protected"`
	Masked            *bool             `json:"masked"`
	Raw               *bool             `json:"raw"`
	EnvironmentScope  string            `json:"environment_scope"`
	Description       string            `json:"description"`
}

func RegisterVariableTools(s *server.MCPServer) {
	// Consolidated group variable tool
	groupVariableTool := mcp.NewTool("manage_group_variable",
		mcp.WithDescription("Manage GitLab group variables with different actions: list, get, create, update, remove"),
		mcp.WithString("action", 
			mcp.Required(), 
			mcp.Description("Action to perform: list, get, create, update, remove")),
		mcp.WithString("group_id", 
			mcp.Required(), 
			mcp.Description("GitLab group ID or path")),
		mcp.WithString("key", 
			mcp.Description("Variable key name (required for get, create, update, remove actions)")),
		mcp.WithString("value", 
			mcp.Description("Variable value (required for create action, optional for update)")),
		mcp.WithString("variable_type", 
			mcp.Description("Variable type: env_var (default) or file")),
		mcp.WithBoolean("protected", 
			mcp.Description("Whether the variable is protected")),
		mcp.WithBoolean("masked", 
			mcp.Description("Whether the variable is masked")),
		mcp.WithBoolean("raw", 
			mcp.Description("Whether the variable is raw")),
		mcp.WithString("environment_scope", 
			mcp.Description("Environment scope (default: *)")),
		mcp.WithString("description", 
			mcp.Description("Variable description")),
	)
	s.AddTool(groupVariableTool, mcp.NewTypedToolHandler(groupVariableHandler))
}

func groupVariableHandler(ctx context.Context, request mcp.CallToolRequest, args GroupVariableArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "list":
		return listGroupVariables(args)
	case "get":
		return getGroupVariable(args)
	case "create":
		return createGroupVariable(args)
	case "update":
		return updateGroupVariable(args)
	case "remove":
		return removeGroupVariable(args)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action: %s. Valid actions are: list, get, create, update, remove", args.Action)), nil
	}
}

func listGroupVariables(args GroupVariableArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListGroupVariablesOptions{}

	variables, _, err := util.GitlabClient().GroupVariables.ListVariables(args.GroupID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list group variables: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Variables in group %s:\n\n", args.GroupID))

	if len(variables) == 0 {
		result.WriteString("No variables found in this group.\n")
		return mcp.NewToolResultText(result.String()), nil
	}

	for _, variable := range variables {
		result.WriteString(fmt.Sprintf("Key: %s\n", variable.Key))
		result.WriteString(fmt.Sprintf("Variable Type: %s\n", variable.VariableType))
		result.WriteString(fmt.Sprintf("Protected: %t\n", variable.Protected))
		result.WriteString(fmt.Sprintf("Masked: %t\n", variable.Masked))
		result.WriteString(fmt.Sprintf("Raw: %t\n", variable.Raw))
		result.WriteString(fmt.Sprintf("Environment Scope: %s\n", variable.EnvironmentScope))
		
		if variable.Description != "" {
			result.WriteString(fmt.Sprintf("Description: %s\n", variable.Description))
		}
		
		// Don't show the actual value for security reasons, just indicate if it exists
		if variable.Value != "" {
			result.WriteString("Value: [HIDDEN]\n")
		} else {
			result.WriteString("Value: [EMPTY]\n")
		}
		
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getGroupVariable(args GroupVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for get action"), nil
	}

	variable, _, err := util.GitlabClient().GroupVariables.GetVariable(args.GroupID, args.Key, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get group variable: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Variable details for key '%s' in group %s:\n\n", args.Key, args.GroupID))
	result.WriteString(fmt.Sprintf("Key: %s\n", variable.Key))
	result.WriteString(fmt.Sprintf("Variable Type: %s\n", variable.VariableType))
	result.WriteString(fmt.Sprintf("Protected: %t\n", variable.Protected))
	result.WriteString(fmt.Sprintf("Masked: %t\n", variable.Masked))
	result.WriteString(fmt.Sprintf("Raw: %t\n", variable.Raw))
	result.WriteString(fmt.Sprintf("Environment Scope: %s\n", variable.EnvironmentScope))
	
	if variable.Description != "" {
		result.WriteString(fmt.Sprintf("Description: %s\n", variable.Description))
	}
	
	// For security, only show if value exists but not the actual value
	if variable.Value != "" {
		result.WriteString("Value: [HIDDEN - Use with caution]\n")
	} else {
		result.WriteString("Value: [EMPTY]\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func createGroupVariable(args GroupVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for create action"), nil
	}
	if args.Value == "" {
		return mcp.NewToolResultError("value is required for create action"), nil
	}

	opt := &gitlab.CreateGroupVariableOptions{
		Key:   gitlab.Ptr(args.Key),
		Value: gitlab.Ptr(args.Value),
	}

	// Set variable type (default to env_var)
	if args.VariableType != "" {
		if args.VariableType == "env_var" || args.VariableType == "file" {
			opt.VariableType = gitlab.Ptr(gitlab.VariableTypeValue(args.VariableType))
		} else {
			return mcp.NewToolResultError("variable_type must be either 'env_var' or 'file'"), nil
		}
	}

	// Set optional parameters
	if args.Protected != nil {
		opt.Protected = args.Protected
	}
	if args.Masked != nil {
		opt.Masked = args.Masked
	}
	if args.Raw != nil {
		opt.Raw = args.Raw
	}
	if args.EnvironmentScope != "" {
		opt.EnvironmentScope = gitlab.Ptr(args.EnvironmentScope)
	}
	if args.Description != "" {
		opt.Description = gitlab.Ptr(args.Description)
	}

	variable, _, err := util.GitlabClient().GroupVariables.CreateVariable(args.GroupID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create group variable: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("✅ Successfully created variable '%s' in group %s\n\n", args.Key, args.GroupID))
	result.WriteString(fmt.Sprintf("Key: %s\n", variable.Key))
	result.WriteString(fmt.Sprintf("Variable Type: %s\n", variable.VariableType))
	result.WriteString(fmt.Sprintf("Protected: %t\n", variable.Protected))
	result.WriteString(fmt.Sprintf("Masked: %t\n", variable.Masked))
	result.WriteString(fmt.Sprintf("Raw: %t\n", variable.Raw))
	result.WriteString(fmt.Sprintf("Environment Scope: %s\n", variable.EnvironmentScope))
	
	if variable.Description != "" {
		result.WriteString(fmt.Sprintf("Description: %s\n", variable.Description))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func updateGroupVariable(args GroupVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for update action"), nil
	}

	opt := &gitlab.UpdateGroupVariableOptions{}

	// Only set fields that were provided
	if args.Value != "" {
		opt.Value = gitlab.Ptr(args.Value)
	}
	if args.VariableType != "" {
		if args.VariableType == "env_var" || args.VariableType == "file" {
			opt.VariableType = gitlab.Ptr(gitlab.VariableTypeValue(args.VariableType))
		} else {
			return mcp.NewToolResultError("variable_type must be either 'env_var' or 'file'"), nil
		}
	}
	if args.Protected != nil {
		opt.Protected = args.Protected
	}
	if args.Masked != nil {
		opt.Masked = args.Masked
	}
	if args.Raw != nil {
		opt.Raw = args.Raw
	}
	if args.EnvironmentScope != "" {
		opt.EnvironmentScope = gitlab.Ptr(args.EnvironmentScope)
	}
	if args.Description != "" {
		opt.Description = gitlab.Ptr(args.Description)
	}

	variable, _, err := util.GitlabClient().GroupVariables.UpdateVariable(args.GroupID, args.Key, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update group variable: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("✅ Successfully updated variable '%s' in group %s\n\n", args.Key, args.GroupID))
	result.WriteString(fmt.Sprintf("Key: %s\n", variable.Key))
	result.WriteString(fmt.Sprintf("Variable Type: %s\n", variable.VariableType))
	result.WriteString(fmt.Sprintf("Protected: %t\n", variable.Protected))
	result.WriteString(fmt.Sprintf("Masked: %t\n", variable.Masked))
	result.WriteString(fmt.Sprintf("Raw: %t\n", variable.Raw))
	result.WriteString(fmt.Sprintf("Environment Scope: %s\n", variable.EnvironmentScope))
	
	if variable.Description != "" {
		result.WriteString(fmt.Sprintf("Description: %s\n", variable.Description))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func removeGroupVariable(args GroupVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for remove action"), nil
	}

	_, err := util.GitlabClient().GroupVariables.RemoveVariable(args.GroupID, args.Key, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to remove group variable: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Successfully removed variable '%s' from group %s", args.Key, args.GroupID)
	return mcp.NewToolResultText(result), nil
}
