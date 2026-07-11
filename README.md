# Zé Agent

Zé is a high-performance, autonomous programming agent built in pure Go. It is designed to run locally using `llama.cpp` (via `llama-server`) and provides a minimal, powerful ANSI-based TUI.

## 🚀 Key Features

- **Zero Dependencies:** Built entirely with the Go standard library. No external bloat or supply chain risks.
- **Autonomous Agent Loop:** A multi-step reasoning loop that allows the agent to plan, execute tools, observe results, and iterate until a task is complete.
- **Context-Aware:** Automatically integrates context from `~/.agents/AGENTS.md` (global) and `./AGENTS.md` (local) into its system prompt.
- **Smart Model Management:** Automatically detects and selects the best available model on your server, with a strong preference for **Gemma 4**.
- **Rich TUI:** A minimal terminal interface with support for:
    - **Thinking Visibility:** View the agent's reasoning process using the `--show-thinking` flag.
    - **Verbose Mode:** Detailed output for tool executions and raw API communication.
- **Built-in Developer Tools:** Equipped with a suite of tools for file manipulation, Go code inspection (`go_doc`), testing (`go_test`), and diffing.

## 🛠 Installation

### Local Build
```bash
go build ./cmd/ze
```

### Multi-platform Build (via GoReleaser)
```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Build all platforms
goreleaser build --snapshot --clean
```

## 📖 Usage

1. **Start your LLM server** (e.g., `llama-server`):
   ```bash
   llama-server --hf-repo google/gemma-4-12B-it-qat-q4_0-gguf --port 8084
   ```

2. **Run Zé**:
   ```bash
   ./ze --url http://localhost:8084
   ```

## ⚙️ Configuration

### Command Line Flags

| Flag | Environment Variable | Default | Description |
|---|---|---|---|
| `--url` | `LLAMA_URL` | `http://localhost:8084` | Llama server URL |
| `--timeout` | `LLAMA_TIMEOUT` | `5m` | Timeout duration |
| `--verbose` | - | `false` | Enable verbose tool output |
| `--verbose-api-calls` | - | `false` | Log raw API requests/responses |
| `--max-iterations` | - | `50` | Maximum number of agent iterations |
| `--show-thinking` | - | `false` | Show reasoning process in the UI |
| `--version` / `-v` | - | - | Show version information |

### Slash Commands

While inside the TUI, you can use the following commands:

- `/help`: Show available commands.
- `/quit` or `/exit`: Exit the session.

## 🛠 Agent Capabilities (Tools)

The agent can interact with your environment using the following tools:

- `read_file`: Read the content of a file.
- `write_file`: Write or overwrite a file.
- `list_files`: List files in a directory.
- `remove_file`: Delete a file.
- `edit_file`: Perform precise, atomic edits on files.
- `go_doc`: Inspect Go documentation for packages and functions.
- `go_test`: Run Go tests in the current directory.
- `diff`: Show changes between files or current state.

## 📜 License

MIT
