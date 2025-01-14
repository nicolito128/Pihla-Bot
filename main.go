package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nicolito128/Pihla-Bot/client"
)

var (
	name     = flag.String("name", "", "bot name")
	password = flag.String("pass", "", "bot password")
	rooms    = flag.String("rooms", "", "bot initial chat rooms")
	avatar   = flag.String("avatar", "", "bot avatar")
	debug    = flag.Bool("debug", false, "output messages to console")
)

func init() {
	flag.Parse()

	var err error

	_, err = os.Stat(".env")
	if err != nil {
		return
	}

	err = godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	*name = os.Getenv("BOT_NAME")
	*password = os.Getenv("BOT_PASSWORD")
	*rooms = os.Getenv("BOT_ROOMS")
	*avatar = os.Getenv("BOT_AVATAR")
}

func main() {
	bot := client.NewClient(UseFlags)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errch := bot.Start(ctx)

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

outer:
	for {
		select {
		case err := <-errch:
			log.Println("Something went wrong, ending the process with the following error:", err)
			break outer

		case <-stopch:
			bot.Stop("Received a stop signal.")
			break outer
		}
	}
}

func UseFlags(cc *client.ClientConfig) {
	cc.Debug = *debug
	cc.Bot.Username = *name
	cc.Bot.Password = *password
	cc.Bot.Avatar = *avatar
	cc.Bot.Rooms = strings.Split(*rooms, ",")
}
