# Zé Agent

Zé é um agente de programação local que opera via `llama.cpp` (llama-server) com uma TUI minimalista em Go puro — **zero dependências externas**.

## Recursos

- TUI com chat e input (ANSI escape codes, sem bibliotecas)
- Loop multi-step com tool-use (`read_file`, `write_file`, `edit_file`, `list_files`, `go_doc`)
- Detecção automática de hardware e seleção de melhor modelo carregado no llama-server
- System prompt otimizado para Gemma 4
- Build multi-plataforma via GoReleaser

## Instalação

```bash
# Via GoReleaser (release)
go install github.com/goreleaser/goreleaser/v2@latest

# Build local
go build ./cmd/ze

# Build multi-plataforma (snapshot)
goreleaser build --snapshot --clean
```

## Uso

```bash
# Inicie o llama-server primeiro
llama-server --hf-repo google/gemma-4-12B-it-qat-q4_0-gguf --port 8080

# Execute o Zé
./ze --url http://localhost:8080
```

## Configuração

| Método | Valor Padrão |
|---|---|
| CLI flag `-url` | `http://localhost:8080` |
| Env `LLAMA_URL` | `http://localhost:8080` |
| Default | `http://localhost:8080` |

| Método | Valor Padrão |
|---|---|
| CLI flag `-timeout` | `60s` |
| Env `LLAMA_TIMEOUT` | `60s` |
| Default | `60s` |

## Build Multi-Plataforma

```bash
goreleaser release --snapshot
```

Gera executáveis para:
- Linux (amd64, arm64)
- macOS (arm64)
- Windows (amd64, arm64)

## Licença

MIT
