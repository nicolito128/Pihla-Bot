package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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
	// Plugins
	//commands map[string]*plugins.Command

	// Bot chat rooms.
	Rooms map[string]*Room
}

func New(opts ...Opt) *Client {
	c := new(Client)

	c.config = DefaultClientConfig()
	for _, opt := range opts {
		opt(c.config)
	}

	c.Rooms = make(map[string]*Room)
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

func (c *Client) Parse(data []byte) error {
	parts := c.parseRawData(data)

	switch parts[1] {
	case "challstr":
		id, str := parts[2], parts[3]
		if err := c.login(id, str); err != nil {
			return err
		}

	case "init":
		id := toID(parts[2])
		if id == "chat" {
			if err := c.initChat(parts[3:]); err != nil {
				return err
			}
		}

	case "c:":
		roomId := toID(strings.TrimPrefix(parts[0], ">"))
		room := c.Rooms[roomId]

		i, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return fmt.Errorf("error trying to parse timestamp: %w", err)
		}

		tstamp, username, content := time.Unix(i, 0), parts[3], parts[4]
		user := room.Users[toID(username)]

		msg := &Message{
			client:    c,
			Room:      room,
			User:      user,
			Timestamp: tstamp,
			Content:   content,
			PM:        false,
		}

		c.handleChatMessage(msg)

	case "pm":
		username := parts[2][1:]
		content := parts[4]

		msg := &Message{
			client:  c,
			User:    NewUser(c, username),
			Content: content,
			PM:      true,
		}

		c.handleChatMessage(msg)

	case "join", "j", "J":
		roomId := toID(strings.TrimPrefix(parts[0], ">"))
		room := c.Rooms[roomId]

		username := parts[2]
		userid := toID(username)

		if room != nil {
			user := room.Users[toID(username)]
			if user == nil {
				room.Users[userid] = NewUser(c, username)
				user = room.Users[userid]
			} else {
				user.updateProfile(username)
			}

			user.Chatrooms = append(user.Chatrooms, roomId)
		}

	case "leave", "l", "L":
		roomId := toID(strings.TrimPrefix(parts[0], ">"))

		var room *Room
		for i := range c.Rooms {
			if c.Rooms[i].ID == roomId {
				room = c.Rooms[i]
				break
			}
		}

		username := parts[2]
		if room != nil {
			user := room.Users[toID(username)]
			if user != nil {
				ind := slices.Index(user.Chatrooms, room.ID)
				if ind != -1 {
					user.Chatrooms = append(user.Chatrooms[:ind], user.Chatrooms[ind+1:]...)
				}
			}
		}
	}

	return nil
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

func (c *Client) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.config.ConnectionURL, nil)
	if err != nil {
		return fmt.Errorf("connection failed with error: %w", err)
	}

	c.ws = conn
	c.connected = true
	return nil
}

func (c *Client) login(id, str string) error {
	u, err := url.Parse(c.config.ActionURL)
	if err != nil {
		return fmt.Errorf("invalid url parsing for client action url: %w", err)
	}

	q := u.Query()
	q.Set("act", "login")
	q.Set("name", toID(c.config.Bot.Username))
	q.Set("pass", c.config.Bot.Password)
	q.Set("challengekeyid", id)
	q.Set("challstr", str)

	u.RawQuery = q.Encode()
	res, err := http.Post(u.String(), "application/x-www-form-urlencoded; encoding=UTF-8", nil)
	if err != nil {
		return fmt.Errorf("post request error when login: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading body of login response request: %w", err)
	}

	var login Login
	if err = json.Unmarshal(body[1:], &login); err != nil {
		return fmt.Errorf("json unmarshal of login session error: %w", err)
	}

	d := []byte(fmt.Sprintf("|/trn %s,%s,%s", c.config.Bot.Username, c.config.Bot.Avatar, login.Assertion))
	err = c.ws.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		return fmt.Errorf("websocket writing /trn error: %w", err)
	}

	for _, room := range c.config.Bot.Rooms {
		if err = c.ws.WriteMessage(websocket.TextMessage, []byte("|/j "+toID(room))); err != nil {
			return fmt.Errorf("error trying to join to room `%s` at loign: %w", room, err)
		}
	}

	return nil
}

func (c *Client) initChat(msg []string) error {
	var room *Room

	if msg[0] == "title" {
		room = NewRoom(c, msg[1])
	}
	msg = msg[2:]

	if msg[0] == "users" {
		userlist := strings.Split(msg[1], ",")[1:]

		for i := range userlist {
			username := userlist[i]

			u := NewUser(c, username)
			room.Users[u.ID] = u
		}
	}

	c.Rooms[room.ID] = room
	return nil
}

func (c *Client) parseRawData(data []byte) []string {
	d := bytes.Split(data, []byte("|"))

	s := make([]string, 0)
	for i := range d {
		s = append(s, string(d[i]))
	}
	return s
}

func toID(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSuffix(s, "\n")
	rg := regexp.MustCompile("/[^a-z0-9]+/g")
	s = rg.ReplaceAllString(s, "")
	return s
}
