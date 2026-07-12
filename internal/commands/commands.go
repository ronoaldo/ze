package commands

import (
	"errors"
	"fmt"
	"strings"
)

var ErrQuit = errors.New("quit")

type Command struct {
	Name        string
	Description string
	Execute     func(args []string) (string, error)
}

var Registry = make(map[string]Command)

func Register(cmd Command) {
	Registry[cmd.Name] = cmd
}

func RegisterCommands() {
	Register(Command{
		Name:        "/quit",
		Description: "Exit the session",
		Execute: func(args []string) (string, error) {
			return "", ErrQuit
		},
	})
	Register(Command{
		Name:        "/exit",
		Description: "Exit the session",
		Execute: func(args []string) (string, error) {
			return "", ErrQuit
		},
	})
	Register(Command{
		Name:        "/help",
		Description: "Show available commands",
		Execute: func(args []string) (string, error) {
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
		Execute: func(args []string) (string, error) {
			return "Multiline mode enabled. Type your message and end with '/send' to send.", nil
		},
	})
	Register(Command{
		Name:        "/send",
		Description: "Send the accumulated multiline input",
		Execute: func(args []string) (string, error) {
			return "Use '/send' within multiline mode to send your message.", nil
		},
	})
}

func ExecuteCommand(input string) (string, error) {
	parts := strings.SplitN(input, " ", 2)
	cmdName := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = strings.Split(parts[1], " ")
	}

	if cmd, ok := Registry[cmdName]; ok {
		return cmd.Execute(args)
	}

	return "", fmt.Errorf("unknown command: %s", cmdName)
}
