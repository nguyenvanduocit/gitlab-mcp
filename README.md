# GitLab MCP

A tool for interacting with GitLab API through MCP.

## Features

- List projects and their details
- List and manage merge requests
- Get file content from GitLab repositories
- List and analyze pipelines
- Search and list commits
- View user events and group members

## Installation

There are several ways to install the GitLab Tool:

### Option 1: Download from GitHub Releases

1. Visit the [GitHub Releases](https://github.com/yourusername/gitlab-mcp/releases) page
2. Download the binary for your platform:
   - `gitlab-mcp_linux_amd64` for Linux
   - `gitlab-mcp_darwin_amd64` for macOS
   - `gitlab-mcp_windows_amd64.exe` for Windows
3. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x gitlab-mcp_*
   ```
4. Move it to your PATH (Linux/macOS):
   ```bash
   sudo mv gitlab-mcp_* /usr/local/bin/gitlab-mcp
   ```

### Option 2: Go install
```
go install github.com/yourusername/gitlab-mcp@latest
```

### Option 3: Docker

#### Using Docker directly
1. Pull the pre-built image from GitHub Container Registry:
```bash
docker pull ghcr.io/yourusername/gitlab-mcp:latest
```

2. Or build the Docker image locally:
```bash
docker build -t gitlab-mcp .
```

3. Run using Docker:
```bash
docker run --rm -i \
  -e GITLAB_TOKEN=your_gitlab_token \
  -e GITLAB_HOST=your_gitlab_host_url \
  ghcr.io/yourusername/gitlab-mcp:latest
```

## Config

### Environment

1. Set up environment variables in `.env` file:
```
GITLAB_TOKEN=your_gitlab_token
GITLAB_HOST=your_gitlab_host_url
```

### Claude, cursor
```
{
  "mcpServers": {
    "gitlab": {
      "command": "/path-to/gitlab-mcp",
      "args": ["-env", "path-to-env-file"]
    }
  }
}
```

#### Using Docker with Cursor
```json
{
  "mcpServers": {
    "gitlab": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-e",
        "GITLAB_TOKEN=your_gitlab_token",
        "-e",
        "GITLAB_HOST=your_gitlab_host_url",
        "ghcr.io/yourusername/gitlab-mcp:latest"
      ]
    }
  }
}
```

## Development

Run the tool in SSE mode:
```
just dev
```

Or build and install:
```
just build
just install
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
