# Refactor: Tool-Self-Reporting for Argument Summarization

## Overview
This document describes the refactoring plan to move the responsibility of summarizing tool arguments from the TUI (Terminal User Interface) to the tools themselves. 

Currently, the `TUI` package contains a centralized `summarizeArgs` method with a large `switch` statement that knows the internal structure of every tool's arguments (e.g., knowing that `git_tool` has an `action` field or `move_file` has `old_path`). This violates the Single Responsibility Principle and makes the system fragile and hard to maintain as new tools are added.

## Objective
Each tool should be responsible for defining how its arguments are reported in a human-readable summary. The `Agent` and `TUI` should remain generic, simply displaying the string provided by the tool.
This should prevent things like TUI adding quotes to arguments when not needed: eacho tool decides if and how arguments should have quotes.

## Proposed Changes

### 1. Update `internal/tools/types.go`
- Modify the `Tool` interface to include a new method:
  `SummarizeArgs(args map[string]interface{}) string`
- This method will return a short, human-readable summary of the input parameters (e.g., `'main.go'`, `add(file.go)`, or `move(old -> new)`).

### 2. Implement `SummarizeArgs` in each Tool
Each tool implementation in `internal/tools/` will implement its own summary logic:
- **File tools (`read`, `write`, `remove`, `edit`)**: Return the value of `path`.
- **`list_files`**: Return `path` or the search `pattern`.
- **`move_file`**: Return `old_path -> new_path`.
- **`git_tool`**:
    - `action: "add"` $\rightarrow$ `add(file1, file2...)`
    - `action: "diff"` $\rightarrow$ `diff`
    - `action: "commit"` $\rightarrow$ `commit(message)`
- **`go_tool`**:
    - `action: "doc"` $\rightarrow$ `doc(package_name)`
    - `action: "test"` $\rightarrow$ `test(path)`
- **`web_fetch`**: Return the `url`.

### 3. Update `internal/agent/loop.go`
- **`AgentReporter` Interface**: Change the signature of `ReportToolExecution` from:
  `ReportToolExecution(toolName string, args string, res tools.ToolResult, err error)`
  to:
  `ReportToolExecution(toolName string, summary string, res tools.ToolResult, err error)`
- **`handleToolCalls` Method**:
    - After unmarshaling the tool arguments, the agent will call `tool.SummarizeArgs(args)`.
    - The resulting summary string will be passed to `a.Reporter.ReportToolExecution`.

### 4. Simplify `internal/tui/tui.go`
- **Remove** the `summarizeArgs(toolName string, argsJSON string) string` method entirely.
- **Update** `ReportToolExecution` to use the `summary` parameter directly to format the output line.

## Benefits
- **Extensibility**: Adding a new tool only requires implementing its specific summary logic within its own package.
- **Decoupling**: The `TUI` no longer needs to know the internal field names of any tool.
- **Maintainability**: Logic for data formatting is encapsulated within the tool that owns the data.

## Implementation Steps
1. Modify `internal/tools/types.go` to update the `Tool` interface.
2. Implement `SummarizeArgs` for all existing tools in `internal/tools/`.
3. Update `internal/agent/loop.go` to use the new method.
4. Refactor `internal/tui/tui.go` to remove the old logic and use the new interface.
5. Run all tests: `go test ./... -v`.

## Status
**Completed**
