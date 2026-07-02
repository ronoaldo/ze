# Zé Agent

Zé é um agente de programação local que opera via `llama.cpp` (llama-server) com uma TUI minimalista em Go puro — **zero dependências externas**.

## Recursos

- TUI com chat e input (ANSI escape codes, sem bibliotecas)
- Loop multi-step com tool-use (write, read, list, go doc)
- Detecção automática de modelo carregado no llama-server
- System prompt otimizado para Gemma 4
- Build multi-plataforma via GoReleaser

## Instalação

```bash
# Via GoReleaser (release)
go install github.com/goreleaser/goreleaser/v2@latest

# Build local
goreleaser build --snapshot

# Release completo
goreleaser release --snapshot
```

## Uso

```bash
# Inicie o llama-server primeiro
llama-server --hf-repo google/gemma-4-12B-it-qat-q4_0-gguf --port 8080

# Execute o Zé
./ze
```

## Configuração

| Método | Valor Padrão |
|---|---|
| CLI flag `-url` | `http://localhost:8080` |
| Env `LLAMA_URL` | `http://localhost:8080` |
| Default | `http://localhost:8080` |

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
