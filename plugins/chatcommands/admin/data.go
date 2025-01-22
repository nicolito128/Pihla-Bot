package admin

import (
	"fmt"

	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
)

var DataCommand = &commands.Command[*client.Message]{
	Name: "data",

	Description: "Get information about anything. Admin only.",

	Permissions: commands.AdminPermission,

	AllowPM: true,

	Handler: func(m *client.Message) error {
		return nil
	},

	SubCommands: map[string]*commands.Command[*client.Message]{
		"rooms": {
			Handler: func(m *client.Message) error {
				for _, room := range m.Client().Rooms {
					fmt.Println(room.ID, room.Title, len(room.Users), room.Users)
				}
				return nil
			},
		},

		"users": {
			Handler: func(m *client.Message) error {
				for _, user := range m.Client().Users {
					fmt.Println(user.ID, user.Name, user.Busy, user.Alts, user.Chatrooms)
				}
				return nil
			},
		},
	},
}
