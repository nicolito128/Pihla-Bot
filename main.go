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
	_ "github.com/nicolito128/Pihla-Bot/commands"
)

// Login information.
var (
	name     = flag.String("name", "", "bot name")
	password = flag.String("pass", "", "bot password")
	prefix   = flag.String("prefix", "", "bot prefix for commands")
	rooms    = flag.String("rooms", "", "bot chat rooms")
	admins   = flag.String("admins", "", "bot owners")
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
	*prefix = os.Getenv("BOT_PREFIX")
	*rooms = os.Getenv("BOT_ROOMS")
	*admins = os.Getenv("BOT_ADMINS")
	*avatar = os.Getenv("BOT_AVATAR")
}

func main() {
	bot := client.New(UseBotData)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot.Println("Loading commands...")
	loadCommands(bot)

	errch := bot.Start(ctx)

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

outer:
	for {
		select {
		case err := <-errch:
			log.Println("Something went wrong, ending the process with the following error:", err)
			break outer

		case <-ctx.Done():
			bot.Stop("Application context job is done.")

		case <-stopch:
			bot.Stop("Received a stop signal.")
			break outer
		}
	}
}

func UseBotData(cc *client.ClientConfig) {
	cc.Debug = *debug
	cc.Bot.Username = *name
	cc.Bot.Password = *password
	cc.Bot.Prefix = *prefix
	cc.Bot.Avatar = *avatar
	cc.Bot.Rooms = strings.Split(*rooms, ",")
	cc.Bot.Admins = strings.Split(*admins, ",")
}
