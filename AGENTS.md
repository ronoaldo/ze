# AGENTS.md — Zé Agent

## Overview
Zé is a pure Go CLI AI agent (zero external dependencies) that connects to a `llama.cpp` server via an OpenAI-compatible API. It features a multi-stage loop with tool-use and an ANSI-based TUI.

## Tech Stack
- **Language:** Go 1.25+ (Standard Library preferred)
- **Build:** `go build ./...`
- **Test:** `go test ./... -v`
- **Release:** GoReleaser v2
- **LLM:** llama.cpp `llama-server` (OpenAI-compatible)

## Essential Commands
```bash
go build ./cmd/ze          # Build the project
go test ./... -v            # Run all tests
./ze --url <url>            # Run the agent
```

## Project Structure
- `cmd/ze/`: Application entry point.
- `internal/agent/`: Core agent logic, including the reasoning loop and shell interaction.
- `internal/commands/`: Handles slash commands within the TUI.
- `internal/llm/`: Manages communication with the LLM server (OpenAI-compatible API).
- `internal/prompt/`: Contains system prompts and conversation management.
- `internal/tools/`: Implementations of all agent capabilities (e.g., file, git, web, go).
- `internal/tui/`: Terminal User Interface and ANSI-based rendering.

## Tooling Protocols (CRITICAL)

### `edit_file` Protocol
- Always call `read_file` before `edit_file`.
- `oldString` must be a bit-for-bit copy of the original content (including whitespace/tabs/newlines).
- Use a sufficiently long `oldString` to ensure it is unique within the file.
- Perform small, focused edits. Avoid large blocks.
- When providing multiple edits, order them from top to bottom.
- Maintain original indentation (use Tabs if the file uses Tabs).
- If the change is >10 lines, use `write_file` instead.

### `git_commit` Protocol
- **NEVER** call `git_commit` without explicit user approval of the commit message and the action.
- Do not use `git_commit` just to generate a message; use it to actually perform the commit.
- Always verify the status with `git diff` or `git status` before committing to ensure all intended changes are staged.

### `go_test` Protocol
- **NEVER** call it with '.': the root folder has no go code.
- Prefer to run all tests with './...' as argument since they run fast and this detects any compilation errors.

## Agent Capabilities (Tools)

The agent can interact with the environment using the following tools:

- `read_file`: Read the content of a file.
- `write_file`: Write or overwrite a file.
- `list_files`: List files and directories (supports glob patterns and recursive search).
- `remove_file`: Delete a file.
- `edit_file`: Perform precise, atomic edits on files.
- `go_tool`: Inspect Go code (docs or testing).
- `web_fetch`: Fetch content from web URLs.
- `git_tool`: Manage git operations (diff, add, commit).
- `move_file`: Rename or move files and directories.

## Boundaries
### ALWAYS:
- Run `go test ./... -v` before finalizing changes.
- Keep zero external dependencies.
- Isolate file tests using `t.TempDir()`.

### ASK BEFORE:
- Adding new external dependencies.
- Modifying `docs/` or `README.md`.
- Changing the tool structure.

### NEVER:
- Touch secrets, `.env`, or credentials.
- Commit binaries (`dist/`, `*.exe`).
- Create `package main` outside of `cmd/`.
- Use frontend frameworks (React, Vue, etc.).
- Ignore Go errors.
