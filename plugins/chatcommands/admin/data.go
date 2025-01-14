package admin

import (
	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
)

var DataCommand = &commands.Command[*client.Message]{
	Name: "data",

	Description: "Get information about anything. Admin only.",

	Permissions: commands.AdminPermission,

	Handler: func(m *client.Message) error {
		m.Room.Send("Admin command.")
		return nil
	},
}
