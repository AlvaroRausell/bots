# Changelog

All notable changes to this project are documented here.

## [0.1.0] - 2026-04-26

### Added
- Added structured workspace-aware stores for checkpoints, logs, tasks, installs, and git integration.
- Added test coverage for checkpoint, log, task, install config, MCP catalog, project initialization, git, document, and workspace behavior.
- Added project documentation under `docs/`.

### Changed
- Refactored MCP tooling around a shared catalog so schemas and handlers stay co-located.
- Refactored project initialization and document updates to reuse shared document/workspace helpers.
- `bots log append` now clusters entries under an existing date section instead of creating duplicate same-day sections.
- Log/task lookup now rejects blank slugs instead of matching the first available file.

### Fixed
- Log append date clustering now ignores markdown headings inside fenced code blocks.
- MCP string argument validation now rejects blank required strings while preserving intentionally empty checkpoint content.

## [0.0.1] - 2026-04-23

### Added
- Initial release of the Bots CLI, MCP server, GoReleaser configuration, and GitHub Actions release workflow.
