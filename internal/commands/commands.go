package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ronoaldo/ze/internal/agent"
)

var ErrQuit = errors.New("quit")

type Command struct {
	Name        string
	Description string
	Execute     func(agent *agent.Agent, args []string) (string, error)
}

var Registry = make(map[string]Command)

func Register(cmd Command) {
	Registry[cmd.Name] = cmd
}

func RegisterCommands() {
	Register(Command{
		Name:        "/quit",
		Description: "Exit the session",
		Execute: func(agent *agent.Agent, args []string) (string, error) {
			return "", ErrQuit
		},
	})
	Register(Command{
		Name:        "/exit",
		Description: "Exit the session",
		Execute: func(agent *agent.Agent, args []string) (string, error) {
			return "", ErrQuit
		},
	})
	Register(Command{
		Name:        "/help",
		Description: "Show available commands",
		Execute: func(agent *agent.Agent, args []string) (string, error) {
			var sb strings.Builder
			sb.WriteString("Available commands:\n")
			for name, cmd := range Registry {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", name, cmd.Description))
			}
			return sb.String(), nil
		},
	})
	Register(Command{
		Name:        "/multiline",
		Description: "Enable multiline input mode",
		Execute: func(agent *agent.Agent, args []string) (string, error) {
			return "", nil
		},
	})
	Register(Command{
		Name:        "/send",
		Description: "Send the accumulated multiline input",
		Execute: func(agent *agent.Agent, args []string) (string, error) {
			return "Use '/send' within multiline mode to send your message.", nil
		},
	})
	Register(Command{
		Name:        "/models",
		Description: "List available models",
		Execute: func(agent *agent.Agent, args []string) (string, error) {
			models, err := agent.ListModels()
			if err != nil {
				return fmt.Sprintf("Error listing models: %v", err), nil
			}
			if len(models) == 0 {
				return "No models found.", nil
			}
			var sb strings.Builder
			sb.WriteString("Available models:\n")
			for _, m := range models {
				sb.WriteString(fmt.Sprintf("- %s (Status: %s)\n", m.ID, m.Status))
			}
			return sb.String(), nil
		},
	})
	Register(Command{
		Name:        "/model",
		Description: "Switch the current model",
		Execute: func(agent *agent.Agent, args []string) (string, error) {
			if len(args) == 0 {
				return "Usage: /model <model_id>", nil
			}
			modelID := args[0]
			if err := agent.SetModel(modelID); err != nil {
				return fmt.Sprintf("Error: %v", err), nil
			}
			return fmt.Sprintf("Model switched to: %s", modelID), nil
		},
	})
}

func ExecuteCommand(agent *agent.Agent, input string) (string, error) {
	parts := strings.SplitN(input, " ", 2)
	cmdName := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = strings.Split(parts[1], " ")
	}

	if cmd, ok := Registry[cmdName]; ok {
		return cmd.Execute(agent, args)
	}

	return "", fmt.Errorf("unknown command: %s", cmdName)
}
