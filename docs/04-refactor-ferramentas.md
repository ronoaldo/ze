# Plano de Refatoração e Expansão de Ferramentas (Zé Agent)

## Objetivo Geral
Simplificar o conjunto de ferramentas (toolset) disponíveis para o agente, reduzindo a carga cognitiva do LLM e a redundância de definições, ao mesmo tempo em que aumentamos a capacidade de manipulação de arquivos e navegação no sistema.

---

## 🛡️ Fase 0: Preservação e Refatoração de Testes (Obrigatório)
**Antes de qualquer alteração no código funcional, devemos garantir a integridade do sistema.**

1.  **Mapeamento de Cobertura:** Analisar todos os arquivos `_test.go` atuais para entender os casos de uso de cada ferramenta (`diff`, `git_add`, `git_commit`, `go_doc`, `go_test`).
2.  **Refatoração de Testes de Unidade:** 
    *   Transformar os testes existentes em "Testes de Regressão".
    *   Atualizar os testes de unidade para que eles não dependam mais das funções/structs antigas, mas sim das novas implementações unificadas (`git_tool` e `go_tool`).
    *   **Regra de Ouro:** O código de produção não deve ser alterado até que os testes atuais estejam preparados para validar a nova estrutura.
3.  **Criação de Testes de Integração:** Criar um diretório `testdata` com repositórios Git e projetos Go reais para garantir que a lógica de comandos (como o `go list | while read...` do `go_doc`) continue funcionando perfeitamente na nova estrutura.

---

## 🛠️ Fase 1: Consolidação de Ferramentas Git (`git_tool.go`)
Unificar as ações de Git em uma única interface de comando.

*   **Nova Ferramenta:** `git_tool`
*   **Parâmetro `action` (enum):**
    *   `diff`: Retorna o status e as mudanças (staged/unstaged/untracked).
    *   `add`: Adiciona arquivos ao staging area (incluindo suporte para `git add .`).
    *   `commit`: Realiza o commit com uma mensagem específica.

---

## 🐹 Fase 2: Consolidação de Ferramentas Go (`go_tool.go`)
Unificar ferramentas de inspeção e validação de código Go.

*   **Nova Ferramenta:** `go_tool`
*   **Parâmetro `action` (enum):**
    *   `doc`: Consulta documentação via `go doc` (incluindo o suporte para `all`).
    *   `test`: Executa testes (`go test`) no caminho especificado.

---

## 🔍 Fase 3: Evolução do `list_files`
Transformar a listagem simples em uma ferramenta de busca poderosa.

*   **Novos Parâmetros:**
    *   `pattern` (string, opcional): Suporte a **Glob Patterns** (ex: `**/*.go`, `config/*.yaml`).
    *   `recursive` (boolean, padrão `false`): Se `true`, realiza a busca em toda a árvore de diretórios a partir do `path` fornecido.

---

## 📂 Fase 4: Nova Ferramenta `move_file`
Adicionar capacidade de manipulação de sistema de arquivos para renomeação e movimentação.

*   **Nova Ferramenta:** `move_file`
*   **Parâmetros:**
    *   `old_path` (string): Caminho atual do arquivo ou diretório.
    *   `new_path` (string): Novo caminho ou nome de destino.

---

## 🧹 Fase 5: Limpeza e Validação Final

1.  **Remoção de Legado:** Deletar os arquivos antigos que foram unificados:
    *   `diff.go`, `git_add.go`, `git_commit.go`
    *   `godoc.go`, `gotest.go`
2.  **Verificação de Tipos:** Atualizar `internal/tools/types.go` com as novas structs de argumentos.
3.  **Validação de Sistema:**
    *   Executar `go test ./... -v` para garantir que todas as ferramentas (incluindo as novas) estão operacionais.

---

## 📝 Fase 6: Atualização de Documentação

Para garantir que tanto usuários humanos quanto o próprio agente (através do contexto de sistema) estejam alinhados com as novas capacidades.

1.  **Atualização de README / README_pt-BR:**
    *   Atualizar a seção de "Capacidades" ou "Ferramentas" para listar as novas ferramentas unificadas (`git_tool`, `go_tool`, `move_file`) e a versão aprimorada do `list_files`.
    *   Exemplos de uso atualizados para as novas ferramentas.
2.  **Atualização do `AGENTS.md`:**
    *   **CRÍTICO:** Atualizar a seção `Agent Capabilities (Tools)` para refletir a nova estrutura de JSON Schema das ferramentas. O agente utiliza este arquivo para entender o que ele é capaz de fazer.
    *   Se o `AGENTS.md` detalha os parâmetros, os novos campos (como `pattern` e `recursive` no `list_files`) devem ser documentados aqui.
