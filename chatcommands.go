package main

import (
	"fmt"

	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
	"github.com/nicolito128/Pihla-Bot/plugins/chatcommands/admin"
	"github.com/nicolito128/Pihla-Bot/plugins/chatcommands/basic"
	"github.com/nicolito128/Pihla-Bot/plugins/chatcommands/info"
)

type CommandHandler interface {
	HandleCommand(string, *commands.Command[*client.Message])
}

var chatCommands = make([]*commands.Command[*client.Message], 0)

func init() {
	chatCommands = append(
		chatCommands,
		basic.PingCommand,
		basic.SayCommand,
		info.HelpCommand,
		admin.DataCommand,
		admin.SayRoomCommand,
	)
}

func loadCommands(handler CommandHandler) {
	for _, cmd := range chatCommands {
		fmt.Printf("-- Command `%s` loaded!\n", cmd.Name)
		handler.HandleCommand(cmd.Name, cmd)
	}
}
