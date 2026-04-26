# Bots Domain Language

Bots is a CLI and MCP server for AI agent coordination through project-local state.

## Terms

- **Project workspace** — the repository or directory tree that owns a `.bots/` directory.
- **Project state** — the files under `.bots/` that capture what agents need to continue work.
- **Checkpoint** — `.bots/CHECKPOINTS.md`, the living project document that records current state, phases, decisions, and open questions.
- **Session log** — a dated Markdown file under `.bots/logs/` that records decisions from one agent session.
- **Task handoff** — a Markdown file under `.bots/tasks/` that carries work status between agents.
- **AI agent instructions** — `.bots/AGENTS.md` plus root pointers that tell agents how to use Bots in a project workspace.
- **MCP server** — the stdio protocol adapter that exposes checkpoint, session log, task handoff, and git integration operations to AI agents.
- **MCP tool** — a named operation in the MCP server catalog, with a schema, validation, and handler.
- **AI agent installer** — the interactive installer that writes MCP server configuration for supported AI agents.
- **Agent config document** — an AI agent configuration file that can contain a Bots MCP server entry.
- **Commit checkpoint** — the git integration operation that stages `.bots/` project state and commits it.
