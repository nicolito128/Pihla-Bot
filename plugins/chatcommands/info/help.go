package info

import (
	"errors"
	"fmt"

	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
)

var HelpCommand = &commands.Command[*client.Message]{
	Name: "help",

	Description: "Get information about a command.",

	Handler: func(m *client.Message) error {
		content := m.Content

		cmd, ok := m.Client().FindCommand(content)
		if !ok {
			return errors.New("error: cmd not found")
		}

		s := fmt.Sprintf("Command name: ``%s`` | Permissions: ``0`` | Description: ``%s``", cmd.Name, cmd.Description)
		m.Room.Send(s)
		return nil
	},
}
