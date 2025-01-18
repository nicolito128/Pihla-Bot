package admin

import (
	"errors"
	"strings"

	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
	"github.com/nicolito128/Pihla-Bot/utils"
)

const sayRoomUsage = "sayroom [room], [phrase]"

var SayRoomCommand = &commands.Command[*client.Message]{
	Name: "sayroom",

	Description: "Say something with the bot to a specific room.",

	Usage: sayRoomUsage,

	AllowPM: true,

	Permissions: commands.AdminPermission,

	Handler: func(m *client.Message) error {
		parts := strings.Split(m.Content, ",")
		if len(parts) < 2 {
			return errors.New("invalid usage. Usage: " + sayRoomUsage)
		}

		roomId := utils.ToID((parts[0]))
		text := strings.Join(parts[1:], " ")

		if ok := m.Client().HasRoom(roomId); !ok {
			return errors.New("room not found")
		}

		m.Client().Room(roomId).Send(text)
		return nil
	},
}
