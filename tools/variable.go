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
	Confirmed         bool              `json:"confirmed,omitempty"`
}

// ProjectVariableArgs defines the consolidated arguments for all project variable operations
type ProjectVariableArgs struct {
	Action            string            `json:"action" validate:"required,oneof=list get create update remove"`
	ProjectID         string            `json:"project_id" validate:"required"`
	Key               string            `json:"key" validate:"required_unless=Action list"`
	Value             string            `json:"value" validate:"required_if=Action create"`
	VariableType      string            `json:"variable_type" validate:"omitempty,oneof=env_var file"`
	Protected         *bool             `json:"protected"`
	Masked            *bool             `json:"masked"`
	Raw               *bool             `json:"raw"`
	EnvironmentScope  string            `json:"environment_scope"`
	Description       string            `json:"description"`
	Confirmed         bool              `json:"confirmed,omitempty"`
}

// getAncestorGroups returns all ancestor groups of a project, starting from immediate parent
func getAncestorGroups(projectID string) ([]*gitlab.Group, error) {
	project, _, err := util.GitlabClient().Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	var ancestors []*gitlab.Group
	
	if project.Namespace != nil && project.Namespace.Kind == "group" {
		// Get the immediate parent group
		group, _, err := util.GitlabClient().Groups.GetGroup(project.Namespace.ID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get group: %v", err)
		}
		ancestors = append(ancestors, group)
		
		// Get all ancestor groups
		currentGroup := group
		for currentGroup.ParentID != 0 {
			parentGroup, _, err := util.GitlabClient().Groups.GetGroup(currentGroup.ParentID, nil)
			if err != nil {
				break // Stop if we can't fetch the parent
			}
			ancestors = append(ancestors, parentGroup)
			currentGroup = parentGroup
		}
	}
	
	return ancestors, nil
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
		mcp.WithBoolean("confirmed", 
			mcp.Description("Confirmation required for create, update, and remove actions")),
	)
	s.AddTool(groupVariableTool, mcp.NewTypedToolHandler(groupVariableHandler))

	// Consolidated project variable tool
	projectVariableTool := mcp.NewTool("manage_project_variable",
		mcp.WithDescription("Manage GitLab project variables with different actions: list, get, create, update, remove. Shows inheritance from group variables and detailed variable properties."),
		mcp.WithString("action", 
			mcp.Required(), 
			mcp.Description("Action to perform: list, get, create, update, remove")),
		mcp.WithString("project_id", 
			mcp.Required(), 
			mcp.Description("GitLab project ID or path")),
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
		mcp.WithBoolean("confirmed", 
			mcp.Description("Confirmation required for create, update, and remove actions")),
	)
	s.AddTool(projectVariableTool, mcp.NewTypedToolHandler(projectVariableHandler))
}

func groupVariableHandler(ctx context.Context, request mcp.CallToolRequest, args GroupVariableArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "list":
		return listGroupVariables(args)
	case "get":
		return getGroupVariable(args)
	case "create":
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with creating a group variable."), nil
		}
		return createGroupVariable(args)
	case "update":
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with updating a group variable."), nil
		}
		return updateGroupVariable(args)
	case "remove":
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with removing a group variable."), nil
		}
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
		
		// Show the actual value
		if variable.Value != "" {
			result.WriteString(fmt.Sprintf("Value: %s\n", variable.Value))
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
	
	// Show the actual value
	if variable.Value != "" {
		result.WriteString(fmt.Sprintf("Value: %s\n", variable.Value))
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

func projectVariableHandler(ctx context.Context, request mcp.CallToolRequest, args ProjectVariableArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "list":
		return listProjectVariables(args)
	case "get":
		return getProjectVariable(args)
	case "create":
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with creating a project variable."), nil
		}
		return createProjectVariable(args)
	case "update":
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with updating a project variable."), nil
		}
		return updateProjectVariable(args)
	case "remove":
		if !args.Confirmed {
			return mcp.NewToolResultError("This operation requires confirmation. Please set 'confirmed: true' to proceed with removing a project variable."), nil
		}
		return removeProjectVariable(args)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action: %s. Valid actions are: list, get, create, update, remove", args.Action)), nil
	}
}

func listProjectVariables(args ProjectVariableArgs) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListProjectVariablesOptions{}

	variables, _, err := util.GitlabClient().ProjectVariables.ListVariables(args.ProjectID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list project variables: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Variables in project %s:\n\n", args.ProjectID))

	if len(variables) == 0 {
		result.WriteString("No variables found in this project.\n")
		return mcp.NewToolResultText(result.String()), nil
	}

	// Get project details to show inheritance information
	project, _, err := util.GitlabClient().Projects.GetProject(args.ProjectID, nil)
	if err == nil && project.Namespace != nil {
		result.WriteString(fmt.Sprintf("📁 Project: %s\n", project.Name))
		result.WriteString(fmt.Sprintf("🏢 Namespace: %s (ID: %d)\n\n", project.Namespace.Name, project.Namespace.ID))
	}

	for i, variable := range variables {
		result.WriteString(fmt.Sprintf("Variable %d:\n", i+1))
		result.WriteString(fmt.Sprintf("  Key: %s\n", variable.Key))
		result.WriteString(fmt.Sprintf("  Variable Type: %s\n", variable.VariableType))
		result.WriteString(fmt.Sprintf("  Protected: %t\n", variable.Protected))
		result.WriteString(fmt.Sprintf("  Masked: %t\n", variable.Masked))
		result.WriteString(fmt.Sprintf("  Raw: %t\n", variable.Raw))
		result.WriteString(fmt.Sprintf("  Environment Scope: %s\n", variable.EnvironmentScope))
		
		if variable.Description != "" {
			result.WriteString(fmt.Sprintf("  Description: %s\n", variable.Description))
		}
		
		// Show the actual value
		if variable.Value != "" {
			result.WriteString(fmt.Sprintf("  Value: %s\n", variable.Value))
		} else {
			result.WriteString("  Value: [EMPTY]\n")
		}
		
		result.WriteString("\n")
	}

	// Show inherited variables from all ancestor groups
	ancestors, ancestorErr := getAncestorGroups(args.ProjectID)
	if ancestorErr == nil && len(ancestors) > 0 {
		result.WriteString("🏢 Inherited Variables from Ancestor Groups:\n")
		
		for groupLevel, group := range ancestors {
			groupVariables, _, groupErr := util.GitlabClient().GroupVariables.ListVariables(fmt.Sprintf("%d", group.ID), &gitlab.ListGroupVariablesOptions{})
			if groupErr == nil && len(groupVariables) > 0 {
				// Show hierarchy level
				indentLevel := ""
				hierarchyLabel := "Parent"
				if groupLevel > 0 {
					indentLevel = strings.Repeat("  ", groupLevel)
					hierarchyLabel = fmt.Sprintf("Ancestor (Level %d)", groupLevel+1)
				}
				
				result.WriteString(fmt.Sprintf("%s📁 %s Group: %s (ID: %d)\n", indentLevel, hierarchyLabel, group.Name, group.ID))
				
				for i, groupVar := range groupVariables {
					// Check if this group variable is overridden by a project variable or higher-level group variable
					overridden := false
					overrideSource := ""
					
					// Check project variables first
					for _, projectVar := range variables {
						if projectVar.Key == groupVar.Key && projectVar.EnvironmentScope == groupVar.EnvironmentScope {
							overridden = true
							overrideSource = "PROJECT"
							break
						}
					}
					
					// Check higher-level groups (closer to project)
					if !overridden {
						for j := groupLevel - 1; j >= 0; j-- {
							higherGroupVars, _, err := util.GitlabClient().GroupVariables.ListVariables(fmt.Sprintf("%d", ancestors[j].ID), &gitlab.ListGroupVariablesOptions{})
							if err == nil {
								for _, higherVar := range higherGroupVars {
									if higherVar.Key == groupVar.Key && higherVar.EnvironmentScope == groupVar.EnvironmentScope {
										overridden = true
										overrideSource = fmt.Sprintf("GROUP (ID: %d)", ancestors[j].ID)
										break
									}
								}
							}
							if overridden {
								break
							}
						}
					}
					
					result.WriteString(fmt.Sprintf("%s  Variable %d:\n", indentLevel, i+1))
					result.WriteString(fmt.Sprintf("%s    Key: %s", indentLevel, groupVar.Key))
					if overridden {
						result.WriteString(fmt.Sprintf(" [OVERRIDDEN BY %s]", overrideSource))
					}
					result.WriteString("\n")
					result.WriteString(fmt.Sprintf("%s    Source: Group (ID: %d)\n", indentLevel, group.ID))
					result.WriteString(fmt.Sprintf("%s    Variable Type: %s\n", indentLevel, groupVar.VariableType))
					result.WriteString(fmt.Sprintf("%s    Protected: %t\n", indentLevel, groupVar.Protected))
					result.WriteString(fmt.Sprintf("%s    Masked: %t\n", indentLevel, groupVar.Masked))
					result.WriteString(fmt.Sprintf("%s    Raw: %t\n", indentLevel, groupVar.Raw))
					result.WriteString(fmt.Sprintf("%s    Environment Scope: %s\n", indentLevel, groupVar.EnvironmentScope))
					
					if groupVar.Description != "" {
						result.WriteString(fmt.Sprintf("%s    Description: %s\n", indentLevel, groupVar.Description))
					}
					
					if groupVar.Value != "" {
						result.WriteString(fmt.Sprintf("%s    Value: %s\n", indentLevel, groupVar.Value))
					} else {
						result.WriteString(fmt.Sprintf("%s    Value: [EMPTY]\n", indentLevel))
					}
					
					result.WriteString("\n")
				}
			} else if groupErr == nil {
				indentLevel := strings.Repeat("  ", groupLevel)
				hierarchyLabel := "Parent"
				if groupLevel > 0 {
					hierarchyLabel = fmt.Sprintf("Ancestor (Level %d)", groupLevel+1)
				}
				result.WriteString(fmt.Sprintf("%s📁 %s Group: %s (ID: %d) - No variables\n", indentLevel, hierarchyLabel, group.Name, group.ID))
			}
		}
		result.WriteString("\n")
	}

	result.WriteString("💡 Note: Project variables override group variables with the same key and environment scope.\n")
	return mcp.NewToolResultText(result.String()), nil
}

func getProjectVariable(args ProjectVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for get action"), nil
	}

	// Get the specific project variable
	variable, _, err := util.GitlabClient().ProjectVariables.GetVariable(args.ProjectID, args.Key, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project variable: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Variable details for key '%s' in project %s:\n\n", args.Key, args.ProjectID))
	
	// Get project details for inheritance context
	project, _, projectErr := util.GitlabClient().Projects.GetProject(args.ProjectID, nil)
	if projectErr == nil && project.Namespace != nil {
		result.WriteString(fmt.Sprintf("📁 Project: %s\n", project.Name))
		result.WriteString(fmt.Sprintf("🏢 Namespace: %s (ID: %d)\n\n", project.Namespace.Name, project.Namespace.ID))
	}

	result.WriteString("🔧 Variable Properties:\n")
	result.WriteString(fmt.Sprintf("  Key: %s\n", variable.Key))
	result.WriteString(fmt.Sprintf("  Variable Type: %s\n", variable.VariableType))
	result.WriteString(fmt.Sprintf("  Protected: %t\n", variable.Protected))
	result.WriteString(fmt.Sprintf("  Masked: %t\n", variable.Masked))
	result.WriteString(fmt.Sprintf("  Raw: %t\n", variable.Raw))
	result.WriteString(fmt.Sprintf("  Environment Scope: %s\n", variable.EnvironmentScope))
	
	if variable.Description != "" {
		result.WriteString(fmt.Sprintf("  Description: %s\n", variable.Description))
	}
	
	// Show the actual value
	if variable.Value != "" {
		result.WriteString(fmt.Sprintf("  Value: %s\n", variable.Value))
	} else {
		result.WriteString("  Value: [EMPTY]\n")
	}

	result.WriteString("\n")

	// Check for inheritance from all ancestor groups
	result.WriteString("🔍 Inheritance Information:\n")
	result.WriteString("  Source: Project-level variable\n")
	
	ancestors, ancestorErr := getAncestorGroups(args.ProjectID)
	if ancestorErr == nil && len(ancestors) > 0 {
		result.WriteString(fmt.Sprintf("  Hierarchy: Project → %s", ancestors[0].Name))
		for i := 1; i < len(ancestors); i++ {
			result.WriteString(fmt.Sprintf(" → %s", ancestors[i].Name))
		}
		result.WriteString("\n\n")
		
		// Check for variables with the same key in all ancestor groups
		foundConflicts := false
		for groupLevel, group := range ancestors {
			groupVariable, _, groupErr := util.GitlabClient().GroupVariables.GetVariable(fmt.Sprintf("%d", group.ID), args.Key, nil)
			if groupErr == nil {
				if !foundConflicts {
					result.WriteString("  ⚠️  Note: Group variables with the same key exist in ancestor groups.\n")
					result.WriteString("      Project variables override group variables. Closer groups override distant ones.\n\n")
					foundConflicts = true
				}
				
				indentLevel := strings.Repeat("  ", groupLevel+1)
				hierarchyLabel := "Parent"
				if groupLevel > 0 {
					hierarchyLabel = fmt.Sprintf("Ancestor (Level %d)", groupLevel+1)
				}
				
				result.WriteString(fmt.Sprintf("%s🏢 %s Group Variable Properties (%s - ID: %d):\n", indentLevel, hierarchyLabel, group.Name, group.ID))
				result.WriteString(fmt.Sprintf("%s    Key: %s\n", indentLevel, groupVariable.Key))
				result.WriteString(fmt.Sprintf("%s    Source: Group (ID: %d)\n", indentLevel, group.ID))
				result.WriteString(fmt.Sprintf("%s    Variable Type: %s\n", indentLevel, groupVariable.VariableType))
				result.WriteString(fmt.Sprintf("%s    Protected: %t\n", indentLevel, groupVariable.Protected))
				result.WriteString(fmt.Sprintf("%s    Masked: %t\n", indentLevel, groupVariable.Masked))
				result.WriteString(fmt.Sprintf("%s    Raw: %t\n", indentLevel, groupVariable.Raw))
				result.WriteString(fmt.Sprintf("%s    Environment Scope: %s\n", indentLevel, groupVariable.EnvironmentScope))
				if groupVariable.Description != "" {
					result.WriteString(fmt.Sprintf("%s    Description: %s\n", indentLevel, groupVariable.Description))
				}
				if groupVariable.Value != "" {
					result.WriteString(fmt.Sprintf("%s    Value: %s\n", indentLevel, groupVariable.Value))
				} else {
					result.WriteString(fmt.Sprintf("%s    Value: [EMPTY]\n", indentLevel))
				}
				result.WriteString("\n")
			}
		}
		
		if !foundConflicts {
			result.WriteString("  ✅ No conflicting group variables found in ancestor groups.\n")
		}
	} else {
		result.WriteString("  Parent: No parent group or personal namespace\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func createProjectVariable(args ProjectVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for create action"), nil
	}
	if args.Value == "" {
		return mcp.NewToolResultError("value is required for create action"), nil
	}

	opt := &gitlab.CreateProjectVariableOptions{
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

	variable, _, err := util.GitlabClient().ProjectVariables.CreateVariable(args.ProjectID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create project variable: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("✅ Successfully created variable '%s' in project %s\n\n", args.Key, args.ProjectID))
	
	result.WriteString("🔧 Variable Properties:\n")
	result.WriteString(fmt.Sprintf("  Key: %s\n", variable.Key))
	result.WriteString(fmt.Sprintf("  Variable Type: %s\n", variable.VariableType))
	result.WriteString(fmt.Sprintf("  Protected: %t\n", variable.Protected))
	result.WriteString(fmt.Sprintf("  Masked: %t\n", variable.Masked))
	result.WriteString(fmt.Sprintf("  Raw: %t\n", variable.Raw))
	result.WriteString(fmt.Sprintf("  Environment Scope: %s\n", variable.EnvironmentScope))
	
	if variable.Description != "" {
		result.WriteString(fmt.Sprintf("  Description: %s\n", variable.Description))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func updateProjectVariable(args ProjectVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for update action"), nil
	}

	opt := &gitlab.UpdateProjectVariableOptions{}

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

	variable, _, err := util.GitlabClient().ProjectVariables.UpdateVariable(args.ProjectID, args.Key, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update project variable: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("✅ Successfully updated variable '%s' in project %s\n\n", args.Key, args.ProjectID))
	
	result.WriteString("🔧 Updated Variable Properties:\n")
	result.WriteString(fmt.Sprintf("  Key: %s\n", variable.Key))
	result.WriteString(fmt.Sprintf("  Variable Type: %s\n", variable.VariableType))
	result.WriteString(fmt.Sprintf("  Protected: %t\n", variable.Protected))
	result.WriteString(fmt.Sprintf("  Masked: %t\n", variable.Masked))
	result.WriteString(fmt.Sprintf("  Raw: %t\n", variable.Raw))
	result.WriteString(fmt.Sprintf("  Environment Scope: %s\n", variable.EnvironmentScope))
	
	if variable.Description != "" {
		result.WriteString(fmt.Sprintf("  Description: %s\n", variable.Description))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func removeProjectVariable(args ProjectVariableArgs) (*mcp.CallToolResult, error) {
	if args.Key == "" {
		return mcp.NewToolResultError("key is required for remove action"), nil
	}

	_, err := util.GitlabClient().ProjectVariables.RemoveVariable(args.ProjectID, args.Key, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to remove project variable: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Successfully removed variable '%s' from project %s", args.Key, args.ProjectID)
	return mcp.NewToolResultText(result), nil
}
