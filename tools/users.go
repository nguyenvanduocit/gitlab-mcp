package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/gitlab-mcp/util"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type ListUserEventsArgs struct {
	Username string `json:"username"`
	Since    string `json:"since"`
	Until    string `json:"until"`
}

func RegisterUserTools(s *server.MCPServer) {
	userEventsTool := mcp.NewTool("list_user_contribution_events",
		mcp.WithDescription("List GitLab user contribution events within a date range"),
		mcp.WithString("username", mcp.Required(), mcp.Description("GitLab username")),
		mcp.WithString("since", mcp.Required(), mcp.Description("Start date (YYYY-MM-DD)")),
		mcp.WithString("until", mcp.Description("End date (YYYY-MM-DD). If not provided, defaults to current date")),
	)
	s.AddTool(userEventsTool, mcp.NewTypedToolHandler(listUserEventsHandler))
}

func listUserEventsHandler(ctx context.Context, request mcp.CallToolRequest, args ListUserEventsArgs) (*mcp.CallToolResult, error) {
	until := args.Until
	if until == "" {
		until = time.Now().Format("2006-01-02")
	}

	sinceTime, err := time.Parse("2006-01-02", args.Since)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid since date: %v", err)), nil
	}

	untilTime, err := time.Parse("2006-01-02 15:04:05", until+" 23:59:59")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid until date: %v", err)), nil
	}

	opt := &gitlab.ListContributionEventsOptions{
		After:  gitlab.Ptr(gitlab.ISOTime(sinceTime)),
		Before: gitlab.Ptr(gitlab.ISOTime(untilTime)),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	events, _, err := util.GitlabClient().Users.ListUserContributionEvents(args.Username, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list user events: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Events for user %s between %s and %s:\n\n",
		args.Username, args.Since, until))

	for _, event := range events {
		result.WriteString(fmt.Sprintf("Date: %s\n", event.CreatedAt.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("Action: %s\n", event.ActionName))

		if event.PushData.CommitCount != 0 {
			result.WriteString(fmt.Sprintf("Ref: %s\n", event.PushData.Ref))
			result.WriteString(fmt.Sprintf("Commit Count: %d\n", event.PushData.CommitCount))
			result.WriteString(fmt.Sprintf("Commit Title: %s\n", event.PushData.CommitTitle))
			result.WriteString(fmt.Sprintf("Commit From: %s\n", event.PushData.CommitFrom))
			result.WriteString(fmt.Sprintf("Commit To: %s\n", event.PushData.CommitTo))
		}

		if len(event.TargetType) > 0 {
			result.WriteString(fmt.Sprintf("Target Type: %s\n", event.TargetType))
		}

		if event.TargetIID != 0 {
			result.WriteString(fmt.Sprintf("Target IID: %d\n", event.TargetIID))
		}

		if event.ProjectID != 0 {
			result.WriteString(fmt.Sprintf("Project ID: %d\n", event.ProjectID))
		}

		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
} 