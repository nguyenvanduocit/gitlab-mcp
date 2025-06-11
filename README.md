# GitLab MCP

A Go-based MCP (Model Control Protocol) connector for GitLab that enables AI assistants like Claude to interact with GitLab repositories and projects. This tool provides a seamless interface for AI models to perform comprehensive GitLab operations through natural language.

## WHY

While GitLab provides various API integrations, our MCP implementation offers **superior AI-native interaction and comprehensive workflow management**. We've built this connector to address the real challenges developers and project managers face when working with GitLab through AI assistants.

**Key Advantages:**
- **Comprehensive GitLab Integration**: Full access to projects, merge requests, pipelines, commits, and user management
- **AI-Optimized Interface**: Designed specifically for natural language interactions with AI assistants
- **Real-World Workflow Focus**: Built to solve actual daily problems like code review, CI/CD monitoring, and project management
- **Enhanced Productivity**: Seamless integration allowing AI to help with GitLab operations without context switching
- **Developer-Friendly**: Tools designed for actual development workflows, not just basic API calls

## Features

### Project & Repository Management
- **List and explore projects** with detailed information and access levels
- **Get file content** from any repository with branch/commit support
- **Browse repository structure** and navigate codebases
- **Access project metadata** including statistics and configuration

### Merge Request Operations
- **List merge requests** with filtering by state, author, and target branch
- **Get detailed MR information** including changes, discussions, and approvals
- **Monitor merge request status** and review progress
- **Track MR lifecycle** from creation to merge

### Pipeline & CI/CD Management
- **List and analyze pipelines** with status and duration information
- **Monitor pipeline execution** and job details
- **Track CI/CD performance** across projects
- **Get pipeline artifacts** and build information

### Commit & History Tracking
- **Search and list commits** with author, date, and message filtering
- **Get detailed commit information** including changes and statistics
- **Track repository history** and development progress
- **Analyze commit patterns** and contributor activity

### User & Group Management
- **View user events** and activity streams
- **List group members** and their roles
- **Track user contributions** across projects
- **Monitor team activity** and collaboration

### Advanced Features
- **Flexible filtering** across all operations
- **Rich data formatting** optimized for AI consumption
- **Error handling** with detailed debugging information
- **Multi-project support** for complex workflows

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
4. Select scopes: `api`, `read_user`, `read_repository`
5. Set expiration date (optional)
6. Click **"Create personal access token"**
7. **Copy the token** (you won't see it again!)

### Step 2: Choose Your Installation Method

We recommend **Docker** for the easiest setup:

#### 🐳 Option A: Docker (Recommended)

```bash
# Pull the latest image
docker pull ghcr.io/yourusername/gitlab-mcp:latest

# Test it works (replace with your details)
docker run --rm \
  -e GITLAB_HOST=https://gitlab.com \
  -e GITLAB_TOKEN=your-personal-access-token \
  ghcr.io/yourusername/gitlab-mcp:latest
```

#### 📦 Option B: Download Binary

1. Go to [GitHub Releases](https://github.com/yourusername/gitlab-mcp/releases)
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
go install github.com/yourusername/gitlab-mcp@latest
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
        "-e", "GITLAB_HOST=https://gitlab.com",
        "-e", "GITLAB_TOKEN=your-personal-access-token",
        "ghcr.io/yourusername/gitlab-mcp:latest"
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
        "GITLAB_HOST": "https://gitlab.com",
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
List my GitLab projects
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
GITLAB_HOST=https://gitlab.com
GITLAB_TOKEN=your-personal-access-token
```

Then use it:
```bash
# With binary
gitlab-mcp -env .env

# With Docker
docker run --rm -i --env-file .env ghcr.io/yourusername/gitlab-mcp:latest
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
- *"Show me all my GitLab projects"*
- *"Get the README file from the main branch of project X"*
- *"List all files in the src directory of repository Y"*
- *"What are the recent changes in project Z?"*

### Merge Request Operations
- *"Show me all open merge requests assigned to me"*
- *"What's the status of MR #123 in project ABC?"*
- *"List merge requests that need my review"*
- *"Show me the changes in the latest merge request"*

### Pipeline & CI/CD Monitoring
- *"What's the status of the latest pipeline in project X?"*
- *"Show me failed pipelines from the last week"*
- *"List all running pipelines across my projects"*
- *"Get details of pipeline #456 in project Y"*

### Development Insights
- *"Show me recent commits by author John"*
- *"What commits were made to the main branch today?"*
- *"List all commits with 'fix' in the message"*
- *"Show me user activity for the development team"*

## 🛠️ Troubleshooting

### Common Issues

**❌ "Connection failed" or "Authentication error"**
- Double-check your `GITLAB_HOST` (should include https://)
- Verify your personal access token is correct and not expired
- Make sure your token has the required scopes (`api`, `read_user`, `read_repository`)

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

### Getting Help

1. **Check the logs**: Run with `-http_port` to see detailed error messages
2. **Test your credentials**: Try the Docker test command from Step 2
3. **Verify Cursor config**: The app will show you the exact configuration to use
4. **Check GitLab permissions**: Ensure your token has access to the resources you need

## 📚 Development

For local development and contributing:

```bash
# Clone the repository
git clone https://github.com/yourusername/gitlab-mcp.git
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
