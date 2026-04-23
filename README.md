# Bots - Session Persistence & Decision Tracking

A CLI tool and MCP server for maintaining living project documentation, session logs, and structured task handoffs for AI agents.

## What This Does

**Bots** provides a structured system for AI agents to:

- **Track project state** in `.bots/CHECKPOINTS.md` - a living document that evolves with your project
- **Log decisions** in per-session markdown files under `.bots/logs/`
- **Hand off tasks** between master/slave models with a structured review protocol
- **Integrate with AI agents** via MCP (Model Context Protocol) for programmatic access
- **Link to git** for commit tracking and branch management

Inspired by the living document pattern used in projects like [guia](https://github.com/AlvaroRausell/guia), where `README.md` serves as an evolving project plan with decision logs.

## How It Works

### Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  AI Agent   │────▶│  MCP Server  │────▶│  CLI Tools  │
│  (opencode, │     │  (stdio)     │     │  (bots)     │
│   claude)   │◀────│              │◀────│             │
└─────────────┘     └──────────────┘     └─────────────┘
                                                │
                                                ▼
                                         ┌─────────────┐
                                         │  .bots/     │
                                         │  - CHECKP   │
                                         │  - logs/    │
                                         │  - tasks/   │
                                         └─────────────┘
```

### File Structure

```
project/
├── .bots/
│   ├── AGENTS.md               # AI agent instructions
│   ├── CHECKPOINTS.md          # Current project state + startup checklist
│   ├── logs/                   # Session decision logs
│   │   └── 2026-04-12-api-design.md
│   ├── tasks/                  # Task handoff files
│   │   └── phase-1.5.md
│   └── skills/
│       └── session-persistence/
│           └── SKILL.md        # Bot usage instructions
├── cmd/
│   └── bots/
│       └── main.go             # CLI entry point
├── internal/                   # Go packages
├── go.mod
└── Makefile
```

### Workflow

1. **Agent initializes project context** → Reads checkpoint and completes the startup checklist
2. **Agent plans the work** → Defines project phases before implementation
3. **Agent starts session** → Creates session log once planning is current
4. **Agent makes decisions** → Logs each decision with rationale
5. **Agent completes work** → Updates checkpoint, commits changes
6. **Next session** → New agent reads checkpoint + logs for context

## Installation

### Homebrew (macOS & Linux)

```bash
brew install AlvaroRausell/tap/bots
```

### From Source

```bash
git clone https://github.com/AlvaroRausell/bots.git
cd bots
make build      # Binary goes to dist/bots
make install    # Installs to ~/.local/bin (override with GOBIN=)
```

### Note for mise users

This repository sets `go_set_gobin = false` in `mise.toml` so `go install ./cmd/bots`
installs to the normal Go bin directory (`$GOPATH/bin`, usually `~/go/bin`) instead of a
version-specific mise shim path. That keeps `bots` runnable outside this repo too.

## Usage

### Quick Start

```bash
# Initialize a new project
bots init "my-project"

# Install MCP to your AI agent
bots install
```

### CLI Commands

#### Project Initialization

```bash
# Initialize .bots directory structure
bots init "project-name"

# View init help
bots init help
```

This creates:
- `.bots/AGENTS.md` - AI agent instructions
- `.bots/CHECKPOINTS.md` - Living project state document with a startup checklist
- `.bots/logs/` - Session decision logs
- `.bots/tasks/` - Task handoff files
- `.bots/skills/` - AI agent skills

Before the first log is created, fill the startup checklist in `.bots/CHECKPOINTS.md`
and define the project phases.

#### Checkpoints

```bash
# Read current project state
bots checkpoint read

# Update a section
bots checkpoint update "Current Checkpoint" "- Completed Phase 1\n- Starting Phase 2"

# List all sections
bots checkpoint list
```

#### Session Logs

Start session logs after the startup checklist and project phases are current.

```bash
# Start a new session log
bots log start "api-redesign"

# Append a decision
bots log append api-redesign "Decision: using gin router for middleware support"

# Search across all logs
bots log search "middleware"

# Generate decision summary
bots log summarize api-redesign

# List all logs
bots log list
```

#### Task Handoff

```bash
# Create a new task
bots task create "phase-1.5"

# Read task file
bots task read "phase-1.5"

# Update status
bots task update "phase-1.5" IN_PROGRESS
bots task update "phase-1.5" READY_FOR_REVIEW
bots task update "phase-1.5" DONE

# List all tasks with status
bots task list
```

#### MCP Server

```bash
# Start MCP server (stdio)
bots mcp serve

# Install to AI agents (interactive TUI)
bots install
```

### MCP Tools

When configured, AI agents can access these tools:

| Tool | Description |
|------|-------------|
| `checkpoint_read` | Read CHECKPOINTS.md content |
| `checkpoint_update` | Update a checkpoint section |
| `checkpoint_list` | List all checkpoint sections |
| `log_create` | Create a new session log |
| `log_append` | Append entry to session log |
| `log_search` | Search across all logs |
| `log_summarize` | Generate decision summary |
| `log_list` | List all session logs |
| `task_create` | Create a new task file |
| `task_read` | Read a task file |
| `task_update_status` | Update task status |
| `git_commit_checkpoint` | Commit .bots changes to git |
| `git_branch_info` | Get current branch info |

### Makefile Targets

```bash
make build       # Build the CLI
make test        # Run tests
make install     # Install to ~/.local/bin (default)
make clean       # Remove build artifacts
make install-mcp # Run interactive installer
make mcp         # Start MCP server
make init        # Create .bots directory structure
make fmt         # Format Go code
make lint        # Lint Go code
make check       # Run fmt, lint, test, build
make help        # Show help
```

## Examples

### Initialize a Project

```bash
bots init "my-app"

# Then fill .bots/CHECKPOINTS.md before starting work:
# - complete the Project Startup Checklist
# - define the Project Phases
```

### Starting a New Feature

```bash
# Make sure planning is current in .bots/CHECKPOINTS.md first
# Then start a session for the feature
bots log start "user-auth"

# Work session - decisions logged in real-time
bots log append user-auth "Decision: JWT with 15min expiry"
bots log append user-auth "Decision: bcrypt cost 12 for password hashing"
bots log append user-auth "Architecture: auth middleware in internal/http/middleware/"

# Update checkpoint when done
bots checkpoint update "Current Checkpoint" "- User auth complete\n- JWT + bcrypt"

# Commit
bots git_commit_checkpoint "Add user authentication"
```

### Task Handoff Protocol

This protocol is particularly useful for splitting planning and execution between models. For example, an expensive, highly capable model (like Opus 4.6, running on claude code) can be used to plan the task and review the results, while a cheaper model (running via a harness like opencode) performs the actual code writing.

**Master model creates task:**

```bash
bots task create "api-refactor"
# Edit .bots/tasks/api-refactor.md to add description and criteria
bots task update "api-refactor" PENDING
```

**Slave model executes:**

```bash
bots task read "api-refactor"
bots task update "api-refactor" IN_PROGRESS
# ... do work, logging decisions ...
bots task update "api-refactor" READY_FOR_REVIEW
# Add progress section to task file
```

**Master model reviews:**

```bash
bots task read "api-refactor"
git diff HEAD
# If acceptable:
bots task update "api-refactor" DONE
# If changes needed:
bots task update "api-refactor" CHANGES_REQUESTED
# Add review section with feedback
```

### AI Agent Integration

Configure your AI agent to use the MCP server:

```bash
./bots install
```

Select your agent (opencode, Claude Code, Codex) from the TUI. The installer will create an `mcp.json` configuration file.

**Example `~/.opencode/mcp.json`:**

```json
{
  "mcpServers": {
    "bots": {
      "command": "/path/to/bots",
      "args": ["mcp", "serve"]
    }
  }
}
```

Restart your AI agent to enable MCP tools.

## Skill Integration

Add the skill to your project:

```bash
mkdir -p .bots/skills/session-persistence
cp .bots/skills/session-persistence/SKILL.md <your-project>/.bots/skills/session-persistence/
```

The skill document tells AI agents how to use this system. See `.bots/skills/session-persistence/SKILL.md` for the full documentation.

## Best Practices

1. **Plan first** - Complete the startup checklist and define project phases before the first work session
2. **Log decisions in real-time** - Don't wait until session end
3. **Be specific** - Include rationale, not just what was decided
4. **Link to files** - Reference specific paths when relevant
5. **Update checkpoints incrementally** - Keep state current
6. **Use task handoff for complex work** - Simple fixes don't need full protocol
7. **One session per log** - Keep logs atomic for easy navigation
8. **Commit checkpoints frequently** - Track state changes in git

## Development

### Build from Source

```bash
git clone https://github.com/AlvaroRausell/bots.git
cd bots
make build    # Output: dist/bots
```

### Run Tests

```bash
make test
```

### Full Check

```bash
make check  # fmt + lint + test + build
```

### MCP Server Testing

```bash
# Terminal 1: Start MCP server
bots mcp serve

# Terminal 2: Test with CLI
bots checkpoint read
bots log start test
```

## License

MIT

## Contributing

1. Fork the repo
2. Create a feature branch
3. Make your changes
4. Run `make check`
5. Submit a PR

---

**Built with** Go, [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [MCP](https://modelcontextprotocol.io)
