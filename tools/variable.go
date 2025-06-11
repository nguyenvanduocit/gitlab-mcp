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

type ListGroupVariablesArgs struct {
	GroupID string `json:"group_id"`
}

type GetGroupVariableArgs struct {
	GroupID string `json:"group_id"`
	Key     string `json:"key"`
}

type CreateGroupVariableArgs struct {
	GroupID           string `json:"group_id"`
	Key               string `json:"key"`
	Value             string `json:"value"`
	VariableType      string `json:"variable_type,omitempty"`      // env_var or file
	Protected         bool   `json:"protected,omitempty"`
	Masked            bool   `json:"masked,omitempty"`
	Raw               bool   `json:"raw,omitempty"`
	EnvironmentScope  string `json:"environment_scope,omitempty"`
	Description       string `json:"description,omitempty"`
}

type UpdateGroupVariableArgs struct {
	GroupID           string `json:"group_id"`
	Key               string `json:"key"`
	Value             string `json:"value,omitempty"`
	VariableType      string `json:"variable_type,omitempty"`      // env_var or file
	Protected         *bool  `json:"protected,omitempty"`
	Masked            *bool  `json:"masked,omitempty"`
	Raw               *bool  `json:"raw,omitempty"`
	EnvironmentScope  string `json:"environment_scope,omitempty"`
	Description       string `json:"description,omitempty"`
}

type RemoveGroupVariableArgs struct {
	GroupID string `json:"group_id"`
	Key     string `json:"key"`
}

func RegisterVariableTools(s *server.MCPServer) {
	// List group variables
	listVariablesTool := mcp.NewTool("list_group_variables",
		mcp.WithDescription("List all variables in a GitLab group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("GitLab group ID or path")),
	)
	s.AddTool(listVariablesTool, mcp.NewTypedToolHandler(listGroupVariablesHandler))

	// Get specific group variable
	getVariableTool := mcp.NewTool("get_group_variable",
		mcp.WithDescription("Get a specific variable from a GitLab group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("GitLab group ID or path")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Variable key name")),
	)
	s.AddTool(getVariableTool, mcp.NewTypedToolHandler(getGroupVariableHandler))

	// Create group variable
	createVariableTool := mcp.NewTool("create_group_variable",
		mcp.WithDescription("Create a new variable in a GitLab group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("GitLab group ID or path")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Variable key name")),
		mcp.WithString("value", mcp.Required(), mcp.Description("Variable value")),
		mcp.WithString("variable_type", mcp.Description("Variable type: env_var (default) or file")),
		mcp.WithBoolean("protected", mcp.Description("Whether the variable is protected (default: false)")),
		mcp.WithBoolean("masked", mcp.Description("Whether the variable is masked (default: false)")),
		mcp.WithBoolean("raw", mcp.Description("Whether the variable is raw (default: false)")),
		mcp.WithString("environment_scope", mcp.Description("Environment scope (default: *)")),
		mcp.WithString("description", mcp.Description("Variable description")),
	)
	s.AddTool(createVariableTool, mcp.NewTypedToolHandler(createGroupVariableHandler))

	// Update group variable
	updateVariableTool := mcp.NewTool("update_group_variable",
		mcp.WithDescription("Update an existing variable in a GitLab group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("GitLab group ID or path")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Variable key name")),
		mcp.WithString("value", mcp.Description("New variable value")),
		mcp.WithString("variable_type", mcp.Description("Variable type: env_var or file")),
		mcp.WithBoolean("protected", mcp.Description("Whether the variable is protected")),
		mcp.WithBoolean("masked", mcp.Description("Whether the variable is masked")),
		mcp.WithBoolean("raw", mcp.Description("Whether the variable is raw")),
		mcp.WithString("environment_scope", mcp.Description("Environment scope")),
		mcp.WithString("description", mcp.Description("Variable description")),
	)
	s.AddTool(updateVariableTool, mcp.NewTypedToolHandler(updateGroupVariableHandler))

	// Remove group variable
	removeVariableTool := mcp.NewTool("remove_group_variable",
		mcp.WithDescription("Remove a variable from a GitLab group"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("GitLab group ID or path")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Variable key name to remove")),
	)
	s.AddTool(removeVariableTool, mcp.NewTypedToolHandler(removeGroupVariableHandler))
}

func listGroupVariablesHandler(ctx context.Context, request mcp.CallToolRequest, args ListGroupVariablesArgs) (*mcp.CallToolResult, error) {
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

func getGroupVariableHandler(ctx context.Context, request mcp.CallToolRequest, args GetGroupVariableArgs) (*mcp.CallToolResult, error) {
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

func createGroupVariableHandler(ctx context.Context, request mcp.CallToolRequest, args CreateGroupVariableArgs) (*mcp.CallToolResult, error) {
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
	if args.Protected {
		opt.Protected = gitlab.Ptr(args.Protected)
	}
	if args.Masked {
		opt.Masked = gitlab.Ptr(args.Masked)
	}
	if args.Raw {
		opt.Raw = gitlab.Ptr(args.Raw)
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

func updateGroupVariableHandler(ctx context.Context, request mcp.CallToolRequest, args UpdateGroupVariableArgs) (*mcp.CallToolResult, error) {
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

func removeGroupVariableHandler(ctx context.Context, request mcp.CallToolRequest, args RemoveGroupVariableArgs) (*mcp.CallToolResult, error) {
	_, err := util.GitlabClient().GroupVariables.RemoveVariable(args.GroupID, args.Key, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to remove group variable: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Successfully removed variable '%s' from group %s", args.Key, args.GroupID)
	return mcp.NewToolResultText(result), nil
}
