package main

import (
	"fmt"

	"github.com/nicolito128/Pihla-Bot/internal/client"
	"github.com/nicolito128/Pihla-Bot/internal/commands"
	"github.com/nicolito128/Pihla-Bot/internal/plugins/chatcommands/admin"
	"github.com/nicolito128/Pihla-Bot/internal/plugins/chatcommands/basic"
	"github.com/nicolito128/Pihla-Bot/internal/plugins/chatcommands/info"
)

type CommandHandler interface {
	HandleCommand(string, *commands.Command[*client.Message])
}

var chatCommands = make([]*commands.Command[*client.Message], 0)

func init() {
	chatCommands = append(
		chatCommands,
		basic.CalcCommand,
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
