package info

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nicolito128/Pihla-Bot/internal/client"
	"github.com/nicolito128/Pihla-Bot/internal/commands"
)

var HelpCommand = &commands.Command[*client.Message]{
	Name: "help",

	Description: "Get information about a command.",

	Usage: "help [command name]",

	AllowPM: true,

	Handler: func(m *client.Message) error {
		if len(m.Content) == 0 || strings.Trim(m.Content, " ") == "" {
			return errors.New("error: invalid usage. Usage: help [command name]")
		}

		parts := strings.Split(m.Content, " ")

		baseCmd, ok := m.Client().FindCommand(parts[0])
		if !ok {
			return errors.New("error: command not found")
		}

		cmd, _ := commands.FindDeeperSubCommand(baseCmd, parts[1:])

		s := fmt.Sprintf("**Name**: ``%s``", cmd.Name)
		s = fmt.Sprintf("%s | Permissions: ``%v``", s, cmd.Permissions.String())

		if cmd.AllowPM {
			s = fmt.Sprintf("%s | PM allowed: ``yes``", s)
		} else {
			s = fmt.Sprintf("%s | PM allowed: ``no``", s)
		}

		if cmd.Usage != "" {
			s = fmt.Sprintf("%s | Usage: ``%s%s``", s, m.Client().Prefix(), cmd.Usage)
		}

		if cmd.Description != "" {
			s = fmt.Sprintf("%s | Description: ``%s``", s, cmd.Description)
		}

		m.Send(s)
		return nil
	},
}
