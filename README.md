# GitLab MCP

A comprehensive Go-based MCP (Model Context Protocol) connector for GitLab that enables AI assistants like Claude to interact seamlessly with GitLab repositories, projects, and workflows. This tool provides a powerful interface for AI models to perform comprehensive GitLab operations through natural language.

## WHY

While GitLab provides various API integrations, our MCP implementation offers **superior AI-native interaction and comprehensive workflow management**. We've built this connector to address the real challenges developers and project managers face when working with GitLab through AI assistants.

**Key Advantages:**
- **Comprehensive GitLab Integration**: Full access to projects, merge requests, pipelines, commits, CI/CD jobs, and user management
- **AI-Optimized Interface**: Designed specifically for natural language interactions with AI assistants
- **Real-World Workflow Focus**: Built to solve actual daily problems like code review, CI/CD monitoring, and project management
- **Enhanced Productivity**: Seamless integration allowing AI to help with GitLab operations without context switching
- **Developer-Friendly**: Tools designed for actual development workflows, not just basic API calls

## 🚀 Features Overview

### 📁 Project & Repository Management
- **List and explore projects** with detailed information and access levels
- **Get file content** from any repository with branch/commit support
- **Browse repository structure** and navigate codebases
- **Access project metadata** including statistics and configuration

### 🔀 Merge Request Operations
- **List merge requests** with filtering by state, author, and target branch
- **Get detailed MR information** including changes, discussions, and approvals
- **Create new merge requests** with custom titles and descriptions
- **Comment on merge requests** and manage discussions
- **Monitor merge request status** and review progress
- **Get MR pipelines and commits** for comprehensive review
- **Rebase merge requests** when needed

### 🔄 Pipeline & CI/CD Management
- **List and analyze pipelines** with status and duration information
- **Get detailed pipeline information** including jobs and artifacts
- **Trigger new pipelines** with custom variables
- **Monitor pipeline execution** and job details
- **Track CI/CD performance** across projects

### 💼 Job Management
- **List project and pipeline jobs** with filtering capabilities
- **Get detailed job information** including logs and artifacts
- **Cancel running jobs** when needed
- **Retry failed jobs** for quick recovery
- **Monitor job status** across pipelines

### 📝 Commit & History Tracking
- **Search and list commits** with author, date, and message filtering
- **Get detailed commit information** including changes and statistics
- **Search commits by author and file path** for targeted analysis
- **Comment on commits** for code review
- **Cherry-pick and revert commits** for branch management
- **Track repository history** and development progress

### 🌊 Git Flow Workflow Support
- **Create release branches** following Git Flow conventions
- **Create feature branches** with proper naming
- **Create hotfix branches** for urgent fixes
- **Finish releases** by creating merge requests to develop and master
- **Finish features** by creating merge requests to develop
- **Finish hotfixes** by creating merge requests to both branches
- **List Git Flow branches** by type (feature, release, hotfix)

### 👥 User & Group Management
- **View user contribution events** and activity streams
- **List group members** and their roles
- **List accessible groups** with filtering options
- **Track user contributions** across projects
- **Monitor team activity** and collaboration

### 🔐 Variable & Security Management
- **List group variables** with security information
- **Create and update group variables** with proper scoping
- **Manage variable security** (protected, masked, raw)
- **Remove variables** when no longer needed
- **Handle environment-specific variables**

### 🚀 Deployment & Token Management
- **List deploy tokens** for projects and groups
- **Create deploy tokens** with specific scopes
- **Manage token permissions** and expiration
- **Delete tokens** when no longer needed
- **Handle both project and group-level tokens**

### 🔍 Advanced Search Capabilities
- **Global search** across all GitLab content
- **Group-specific search** within organizations
- **Project-specific search** for targeted results
- **Search by scope** (projects, merge requests, commits, users, code)
- **Specialized search tools** for issues, merge requests, commits, and code

## 🚀 Quick Start Guide

### Prerequisites

Before you begin, you'll need:
1. **GitLab Account** with access to your GitLab instance (GitLab.com or self-hosted)
2. **Personal Access Token** from GitLab (we'll help you get this)
3. **Cursor IDE** with Claude integration

### Step 1: Get Your GitLab Personal Access Token

1. Go to your GitLab instance → **User Settings** → **Access Tokens**
2. Click **"Add new token"**
3. Give it a name like "GitLab MCP Connector"
4. Select scopes: `api`, `read_user`, `read_repository`, `write_repository`
5. Set expiration date (optional)
6. Click **"Create personal access token"**
7. **Copy the token** (you won't see it again!)

### Step 2: Choose Your Installation Method

We recommend **Docker** for the easiest setup:

#### 🐳 Option A: Docker (Recommended)

```bash
# Pull the latest image
docker pull ghcr.io/nguyenvanduocit/gitlab-mcp:latest

# Test it works (replace with your details)
docker run --rm \
  -e GITLAB_URL=https://gitlab.com \
  -e GITLAB_TOKEN=your-personal-access-token \
  ghcr.io/nguyenvanduocit/gitlab-mcp:latest
```

#### 📦 Option B: Download Binary

1. Go to [GitHub Releases](https://github.com/nguyenvanduocit/gitlab-mcp/releases)
2. Download for your platform:
   - **macOS**: `gitlab-mcp_darwin_amd64`
   - **Linux**: `gitlab-mcp_linux_amd64`  
   - **Windows**: `gitlab-mcp_windows_amd64.exe`
3. Make it executable (macOS/Linux):
   ```bash
   chmod +x gitlab-mcp_*
   sudo mv gitlab-mcp_* /usr/local/bin/gitlab-mcp
   ```

#### 🛠️ Option C: Build from Source

```bash
go install github.com/nguyenvanduocit/gitlab-mcp@latest
```

### Step 3: Configure Cursor

1. **Open Cursor**
2. **Go to Settings** → **Features** → **Model Context Protocol**
3. **Add a new MCP server** with this configuration:

#### For Docker Users:
```json
{
  "mcpServers": {
    "gitlab": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "GITLAB_URL=https://gitlab.com",
        "-e", "GITLAB_TOKEN=your-personal-access-token",
        "ghcr.io/nguyenvanduocit/gitlab-mcp:latest"
      ]
    }
  }
}
```

#### For Binary Users:
```json
{
  "mcpServers": {
    "gitlab": {
      "command": "/usr/local/bin/gitlab-mcp",
      "env": {
        "GITLAB_URL": "https://gitlab.com",
        "GITLAB_TOKEN": "your-personal-access-token"
      }
    }
  }
}
```

#### For Self-Hosted GitLab:
Replace `https://gitlab.com` with your GitLab instance URL (e.g., `https://gitlab.yourcompany.com`)

### Step 4: Test Your Setup

1. **Restart Cursor** completely
2. **Open a new chat** with Claude
3. **Try these test commands**:

```
List my GitLab groups
```

```
Show me recent merge requests in my projects
```

```
What's the status of the latest pipeline?
```

If you see GitLab data, **congratulations! 🎉** You're all set up.

## 🔧 Advanced Configuration

### Using Environment Files

Create a `.env` file for easier management:

```bash
# .env file
GITLAB_URL=https://gitlab.com
GITLAB_TOKEN=your-personal-access-token
```

Then use it:
```bash
# With binary
gitlab-mcp -env .env

# With Docker
docker run --rm -i --env-file .env ghcr.io/nguyenvanduocit/gitlab-mcp:latest
```

### HTTP Mode for Development

For development and testing, you can run in HTTP mode:

```bash
# Start HTTP server on port 3000
gitlab-mcp -env .env -http_port 3000
```

Then configure Cursor with:
```json
{
  "mcpServers": {
    "gitlab": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

## 🎯 Usage Examples

Once configured, you can ask Claude to help with GitLab tasks using natural language:

### Project & Repository Management
- *"Show me all my GitLab groups"*
- *"List projects in group 'my-team'"*
- *"Get the README file from the main branch of project X"*
- *"What are the recent changes in project Z?"*

### Merge Request Operations
- *"Show me all open merge requests in project ABC"*
- *"Create a merge request from feature-branch to develop"*
- *"What's the status of MR #123 in project ABC?"*
- *"Add a comment to merge request #456 saying 'LGTM'"*
- *"Show me the pipeline status for MR #789"*

### Pipeline & CI/CD Management
- *"What's the status of the latest pipeline in project X?"*
- *"Trigger a new pipeline on the develop branch"*
- *"Show me failed pipelines from the last week"*
- *"List all running jobs in project Y"*
- *"Cancel job #123 in project Z"*

### Git Flow Workflows
- *"Create a release branch for version 1.2.0"*
- *"Create a feature branch called user-authentication"*
- *"Finish the release 1.2.0 and create merge requests"*
- *"List all feature branches in the project"*

### Search & Discovery
- *"Search for commits containing 'bug fix' across all projects"*
- *"Find merge requests related to authentication in group X"*
- *"Search for code containing 'API_KEY' in project Y"*
- *"Show me all users in group Z"*

### Variable & Security Management
- *"List all variables in group 'production'"*
- *"Create a new variable DATABASE_URL in group X"*
- *"Show me deploy tokens for project Y"*

### Development Insights
- *"Show me recent commits by author John"*
- *"What commits were made to the main branch today?"*
- *"List all commits with 'fix' in the message"*
- *"Show me user activity for the development team"*

## 🛠️ Available Tools Reference

### Project Tools
- `list_projects` - List projects in a group
- `get_project` - Get detailed project information

### Merge Request Tools
- `list_mrs` - List merge requests with filtering
- `get_mr_details` - Get detailed MR information
- `create_mr` - Create new merge requests
- `create_mr_note` - Add comments to merge requests
- `list_mr_comments` - List all MR comments
- `get_mr_pipelines` - Get MR pipeline information
- `get_mr_commits` - Get MR commit history
- `create_mr_pipeline` - Trigger new MR pipeline
- `rebase_mr` - Rebase merge requests

### Repository Tools
- `get_file_content` - Get file content from repositories
- `list_commits` - List commits with date filtering
- `get_commit_details` - Get detailed commit information
- `search_commits` - Search commits by author/path/date
- `get_commit_comments` - Get commit comments
- `post_commit_comment` - Add comments to commits
- `get_commit_merge_requests` - Get MRs associated with commits
- `cherry_pick_commit` - Cherry-pick commits to other branches
- `revert_commit` - Revert commits

### Pipeline Tools
- `list_pipelines` - List project pipelines
- `get_pipeline` - Get detailed pipeline information
- `trigger_pipeline` - Trigger new pipelines with variables

### Job Tools
- `list_project_jobs` - List all project jobs
- `list_pipeline_jobs` - List jobs for specific pipeline
- `get_job` - Get detailed job information
- `cancel_job` - Cancel running jobs
- `retry_job` - Retry failed jobs

### Git Flow Tools
- `gitflow_create_release` - Create release branches
- `gitflow_finish_release` - Finish releases with MRs
- `gitflow_create_feature` - Create feature branches
- `gitflow_finish_feature` - Finish features with MRs
- `gitflow_create_hotfix` - Create hotfix branches
- `gitflow_finish_hotfix` - Finish hotfixes with MRs
- `gitflow_list_branches` - List Git Flow branches

### User & Group Tools
- `list_user_contribution_events` - List user activity
- `list_group_users` - List group members
- `list_groups` - List accessible groups

### Variable Tools
- `list_group_variables` - List group variables
- `get_group_variable` - Get specific variable details
- `create_group_variable` - Create new variables
- `update_group_variable` - Update existing variables
- `remove_group_variable` - Remove variables

### Deployment Tools
- `list_all_deploy_tokens` - List all deploy tokens (admin)
- `list_project_deploy_tokens` - List project deploy tokens
- `get_project_deploy_token` - Get project token details
- `create_project_deploy_token` - Create project tokens
- `delete_project_deploy_token` - Delete project tokens
- `list_group_deploy_tokens` - List group deploy tokens
- `get_group_deploy_token` - Get group token details
- `create_group_deploy_token` - Create group tokens
- `delete_group_deploy_token` - Delete group tokens

### Search Tools
- `search_global` - Search across all GitLab
- `search_group` - Search within specific groups
- `search_project` - Search within specific projects
- `search_issues_global` - Global issue search
- `search_merge_requests_global` - Global MR search
- `search_commits_global` - Global commit search
- `search_code_global` - Global code search

## 🛠️ Troubleshooting

### Common Issues

**❌ "Connection failed" or "Authentication error"**
- Double-check your `GITLAB_URL` (should include https://)
- Verify your personal access token is correct and not expired
- Make sure your token has the required scopes (`api`, `read_user`, `read_repository`, `write_repository`)

**❌ "No MCP servers found"**
- Restart Cursor completely after adding the configuration
- Check the MCP configuration syntax in Cursor settings
- Verify the binary path is correct (for binary installations)

**❌ "Permission denied" errors**
- Make sure your GitLab account has access to the projects you're trying to access
- Check if your personal access token has the necessary permissions
- Verify you're using the correct GitLab instance URL

**❌ "Rate limit exceeded"**
- GitLab has API rate limits; wait a moment before retrying
- Consider using a more specific query to reduce API calls

**❌ "Missing required environment variables"**
- Ensure both `GITLAB_URL` and `GITLAB_TOKEN` are set
- Check that environment variables are properly loaded from .env file

### Getting Help

1. **Check the logs**: Run with `-http_port` to see detailed error messages
2. **Test your credentials**: Try the Docker test command from Step 2
3. **Verify Cursor config**: The app will show you the exact configuration to use
4. **Check GitLab permissions**: Ensure your token has access to the resources you need

## 📚 Development

For local development and contributing:

```bash
# Clone the repository
git clone https://github.com/nguyenvanduocit/gitlab-mcp.git
cd gitlab-mcp

# Create .env file with your credentials
cp .env.example .env
# Edit .env with your details

# Run in development mode
just dev
# or
go run main.go -env .env

# Build the project
just build

# Install locally
just install

# Test with MCP inspector (if running in HTTP mode)
npx @modelcontextprotocol/inspector http://localhost:3000/mcp
```

### Project Structure

```
gitlab-mcp/
├── bin/          # Compiled binaries
├── tools/        # Tool implementations
│   ├── projects.go      # Project management tools
│   ├── merge_requests.go # MR management tools
│   ├── repositories.go  # Repository and commit tools
│   ├── pipelines.go     # Pipeline management tools
│   ├── job.go          # CI/CD job tools
│   ├── flow.go         # Git Flow workflow tools
│   ├── users.go        # User management tools
│   ├── groups.go       # Group management tools
│   ├── variable.go     # Variable management tools
│   ├── deploy.go       # Deployment token tools
│   └── search.go       # Search functionality tools
├── util/         # Utility functions
├── main.go       # Application entry point
├── go.mod        # Go module definition
├── justfile      # Build automation
└── README.md     # This file
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository**
2. **Create your feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit your changes** (`git commit -m 'feat: add some amazing feature'`)
4. **Push to the branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

### Contribution Guidelines

- Follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Need help?** Check our [CHANGELOG.md](./CHANGELOG.md) for recent updates or open an issue on GitHub.
