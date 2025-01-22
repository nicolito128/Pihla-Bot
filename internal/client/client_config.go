package client

import (
	"fmt"
	"log"
)

const (
	DefaultServer        string = "sim.psim.us"
	DefaultConnectionURL string = "wss://sim3.psim.us/showdown/websocket"
	DefaultActionURL     string = "https://play.pokemonshowdown.com/~~" + DefaultServer + "/action.php"
)

type Opt func(*ClientConfig)

type ClientConfig struct {
	ConnectionURL string
	ActionURL     string
	Server        string
	Debug         bool
	Bot           *BotConfig
	Logs          *log.Logger
}

type BotConfig struct {
	Username string
	Password string
	Prefix   string
	Avatar   string
	Status   string
	Rooms    []string
	Admins   []string
}

func DefaultClientConfig() *ClientConfig {
	c := new(ClientConfig)
	c.ConnectionURL = DefaultConnectionURL
	c.Server = DefaultServer
	c.ActionURL = DefaultActionURL
	c.Debug = false
	c.Bot = &BotConfig{}
	c.Logs = log.Default()

	return c
}

func GetActionURL(server string) string {
	return fmt.Sprintf("https://play.pokemonshowdown.com/~~%s/action.php", server)
}
