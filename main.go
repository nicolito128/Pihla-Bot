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
	"github.com/nicolito128/Pihla-Bot/internal/application"
	"github.com/nicolito128/Pihla-Bot/internal/client"
)

var (
	addr     = flag.String("addr", ":8080", "addr for web server")
	name     = flag.String("name", "", "bot name")
	password = flag.String("pass", "", "bot password")
	prefix   = flag.String("prefix", "", "bot prefix for commands")
	rooms    = flag.String("rooms", "", "bot chat rooms")
	admins   = flag.String("admins", "", "bot owners")
	avatar   = flag.String("avatar", "", "bot avatar")

	debug = flag.Bool("debug", false, "output messages to console")
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
	bot := client.New(UseFlags)

	log.Println("Loading commands...")
	loadCommands(bot)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := application.New(*addr, bot)
	app.Run(ctx)

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stopch
}

func UseFlags(cc *client.ClientConfig) {
	cc.Debug = *debug
	cc.Bot.Username = *name
	cc.Bot.Password = *password
	cc.Bot.Prefix = *prefix
	cc.Bot.Avatar = *avatar
	cc.Bot.Rooms = strings.Split(*rooms, ",")
	cc.Bot.Admins = strings.Split(*admins, ",")
}
