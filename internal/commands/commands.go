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
