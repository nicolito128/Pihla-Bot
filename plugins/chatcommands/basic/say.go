package basic

import (
	"fmt"
	"strings"

	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
)

var SayCommand = &commands.Command[*client.Message]{
	Name: "say",

	Description: "Say something with the bot.",

	Usage: "say [phrase]",

	AllowPM: true,

	Permissions: commands.VoicerPermission,

	Handler: func(m *client.Message) error {
		content := m.Content
		if strings.HasPrefix(content, "/") {
			return fmt.Errorf("error: no estoy autorizada a utilizar otros comandos")
		}

		if m.PM {
			m.User.Send(content)
		} else {
			m.Room.Send(content)
		}

		return nil
	},
}
