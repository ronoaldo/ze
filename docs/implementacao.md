# Plano de Implementação: Zé (Versão Zero)

O Zé é um agente de programação desenhado para atuar com modelos locais através do `llama.cpp`, focado em eficiência e simplicidade.

## Avaliação do Documento Base
O projeto é bem definido e focado em execução prática. O diferencial do Zé é a otimização para modelos locais (`llama.cpp`) e a curadoria de ferramentas específicas para engenharia de software (Go doc, Git, Exa AI, MCP). A escolha de Go para a TUI e a meta de "mínimas dependências" sugerem um agente de alta performance e baixa latência.

---

## Fases de Desenvolvimento

### Fase 1: Core e Infraestrutura (Base)
1.  **Setup do Projeto Go:** Estruturação de workspace, gerenciamento de dependências e configuração de logs.
2.  **Motor de TUI:** Implementação de uma interface de terminal responsiva para exibição de chat, status do modelo e logs de ferramentas.
3.  **Integração llama.cpp:** Camada de abstração para comunicação com o `llama.cpp` via CLI ou bindings, suportando streaming de tokens.
4.  **Gerenciador de Modelos:**
    *   Módulo de detecção de hardware (CPU cores, GPU, VRAM).
    *   Catálogo de modelos Gemma 4 (URLs de download e mapeamento de quantizações).
    *   Downloader paralelo com suporte a *resume*.

### Fase 2: O "Cérebro" (Orquestração e Prompt)
1.  **Sistema de Prompt:** Implementação do System Prompt otimizado para os modelos Gemma 4 (focado em *chain-of-thought* para código e formato estrito de *tool calling*).
2.  **Loop de Agente:** Implementação do ciclo `Input -> LLM -> Tool Selection -> Execution -> Observation -> LLM`.
3.  **Gerenciamento de Contexto:** Janelamento de histórico e resumo de logs de ferramentas para manter a coerência.

### Fase 3: Ecossistema de Ferramentas (Toolkits)
1.  **File System Tool:** Operações CRUD em arquivos, leitura de diretórios e busca por regex.
2.  **Git Tool:** Wrapper para comandos `git` (commit, push, pull, log, status).
3.  **Web Search (Exa AI):** Integração com API para busca semântica de documentação.
4.  **Go Doc Tool:** Integração com `go doc` para inspeção de bibliotecas Go.
5.  **MCP Client:** Implementação do protocolo Model Context Protocol para expansão de ferramentas.

### Fase 4: Refinamento e Otimização
1.  **Geração "Lean":** Ajustes finos no prompt para evitar explicações excessivas e focar em código funcional.
2.  **Multi-platform Build:** Configuração do Go Release para compilação cruzada (Windows, Linux, macOS).

### Fase 5: Verificação e Deploy
1.  **Testes de Integração:** Testes unitários para cada ferramenta individualmente.
2.  **E2E Test:** Execução do cenário do servidor Minecraft.

---

## Critérios de Aceitação Detalhados (E2E)

Para validar a conclusão da "Versão Zero", o projeto deve cumprir os seguintes requisitos técnicos:

**1. Automação de Modelos:**
*   O usuário deve iniciar o Zé sem modelos baixados.
*   O Zé deve detectar a VRAM (ex: 8GB) e sugerir/baixar automaticamente a quantização correta do Gemma 4.
*   O download deve ser resiliente a interrupções.

**2. Capacidades de Código:**
*   O Zé deve criar uma pasta de projeto e gerar um arquivo `main.go` funcional.
*   O Zé deve utilizar a ferramenta de Git para iniciar um repositório e realizar o primeiro commit automaticamente após a criação do código.

**3. Uso de Ferramentas:**
*   **Busca:** O Zé deve ser capaz de buscar a documentação da biblioteca `net/http` via Exa AI se encontrar dúvidas sobre conexões.
*   **Go Doc:** O Zé deve conseguir listar as funções de um pacote Go específico usando `go doc`.

**4. Teste Final de Verificação (Obrigatório):**
*   **Input:** "Crie uma API em Go que receba um endereço de servidor Minecraft via query string e retorne um JSON com status e os 18 primeiros jogadores."
*   **Resultado Esperado:** 
    *   O Zé deve planejar a estrutura.
    *   O Zé deve criar o código.
    *   O Zé deve rodar o código (ou fornecer instruções claras de como rodar).
    *   O código deve efetivamente conectar ao servidor e retornar o JSON correto.
*   **Sucesso:** A API deve estar funcional e o JSON deve ser válido conforme o esquema solicitado.
