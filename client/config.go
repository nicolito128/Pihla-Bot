package client

import "fmt"

const (
	DefaultServerID      string = "sim"
	DefaultConnectionURL string = "wss://sim3.psim.us/showdown/websocket"
	DefaultActionURL     string = "https://play.pokemonshowdown.com/~~" + DefaultServerID + ".psim.us/action.php"
)

type Opt func(*ClientConfig)

type ClientConfig struct {
	ConnectionURL string
	ServerID      string
	ActionURL     string
	Debug         bool
	Bot           *BotConfig
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
	c.ServerID = DefaultServerID
	c.ActionURL = DefaultActionURL
	c.Debug = false
	c.Bot = &BotConfig{}

	return c
}

func GetActionURL(serverId string) string {
	return fmt.Sprintf("https://play.pokemonshowdown.com/~~%s.psim.us/action.php", serverId)
}
