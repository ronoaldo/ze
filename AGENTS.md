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
- `cmd/ze/`: Entry point.
- `internal/agent/`: Core agent logic (loop, history).
- `internal/commands/`: Slash commands handler.
- `internal/llm/`: LLM client & hardware detection.
- `internal/prompt/`: System prompts.
- `internal/tools/`: Tool implementations (file operations, etc.).
- `internal/tui/`: ANSI-based TUI.

## Tooling: `edit_file` Protocol (CRITICAL)
When using `edit_file`, follow these rules strictly:
1. **Exact Match:** `oldString` must be a bit-for-bit copy of the original content (including whitespace/tabs/newlines).
2. **Uniqueness:** Use a sufficiently long `oldString` to ensure it is unique within the file.
3. **Atomicity:** Perform small, focused edits. Avoid large blocks.
4. **Ordering:** When providing multiple edits, order them from top to bottom.
5. **Indentation:** Maintain original indentation (use Tabs if the file uses Tabs).
6. **Pre-requisite:** Always call `read_file` before `edit_file`. If the change is >10 lines, use `write_file` instead.

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
