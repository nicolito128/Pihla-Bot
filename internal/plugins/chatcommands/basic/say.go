package basic

import (
	"fmt"
	"strings"

	"github.com/nicolito128/Pihla-Bot/internal/client"
	"github.com/nicolito128/Pihla-Bot/internal/commands"
)

var SayCommand = &commands.Command[*client.Message]{
	Name: "say",

	Description: "Say something with the bot.",

	Usage: "say [phrase]",

	AllowPM: true,

	Permissions: commands.VoicerPermission,

	Handler: func(m *client.Message) error {
		content := m.Content
		if strings.HasPrefix(content, "/") || strings.HasPrefix(content, "!") {
			return fmt.Errorf("invalid message content: I'm not authorized to use any other commands")
		}

		if len(content) == 0 || strings.Trim(content, " ") == "" {
			return fmt.Errorf("error: invalid usage. Usage: say [phrase]")
		}

		m.Send(content)
		return nil
	},
}
