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

## Agent Capabilities (Tools)

The agent can interact with the environment using the following tools:

- `read_file`: Read the content of a file.
- `write_file`: Write or overwrite a file.
- `list_files`: List files in a directory.
- `remove_file`: Delete a file.
- `edit_file`: Perform precise, atomic edits on files.
- `go_doc`: Inspect Go documentation.
- `go_test`: Run Go tests.
- `diff`: Show detailed statistics of changes.
- `web_fetch`: Fetch content from web URLs.
- `git_add`: Add files to the git staging area.
- `git_commit`: Commit staged changes (requires explicit user approval).

## Coding Standards
- **Error Handling:** Always check `if err != nil`. Never ignore errors.
- **Dependencies:** Use the Go standard library. Avoid external dependencies.
- **Naming:** Use meaningful names; exported functions should use verbs.
- **Tests:** Use `*_test.go` files in the same package. Use `t.Helper()`.
- **Test Preservation:** NUNCA remova testes existentes, a menos que eles estejam verificando uma funcionalidade que foi removida. Testes devem sempre ser ajustados conforme necessário à medida que o projeto evolui.

## Testing & Git
- **Testing:** Always run `go test ./... -v` before committing. Use `t.TempDir()` for file-related tests.
- **Commits:** Use Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`). One commit per logical change.
- **Make AI use explicit:** Add the Co-Authored-By: Tool (with Model Name). Replace Tool by the Agent name and Model Name by the model Version.
 - Example: if using Claude with Fable 5: Co-Authored-By: Claude (with Fable 5)
 - Example: if using Gemma 4 with Ze: Co-Authored-By: Ze (with Gemma 4)
 - Example: if using Gemini with Antigravity: Co-Authored-By: Antigravity (with Gemini 3.5 Flash)

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
