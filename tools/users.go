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

func RegisterUserTools(s *server.MCPServer) {
	userEventsTool := mcp.NewTool("gitlab_list_user_events",
		mcp.WithDescription("List GitLab user events within a date range"),
		mcp.WithString("username", mcp.Required(), mcp.Description("GitLab username")),
		mcp.WithString("since", mcp.Required(), mcp.Description("Start date (YYYY-MM-DD)")),
		mcp.WithString("until", mcp.Description("End date (YYYY-MM-DD). If not provided, defaults to current date")),
	)
	s.AddTool(userEventsTool, util.ErrorGuard(listUserEventsHandler))
}

func listUserEventsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username := request.Params.Arguments["username"].(string)
	since, ok := request.Params.Arguments["since"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: since")
	}

	until := time.Now().Format("2006-01-02")
	if value, ok := request.Params.Arguments["until"]; ok {
		until = value.(string)
	}

	sinceTime, err := time.Parse("2006-01-02", since)
	if err != nil {
		return nil, fmt.Errorf("invalid since date: %v", err)
	}

	untilTime, err := time.Parse("2006-01-02 15:04:05", until+" 23:59:59")
	if err != nil {
		return nil, fmt.Errorf("invalid until date: %v", err)
	}

	opt := &gitlab.ListContributionEventsOptions{
		After:  gitlab.Ptr(gitlab.ISOTime(sinceTime)),
		Before: gitlab.Ptr(gitlab.ISOTime(untilTime)),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	events, _, err := util.GitlabClient().Users.ListUserContributionEvents(username, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to list user events: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Events for user %s between %s and %s:\n\n",
		username, since, until))

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