# AGENTS.md — Zé Agent

## O que é

Zé é um agente CLI de IA escrito em Go puro (zero dependências externas) que se conecta a um servidor `llama.cpp` via API OpenAI-compatível. Ele executa loops multi-estágio com tool-use, possui uma TUI baseada em ANSI e detecta hardware local para selecionar o melhor modelo GGUF.

## Stack

- **Linguagem:** Go 1.25+
- **Build:** `go build ./...` (zero dependências externas)
- **Testes:** `go test ./... -v` (21 testes unitários)
- **Release:** GoReleaser v2 (`goreleaser build --snapshot --clean`)
- **LLM:** llama.cpp `llama-server` (API OpenAI-compatível, padrão `http://localhost:8080`)
- **Modelo:** GGUF (Gemma 4 instruction-tuned recomendado)

## Comandos essenciais

```bash
# Compilar
go build ./cmd/ze

# Testar
go test ./... -v

# Build multi-plataforma (snapshot)
goreleaser build --snapshot --clean

# Executar (requer terminal TTY real)
./ze --url http://localhost:8080

# Versão
./ze --version
```

## Estrutura do projeto

```
cmd/ze/main.go          ← Entry point: CLI flags, hardware detection, model selection, TUI
internal/agent/loop.go  ← Agent loop multi-estágio com tool-use (guarda 20 iterações)
internal/llm/client.go  ← Cliente OpenAI-compatível para llama.cpp
internal/llm/hardware.go← Detecção de GPU/CPU/RAM + seleção de modelo GGUF
internal/prompt/prompt.go ← System prompt otimizado para Gemma 4
internal/tools/         ← Tools: file_read, file_write, list_files, go_doc
internal/tui/           ← TUI ANSI: raw mode, line editing, scroll buffer, SIGWINCH
internal/tui/tui_linux.go   ← Linux: TCGETS/TCSETS, TIOCGWINSZ
internal/tui/tui_darwin.go  ← Darwin: ioctl constants nativas
internal/tui/tui_windows.go ← Windows: stub (raw mode não suportado)
```

## Padrões de código

- **Erros:** sempre verifique `if err != nil` — nunca ignore erros
- **Importação:** organize em grupos (padrão, internal, third-party)
- **Nomes:** funcões exportadas com verbos (`NewAgent`, `SelectBestModel`)
- **Testes:** coloque em `*_test.go` no mesmo pacote; use `t.Helper()` em helpers
- **TUI:** use apenas ANSI escape codes nativos — nenhuma biblioteca externa
- **Agent loop:** limite de 20 iterações; formato `TOOL_CALL:tool_name{json}`
- **Config:** prioridade `-url` flag > `LLAMA_URL` env > `http://localhost:8080`
- **Modelo:** prioridade `status.value == "loaded"` > hardware detection fallback

## Testes

- Execute `go test ./... -v` antes de qualquer commit
- Todos os 21 testes devem passar
- Ferramentas de arquivo usam `BaseDir` para isolar testes do host filesystem
- Não adicione junk files (`a.go`, `b.go`, etc.) em `internal/agent/`

## Git workflow

- Mensagens de commit: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`
- Um commit = uma mudança lógica
- Nunca commitar segredos, chaves ou arquivos gerados (`dist/`, `*.exe`)

## Fronteiras

**SEMPRE:**
- Executar `go test ./... -v` antes de finalizar qualquer mudança
- Manter zero dependências externas (exceto Go padrão)
- Isolar testes com `t.TempDir()` para ferramentas de arquivo

**PEDIR ANTES:**
- Adicionar novas dependências externas
- Modificar arquivos em `docs/` ou `README.md`
- Alterar a estrutura de ferramentas do agent

**NUNCA:**
- Touch segredos, `.env`, ou credenciais
- Commitar binários em `dist/`
- Criar arquivos com `package main` fora de `cmd/`
- Usar React, Vue, Tailwind ou bibliotecas de frontend (este projeto é Go puro)
- Ignorar erros de Go (`if err != nil` é obrigatório)
