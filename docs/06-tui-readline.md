# Plano de Implementação: Zé-Readline (TUI Advanced Input)

Este documento detalha o plano de implementação de uma nova camada de entrada interativa (readline) para o Zé, permitindo funcionalidades avançadas de edição de linha, histórico e navegação, mantendo a filosofia de **zero dependências externas** e alta performance.

## 1. Objetivos

Implementar uma interface de linha de comando (CLI) interativa que suporte:
- **Edição de linha:** Mover cursor (setas), apagar caracteres (Backspace/Delete), apagar palavras, mover para início/fim da linha (Home/End).
- **Histórico de comandos:** Navegação por setas (Cima/Baixo) entre comandos executados anteriormente.
- **Suporte a sequências de escape:** Captura correta de teclas especiais (setas, Home, End, Del).
- **Controle de linha:** Atalhos de teclado como `Ctrl+L` (limpar tela) e `Ctrl+C` (interromper).
- **Robustez:** Garantir que o terminal retorne ao modo normal mesmo em caso de erro ou pânico.
- **Restrições:** Implementação em Go puro utilizando apenas a biblioteca padrão (`os`, `syscall`, `fmt`, `io`).

## 2. Arquitetura do Componente

O sistema será dividido em quatro camadas principais para garantir modularidade e testabilidade.

### A. Camada de Terminal (Raw Mode)
Responsável por alternar o terminal do modo "cooked" (padrão) para o modo "raw" (cru).
- **Mecanismo:** Uso de `syscall.IoctlSetTermios` para desabilitar `ECHO` (evita duplicidade de caracteres) e `ICANON` (permite leitura caractere a caractere sem esperar pelo Enter).
- **Segurança:** Implementação de um mecanismo de recuperação (`recover`) para restaurar o terminal original em caso de crash.

### B. Camada de Buffer de Entrada (`InputBuffer`)
Gerencia o estado do texto digitado e a posição do cursor.
- **Estado:**
    - `content`: `[]rune` (para suporte nativo a UTF-8 e caracteres especiais).
    - `cursorPos`: `int` (índice atual do cursor dentro do buffer).
- **Métodos:** `Insert(r rune)`, `DeleteChar()`, `DeleteWord()`, `MoveCursor(pos int)`, `GetLine() string`.

### C. Camada de Histórico (`HistoryManager`)
Armazena e gerencia a lista de comandos executados.
- **Estado:** `entries []string` e `currentIndex int`.
- **Métodos:** `Add(line string)`, `Prev()`, `Next()`.

### D. Camada de Renderização (ANSI Engine)
Responsável por desenhar o prompt e a linha atual no terminal de forma eficiente.
- **Técnica:** Uso de sequências de escape ANSI (ex: `\x1b[2K` para limpar linha, `\x1b[G` para mover cursor) para evitar o efeito de *flickering* (piscadas na tela).

## 3. Plano de Execução

### Fase 1: O Core do Terminal (Baixo Nível)
- Implementar o controle de modo `raw` via `syscall` (compatível com ambientes POSIX/Wine).
- Implementar o handler de sinais para captura de `SIGINT`.

### Fase 2: O Buffer e Lógica de Teclas
- Implementar o loop principal de leitura de bytes (`os.Stdin.Read`).
- Implementar o *parser* de sequências de escape (identificação de setas, Home, End).
- Implementar a lógica de edição de texto no buffer de `[]rune`.

### Fase 3: Histórico e Interatividade
- Implementar o `HistoryManager`.
- Integrar a navegação de histórico com as teclas de seta.
- Implementar atalhos de teclado (`Ctrl+L`, `Ctrl+C`).

### Fase 4: Integração e Testes
- Substituir o input atual do Zé pelo novo `Readline`.
- Executar a suíte completa de testes de integração.

## 4. Desafios Técnicos e Soluções

| Desafio | Solução Proposta |
| :--- | :--- |
| **Caracteres Multi-byte (UTF-8)** | Uso de `[]rune` para evitar quebra de caracteres acentuados ou emojis. |
| **Flickering (Piscadas)** | Redesenho parcial da linha usando sequências ANSI de limpeza de linha (`\x1b[K`). |
| **Compatibilidade Windows/Wine** | Focar em chamadas de sistema POSIX via `syscall`, emuladas pelo terminal do Wine. |
| **Segurança de Terminal** | Uso rigoroso de `defer` e `recover` para garantir o retorno ao modo normal. |

## 5. Estratégia de Testes Automáticos

Utilizaremos uma abordagem de "Simulação de Terminal" em três níveis:

### Nível 1: Testes de Unidade do Buffer (`InputBuffer`)
Testa a lógica matemática de manipulação de strings e cursor sem depender de I/O.
- *Exemplo:* Validar que `Insert('a'), MoveCursor(0), Insert('b')` resulta em `"ba"`.

### Nível 2: Testes de Integração com Mock de I/O
Criação de um `MockTerminal` que implementa a interface de leitura/escrita.
- **Simulação:** Alimentar o mock com sequências de bytes (ex: `\x1b[A`) e validar se o output emitido contém as sequências ANSI de renderização esperadas.

### Nível 3: Testes de "Golden Files" (Renderização Visual)
Garante que a interface visual permaneça consistente.
- **Mecanismo:** Salvar o fluxo completo de bytes (incluindo códigos ANSI) gerado por uma sequência de comandos em arquivos `.golden`. Testes subsequentes comparam o output atual com o arquivo salvo.
