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
internal/agent/         ← Core do agente: loop multi-estágio, gestão de histórico e estatísticas
internal/commands/      ← Gerenciador de comandos de barra (ex: /quit, /help)
internal/llm/           ← Cliente OpenAI-compatível e detecção de hardware/modelos
internal/prompt/        ← System prompts otimizados para o modelo (Gemma 4)
internal/tools/         ← Implementação das ferramentas (file_read, file_write, edit_file, etc.)
internal/tui/           ← TUI ANSI: raw mode, line editing, scroll buffer, SIGWINCH
```

## Ferramentas (Tools)

O agente utiliza ferramentas baseadas em JSON Schema para interagir com o sistema:

- `read_file`: Lê o conteúdo de um arquivo.
- `write_file`: Escreve conteúdo em um arquivo (sobrescreve).
- `edit_file`: Aplica edições parciais em arquivos usando substituição de strings. REQUER CORRESPONDÊNCIA EXATA (caractere por caractere) de `oldString` para garantir sucesso. Requer atenção absoluta a espaços, tabs e quebras de linha.
- `list_files`: Lista arquivos e diretórios.
- `go_doc`: Recupera documentação de pacotes Go via `go doc`.

### Protocolo de Edição de Arquivos (CRÍTICO)

Ao utilizar `edit_file`, siga RIGOROSAMENTE estas regras:

1. **CORRESPONDÊNCIA EXATA (Exact Match):** O `oldString` deve ser uma cópia bit-a-bit do conteúdo original, incluindo todos os espaços, tabs e quebras de linha. Mesmo um único caractere divergente causará falha.
2. **UNICIDADE (Uniqueness):** Escolha um `oldString` longo o suficiente para ser único no arquivo. Evite palavras comuns ou linhas curtas que se repetem. Use o contexto ao redor para tornar o identificador unívoco.
3. **ATOMICIDADE (Atomicity):** Realize edições pequenas e focadas. Não tente reescrever grandes blocos; quebre alterações complexas em vários `edit_file` menores e sequenciais.
4. **ORDENAÇÃO (Ordering):** Ao enviar múltiplos edits no array `edits`, organize-os da primeira ocorrência para a última (ordem ascendente no arquivo) para evitar deslocamentos de índice.
5. **PRESERVAÇÃO DE INDENTAÇÃO:** O `newString` deve manter a indentação exata do código original (utilize Tabs se o arquivo usar Tabs).
6. **ESTRATÉGIA DE SELEÇÃO:** Antes de usar `edit_file`, chame obrigatoriamente `read_file` para obter a representação exata do texto. Se a alteração for muito extensa (>10 linhas), utilize `write_file` com o conteúdo completo para maior segurança.

## Arquitetura e Design

### Agent Loop & Reporter
O agente opera em um loop de até 20 iterações. Para manter a interface de usuário (TUI) informada sem acoplar o agente diretamente à UI, utiliza-se a interface `AgentReporter`.
- O agente chama métodos como `ReportToolCall` e `ReportToolResult`.
- A TUI implementa essa interface para renderizar o progresso em tempo real.

### Slash Commands
O sistema de comandos de barra (`/command`) é processado antes do loop do agente. Se um comando for detectado, ele é executado pelo pacote `internal/commands`. Se o comando retornar `ErrQuit`, o programa encerra; caso contrário, o input é tratado como uma mensagem para o agente.

### Model Selection
A seleção de modelo segue uma hierarquia de prioridade:
1. Modelo Gemma 4 já carregado no servidor.
2. Qualquer outro modelo carregado no servidor.
3. Fallback para detecção de hardware local para sugerir o melhor modelo GGUF disponível.

## Padrões de código

- **Erros:** sempre verifique `if err != nil` — nunca ignore erros.
- **Importação:** organize em grupos (padrão, internal, third-party).
- **Nomes:** funções exportadas com verbos (`NewAgent`, `SelectBestModel`).
- **Testes:** coloque em `*_test.go` no mesmo pacote; use `t.Helper()` em helpers.
- **TUI:** use apenas ANSI escape codes nativos — nenhuma biblioteca externa.
- **Agent loop:** limite de 20 iterações; formato de tool call via JSON.
- **Config:** prioridade `-url` flag > `LLAMA_URL` env > `http://localhost:8080`.

## Testes

- Execute `go test ./... -v` antes de qualquer commit.
- Todos os testes devem passar.
- Ferramentas de arquivo usam `BaseDir` para isolar testes do host filesystem.
- Não adicione junk files (`a.go`, `b.go`, etc.) em `internal/agent/`.

## Git workflow

- Mensagens de commit: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`
- Um commit = uma mudança lógica.
- Nunca commitar segredos, chaves ou arquivos gerados (`dist/`, `*.exe`).

## Fronteiras

**SEMPRE:**
- Executar `go test ./... -v` antes de finalizar qualquer mudança.
- Manter zero dependências externas (exceto Go padrão).
- Isolar testes com `t.TempDir()` para ferramentas de arquivo.

**PEDIR ANTES:**
- Adicionar novas dependências externas.
- Modificar arquivos em `docs/` ou `README.md`.
- Alterar a estrutura de ferramentas do agent.

**NUNCA:**
- Touch segredos, `.env`, ou credenciais.
- Commitar binários em `dist/`.
- Criar arquivos com `package main` fora de `cmd/`.
- Usar React, Vue, Tailwind ou bibliotecas de frontend.
- Ignorar erros de Go (`if err != nil` é obrigatório).
