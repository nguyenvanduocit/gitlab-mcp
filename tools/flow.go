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

// Git Flow tool argument structures
type CreateReleaseBranchArgs struct {
	ProjectPath       string `json:"project_path"`
	ReleaseVersion    string `json:"release_version"`
	BaseBranch        string `json:"base_branch"`
	DevelopmentBranch string `json:"development_branch"`
}

type CreateFeatureBranchArgs struct {
	ProjectPath       string `json:"project_path"`
	FeatureName       string `json:"feature_name"`
	BaseBranch        string `json:"base_branch"`
	DevelopmentBranch string `json:"development_branch"`
}

type CreateHotfixBranchArgs struct {
	ProjectPath      string `json:"project_path"`
	HotfixVersion    string `json:"hotfix_version"`
	BaseBranch       string `json:"base_branch"`
	ProductionBranch string `json:"production_branch"`
}

type FinishReleaseArgs struct {
	ProjectPath       string `json:"project_path"`
	ReleaseVersion    string `json:"release_version"`
	DeleteBranch      bool   `json:"delete_branch"`
	DevelopmentBranch string `json:"development_branch"`
	ProductionBranch  string `json:"production_branch"`
}

type FinishFeatureArgs struct {
	ProjectPath       string `json:"project_path"`
	FeatureName       string `json:"feature_name"`
	TargetBranch      string `json:"target_branch"`
	DeleteBranch      bool   `json:"delete_branch"`
	DevelopmentBranch string `json:"development_branch"`
}

type FinishHotfixArgs struct {
	ProjectPath       string `json:"project_path"`
	HotfixVersion     string `json:"hotfix_version"`
	DeleteBranch      bool   `json:"delete_branch"`
	DevelopmentBranch string `json:"development_branch"`
	ProductionBranch  string `json:"production_branch"`
}

type ListFlowBranchesArgs struct {
	ProjectPath string `json:"project_path"`
	BranchType  string `json:"branch_type"`
}

type CreateFlowMRsArgs struct {
	ProjectPath      string `json:"project_path"`
	SourceBranch     string `json:"source_branch"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	CreateToDevelop  bool   `json:"create_to_develop"`
	CreateToMaster   bool   `json:"create_to_master"`
	DevelopmentBranch string `json:"development_branch"`
	ProductionBranch  string `json:"production_branch"`
}

// RegisterFlowTools registers all Git Flow related tools
func RegisterFlowTools(s *server.MCPServer) {
	// Release branch tools
	createReleaseTool := mcp.NewTool("gitflow_create_release",
		mcp.WithDescription("Create a new release branch following Git Flow conventions"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("release_version", mcp.Required(), mcp.Description("Release version (e.g., 1.2.0)")),
		mcp.WithString("base_branch", mcp.DefaultString("develop"), mcp.Description("Base branch to create release from")),
		mcp.WithString("development_branch", mcp.DefaultString("develop"), mcp.Description("Development branch name")),
	)

	finishReleaseTool := mcp.NewTool("gitflow_finish_release",
		mcp.WithDescription("Finish a release by creating MRs to develop and master"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("release_version", mcp.Required(), mcp.Description("Release version")),
		mcp.WithBoolean("delete_branch", mcp.Description("Delete release branch after creating MRs")),
		mcp.WithString("development_branch", mcp.DefaultString("develop"), mcp.Description("Development branch name")),
		mcp.WithString("production_branch", mcp.DefaultString("master"), mcp.Description("Production branch name")),
	)

	// Feature branch tools
	createFeatureTool := mcp.NewTool("gitflow_create_feature",
		mcp.WithDescription("Create a new feature branch following Git Flow conventions"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("feature_name", mcp.Required(), mcp.Description("Feature name (e.g., user-authentication)")),
		mcp.WithString("base_branch", mcp.DefaultString("develop"), mcp.Description("Base branch to create feature from")),
		mcp.WithString("development_branch", mcp.DefaultString("develop"), mcp.Description("Development branch name")),
	)

	finishFeatureTool := mcp.NewTool("gitflow_finish_feature",
		mcp.WithDescription("Finish a feature by creating MR to develop"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("feature_name", mcp.Required(), mcp.Description("Feature name")),
		mcp.WithString("target_branch", mcp.DefaultString("develop"), mcp.Description("Target branch for merge request")),
		mcp.WithBoolean("delete_branch", mcp.Description("Delete feature branch after creating MR")),
		mcp.WithString("development_branch", mcp.DefaultString("develop"), mcp.Description("Development branch name")),
	)

	// Hotfix branch tools
	createHotfixTool := mcp.NewTool("gitflow_create_hotfix",
		mcp.WithDescription("Create a new hotfix branch following Git Flow conventions"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("hotfix_version", mcp.Required(), mcp.Description("Hotfix version (e.g., 1.2.1)")),
		mcp.WithString("base_branch", mcp.DefaultString("master"), mcp.Description("Base branch to create hotfix from")),
		mcp.WithString("production_branch", mcp.DefaultString("master"), mcp.Description("Production branch name")),
	)

	finishHotfixTool := mcp.NewTool("gitflow_finish_hotfix",
		mcp.WithDescription("Finish a hotfix by creating MRs to develop and master"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("hotfix_version", mcp.Required(), mcp.Description("Hotfix version")),
		mcp.WithBoolean("delete_branch", mcp.Description("Delete hotfix branch after creating MRs")),
		mcp.WithString("development_branch", mcp.DefaultString("develop"), mcp.Description("Development branch name")),
		mcp.WithString("production_branch", mcp.DefaultString("master"), mcp.Description("Production branch name")),
	)

	// Utility tools
	listFlowBranchesTool := mcp.NewTool("gitflow_list_branches",
		mcp.WithDescription("List Git Flow branches (feature, release, hotfix)"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Project/repo path")),
		mcp.WithString("branch_type", mcp.DefaultString("all"), mcp.Description("Branch type to list (feature, release, hotfix, all)")),
	)
	// Register all tools
	s.AddTool(createReleaseTool, mcp.NewTypedToolHandler(createReleaseBranchHandler))
	s.AddTool(finishReleaseTool, mcp.NewTypedToolHandler(finishReleaseHandler))
	s.AddTool(createFeatureTool, mcp.NewTypedToolHandler(createFeatureBranchHandler))
	s.AddTool(finishFeatureTool, mcp.NewTypedToolHandler(finishFeatureHandler))
	s.AddTool(createHotfixTool, mcp.NewTypedToolHandler(createHotfixBranchHandler))
	s.AddTool(finishHotfixTool, mcp.NewTypedToolHandler(finishHotfixHandler))
	s.AddTool(listFlowBranchesTool, mcp.NewTypedToolHandler(listFlowBranchesHandler))
}

// Release branch handlers
func createReleaseBranchHandler(ctx context.Context, request mcp.CallToolRequest, args CreateReleaseBranchArgs) (*mcp.CallToolResult, error) {
	baseBranch := args.BaseBranch
	if baseBranch == "" {
		developmentBranch := args.DevelopmentBranch
		if developmentBranch == "" {
			developmentBranch = "develop"
		}
		baseBranch = developmentBranch
	}

	releaseBranch := fmt.Sprintf("release/%s", args.ReleaseVersion)

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
	result.WriteString(fmt.Sprintf("3. Use 'gitflow_finish_release' with version '%s' to create MRs\n", args.ReleaseVersion))

	return mcp.NewToolResultText(result.String()), nil
}

func finishReleaseHandler(ctx context.Context, request mcp.CallToolRequest, args FinishReleaseArgs) (*mcp.CallToolResult, error) {
	releaseBranch := fmt.Sprintf("release/%s", args.ReleaseVersion)
	
	// Get branch names with defaults
	developmentBranch := args.DevelopmentBranch
	if developmentBranch == "" {
		developmentBranch = "develop"
	}
	
	productionBranch := args.ProductionBranch
	if productionBranch == "" {
		productionBranch = "master"
	}
	
	// Verify release branch exists
	_, _, err := util.GitlabClient().Branches.GetBranch(args.ProjectPath, releaseBranch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("release branch '%s' not found: %v", releaseBranch, err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🚀 Finishing release %s\n\n", args.ReleaseVersion))

	// Create MR to development branch
	developMR, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Release %s", args.ReleaseVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Release %s ready for merge to %s\n\n- [ ] Code review completed\n- [ ] Tests passing\n- [ ] Documentation updated", args.ReleaseVersion, developmentBranch)),
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
		Title:        gitlab.Ptr(fmt.Sprintf("Release %s", args.ReleaseVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Release %s ready for production\n\n- [ ] Release notes prepared\n- [ ] Deployment plan reviewed\n- [ ] Rollback plan confirmed", args.ReleaseVersion)),
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
	if args.DeleteBranch {
		_, err := util.GitlabClient().Branches.DeleteBranch(args.ProjectPath, releaseBranch)
		if err != nil {
			result.WriteString(fmt.Sprintf("⚠️  Failed to delete release branch: %v\n", err))
		} else {
			result.WriteString(fmt.Sprintf("🗑️  Deleted release branch: %s\n", releaseBranch))
		}
	}

	result.WriteString(fmt.Sprintf("\n📋 Release %s is ready for review and merge!\n", args.ReleaseVersion))

	return mcp.NewToolResultText(result.String()), nil
}

// Feature branch handlers
func createFeatureBranchHandler(ctx context.Context, request mcp.CallToolRequest, args CreateFeatureBranchArgs) (*mcp.CallToolResult, error) {
	baseBranch := args.BaseBranch
	if baseBranch == "" {
		developmentBranch := args.DevelopmentBranch
		if developmentBranch == "" {
			developmentBranch = "develop"
		}
		baseBranch = developmentBranch
	}

	featureBranch := fmt.Sprintf("feature/%s", args.FeatureName)

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
	result.WriteString(fmt.Sprintf("3. Use 'gitflow_finish_feature' with name '%s' to create MR\n", args.FeatureName))

	return mcp.NewToolResultText(result.String()), nil
}

func finishFeatureHandler(ctx context.Context, request mcp.CallToolRequest, args FinishFeatureArgs) (*mcp.CallToolResult, error) {
	featureBranch := fmt.Sprintf("feature/%s", args.FeatureName)
	targetBranch := args.TargetBranch
	if targetBranch == "" {
		developmentBranch := args.DevelopmentBranch
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
	result.WriteString(fmt.Sprintf("🚀 Finishing feature %s\n\n", args.FeatureName))

	// Create MR to target branch (usually develop)
	mr, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Feature: %s", args.FeatureName)),
		Description:  gitlab.Ptr(fmt.Sprintf("Feature implementation: %s\n\n- [ ] Code review completed\n- [ ] Tests added/updated\n- [ ] Documentation updated\n- [ ] Ready for merge", args.FeatureName)),
		SourceBranch: gitlab.Ptr(featureBranch),
		TargetBranch: gitlab.Ptr(targetBranch),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create MR: %v", err)), nil
	}

	result.WriteString(fmt.Sprintf("✅ Created MR to %s: !%d\n", targetBranch, mr.IID))
	result.WriteString(fmt.Sprintf("   URL: %s\n", mr.WebURL))

	// Delete branch if requested
	if args.DeleteBranch {
		_, err := util.GitlabClient().Branches.DeleteBranch(args.ProjectPath, featureBranch)
		if err != nil {
			result.WriteString(fmt.Sprintf("⚠️  Failed to delete feature branch: %v\n", err))
		} else {
			result.WriteString(fmt.Sprintf("🗑️  Deleted feature branch: %s\n", featureBranch))
		}
	}

	result.WriteString(fmt.Sprintf("\n📋 Feature %s is ready for review!\n", args.FeatureName))

	return mcp.NewToolResultText(result.String()), nil
}

// Hotfix branch handlers
func createHotfixBranchHandler(ctx context.Context, request mcp.CallToolRequest, args CreateHotfixBranchArgs) (*mcp.CallToolResult, error) {
	baseBranch := args.BaseBranch
	if baseBranch == "" {
		productionBranch := args.ProductionBranch
		if productionBranch == "" {
			productionBranch = "master"
		}
		baseBranch = productionBranch
	}

	hotfixBranch := fmt.Sprintf("hotfix/%s", args.HotfixVersion)

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
	result.WriteString(fmt.Sprintf("3. Use 'gitflow_finish_hotfix' with version '%s' to create MRs\n", args.HotfixVersion))

	return mcp.NewToolResultText(result.String()), nil
}

func finishHotfixHandler(ctx context.Context, request mcp.CallToolRequest, args FinishHotfixArgs) (*mcp.CallToolResult, error) {
	hotfixBranch := fmt.Sprintf("hotfix/%s", args.HotfixVersion)
	
	// Get branch names with defaults
	developmentBranch := args.DevelopmentBranch
	if developmentBranch == "" {
		developmentBranch = "develop"
	}
	
	productionBranch := args.ProductionBranch
	if productionBranch == "" {
		productionBranch = "master"
	}
	
	// Verify hotfix branch exists
	_, _, err := util.GitlabClient().Branches.GetBranch(args.ProjectPath, hotfixBranch)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("hotfix branch '%s' not found: %v", hotfixBranch, err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🚨 Finishing hotfix %s\n\n", args.HotfixVersion))

	// Create MR to production branch
	masterMR, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(args.ProjectPath, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Hotfix %s", args.HotfixVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Critical hotfix %s\n\n- [ ] Fix verified\n- [ ] Tests passing\n- [ ] Ready for immediate deployment", args.HotfixVersion)),
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
		Title:        gitlab.Ptr(fmt.Sprintf("Hotfix %s", args.HotfixVersion)),
		Description:  gitlab.Ptr(fmt.Sprintf("Hotfix %s merge to %s\n\n- [ ] Conflicts resolved\n- [ ] Tests updated if needed", args.HotfixVersion, developmentBranch)),
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
	if args.DeleteBranch {
		_, err := util.GitlabClient().Branches.DeleteBranch(args.ProjectPath, hotfixBranch)
		if err != nil {
			result.WriteString(fmt.Sprintf("⚠️  Failed to delete hotfix branch: %v\n", err))
		} else {
			result.WriteString(fmt.Sprintf("🗑️  Deleted hotfix branch: %s\n", hotfixBranch))
		}
	}

	result.WriteString(fmt.Sprintf("\n🚨 Hotfix %s is ready for urgent review and deployment!\n", args.HotfixVersion))

	return mcp.NewToolResultText(result.String()), nil
}

// Utility handlers
func listFlowBranchesHandler(ctx context.Context, request mcp.CallToolRequest, args ListFlowBranchesArgs) (*mcp.CallToolResult, error) {
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