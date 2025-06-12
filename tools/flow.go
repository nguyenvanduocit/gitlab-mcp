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

// Unified Git Flow tool argument structures
type GitFlowCreateBranchArgs struct {
	Action      string `json:"action" validate:"required,oneof=create_release create_feature create_hotfix"`
	ProjectPath string `json:"project_path" validate:"required,min=1,max=200"`
	
	// Branch creation options
	CreateOptions struct {
		// For release branches
		ReleaseVersion string `json:"release_version" validate:"required_if=Action create_release,min=1,max=50"`
		
		// For feature branches  
		FeatureName string `json:"feature_name" validate:"required_if=Action create_feature,min=1,max=100"`
		
		// For hotfix branches
		HotfixVersion string `json:"hotfix_version" validate:"required_if=Action create_hotfix,min=1,max=50"`
		
		// Common branch options
		BaseBranch        string `json:"base_branch" validate:"max=100"`
		DevelopmentBranch string `json:"development_branch" validate:"max=100"`
		ProductionBranch  string `json:"production_branch" validate:"max=100"`
	} `json:"create_options"`
}

type GitFlowFinishBranchArgs struct {
	Action      string `json:"action" validate:"required,oneof=finish_release finish_feature finish_hotfix"`
	ProjectPath string `json:"project_path" validate:"required,min=1,max=200"`
	
	// Branch finishing options
	FinishOptions struct {
		// For release branches
		ReleaseVersion string `json:"release_version" validate:"required_if=Action finish_release,min=1,max=50"`
		
		// For feature branches
		FeatureName  string `json:"feature_name" validate:"required_if=Action finish_feature,min=1,max=100"`
		TargetBranch string `json:"target_branch" validate:"max=100"`
		
		// For hotfix branches
		HotfixVersion string `json:"hotfix_version" validate:"required_if=Action finish_hotfix,min=1,max=50"`
		
		// Common finish options
		DeleteBranch      bool   `json:"delete_branch"`
		DevelopmentBranch string `json:"development_branch" validate:"max=100"`
		ProductionBranch  string `json:"production_branch" validate:"max=100"`
	} `json:"finish_options"`
}

type GitFlowListBranchesArgs struct {
	ProjectPath string `json:"project_path" validate:"required,min=1,max=200"`
	BranchType  string `json:"branch_type" validate:"oneof=all feature release hotfix"`
}

// RegisterFlowTools registers all Git Flow related tools
func RegisterFlowTools(s *server.MCPServer) {
	// Unified branch creation tool
	createBranchTool := mcp.NewTool("gitflow_create_branch",
		mcp.WithDescription("Create a new Git Flow branch (release, feature, or hotfix)"),
		mcp.WithString("action", 
			mcp.Required(), 
			mcp.Description("Action to perform: create_release, create_feature, create_hotfix")),
		mcp.WithString("project_path", 
			mcp.Required(), 
			mcp.Description("Project/repo path")),
		mcp.WithObject("create_options",
			mcp.Description("Branch creation options"),
			mcp.Properties(map[string]any{
				"release_version": map[string]any{
					"type":        "string",
					"description": "Release version (e.g., 1.2.0) - required for create_release",
				},
				"feature_name": map[string]any{
					"type":        "string", 
					"description": "Feature name (e.g., user-authentication) - required for create_feature",
				},
				"hotfix_version": map[string]any{
					"type":        "string",
					"description": "Hotfix version (e.g., 1.2.1) - required for create_hotfix",
				},
				"base_branch": map[string]any{
					"type":        "string",
					"description": "Base branch to create from (defaults: develop for release/feature, master for hotfix)",
				},
				"development_branch": map[string]any{
					"type":        "string",
					"description": "Development branch name (default: develop)",
				},
				"production_branch": map[string]any{
					"type":        "string", 
					"description": "Production branch name (default: master)",
				},
			}),
		),
	)

	// Unified branch finishing tool
	finishBranchTool := mcp.NewTool("gitflow_finish_branch",
		mcp.WithDescription("Finish a Git Flow branch by creating merge requests"),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("Action to perform: finish_release, finish_feature, finish_hotfix")),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Project/repo path")),
		mcp.WithObject("finish_options",
			mcp.Description("Branch finishing options"),
			mcp.Properties(map[string]any{
				"release_version": map[string]any{
					"type":        "string",
					"description": "Release version - required for finish_release",
				},
				"feature_name": map[string]any{
					"type":        "string",
					"description": "Feature name - required for finish_feature",
				},
				"hotfix_version": map[string]any{
					"type":        "string", 
					"description": "Hotfix version - required for finish_hotfix",
				},
				"target_branch": map[string]any{
					"type":        "string",
					"description": "Target branch for feature MR (default: develop)",
				},
				"delete_branch": map[string]any{
					"type":        "boolean",
					"description": "Delete branch after creating MRs",
				},
				"development_branch": map[string]any{
					"type":        "string",
					"description": "Development branch name (default: develop)",
				},
				"production_branch": map[string]any{
					"type":        "string",
					"description": "Production branch name (default: master)",
				},
			}),
		),
	)

	// List branches tool (keeping as is since it's already unified)
	listFlowBranchesTool := mcp.NewTool("gitflow_list_branches",
		mcp.WithDescription("List Git Flow branches (feature, release, hotfix)"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("branch_type", mcp.DefaultString("all"), mcp.Description("Branch type to list (feature, release, hotfix, all)")),
	)

	// Register all tools
	s.AddTool(createBranchTool, mcp.NewTypedToolHandler(gitFlowCreateBranchHandler))
	s.AddTool(finishBranchTool, mcp.NewTypedToolHandler(gitFlowFinishBranchHandler))
	s.AddTool(listFlowBranchesTool, mcp.NewTypedToolHandler(listFlowBranchesHandler))
}

// Unified branch creation handler
func gitFlowCreateBranchHandler(ctx context.Context, request mcp.CallToolRequest, args GitFlowCreateBranchArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "create_release":
		return createReleaseBranch(args)
	case "create_feature":
		return createFeatureBranch(args)
	case "create_hotfix":
		return createHotfixBranch(args)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported action: %s", args.Action)), nil
	}
}

// Unified branch finishing handler
func gitFlowFinishBranchHandler(ctx context.Context, request mcp.CallToolRequest, args GitFlowFinishBranchArgs) (*mcp.CallToolResult, error) {
	switch args.Action {
	case "finish_release":
		return finishReleaseBranch(args)
	case "finish_feature":
		return finishFeatureBranch(args)
	case "finish_hotfix":
		return finishHotfixBranch(args)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported action: %s", args.Action)), nil
	}
}

// Release branch implementation
func createReleaseBranch(args GitFlowCreateBranchArgs) (*mcp.CallToolResult, error) {
	baseBranch := args.CreateOptions.BaseBranch
	if baseBranch == "" {
		developmentBranch := args.CreateOptions.DevelopmentBranch
		if developmentBranch == "" {
			developmentBranch = "develop"
		}
		baseBranch = developmentBranch
	}

	releaseBranch := fmt.Sprintf("release/%s", args.CreateOptions.ReleaseVersion)

	// Check if release branch already exists
	branches, _, err := util.GitlabClient().Branches.ListBranches(args.ProjectPath, &gitlab.ListBranchesOptions{
		Search: gitlab.Ptr(releaseBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to check existing branches: %v", err)), nil
	}

	for _, branch := range branches {
		if branch.Name == releaseBranch {
			return mcp.NewToolResultError(fmt.Sprintf("release branch '%s' already exists", releaseBranch)), nil
		}
	}

	// Create the release branch
	branch, _, err := util.GitlabClient().Branches.CreateBranch(args.ProjectPath, &gitlab.CreateBranchOptions{
		Branch: gitlab.Ptr(releaseBranch),
		Ref:    gitlab.Ptr(baseBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create release branch: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("✅ Release branch created successfully!\n\n"))
	result.WriteString(fmt.Sprintf("Branch: %s\n", branch.Name))
	result.WriteString(fmt.Sprintf("Based on: %s\n", baseBranch))
	result.WriteString(fmt.Sprintf("Commit: %s\n", branch.Commit.ID))
	result.WriteString(fmt.Sprintf("Author: %s\n", branch.Commit.AuthorName))
	result.WriteString(fmt.Sprintf("Message: %s\n\n", branch.Commit.Message))
	
	result.WriteString("🔄 Next steps:\n")
	result.WriteString("1. Make your release changes on this branch\n")
	result.WriteString("2. Test thoroughly\n")
	result.WriteString(fmt.Sprintf("3. Use 'gitflow_finish_branch' with action 'finish_release' and version '%s' to create MRs\n", args.CreateOptions.ReleaseVersion))

	return mcp.NewToolResultText(result.String()), nil
}

func finishReleaseBranch(args GitFlowFinishBranchArgs) (*mcp.CallToolResult, error) {
	releaseBranch := fmt.Sprintf("release/%s", args.FinishOptions.ReleaseVersion)
	
	// Get branch names with defaults
	developmentBranch := args.FinishOptions.DevelopmentBranch
	if developmentBranch == "" {
		developmentBranch = "develop"
	}
	
	productionBranch := args.FinishOptions.ProductionBranch
	if productionBranch == "" {
		productionBranch = "master"
	}
	
	// Verify release branch exists
	_, _, err := util.GitlabClient().Branches.GetBranch(args.ProjectPath, releaseBranch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("release branch '%s' not found: %v", releaseBranch, err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🚀 Finishing release %s\n\n", args.FinishOptions.ReleaseVersion))

	// Create MR to development branch
	developMR, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Release %s", args.FinishOptions.ReleaseVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Release %s ready for merge to %s\n\n- [ ] Code review completed\n- [ ] Tests passing\n- [ ] Documentation updated", args.FinishOptions.ReleaseVersion, developmentBranch)),
		SourceBranch: gitlab.Ptr(releaseBranch),
		TargetBranch: gitlab.Ptr(developmentBranch),
	})
	if err != nil {
		result.WriteString(fmt.Sprintf("❌ Failed to create MR to %s: %v\n", developmentBranch, err))
	} else {
		result.WriteString(fmt.Sprintf("✅ Created MR to %s: !%d\n", developmentBranch, developMR.IID))
		result.WriteString(fmt.Sprintf("   URL: %s\n", developMR.WebURL))
	}

	// Create MR to production branch
	masterMR, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Release %s", args.FinishOptions.ReleaseVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Release %s ready for production\n\n- [ ] Release notes prepared\n- [ ] Deployment plan reviewed\n- [ ] Rollback plan confirmed", args.FinishOptions.ReleaseVersion)),
		SourceBranch: gitlab.Ptr(releaseBranch),
		TargetBranch: gitlab.Ptr(productionBranch),
	})
	if err != nil {
		result.WriteString(fmt.Sprintf("❌ Failed to create MR to %s: %v\n", productionBranch, err))
	} else {
		result.WriteString(fmt.Sprintf("✅ Created MR to %s: !%d\n", productionBranch, masterMR.IID))
		result.WriteString(fmt.Sprintf("   URL: %s\n", masterMR.WebURL))
	}

	// Delete branch if requested
	if args.FinishOptions.DeleteBranch {
		_, err := util.GitlabClient().Branches.DeleteBranch(args.ProjectPath, releaseBranch)
		if err != nil {
			result.WriteString(fmt.Sprintf("⚠️  Failed to delete release branch: %v\n", err))
		} else {
			result.WriteString(fmt.Sprintf("🗑️  Deleted release branch: %s\n", releaseBranch))
		}
	}

	result.WriteString(fmt.Sprintf("\n📋 Release %s is ready for review and merge!\n", args.FinishOptions.ReleaseVersion))

	return mcp.NewToolResultText(result.String()), nil
}

// Feature branch implementation
func createFeatureBranch(args GitFlowCreateBranchArgs) (*mcp.CallToolResult, error) {
	baseBranch := args.CreateOptions.BaseBranch
	if baseBranch == "" {
		developmentBranch := args.CreateOptions.DevelopmentBranch
		if developmentBranch == "" {
			developmentBranch = "develop"
		}
		baseBranch = developmentBranch
	}

	featureBranch := fmt.Sprintf("feature/%s", args.CreateOptions.FeatureName)

	// Check if feature branch already exists
	branches, _, err := util.GitlabClient().Branches.ListBranches(args.ProjectPath, &gitlab.ListBranchesOptions{
		Search: gitlab.Ptr(featureBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to check existing branches: %v", err)), nil
	}

	for _, branch := range branches {
		if branch.Name == featureBranch {
			return mcp.NewToolResultError(fmt.Sprintf("feature branch '%s' already exists", featureBranch)), nil
		}
	}

	// Create the feature branch
	branch, _, err := util.GitlabClient().Branches.CreateBranch(args.ProjectPath, &gitlab.CreateBranchOptions{
		Branch: gitlab.Ptr(featureBranch),
		Ref:    gitlab.Ptr(baseBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create feature branch: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("✅ Feature branch created successfully!\n\n"))
	result.WriteString(fmt.Sprintf("Branch: %s\n", branch.Name))
	result.WriteString(fmt.Sprintf("Based on: %s\n", baseBranch))
	result.WriteString(fmt.Sprintf("Commit: %s\n", branch.Commit.ID))
	result.WriteString(fmt.Sprintf("Author: %s\n", branch.Commit.AuthorName))
	result.WriteString(fmt.Sprintf("Message: %s\n\n", branch.Commit.Message))
	
	result.WriteString("🔄 Next steps:\n")
	result.WriteString("1. Implement your feature on this branch\n")
	result.WriteString("2. Commit your changes regularly\n")
	result.WriteString(fmt.Sprintf("3. Use 'gitflow_finish_branch' with action 'finish_feature' and name '%s' to create MR\n", args.CreateOptions.FeatureName))

	return mcp.NewToolResultText(result.String()), nil
}

func finishFeatureBranch(args GitFlowFinishBranchArgs) (*mcp.CallToolResult, error) {
	featureBranch := fmt.Sprintf("feature/%s", args.FinishOptions.FeatureName)
	targetBranch := args.FinishOptions.TargetBranch
	if targetBranch == "" {
		developmentBranch := args.FinishOptions.DevelopmentBranch
		if developmentBranch == "" {
			developmentBranch = "develop"
		}
		targetBranch = developmentBranch
	}
	
	// Verify feature branch exists
	_, _, err := util.GitlabClient().Branches.GetBranch(args.ProjectPath, featureBranch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("feature branch '%s' not found: %v", featureBranch, err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🚀 Finishing feature %s\n\n", args.FinishOptions.FeatureName))

	// Create MR to target branch (usually develop)
	mr, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Feature: %s", args.FinishOptions.FeatureName)),
		Description:  gitlab.Ptr(fmt.Sprintf("Feature implementation: %s\n\n- [ ] Code review completed\n- [ ] Tests added/updated\n- [ ] Documentation updated\n- [ ] Ready for merge", args.FinishOptions.FeatureName)),
		SourceBranch: gitlab.Ptr(featureBranch),
		TargetBranch: gitlab.Ptr(targetBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create MR: %v", err)), nil
	}

	result.WriteString(fmt.Sprintf("✅ Created MR to %s: !%d\n", targetBranch, mr.IID))
	result.WriteString(fmt.Sprintf("   URL: %s\n", mr.WebURL))

	// Delete branch if requested
	if args.FinishOptions.DeleteBranch {
		_, err := util.GitlabClient().Branches.DeleteBranch(args.ProjectPath, featureBranch)
		if err != nil {
			result.WriteString(fmt.Sprintf("⚠️  Failed to delete feature branch: %v\n", err))
		} else {
			result.WriteString(fmt.Sprintf("🗑️  Deleted feature branch: %s\n", featureBranch))
		}
	}

	result.WriteString(fmt.Sprintf("\n📋 Feature %s is ready for review!\n", args.FinishOptions.FeatureName))

	return mcp.NewToolResultText(result.String()), nil
}

// Hotfix branch implementation
func createHotfixBranch(args GitFlowCreateBranchArgs) (*mcp.CallToolResult, error) {
	baseBranch := args.CreateOptions.BaseBranch
	if baseBranch == "" {
		productionBranch := args.CreateOptions.ProductionBranch
		if productionBranch == "" {
			productionBranch = "master"
		}
		baseBranch = productionBranch
	}

	hotfixBranch := fmt.Sprintf("hotfix/%s", args.CreateOptions.HotfixVersion)

	// Check if hotfix branch already exists
	branches, _, err := util.GitlabClient().Branches.ListBranches(args.ProjectPath, &gitlab.ListBranchesOptions{
		Search: gitlab.Ptr(hotfixBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to check existing branches: %v", err)), nil
	}

	for _, branch := range branches {
		if branch.Name == hotfixBranch {
			return mcp.NewToolResultError(fmt.Sprintf("hotfix branch '%s' already exists", hotfixBranch)), nil
		}
	}

	// Create the hotfix branch
	branch, _, err := util.GitlabClient().Branches.CreateBranch(args.ProjectPath, &gitlab.CreateBranchOptions{
		Branch: gitlab.Ptr(hotfixBranch),
		Ref:    gitlab.Ptr(baseBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create hotfix branch: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🚨 Hotfix branch created successfully!\n\n"))
	result.WriteString(fmt.Sprintf("Branch: %s\n", branch.Name))
	result.WriteString(fmt.Sprintf("Based on: %s\n", baseBranch))
	result.WriteString(fmt.Sprintf("Commit: %s\n", branch.Commit.ID))
	result.WriteString(fmt.Sprintf("Author: %s\n", branch.Commit.AuthorName))
	result.WriteString(fmt.Sprintf("Message: %s\n\n", branch.Commit.Message))
	
	result.WriteString("🔄 Next steps:\n")
	result.WriteString("1. Fix the critical issue on this branch\n")
	result.WriteString("2. Test the fix thoroughly\n")
	result.WriteString(fmt.Sprintf("3. Use 'gitflow_finish_branch' with action 'finish_hotfix' and version '%s' to create MRs\n", args.CreateOptions.HotfixVersion))

	return mcp.NewToolResultText(result.String()), nil
}

func finishHotfixBranch(args GitFlowFinishBranchArgs) (*mcp.CallToolResult, error) {
	hotfixBranch := fmt.Sprintf("hotfix/%s", args.FinishOptions.HotfixVersion)
	
	// Get branch names with defaults
	developmentBranch := args.FinishOptions.DevelopmentBranch
	if developmentBranch == "" {
		developmentBranch = "develop"
	}
	
	productionBranch := args.FinishOptions.ProductionBranch
	if productionBranch == "" {
		productionBranch = "master"
	}
	
	// Verify hotfix branch exists
	_, _, err := util.GitlabClient().Branches.GetBranch(args.ProjectPath, hotfixBranch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hotfix branch '%s' not found: %v", hotfixBranch, err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🚨 Finishing hotfix %s\n\n", args.FinishOptions.HotfixVersion))

	// Create MR to production branch
	masterMR, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Hotfix %s", args.FinishOptions.HotfixVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Critical hotfix %s\n\n- [ ] Fix verified\n- [ ] Tests passing\n- [ ] Ready for immediate deployment", args.FinishOptions.HotfixVersion)),
		SourceBranch: gitlab.Ptr(hotfixBranch),
		TargetBranch: gitlab.Ptr(productionBranch),
	})
	if err != nil {
		result.WriteString(fmt.Sprintf("❌ Failed to create MR to %s: %v\n", productionBranch, err))
	} else {
		result.WriteString(fmt.Sprintf("✅ Created MR to %s: !%d\n", productionBranch, masterMR.IID))
		result.WriteString(fmt.Sprintf("   URL: %s\n", masterMR.WebURL))
	}

	// Create MR to development branch
	developMR, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Hotfix %s", args.FinishOptions.HotfixVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Hotfix %s merge to %s\n\n- [ ] Conflicts resolved\n- [ ] Tests updated if needed", args.FinishOptions.HotfixVersion, developmentBranch)),
		SourceBranch: gitlab.Ptr(hotfixBranch),
		TargetBranch: gitlab.Ptr(developmentBranch),
	})
	if err != nil {
		result.WriteString(fmt.Sprintf("❌ Failed to create MR to %s: %v\n", developmentBranch, err))
	} else {
		result.WriteString(fmt.Sprintf("✅ Created MR to %s: !%d\n", developmentBranch, developMR.IID))
		result.WriteString(fmt.Sprintf("   URL: %s\n", developMR.WebURL))
	}

	// Delete branch if requested
	if args.FinishOptions.DeleteBranch {
		_, err := util.GitlabClient().Branches.DeleteBranch(args.ProjectPath, hotfixBranch)
		if err != nil {
			result.WriteString(fmt.Sprintf("⚠️  Failed to delete hotfix branch: %v\n", err))
		} else {
			result.WriteString(fmt.Sprintf("🗑️  Deleted hotfix branch: %s\n", hotfixBranch))
		}
	}

	result.WriteString(fmt.Sprintf("\n🚨 Hotfix %s is ready for urgent review and deployment!\n", args.FinishOptions.HotfixVersion))

	return mcp.NewToolResultText(result.String()), nil
}

// List branches handler (keeping existing implementation)
func listFlowBranchesHandler(ctx context.Context, request mcp.CallToolRequest, args GitFlowListBranchesArgs) (*mcp.CallToolResult, error) {
	branches, _, err := util.GitlabClient().Branches.ListBranches(args.ProjectPath, &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list branches: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Git Flow Branches for %s:\n\n", args.ProjectPath))

	branchType := strings.ToLower(args.BranchType)
	
	// Categorize branches
	var featureBranches, releaseBranches, hotfixBranches []*gitlab.Branch
	
	for _, branch := range branches {
		switch {
		case strings.HasPrefix(branch.Name, "feature/"):
			featureBranches = append(featureBranches, branch)
		case strings.HasPrefix(branch.Name, "release/"):
			releaseBranches = append(releaseBranches, branch)
		case strings.HasPrefix(branch.Name, "hotfix/"):
			hotfixBranches = append(hotfixBranches, branch)
		}
	}

	// Display branches based on type filter
	if branchType == "all" || branchType == "feature" {
		result.WriteString("🌟 Feature Branches:\n")
		if len(featureBranches) == 0 {
			result.WriteString("  No feature branches found\n")
		} else {
			for _, branch := range featureBranches {
				result.WriteString(fmt.Sprintf("  - %s (last commit: %s)\n", 
					branch.Name, branch.Commit.CreatedAt.Format("2006-01-02 15:04:05")))
			}
		}
		result.WriteString("\n")
	}

	if branchType == "all" || branchType == "release" {
		result.WriteString("🚀 Release Branches:\n")
		if len(releaseBranches) == 0 {
			result.WriteString("  No release branches found\n")
		} else {
			for _, branch := range releaseBranches {
				result.WriteString(fmt.Sprintf("  - %s (last commit: %s)\n", 
					branch.Name, branch.Commit.CreatedAt.Format("2006-01-02 15:04:05")))
			}
		}
		result.WriteString("\n")
	}

	if branchType == "all" || branchType == "hotfix" {
		result.WriteString("🚨 Hotfix Branches:\n")
		if len(hotfixBranches) == 0 {
			result.WriteString("  No hotfix branches found\n")
		} else {
			for _, branch := range hotfixBranches {
				result.WriteString(fmt.Sprintf("  - %s (last commit: %s)\n", 
					branch.Name, branch.Commit.CreatedAt.Format("2006-01-02 15:04:05")))
			}
		}
		result.WriteString("\n")
	}

	result.WriteString(fmt.Sprintf("📊 Summary: %d feature, %d release, %d hotfix branches\n", 
		len(featureBranches), len(releaseBranches), len(hotfixBranches)))

	return mcp.NewToolResultText(result.String()), nil
}