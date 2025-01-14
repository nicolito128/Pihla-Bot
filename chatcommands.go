package main

import (
	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
	"github.com/nicolito128/Pihla-Bot/plugins/chatcommands/basic"
)

type CommandHandler interface {
	HandleCommand(string, *commands.Command[*client.Message])
}

var chatCommands = make([]*commands.Command[*client.Message], 0)

func init() {
	chatCommands = append(
		chatCommands,
		basic.PingCommand,
		basic.DeepCommand,
	)
}

func loadCommands(handler CommandHandler) {
	for _, cmd := range chatCommands {
		handler.HandleCommand(cmd.Name, cmd)
	}
}
