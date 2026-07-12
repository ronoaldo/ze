package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ronoaldo/ze/internal/agent"
	"github.com/ronoaldo/ze/internal/commands"
)

// ErrSkipLine is a special error that tells the TUI to skip the current line.
var ErrSkipLine = errors.New("skip line")

// InputHandler manages the state of user input, supporting normal mode and multiline mode.
type InputHandler struct {
	isMultiline     bool
	multilineBuffer strings.Builder
	// Dependencies injected for testability.
	commandExecutor func(string) (string, error)
	agentExecutor   func(string) (string, agent.AgentStats, error)
}

// NewInputHandler creates a new instance of InputHandler.
func NewInputHandler(
	commandExecutor func(string) (string, error),
	agentExecutor func(string) (string, agent.AgentStats, error),
) *InputHandler {
	return &InputHandler{
		commandExecutor: commandExecutor,
		agentExecutor:   agentExecutor,
	}
}

// IsMultiline returns true if the handler is currently in multiline mode.
func (h *InputHandler) IsMultiline() bool {
	return h.isMultiline
}

// Process handles a single line of input and returns the response, stats, and error.
// In multiline mode, it accumulates input until "/send" is received.
func (h *InputHandler) Process(input string) (string, agent.AgentStats, error) {
	if h.isMultiline {
		// Modo Multiline
		if strings.HasPrefix(input, "/send") {
			content := h.multilineBuffer.String()
			h.multilineBuffer.Reset()
			h.isMultiline = false
			return h.agentExecutor(content)
		}

		// Acumula o conteúdo
		h.multilineBuffer.WriteString(input + "\n")
		return "", agent.AgentStats{}, nil
	}

	// Modo Normal

	// 1. Trim o input para normal mode
	input = strings.TrimSpace(input)

	// 2. Se o input estiver vazio, retornamos ErrSkipLine para que o TUI não faça nada.
	if input == "" {
		return "", agent.AgentStats{}, ErrSkipLine
	}

	// 3. Verifica se é o comando para ativar o modo multiline
	if input == "/multiline" {
		h.isMultiline = true
		return "Multiline mode enabled. Type your message and end with '/send' to send.", agent.AgentStats{}, nil
	}

	// 4. Se for um comando (começa com /)
	if strings.HasPrefix(input, "/") {
		resp, err := h.commandExecutor(input)
		if err == nil {
			return resp, agent.AgentStats{}, nil
		}

		// Se for o comando de sair, propaga o erro
		if errors.Is(err, commands.ErrQuit) {
			return "", agent.AgentStats{}, err
		}

		// Se for um comando desconhecido ou com erro, retorna o erro como resposta (seguindo o comportamento atual)
		return fmt.Sprintf("Error: %v", err), agent.AgentStats{}, nil
	}

	// 5. Caso contrário, é uma mensagem para o agente
	return h.agentExecutor(input)
}
