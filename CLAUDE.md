# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build & Run
```bash
# Build the binary
just build
# or
CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/gitlab-mcp ./main.go

# Run in development mode (HTTP server on port 3001)
just dev
# or
go run main.go --env .env --sse_port 3001

# Install globally
just install
# or
go install ./...
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Architecture Overview

This is a Go-based MCP (Model Context Protocol) server that provides comprehensive GitLab API integration. The codebase follows a clear modular structure:

### Core Components

1. **Main Entry Point** (`main.go`)
   - Handles environment variable validation (GITLAB_TOKEN, GITLAB_URL required)
   - Supports both stdio mode (default) and HTTP mode (--http_port flag)
   - Registers all tool modules with the MCP server

2. **Tool Modules** (`tools/` directory)
   - Each file represents a logical grouping of GitLab functionality
   - All tools follow consistent patterns:
     - Typed argument structs with validation tags
     - Handler functions that interact with GitLab API
     - Registration functions that add tools to the MCP server
   
3. **Utility Layer** (`util/gitlab.go`)
   - Singleton GitLab client initialization using sync.OnceValue
   - Centralized error handling for missing environment variables

### Tool Organization

- **projects.go**: Project listing and details
- **merge_requests.go**: MR operations (list, create, comment, rebase, pipelines)
- **repositories.go**: File content, commits, comments, cherry-pick/revert
- **branches.go**: Branch protection management (protect, unprotect, list)
- **pipelines.go**: Pipeline listing, details, and triggering
- **job.go**: CI/CD job management (list, cancel, retry)
- **flow.go**: Git Flow workflow automation
- **users.go**: User contribution events
- **groups.go**: Group management and member listing
- **variable.go**: Group and project variable CRUD operations with inheritance detection
- **deploy.go**: Deploy token management
- **search.go**: Global, group, and project-specific search

### New Features

#### Branch Protection Management (branches.go)
Added comprehensive branch protection management tool `manage_branch_protection` that provides:

**Key Features:**
- **Full Branch Protection CRUD**: protect, unprotect, list protected branches, get detailed protection info
- **Access Level Configuration**: Set push, merge, and unprotect access levels (No access, Developer, Maintainer)
- **User-Specific Permissions**: Allow specific users to push, merge, or unprotect branches
- **Code Owner Integration**: Require code owner approval for merges
- **Confirmation Required**: Protect/unprotect operations require explicit confirmation for safety

**Usage Examples:**
```bash
# List all protected branches
manage_branch_protection --action list --project_id "my-project"

# Get detailed protection info for a specific branch
manage_branch_protection --action get_protection --project_id "my-project" --branch_name "main"

# Protect a branch with maintainer-only access
manage_branch_protection --action protect --project_id "my-project" --branch_name "main" --confirmed true --protection_options '{"push_access_level": "40", "merge_access_level": "40"}'

# Protect a branch with developer merge access and code owner approval
manage_branch_protection --action protect --project_id "my-project" --branch_name "develop" --confirmed true --protection_options '{"push_access_level": "40", "merge_access_level": "30", "code_owner_approval_required": true}'

# Allow specific users to push to a protected branch
manage_branch_protection --action protect --project_id "my-project" --branch_name "main" --confirmed true --protection_options '{"allowed_to_push": ["123", "456"]}'

# Unprotect a branch (use with caution)
manage_branch_protection --action unprotect --project_id "my-project" --branch_name "feature-branch" --confirmed true
```

**Access Levels:**
- **0**: No access - completely restricted
- **30**: Developer access - users with Developer role or higher
- **40**: Maintainer access - users with Maintainer role or higher

**Safety Features:**
- **Confirmation Required**: All protect/unprotect operations require `confirmed: true` to prevent accidental changes
- **Detailed Information**: Get comprehensive protection details including access levels and permissions
- **Comprehensive Listing**: View all protected branches with their protection settings

#### Project Variables (variable.go)
Added comprehensive project variable management tool `manage_project_variable` that provides:

**Key Features:**
- **Full CRUD Operations**: list, get, create, update, remove project variables
- **Complete Value Visibility**: Shows actual values of all variables including protected, masked, and inherited ones
- **Multi-Level Inheritance**: Displays full hierarchy of ancestor groups and their variables
- **Override Detection**: Shows which variables are overridden by project or higher-level groups
- **Rich Metadata**: Displays all variable properties (protected, masked, raw, environment_scope, description)
- **Variable Types**: Supports both `env_var` and `file` variable types
- **Parent Group IDs**: Shows exact group IDs for all inherited variables

**Usage Examples:**
```bash
# List all project variables with inheritance info
manage_project_variable --action list --project_id "my-project"

# Get detailed info about a specific variable including inheritance
manage_project_variable --action get --project_id "my-project" --key "API_KEY"

# Create a new protected and masked variable
manage_project_variable --action create --project_id "my-project" --key "SECRET_TOKEN" --value "secret123" --protected true --masked true

# Update variable properties
manage_project_variable --action update --project_id "my-project" --key "API_KEY" --environment_scope "production"

# Remove a variable
manage_project_variable --action remove --project_id "my-project" --key "OLD_KEY"
```

**Inheritance Features:**
- **Multi-Level Hierarchy**: Displays complete chain of ancestor groups (project → parent → grandparent → ...)
- **Complete Variable Visibility**: Shows actual values from all levels including protected/masked variables
- **Override Detection**: Identifies which variables are overridden at each level with source group IDs
- **Group ID Tracking**: Shows exact group IDs for all inherited variables
- **Hierarchy Visualization**: Indented display showing group relationships and inheritance levels
- **Conflict Resolution**: Clear indication of which variables take precedence in the inheritance chain

### Key Design Patterns

1. **Validation**: All tool arguments use struct tags with validator/v10 for input validation
2. **Error Handling**: Consistent error wrapping with contextual messages
3. **Pagination**: GitLab API pagination handled with ListOptions
4. **Type Safety**: Strongly typed argument structs for each tool
5. **Modular Registration**: Each tool module has its own Register* function

### Environment Configuration

Required environment variables:
- `GITLAB_URL`: GitLab instance URL (e.g., https://gitlab.com)
- `GITLAB_TOKEN`: Personal access token with appropriate scopes

Optional:
- `.env` file support via --env flag
- HTTP mode support via --http_port flag for development/testing