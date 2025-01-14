package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nicolito128/Pihla-Bot/commands"
)

type Client struct {
	// Basic client configuration.
	//
	// The bot configuration can be set with an .env
	// file or by application arguments.
	config *ClientConfig
	// The websocket connection.
	ws *websocket.Conn
	// Context for the job created by Client.Start.
	ctx context.Context
	// Channel to handle client errors.
	errCh chan error
	// If the application has already been started.
	started bool
	// If the application had an active websocket connection.
	connected bool
	// Chat commands
	chatCommands map[string]*commands.Command[*Message]

	// Bot chat rooms.
	Rooms map[string]*Room
	// Bot users.
	Users map[string]*User
}

func New(opts ...Opt) *Client {
	c := new(Client)

	c.config = DefaultClientConfig()
	for _, opt := range opts {
		opt(c.config)
	}

	c.chatCommands = make(map[string]*commands.Command[*Message])
	c.Rooms = make(map[string]*Room)
	c.Users = make(map[string]*User)
	return c
}

func NewClient(opts ...Opt) *Client {
	return New(opts...)
}

func (c *Client) Start(ctx context.Context) <-chan error {
	if c.started {
		return c.errCh
	}
	c.started = true

	c.errCh = make(chan error)
	c.ctx = ctx

	go func(ch chan<- error) {
		c.Println("Connecting to Pokemon Showdown...")
		if err := c.connect(); err != nil {
			c.errCh <- err
		}

		for {
			typ, p, err := c.ws.ReadMessage()
			if typ == websocket.CloseMessage {
				c.ws.Close()
				c.connected = false

				ticker := time.NewTicker(10 * time.Second)
				counter, limit := 0, 10
				for {
					select {
					case <-ticker.C:
						counter++

						c.Println("Trying to reconnect every 10 seconds...")
						if err := c.connect(); err != nil {
							c.Println("Error when trying to reconnect the application: %w", err)
							c.Printf("Attempts to reconnect: %d\n", counter)
						}

					default:
						if counter == limit {
							c.Stop("It was not possible to reconnect the application.")
							return
						}
					}
				}
			}

			if err != nil && typ != websocket.CloseMessage {
				c.errCh <- err
			}

			if c.connected {
				if c.config.Debug {
					c.Println(string(p))
				}

				if typ == websocket.TextMessage {
					if err = c.Parse(p); err != nil {
						c.errCh <- err
					}
				}
			}
		}
	}(c.errCh)

	return c.errCh
}

// SendRaw writes a new websocket.BinaryMessage to the server.
func (c *Client) SendRaw(data []byte) error {
	return c.ws.WriteMessage(websocket.BinaryMessage, data)
}

// Send writes a new websocket.TextMessage to the server.
func (c *Client) Send(data string) error {
	return c.ws.WriteMessage(websocket.TextMessage, []byte(data))
}

func (c *Client) SendRoomMessage(roomId, message string) error {
	return c.Send(fmt.Sprintf("|/msgroom %s, %s", roomId, message))
}

func (c *Client) SendPrivateMessage(userId, message string) error {
	return c.Send(fmt.Sprintf("|/pm %s, %s", userId, message))
}

func (c *Client) SendCommand(commandName, body string) error {
	s := fmt.Sprintf("|/%s %s", commandName, body)
	return c.Send(s)
}

func (c *Client) Stop(reason string) error {
	if !c.started {
		return errors.New("client has not been started before closing")
	}

	if c.connected {
		c.ws.Close()
	}

	c.Println(reason)
	return nil
}

func (c *Client) HandleCommand(name string, cmd *commands.Command[*Message]) {
	c.chatCommands[name] = cmd
}

func (c *Client) FindCommand(name string) (cmd *commands.Command[*Message], ok bool) {
	cmd, ok = c.chatCommands[name]
	return
}

func (c *Client) Prefix() string {
	return c.config.Bot.Prefix
}

func (c *Client) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.config.ConnectionURL, nil)
	if err != nil {
		return fmt.Errorf("connection failed with error: %w", err)
	}

	c.ws = conn
	c.connected = true
	return nil
}
