package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	// Projects
	case "list-projects":
		runListProjects(os.Args[2:])
	case "get-project":
		runGetProject(os.Args[2:])

	// Merge Requests
	case "list-mrs":
		runListMRs(os.Args[2:])
	case "get-mr":
		runGetMR(os.Args[2:])
	case "create-mr":
		runCreateMR(os.Args[2:])
	case "accept-mr":
		runAcceptMR(os.Args[2:])
	case "rebase-mr":
		runRebaseMR(os.Args[2:])
	case "list-mr-comments":
		runListMRComments(os.Args[2:])
	case "comment-mr":
		runCommentMR(os.Args[2:])
	case "list-mr-pipelines":
		runListMRPipelines(os.Args[2:])
	case "get-mr-commits":
		runGetMRCommits(os.Args[2:])

	// Repositories
	case "get-file":
		runGetFile(os.Args[2:])
	case "list-commits":
		runListCommits(os.Args[2:])
	case "get-commit":
		runGetCommit(os.Args[2:])

	// Branches
	case "manage-branch-protection":
		runManageBranchProtection(os.Args[2:])

	// Pipelines
	case "list-pipelines":
		runListPipelines(os.Args[2:])
	case "get-pipeline":
		runGetPipeline(os.Args[2:])
	case "trigger-pipeline":
		runTriggerPipeline(os.Args[2:])

	// Jobs
	case "list-jobs":
		runListJobs(os.Args[2:])
	case "get-job":
		runGetJob(os.Args[2:])
	case "cancel-job":
		runCancelJob(os.Args[2:])
	case "retry-job":
		runRetryJob(os.Args[2:])

	// Users
	case "list-user-events":
		runListUserEvents(os.Args[2:])

	// Groups
	case "list-groups":
		runListGroups(os.Args[2:])
	case "list-group-users":
		runListGroupUsers(os.Args[2:])

	// Variables
	case "list-group-vars":
		runListGroupVars(os.Args[2:])
	case "get-group-var":
		runGetGroupVar(os.Args[2:])
	case "create-group-var":
		runCreateGroupVar(os.Args[2:])
	case "list-project-vars":
		runListProjectVars(os.Args[2:])
	case "get-project-var":
		runGetProjectVar(os.Args[2:])
	case "create-project-var":
		runCreateProjectVar(os.Args[2:])

	// Search
	case "search":
		runSearch(os.Args[2:])

	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "Run 'gitlab-cli help' for usage.\n")
		os.Exit(1)
	}
}

func loadEnv(envFile string) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load env file %s: %v\n", envFile, err)
		}
	}
}

func outputJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Println(`gitlab-cli - GitLab CLI tool

Usage: gitlab-cli <command> [flags]

Projects:
  list-projects     List projects in a group
  get-project       Get project details

Merge Requests:
  list-mrs          List merge requests
  get-mr            Get merge request details
  create-mr         Create a merge request
  accept-mr         Accept/merge a merge request
  rebase-mr         Rebase a merge request
  list-mr-comments  List MR comments
  comment-mr        Post a comment on an MR
  list-mr-pipelines List MR pipelines
  get-mr-commits    Get MR commits

Repositories:
  get-file          Get file content
  list-commits      List commits
  get-commit        Get commit details

Branches:
  manage-branch-protection  Manage branch protection (--action list|get|protect|unprotect)

Pipelines:
  list-pipelines    List pipelines
  get-pipeline      Get pipeline details
  trigger-pipeline  Trigger a new pipeline

Jobs:
  list-jobs         List jobs
  get-job           Get job details
  cancel-job        Cancel a job
  retry-job         Retry a job

Users:
  list-user-events  List user contribution events

Groups:
  list-groups       List groups
  list-group-users  List group members

Variables:
  list-group-vars   List group variables
  get-group-var     Get a group variable
  create-group-var  Create a group variable
  list-project-vars List project variables
  get-project-var   Get a project variable
  create-project-var Create a project variable

Search:
  search            Search GitLab (--action global|group|project)

Global flags available on every command:
  --env     Path to .env file
  --output  Output format: text (default) | json
`)
}

// ── Projects ─────────────────────────────────────────────────────────────────

func runListProjects(args []string) {
	fs := flag.NewFlagSet("list-projects", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	groupID := fs.String("group-id", "", "Group ID or path (required)")
	search := fs.String("search", "", "Search term")
	fs.Parse(args)
	loadEnv(*env)

	if *groupID == "" {
		die("--group-id is required")
	}

	opt := &gitlab.ListGroupProjectsOptions{
		Archived: gitlab.Ptr(false),
		OrderBy:  gitlab.Ptr("last_activity_at"),
		Sort:     gitlab.Ptr("desc"),
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}
	if *search != "" {
		opt.Search = gitlab.Ptr(*search)
	}

	projects, _, err := util.GitlabClient().Groups.ListGroupProjects(*groupID, opt)
	if err != nil {
		die("failed to list projects: %v", err)
	}

	if *output == "json" {
		outputJSON(projects)
		return
	}
	for _, p := range projects {
		fmt.Printf("%-8d  %-50s  %s\n", p.ID, p.PathWithNamespace, p.LastActivityAt.Format("2006-01-02"))
	}
}

func runGetProject(args []string) {
	fs := flag.NewFlagSet("get-project", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" {
		die("--project is required")
	}

	project, _, err := util.GitlabClient().Projects.GetProject(*projectPath, nil)
	if err != nil {
		die("failed to get project: %v", err)
	}

	if *output == "json" {
		outputJSON(project)
		return
	}
	fmt.Printf("ID:             %d\n", project.ID)
	fmt.Printf("Name:           %s\n", project.Name)
	fmt.Printf("Path:           %s\n", project.PathWithNamespace)
	fmt.Printf("Description:    %s\n", project.Description)
	fmt.Printf("URL:            %s\n", project.WebURL)
	fmt.Printf("Default Branch: %s\n", project.DefaultBranch)
}

// ── Merge Requests ────────────────────────────────────────────────────────────

func runListMRs(args []string) {
	fs := flag.NewFlagSet("list-mrs", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	state := fs.String("state", "opened", "MR state: opened|closed|merged|all")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" {
		die("--project is required")
	}

	opt := &gitlab.ListProjectMergeRequestsOptions{
		State: gitlab.Ptr(*state),
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	mrs, _, err := util.GitlabClient().MergeRequests.ListProjectMergeRequests(*projectPath, opt)
	if err != nil {
		die("failed to list merge requests: %v", err)
	}

	if *output == "json" {
		outputJSON(mrs)
		return
	}
	for _, mr := range mrs {
		fmt.Printf("!%-5d  %-8s  %-12s  %s\n", mr.IID, mr.State, mr.Author.Username, mr.Title)
	}
}

func runGetMR(args []string) {
	fs := flag.NewFlagSet("get-mr", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	mrIID := fs.Int("mr", 0, "MR IID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *mrIID == 0 {
		die("--project and --mr are required")
	}

	mr, _, err := util.GitlabClient().MergeRequests.GetMergeRequest(*projectPath, *mrIID, nil)
	if err != nil {
		die("failed to get merge request: %v", err)
	}

	if *output == "json" {
		outputJSON(mr)
		return
	}
	fmt.Printf("MR !%d: %s\n", mr.IID, mr.Title)
	fmt.Printf("State:         %s\n", mr.State)
	fmt.Printf("Author:        %s\n", mr.Author.Username)
	fmt.Printf("Source Branch: %s\n", mr.SourceBranch)
	fmt.Printf("Target Branch: %s\n", mr.TargetBranch)
	fmt.Printf("Created:       %s\n", mr.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("URL:           %s\n", mr.WebURL)
	if mr.Description != "" {
		fmt.Printf("Description:\n%s\n", mr.Description)
	}
}

func runCreateMR(args []string) {
	fs := flag.NewFlagSet("create-mr", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	sourceBranch := fs.String("source", "", "Source branch (required)")
	targetBranch := fs.String("target", "", "Target branch (required)")
	title := fs.String("title", "", "MR title (required)")
	description := fs.String("description", "", "MR description")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *sourceBranch == "" || *targetBranch == "" || *title == "" {
		die("--project, --source, --target, and --title are required")
	}

	opt := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(*title),
		SourceBranch: gitlab.Ptr(*sourceBranch),
		TargetBranch: gitlab.Ptr(*targetBranch),
	}
	if *description != "" {
		opt.Description = gitlab.Ptr(*description)
	}

	mr, _, err := util.GitlabClient().MergeRequests.CreateMergeRequest(*projectPath, opt)
	if err != nil {
		die("failed to create merge request: %v", err)
	}

	if *output == "json" {
		outputJSON(mr)
		return
	}
	fmt.Printf("Created MR !%d: %s\n", mr.IID, mr.Title)
	fmt.Printf("URL: %s\n", mr.WebURL)
}

func runAcceptMR(args []string) {
	fs := flag.NewFlagSet("accept-mr", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	mrIID := fs.Int("mr", 0, "MR IID (required)")
	squash := fs.Bool("squash", false, "Squash commits")
	removeSourceBranch := fs.Bool("remove-source-branch", false, "Remove source branch after merge")
	mergeWhenPipelineSucceeds := fs.Bool("merge-when-pipeline-succeeds", false, "Merge when pipeline succeeds")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *mrIID == 0 {
		die("--project and --mr are required")
	}

	opt := &gitlab.AcceptMergeRequestOptions{}
	if *squash {
		opt.Squash = gitlab.Ptr(true)
	}
	if *removeSourceBranch {
		opt.ShouldRemoveSourceBranch = gitlab.Ptr(true)
	}
	if *mergeWhenPipelineSucceeds {
		opt.MergeWhenPipelineSucceeds = gitlab.Ptr(true)
	}

	mr, _, err := util.GitlabClient().MergeRequests.AcceptMergeRequest(*projectPath, *mrIID, opt)
	if err != nil {
		die("failed to accept merge request: %v", err)
	}

	if *output == "json" {
		outputJSON(mr)
		return
	}
	fmt.Printf("Accepted MR !%d: %s\n", mr.IID, mr.Title)
	fmt.Printf("State: %s\n", mr.State)
	if mr.MergedAt != nil {
		fmt.Printf("Merged At: %s\n", mr.MergedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("URL: %s\n", mr.WebURL)
}

func runRebaseMR(args []string) {
	fs := flag.NewFlagSet("rebase-mr", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	projectPath := fs.String("project", "", "Project path (required)")
	mrIID := fs.Int("mr", 0, "MR IID (required)")
	skipCI := fs.Bool("skip-ci", false, "Skip CI for rebase")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *mrIID == 0 {
		die("--project and --mr are required")
	}

	opt := &gitlab.RebaseMergeRequestOptions{
		SkipCI: gitlab.Ptr(*skipCI),
	}

	_, err := util.GitlabClient().MergeRequests.RebaseMergeRequest(*projectPath, *mrIID, opt)
	if err != nil {
		die("failed to rebase merge request: %v", err)
	}
	fmt.Printf("MR !%d rebase initiated\n", *mrIID)
}

func runListMRComments(args []string) {
	fs := flag.NewFlagSet("list-mr-comments", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	mrIID := fs.Int("mr", 0, "MR IID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *mrIID == 0 {
		die("--project and --mr are required")
	}

	opt := &gitlab.ListMergeRequestNotesOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
		OrderBy:     gitlab.Ptr("created_at"),
		Sort:        gitlab.Ptr("desc"),
	}

	notes, _, err := util.GitlabClient().Notes.ListMergeRequestNotes(*projectPath, *mrIID, opt)
	if err != nil {
		die("failed to list MR comments: %v", err)
	}

	if *output == "json" {
		outputJSON(notes)
		return
	}
	for _, note := range notes {
		fmt.Printf("--- ID: %d | %s | %s ---\n%s\n\n",
			note.ID, note.Author.Username, note.CreatedAt.Format("2006-01-02 15:04:05"), note.Body)
	}
}

func runCommentMR(args []string) {
	fs := flag.NewFlagSet("comment-mr", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	mrIID := fs.Int("mr", 0, "MR IID (required)")
	comment := fs.String("comment", "", "Comment text (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *mrIID == 0 || *comment == "" {
		die("--project, --mr, and --comment are required")
	}

	opt := &gitlab.CreateMergeRequestNoteOptions{Body: gitlab.Ptr(*comment)}
	note, _, err := util.GitlabClient().Notes.CreateMergeRequestNote(*projectPath, *mrIID, opt)
	if err != nil {
		die("failed to create comment: %v", err)
	}

	if *output == "json" {
		outputJSON(note)
		return
	}
	fmt.Printf("Comment posted (ID: %d) by %s\n", note.ID, note.Author.Username)
}

func runListMRPipelines(args []string) {
	fs := flag.NewFlagSet("list-mr-pipelines", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	mrIID := fs.Int("mr", 0, "MR IID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *mrIID == 0 {
		die("--project and --mr are required")
	}

	pipelines, _, err := util.GitlabClient().MergeRequests.ListMergeRequestPipelines(*projectPath, *mrIID)
	if err != nil {
		die("failed to list MR pipelines: %v", err)
	}

	if *output == "json" {
		outputJSON(pipelines)
		return
	}
	for _, p := range pipelines {
		fmt.Printf("%-8d  %-10s  %s\n", p.ID, p.Status, p.Ref)
	}
}

func runGetMRCommits(args []string) {
	fs := flag.NewFlagSet("get-mr-commits", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	mrIID := fs.Int("mr", 0, "MR IID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *mrIID == 0 {
		die("--project and --mr are required")
	}

	commits, _, err := util.GitlabClient().MergeRequests.GetMergeRequestCommits(*projectPath, *mrIID, nil)
	if err != nil {
		die("failed to get MR commits: %v", err)
	}

	if *output == "json" {
		outputJSON(commits)
		return
	}
	for _, c := range commits {
		fmt.Printf("%s  %s  %s\n", c.ShortID, c.AuthorName, c.Title)
	}
}

// ── Repositories ──────────────────────────────────────────────────────────────

func runGetFile(args []string) {
	fs := flag.NewFlagSet("get-file", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	projectPath := fs.String("project", "", "Project path (required)")
	filePath := fs.String("file", "", "File path in repo (required)")
	ref := fs.String("ref", "main", "Branch, tag, or commit SHA")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *filePath == "" {
		die("--project and --file are required")
	}

	content, _, err := util.GitlabClient().RepositoryFiles.GetRawFile(*projectPath, *filePath, &gitlab.GetRawFileOptions{
		Ref: gitlab.Ptr(*ref),
	})
	if err != nil {
		die("failed to get file: %v", err)
	}
	fmt.Print(string(content))
}

func runListCommits(args []string) {
	fs := flag.NewFlagSet("list-commits", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	ref := fs.String("ref", "main", "Branch, tag, or commit SHA")
	since := fs.String("since", "", "Start date YYYY-MM-DD (required)")
	until := fs.String("until", "", "End date YYYY-MM-DD (default: today)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *since == "" {
		die("--project and --since are required")
	}

	untilStr := *until
	if untilStr == "" {
		untilStr = time.Now().Format("2006-01-02")
	}

	sinceTime, err := time.Parse("2006-01-02", *since)
	if err != nil {
		die("invalid --since date: %v", err)
	}
	untilTime, err := time.Parse("2006-01-02 15:04:05", untilStr+" 23:59:59")
	if err != nil {
		die("invalid --until date: %v", err)
	}

	opt := &gitlab.ListCommitsOptions{
		Since:   gitlab.Ptr(sinceTime),
		Until:   gitlab.Ptr(untilTime),
		RefName: gitlab.Ptr(*ref),
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	commits, _, err := util.GitlabClient().Commits.ListCommits(*projectPath, opt)
	if err != nil {
		die("failed to list commits: %v", err)
	}

	if *output == "json" {
		outputJSON(commits)
		return
	}
	for _, c := range commits {
		fmt.Printf("%s  %-20s  %s\n", c.ShortID, c.AuthorName, c.Title)
	}
}

func runGetCommit(args []string) {
	fs := flag.NewFlagSet("get-commit", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	sha := fs.String("sha", "", "Commit SHA (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *sha == "" {
		die("--project and --sha are required")
	}

	commit, _, err := util.GitlabClient().Commits.GetCommit(*projectPath, *sha, nil)
	if err != nil {
		die("failed to get commit: %v", err)
	}

	if *output == "json" {
		outputJSON(commit)
		return
	}
	fmt.Printf("Commit:  %s\n", commit.ID)
	fmt.Printf("Author:  %s <%s>\n", commit.AuthorName, commit.AuthorEmail)
	fmt.Printf("Date:    %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05"))
	fmt.Printf("Message: %s\n", commit.Title)
	fmt.Printf("URL:     %s\n", commit.WebURL)
}

// ── Branches ─────────────────────────────────────────────────────────────────

func runManageBranchProtection(args []string) {
	fs := flag.NewFlagSet("manage-branch-protection", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	action := fs.String("action", "", "Action: list|get|protect|unprotect (required)")
	projectPath := fs.String("project", "", "Project path (required)")
	branchName := fs.String("branch", "", "Branch name (required for get|protect|unprotect)")
	pushLevel := fs.String("push-level", "40", "Push access level: 0|30|40")
	mergeLevel := fs.String("merge-level", "40", "Merge access level: 0|30|40")
	codeOwner := fs.Bool("code-owner-approval", false, "Require code owner approval")
	fs.Parse(args)
	loadEnv(*env)

	if *action == "" || *projectPath == "" {
		die("--action and --project are required")
	}

	client := util.GitlabClient()

	switch *action {
	case "list":
		branches, _, err := client.ProtectedBranches.ListProtectedBranches(*projectPath, nil)
		if err != nil {
			die("failed to list protected branches: %v", err)
		}
		if *output == "json" {
			outputJSON(branches)
			return
		}
		for _, b := range branches {
			fmt.Printf("%-40s  push=%s  merge=%s\n", b.Name,
				formatAccessLevelCLI(b.PushAccessLevels),
				formatAccessLevelCLI(b.MergeAccessLevels))
		}

	case "get":
		if *branchName == "" {
			die("--branch is required for get action")
		}
		branch, _, err := client.ProtectedBranches.GetProtectedBranch(*projectPath, *branchName)
		if err != nil {
			die("failed to get branch protection: %v", err)
		}
		if *output == "json" {
			outputJSON(branch)
			return
		}
		fmt.Printf("Branch: %s\n", branch.Name)
		fmt.Printf("Push Access:      %s\n", formatAccessLevelCLI(branch.PushAccessLevels))
		fmt.Printf("Merge Access:     %s\n", formatAccessLevelCLI(branch.MergeAccessLevels))
		fmt.Printf("Unprotect Access: %s\n", formatAccessLevelCLI(branch.UnprotectAccessLevels))
		fmt.Printf("Code Owner:       %v\n", branch.CodeOwnerApprovalRequired)

	case "protect":
		if *branchName == "" {
			die("--branch is required for protect action")
		}
		opt := &gitlab.ProtectRepositoryBranchesOptions{
			Name:             gitlab.Ptr(*branchName),
			PushAccessLevel:  parseAccessLevelCLI(*pushLevel),
			MergeAccessLevel: parseAccessLevelCLI(*mergeLevel),
		}
		if *codeOwner {
			opt.CodeOwnerApprovalRequired = gitlab.Ptr(true)
		}
		branch, _, err := client.ProtectedBranches.ProtectRepositoryBranches(*projectPath, opt)
		if err != nil {
			die("failed to protect branch: %v", err)
		}
		if *output == "json" {
			outputJSON(branch)
			return
		}
		fmt.Printf("Protected branch '%s'\n", branch.Name)

	case "unprotect":
		if *branchName == "" {
			die("--branch is required for unprotect action")
		}
		_, err := client.ProtectedBranches.UnprotectRepositoryBranches(*projectPath, *branchName)
		if err != nil {
			die("failed to unprotect branch: %v", err)
		}
		fmt.Printf("Unprotected branch '%s'\n", *branchName)

	default:
		die("unknown action '%s'. Valid: list, get, protect, unprotect", *action)
	}
}

func parseAccessLevelCLI(level string) *gitlab.AccessLevelValue {
	switch level {
	case "0":
		return gitlab.Ptr(gitlab.NoPermissions)
	case "30":
		return gitlab.Ptr(gitlab.DeveloperPermissions)
	case "40":
		return gitlab.Ptr(gitlab.MaintainerPermissions)
	default:
		return gitlab.Ptr(gitlab.MaintainerPermissions)
	}
}

func formatAccessLevelCLI(levels []*gitlab.BranchAccessDescription) string {
	if len(levels) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(levels))
	for _, l := range levels {
		switch l.AccessLevel {
		case 0:
			parts = append(parts, "no-access")
		case 30:
			parts = append(parts, "developer")
		case 40:
			parts = append(parts, "maintainer")
		default:
			parts = append(parts, strconv.Itoa(int(l.AccessLevel)))
		}
	}
	return strings.Join(parts, ",")
}

// ── Pipelines ────────────────────────────────────────────────────────────────

func runListPipelines(args []string) {
	fs := flag.NewFlagSet("list-pipelines", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	status := fs.String("status", "", "Pipeline status filter: running|pending|success|failed|canceled|skipped")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" {
		die("--project is required")
	}

	opt := &gitlab.ListProjectPipelinesOptions{}
	if *status != "" && *status != "all" {
		opt.Status = gitlab.Ptr(gitlab.BuildStateValue(*status))
	}

	pipelines, _, err := util.GitlabClient().Pipelines.ListProjectPipelines(*projectPath, opt)
	if err != nil {
		die("failed to list pipelines: %v", err)
	}

	if *output == "json" {
		outputJSON(pipelines)
		return
	}
	for _, p := range pipelines {
		created := ""
		if p.CreatedAt != nil {
			created = p.CreatedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-8d  %-10s  %-30s  %s\n", p.ID, p.Status, p.Ref, created)
	}
}

func runGetPipeline(args []string) {
	fs := flag.NewFlagSet("get-pipeline", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	pipelineID := fs.Int("pipeline", 0, "Pipeline ID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *pipelineID == 0 {
		die("--project and --pipeline are required")
	}

	pipeline, _, err := util.GitlabClient().Pipelines.GetPipeline(*projectPath, *pipelineID)
	if err != nil {
		die("failed to get pipeline: %v", err)
	}

	if *output == "json" {
		outputJSON(pipeline)
		return
	}
	fmt.Printf("Pipeline #%d\n", pipeline.ID)
	fmt.Printf("Status:   %s\n", pipeline.Status)
	fmt.Printf("Ref:      %s\n", pipeline.Ref)
	fmt.Printf("SHA:      %s\n", pipeline.SHA)
	fmt.Printf("Created:  %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("URL:      %s\n", pipeline.WebURL)
}

func runTriggerPipeline(args []string) {
	fs := flag.NewFlagSet("trigger-pipeline", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	ref := fs.String("ref", "", "Branch, tag, or SHA (required)")
	variables := fs.String("variables", "", "Variables as KEY=VALUE,KEY2=VALUE2")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *ref == "" {
		die("--project and --ref are required")
	}

	opt := &gitlab.CreatePipelineOptions{
		Ref: gitlab.Ptr(*ref),
	}

	if *variables != "" {
		var vars []*gitlab.PipelineVariableOptions
		for _, pair := range strings.Split(*variables, ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				vars = append(vars, &gitlab.PipelineVariableOptions{
					Key:   gitlab.Ptr(strings.TrimSpace(parts[0])),
					Value: gitlab.Ptr(strings.TrimSpace(parts[1])),
				})
			}
		}
		if len(vars) > 0 {
			opt.Variables = &vars
		}
	}

	pipeline, _, err := util.GitlabClient().Pipelines.CreatePipeline(*projectPath, opt)
	if err != nil {
		die("failed to trigger pipeline: %v", err)
	}

	if *output == "json" {
		outputJSON(pipeline)
		return
	}
	fmt.Printf("Triggered Pipeline #%d\n", pipeline.ID)
	fmt.Printf("Status: %s\n", pipeline.Status)
	fmt.Printf("Ref:    %s\n", pipeline.Ref)
	fmt.Printf("URL:    %s\n", pipeline.WebURL)
}

// ── Jobs ─────────────────────────────────────────────────────────────────────

func runListJobs(args []string) {
	fs := flag.NewFlagSet("list-jobs", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	pipelineID := fs.Int("pipeline", 0, "Pipeline ID (optional, lists pipeline jobs if set)")
	scope := fs.String("scope", "", "Job scope filter: created|pending|running|failed|success|canceled|skipped")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" {
		die("--project is required")
	}

	opt := &gitlab.ListJobsOptions{}
	if *scope != "" {
		scopes := []gitlab.BuildStateValue{gitlab.BuildStateValue(*scope)}
		opt.Scope = &scopes
	}

	var jobs []*gitlab.Job
	var err error

	if *pipelineID != 0 {
		jobs, _, err = util.GitlabClient().Jobs.ListPipelineJobs(*projectPath, *pipelineID, opt)
		if err != nil {
			die("failed to list pipeline jobs: %v", err)
		}
	} else {
		jobs, _, err = util.GitlabClient().Jobs.ListProjectJobs(*projectPath, opt)
		if err != nil {
			die("failed to list project jobs: %v", err)
		}
	}

	if *output == "json" {
		outputJSON(jobs)
		return
	}
	for _, j := range jobs {
		fmt.Printf("%-8d  %-10s  %-15s  %s\n", j.ID, j.Status, j.Stage, j.Name)
	}
}

func runGetJob(args []string) {
	fs := flag.NewFlagSet("get-job", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	jobID := fs.Int("job", 0, "Job ID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *jobID == 0 {
		die("--project and --job are required")
	}

	job, _, err := util.GitlabClient().Jobs.GetJob(*projectPath, *jobID)
	if err != nil {
		die("failed to get job: %v", err)
	}

	if *output == "json" {
		outputJSON(job)
		return
	}
	fmt.Printf("Job #%d: %s\n", job.ID, job.Name)
	fmt.Printf("Status: %s\n", job.Status)
	fmt.Printf("Stage:  %s\n", job.Stage)
	fmt.Printf("Ref:    %s\n", job.Ref)
	fmt.Printf("URL:    %s\n", job.WebURL)
}

func runCancelJob(args []string) {
	fs := flag.NewFlagSet("cancel-job", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	jobID := fs.Int("job", 0, "Job ID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *jobID == 0 {
		die("--project and --job are required")
	}

	job, _, err := util.GitlabClient().Jobs.CancelJob(*projectPath, *jobID)
	if err != nil {
		die("failed to cancel job: %v", err)
	}

	if *output == "json" {
		outputJSON(job)
		return
	}
	fmt.Printf("Cancelled Job #%d: %s (status: %s)\n", job.ID, job.Name, job.Status)
}

func runRetryJob(args []string) {
	fs := flag.NewFlagSet("retry-job", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectPath := fs.String("project", "", "Project path (required)")
	jobID := fs.Int("job", 0, "Job ID (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectPath == "" || *jobID == 0 {
		die("--project and --job are required")
	}

	job, _, err := util.GitlabClient().Jobs.RetryJob(*projectPath, *jobID)
	if err != nil {
		die("failed to retry job: %v", err)
	}

	if *output == "json" {
		outputJSON(job)
		return
	}
	fmt.Printf("Retried Job #%d: %s (status: %s)\n", job.ID, job.Name, job.Status)
}

// ── Users ─────────────────────────────────────────────────────────────────────

func runListUserEvents(args []string) {
	fs := flag.NewFlagSet("list-user-events", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	username := fs.String("username", "", "GitLab username (required)")
	since := fs.String("since", "", "Start date YYYY-MM-DD (required)")
	until := fs.String("until", "", "End date YYYY-MM-DD (default: today)")
	fs.Parse(args)
	loadEnv(*env)

	if *username == "" || *since == "" {
		die("--username and --since are required")
	}

	untilStr := *until
	if untilStr == "" {
		untilStr = time.Now().Format("2006-01-02")
	}

	sinceTime, err := time.Parse("2006-01-02", *since)
	if err != nil {
		die("invalid --since date: %v", err)
	}
	untilTime, err := time.Parse("2006-01-02 15:04:05", untilStr+" 23:59:59")
	if err != nil {
		die("invalid --until date: %v", err)
	}

	opt := &gitlab.ListContributionEventsOptions{
		After:  gitlab.Ptr(gitlab.ISOTime(sinceTime)),
		Before: gitlab.Ptr(gitlab.ISOTime(untilTime)),
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	events, _, err := util.GitlabClient().Users.ListUserContributionEvents(*username, opt)
	if err != nil {
		die("failed to list user events: %v", err)
	}

	if *output == "json" {
		outputJSON(events)
		return
	}
	for _, e := range events {
		fmt.Printf("%s  %-20s", e.CreatedAt.Format("2006-01-02 15:04:05"), e.ActionName)
		if e.PushData.CommitCount > 0 {
			fmt.Printf("  ref=%s commits=%d", e.PushData.Ref, e.PushData.CommitCount)
		}
		fmt.Println()
	}
}

// ── Groups ────────────────────────────────────────────────────────────────────

func runListGroups(args []string) {
	fs := flag.NewFlagSet("list-groups", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	search := fs.String("search", "", "Search term")
	owned := fs.Bool("owned", false, "Only owned groups")
	fs.Parse(args)
	loadEnv(*env)

	opt := &gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
		OrderBy:     gitlab.Ptr("name"),
		Sort:        gitlab.Ptr("asc"),
	}
	if *search != "" {
		opt.Search = gitlab.Ptr(*search)
	}
	if *owned {
		opt.Owned = gitlab.Ptr(true)
	}

	groups, _, err := util.GitlabClient().Groups.ListGroups(opt)
	if err != nil {
		die("failed to list groups: %v", err)
	}

	if *output == "json" {
		outputJSON(groups)
		return
	}
	for _, g := range groups {
		fmt.Printf("%-8d  %-40s  %s\n", g.ID, g.FullPath, g.WebURL)
	}
}

func runListGroupUsers(args []string) {
	fs := flag.NewFlagSet("list-group-users", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	groupID := fs.String("group-id", "", "Group ID or path (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *groupID == "" {
		die("--group-id is required")
	}

	opt := &gitlab.ListGroupMembersOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	members, _, err := util.GitlabClient().Groups.ListGroupMembers(*groupID, opt)
	if err != nil {
		die("failed to list group members: %v", err)
	}

	if *output == "json" {
		outputJSON(members)
		return
	}
	for _, m := range members {
		fmt.Printf("%-8d  %-20s  %-30s  %s\n", m.ID, m.Username, m.Name, m.State)
	}
}

// ── Variables ─────────────────────────────────────────────────────────────────

func runListGroupVars(args []string) {
	fs := flag.NewFlagSet("list-group-vars", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	groupID := fs.String("group-id", "", "Group ID or path (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *groupID == "" {
		die("--group-id is required")
	}

	vars, _, err := util.GitlabClient().GroupVariables.ListVariables(*groupID, &gitlab.ListGroupVariablesOptions{})
	if err != nil {
		die("failed to list group variables: %v", err)
	}

	if *output == "json" {
		outputJSON(vars)
		return
	}
	for _, v := range vars {
		fmt.Printf("%-40s  protected=%-5v  masked=%-5v  scope=%s\n",
			v.Key, v.Protected, v.Masked, v.EnvironmentScope)
	}
}

func runGetGroupVar(args []string) {
	fs := flag.NewFlagSet("get-group-var", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	groupID := fs.String("group-id", "", "Group ID or path (required)")
	key := fs.String("key", "", "Variable key (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *groupID == "" || *key == "" {
		die("--group-id and --key are required")
	}

	v, _, err := util.GitlabClient().GroupVariables.GetVariable(*groupID, *key, nil)
	if err != nil {
		die("failed to get group variable: %v", err)
	}

	if *output == "json" {
		outputJSON(v)
		return
	}
	fmt.Printf("Key:   %s\n", v.Key)
	fmt.Printf("Value: %s\n", v.Value)
	fmt.Printf("Type:  %s\n", v.VariableType)
	fmt.Printf("Protected: %v  Masked: %v  Raw: %v\n", v.Protected, v.Masked, v.Raw)
	fmt.Printf("Scope: %s\n", v.EnvironmentScope)
}

func runCreateGroupVar(args []string) {
	fs := flag.NewFlagSet("create-group-var", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	groupID := fs.String("group-id", "", "Group ID or path (required)")
	key := fs.String("key", "", "Variable key (required)")
	value := fs.String("value", "", "Variable value (required)")
	protected := fs.Bool("protected", false, "Mark as protected")
	masked := fs.Bool("masked", false, "Mark as masked")
	varType := fs.String("type", "env_var", "Variable type: env_var|file")
	scope := fs.String("scope", "*", "Environment scope")
	fs.Parse(args)
	loadEnv(*env)

	if *groupID == "" || *key == "" || *value == "" {
		die("--group-id, --key, and --value are required")
	}

	opt := &gitlab.CreateGroupVariableOptions{
		Key:              gitlab.Ptr(*key),
		Value:            gitlab.Ptr(*value),
		Protected:        gitlab.Ptr(*protected),
		Masked:           gitlab.Ptr(*masked),
		VariableType:     gitlab.Ptr(gitlab.VariableTypeValue(*varType)),
		EnvironmentScope: gitlab.Ptr(*scope),
	}

	v, _, err := util.GitlabClient().GroupVariables.CreateVariable(*groupID, opt)
	if err != nil {
		die("failed to create group variable: %v", err)
	}

	if *output == "json" {
		outputJSON(v)
		return
	}
	fmt.Printf("Created variable '%s' in group %s\n", v.Key, *groupID)
}

func runListProjectVars(args []string) {
	fs := flag.NewFlagSet("list-project-vars", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectID := fs.String("project", "", "Project ID or path (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectID == "" {
		die("--project is required")
	}

	vars, _, err := util.GitlabClient().ProjectVariables.ListVariables(*projectID, &gitlab.ListProjectVariablesOptions{})
	if err != nil {
		die("failed to list project variables: %v", err)
	}

	if *output == "json" {
		outputJSON(vars)
		return
	}
	for _, v := range vars {
		fmt.Printf("%-40s  protected=%-5v  masked=%-5v  scope=%s\n",
			v.Key, v.Protected, v.Masked, v.EnvironmentScope)
	}
}

func runGetProjectVar(args []string) {
	fs := flag.NewFlagSet("get-project-var", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectID := fs.String("project", "", "Project ID or path (required)")
	key := fs.String("key", "", "Variable key (required)")
	fs.Parse(args)
	loadEnv(*env)

	if *projectID == "" || *key == "" {
		die("--project and --key are required")
	}

	v, _, err := util.GitlabClient().ProjectVariables.GetVariable(*projectID, *key, nil)
	if err != nil {
		die("failed to get project variable: %v", err)
	}

	if *output == "json" {
		outputJSON(v)
		return
	}
	fmt.Printf("Key:   %s\n", v.Key)
	fmt.Printf("Value: %s\n", v.Value)
	fmt.Printf("Type:  %s\n", v.VariableType)
	fmt.Printf("Protected: %v  Masked: %v  Raw: %v\n", v.Protected, v.Masked, v.Raw)
	fmt.Printf("Scope: %s\n", v.EnvironmentScope)
}

func runCreateProjectVar(args []string) {
	fs := flag.NewFlagSet("create-project-var", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	projectID := fs.String("project", "", "Project ID or path (required)")
	key := fs.String("key", "", "Variable key (required)")
	value := fs.String("value", "", "Variable value (required)")
	protected := fs.Bool("protected", false, "Mark as protected")
	masked := fs.Bool("masked", false, "Mark as masked")
	varType := fs.String("type", "env_var", "Variable type: env_var|file")
	scope := fs.String("scope", "*", "Environment scope")
	fs.Parse(args)
	loadEnv(*env)

	if *projectID == "" || *key == "" || *value == "" {
		die("--project, --key, and --value are required")
	}

	opt := &gitlab.CreateProjectVariableOptions{
		Key:              gitlab.Ptr(*key),
		Value:            gitlab.Ptr(*value),
		Protected:        gitlab.Ptr(*protected),
		Masked:           gitlab.Ptr(*masked),
		VariableType:     gitlab.Ptr(gitlab.VariableTypeValue(*varType)),
		EnvironmentScope: gitlab.Ptr(*scope),
	}

	v, _, err := util.GitlabClient().ProjectVariables.CreateVariable(*projectID, opt)
	if err != nil {
		die("failed to create project variable: %v", err)
	}

	if *output == "json" {
		outputJSON(v)
		return
	}
	fmt.Printf("Created variable '%s' in project %s\n", v.Key, *projectID)
}

// ── Search ────────────────────────────────────────────────────────────────────

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	action := fs.String("action", "global", "Search action: global|group|project")
	query := fs.String("query", "", "Search query (required)")
	scope := fs.String("scope", "projects", "Scope: projects|merge_requests|commits|blobs|users")
	groupID := fs.String("group-id", "", "Group ID (required for group action)")
	projectID := fs.String("project", "", "Project ID (required for project action)")
	ref := fs.String("ref", "", "Ref for blob/commit search")
	fs.Parse(args)
	loadEnv(*env)

	if *query == "" {
		die("--query is required")
	}

	client := util.GitlabClient()
	opt := &gitlab.SearchOptions{ListOptions: gitlab.ListOptions{PerPage: 20}}
	if *ref != "" {
		opt.Ref = ref
	}

	ctx := context.Background()
	_ = ctx // SearchOptions does not take context; kept for pattern clarity

	switch *action {
	case "global":
		runGlobalSearch(client, *query, *scope, opt, *output)
	case "group":
		if *groupID == "" {
			die("--group-id is required for group action")
		}
		runGroupSearch(client, *groupID, *query, *scope, opt, *output)
	case "project":
		if *projectID == "" {
			die("--project is required for project action")
		}
		runProjectSearch(client, *projectID, *query, *scope, opt, *output)
	default:
		die("unknown action '%s'. Valid: global, group, project", *action)
	}
}

func runGlobalSearch(client *gitlab.Client, query, scope string, opt *gitlab.SearchOptions, output string) {
	switch scope {
	case "projects":
		results, _, err := client.Search.Projects(query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, p := range results {
			fmt.Printf("%-8d  %s\n", p.ID, p.PathWithNamespace)
		}
	case "merge_requests":
		results, _, err := client.Search.MergeRequests(query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, mr := range results {
			fmt.Printf("!%-5d  %-8s  %s\n", mr.IID, mr.State, mr.Title)
		}
	case "commits":
		results, _, err := client.Search.Commits(query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, c := range results {
			fmt.Printf("%s  %s\n", c.ShortID, c.Title)
		}
	case "blobs":
		results, _, err := client.Search.Blobs(query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, b := range results {
			fmt.Printf("%-8d  %s\n", b.ProjectID, b.Path)
		}
	case "users":
		results, _, err := client.Search.Users(query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, u := range results {
			fmt.Printf("%-8d  %-20s  %s\n", u.ID, u.Username, u.Name)
		}
	default:
		die("unsupported scope '%s' for global search", scope)
	}
}

func runGroupSearch(client *gitlab.Client, groupID, query, scope string, opt *gitlab.SearchOptions, output string) {
	switch scope {
	case "projects":
		results, _, err := client.Search.ProjectsByGroup(groupID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, p := range results {
			fmt.Printf("%-8d  %s\n", p.ID, p.PathWithNamespace)
		}
	case "merge_requests":
		results, _, err := client.Search.MergeRequestsByGroup(groupID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, mr := range results {
			fmt.Printf("!%-5d  %-8s  %s\n", mr.IID, mr.State, mr.Title)
		}
	case "commits":
		results, _, err := client.Search.CommitsByGroup(groupID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, c := range results {
			fmt.Printf("%s  %s\n", c.ShortID, c.Title)
		}
	case "blobs":
		results, _, err := client.Search.BlobsByGroup(groupID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, b := range results {
			fmt.Printf("%-8d  %s\n", b.ProjectID, b.Path)
		}
	case "users":
		results, _, err := client.Search.UsersByGroup(groupID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, u := range results {
			fmt.Printf("%-8d  %-20s  %s\n", u.ID, u.Username, u.Name)
		}
	default:
		die("unsupported scope '%s' for group search", scope)
	}
}

func runProjectSearch(client *gitlab.Client, projectID, query, scope string, opt *gitlab.SearchOptions, output string) {
	switch scope {
	case "merge_requests":
		results, _, err := client.Search.MergeRequestsByProject(projectID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, mr := range results {
			fmt.Printf("!%-5d  %-8s  %s\n", mr.IID, mr.State, mr.Title)
		}
	case "commits":
		results, _, err := client.Search.CommitsByProject(projectID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, c := range results {
			fmt.Printf("%s  %s\n", c.ShortID, c.Title)
		}
	case "blobs":
		results, _, err := client.Search.BlobsByProject(projectID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, b := range results {
			fmt.Printf("%-8d  %s\n", b.ProjectID, b.Path)
		}
	case "users":
		results, _, err := client.Search.UsersByProject(projectID, query, opt)
		if err != nil {
			die("search failed: %v", err)
		}
		if output == "json" {
			outputJSON(results)
			return
		}
		for _, u := range results {
			fmt.Printf("%-8d  %-20s  %s\n", u.ID, u.Username, u.Name)
		}
	default:
		die("unsupported scope '%s' for project search", scope)
	}
}
