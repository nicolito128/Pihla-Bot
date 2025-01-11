package main

import (
	"flag"
	"log"

	"github.com/nicolito128/Pihla-Bot/client"
)

var (
	name     = flag.String("name", "", "bot name")
	password = flag.String("pass", "", "bot password")
	debug    = flag.Bool("debug", false, "output messages to console")
)

func main() {
	flag.Parse()

	bot := client.NewClient(UseFlagsForLogin)

	log.Println("Bot started succesfully!")
	if err := bot.Start(); err != nil {
		panic(err)
	}
}

func UseFlagsForLogin(cc *client.ClientConfig) {
	cc.Debug = *debug
	cc.Bot.Username = *name
	cc.Bot.Password = *password
	cc.Bot.Rooms = []string{"Hispano", "Bot Development"}
	cc.Bot.Avatar = "211"
}
