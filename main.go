package main

import (
	"flag"
	"log"
	"strings"

	"github.com/nicolito128/Pihla-Bot/client"
)

var (
	name     = flag.String("name", "", "bot name")
	password = flag.String("pass", "", "bot password")
	rooms    = flag.String("rooms", "botdev", "bot initial chat rooms")
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
	cc.Bot.Rooms = strings.Split(*rooms, ",")
	cc.Bot.Avatar = "211"
}
