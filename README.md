# Zé Agent

Zé is a high-performance, autonomous programming agent built in pure Go. It is designed to run locally using `llama.cpp` (via `llama-server`) and provides a minimal, powerful ANSI-based TUI.

## 🚀 Key Features

- **Zero Dependencies:** Built entirely with the Go standard library. No external bloat or supply chain risks.
- **Autonomous Agent Loop:** A multi-step reasoning loop that allows the agent to plan, execute tools, observe results, and iterate until a task is complete.
- **Context-Aware:** Automatically integrates context from `~/.agents/AGENTS.md` (global) and `./AGENTS.md` (local) into its system prompt.
- **Smart Model Management:** Automatically detects and selects the best available model on your server, with a strong preference for **Gemma 4**.
- **Rich TUI & Markdown:** A terminal interface that supports Markdown rendering (tables, lists, formatting) for readable agent responses.
- **Headless Mode:** Automatic detection of non-interactive environments (pipes/redirects), switching to a simplified, plain-text prompt (`prompt > `) without banners or colors.
- **Thinking Visibility:** View the agent's reasoning process using the `--show-thinking` flag.
- **Verbose Mode:** Detailed output for tool executions and raw API communication.

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

### Advanced Interaction
- **Shell Commands:** Run any shell command by prefixing it with `!` (e.g., `!ls`). The command's output is automatically added to the conversation context.
- **Multiline Input:** Use `/multiline` to enter or paste long blocks of text. Type `/send` on a new line to finish and submit.

## ⚙️ Configuration

### Command Line Flags

| Flag | Environment Variable | Default | Description |
|---|---|---|---|
| `--url` | `LLAMA_URL` | `http://localhost:8084` | Llama server URL |
| `--model` | - | | Specify model name |
| `--timeout` | `LLAMA_TIMEOUT` | `5m` | Timeout duration |
| `--verbose` | - | `false` | Enable verbose tool output |
| `--verbose-api-calls` | - | `false` | Log raw API requests/responses |
| `--max-iterations` | - | `50` | Maximum number of agent iterations |
| `--show-thinking` | - | `false` | Show reasoning process in the UI |
| `--no-color` | - | `false` | Disable color output |
| `--version` / `-v` | - | - | Show version information |

### Slash Commands

While inside the TUI, you can use the following commands:

- `/help`: Show available commands.
- `/multiline`: Start a multi-line input session.
- `/quit` or `/exit`: Exit the session.

## 🛠 Agent Capabilities (Tools)

The agent can interact with your environment using the following tools:

- `read_file`: Read the content of a file.
- `write_file`: Write or overwrite a file (reports total bytes written).
- `list_files`: List files in a directory.
- `remove_file`: Delete a file.
- `edit_file`: Perform precise, atomic edits on files (shows change summary).
- `go_doc`: Inspect Go documentation (`go_doc('all')` for full API inspection).
- `go_test`: Run Go tests (displays error output on failure).
- `diff`: Show detailed statistics of changes (staged, unstaged, and untracked).
- `web_fetch`: Fetch content from web URLs (HTML, JSON, Markdown, etc.).
- `git_add`: Add files to the git staging area (supports specific files or all changes via '.').
- `git_commit`: Commit changes (requires explicit user confirmation).

## 🛠 Troubleshooting

- **Connection Error:** Ensure `llama-server` is running and accessible at the URL provided via `--url`.
- **Model Not Found:** Verify that the model name is correct or allow Zé to detect it automatically.
- **Permission Denied:** Ensure the agent has read/write permissions in the directory it is operating in.

## 📜 License

MIT
