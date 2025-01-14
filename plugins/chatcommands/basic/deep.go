package basic

import (
	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
)

var DeepCommand = &commands.Command[*client.Message]{
	Name: "deep",

	Description: "",

	Handler: func(m *client.Message) error {
		m.Room.Send("A")
		return nil
	},

	SubCommands: map[string]*commands.Command[*client.Message]{
		"b": {
			Name: "b",

			Handler: func(m *client.Message) error {
				m.Room.Send("AB")
				return nil
			},

			SubCommands: map[string]*commands.Command[*client.Message]{
				"c": {
					Name: "c",

					Handler: func(m *client.Message) error {
						m.Room.Send("ABC and Message.Content: " + m.Content)
						return nil
					},
				},
			},
		},
	},
}
