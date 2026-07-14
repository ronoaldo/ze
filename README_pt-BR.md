# Zé Agent

Zé é um agente de programação autônomo de alto desempenho construído em Go puro. Ele foi projetado para rodar localmente usando `llama.cpp` (via `llama-server`) e fornece uma TUI minimalista e poderosa baseada em ANSI.

## 🚀 Principais Recursos

- **Zero Dependências:** Construído inteiramente com a biblioteca padrão do Go. Sem excessos externos ou riscos de supply chain.
- **Loop de Agente Autônomo:** Um loop de raciocínio de várias etapas que permite ao agente planejar, executar ferramentas, observar resultados e iterar até que uma tarefa seja concluída.
- **Consciente de Contexto:** Integra automaticamente o contexto de `~/.agents/AGENTS.md` (global) e `./AGENTS.md` (local) em seu prompt de sistema.
- **Gerenciamento Inteligente de Modelos:** Detecta e seleciona automaticamente o melhor modelo disponível em seu servidor, com uma forte preferência por **Gemma 4**.
- **TUI Rica:** Uma interface de terminal minimalista com suporte para:
    - **Visibilidade de Pensamento:** Veja o processo de raciocínio do agente usando a flag `--show-thinking`.
    - **Modo Verbose:** Saída detalhada para execuções de ferramentas e comunicação bruta de API.
- **Ferramentas de Desenvolvedor Integradas:** Equipado com um conjunto de ferramentas para manipulação de arquivos, inspeção de código Go (`go_doc`), testes (`go_test`) e diffing.

## 🛠 Instalação

### Build Local
```bash
go build ./cmd/ze
```

### Build Multiplataforma (via GoReleaser)
```bash
# Instale o GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Build para todas as plataformas
goreleaser build --snapshot --clean
```

## 📖 Uso

1. **Inicie seu servidor LLM** (ex: `llama-server`):
   ```bash
   llama-server --hf-repo google/gemma-4-12B-it-qat-q4_0-gguf --port 8084
   ```

2. **Execute o Zé**:
   ```bash
   ./ze --url http://localhost:8084
   ```

## ⚙️ Configuração

### Flags de Linha de Comando

| Flag | Variável de Ambiente | Padrão | Descrição |
|---|---|---|---|
| `--url` | `LLAMA_URL` | `http://localhost:8084` | URL do servidor Llama |
| `--model` | - | | Especificar nome do modelo |
| `--timeout` | `LLAMA_TIMEOUT` | `5m` | Duração do timeout |
| `--verbose` | - | `false` | Habilita saída detalhada das ferramentas |
| `--verbose-api-calls` | - | `false` | Log de requisições/respostas brutas da API |
| `--max-iterations` | - | `50` | Número máximo de iterações do agente |
| `--show-thinking` | - | `false` | Mostra o processo de raciocínio na interface |
| `--no-color` | - | `false` | Desabilita a saída colorida |
| `--version` / `-v` | - | - | Mostra informações da versão |

### Comandos de Barra (Slash Commands)

Dentro da TUI, você pode usar os seguintes comandos:

- `/help`: Mostra os comandos disponíveis.
- `/quit` ou `/exit`: Sai da sessão.

## 🛠 Capacidades do Agente (Ferramentas)

O agente pode interagir com seu ambiente usando as seguintes ferramentas:

- `read_file`: Lê o conteúdo de um arquivo.
- `write_file`: Escreve ou sobrescreve um arquivo.
- `list_files`: Lista arquivos em um diretório.
- `remove_file`: Deleta um arquivo.
- `edit_file`: Realiza edições precisas e atômicas em arquivos.
- `go_doc`: Inspeciona a documentação Go para pacotes e funções.
- `go_test`: Executa testes Go no diretório atual.
- `diff`: Mostra as mudanças entre arquivos ou no estado atual.

## 🛠 Solução de Problemas

- **Erro de Conexão:** Certifique-se de que o `llama-server` está rodando e é acessível na URL fornecida via `--url`.
- **Modelo não encontrado:** Verifique se o nome do modelo está correto ou permita que o Zé o detecte automaticamente.
- **Permissão Negada:** Certifique-se de que o agente tem permissões de leitura/escrita no diretório em que está operando.

## 📜 Licença

MIT
