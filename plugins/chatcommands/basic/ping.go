package basic

import (
	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
)

var PingCommand = &commands.Command[*client.Message]{
	Name: "ping",

	Description: "Ping pong",

	Handler: func(m *client.Message) error {
		m.Room.Send("Pong! :)")
		return nil
	},
}
