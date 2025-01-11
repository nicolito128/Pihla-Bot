package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Handler interface {
	ServeWS(c *Client)
}

type Client struct {
	config    *ClientConfig
	ws        *websocket.Conn
	connected bool

	Rooms []*Room
}

func NewClient(opts ...Opt) *Client {
	c := new(Client)

	c.config = DefaultClientConfig()
	for _, opt := range opts {
		opt(c.config)
	}

	return c
}

func (c *Client) Start() error {
	log.Println("Starting to listen the websocket connection...")

	log.Println("Connecting to Pokemon Showdown...")
	if err := c.connect(); err != nil {
		return err
	}

outer:
	for {
		typ, p, err := c.ws.ReadMessage()
		if typ == websocket.CloseMessage {
			c.ws.Close()
			c.connected = false

			ticker := time.NewTicker(5 * time.Second)
			counter, limit := 0, 10
			for {
				select {
				case <-ticker.C:
					counter++

					log.Println("Trying to reconnect every 5 seconds...")
					if err := c.connect(); err != nil {
						log.Println("Error when trying to reconnect the application: %w", err)
						log.Printf("Attempts to reconnect: %d\n", counter)
					}

				default:
					if counter == limit {
						log.Println("Shutting down the application.")
						break outer
					}
				}
			}
		}

		if err != nil && typ != websocket.CloseMessage {
			return err
		}

		if c.connected {
			if c.config.Debug {
				log.Println(string(p))
			}

			if typ == websocket.TextMessage {
				if err = c.Parse(p); err != nil {
					return err
				}
			}
		}
	}

	return nil
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

		i, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)
		username := parts[3]
		content := parts[4]

		msg := &Message{c, roomId, toID(username), username, content, tm}

		c.handleChatMessage(msg)
	}

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
	res, err := http.Post(u.String(), "application/x-www-form-urlencoded", nil)
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

	d := []byte(fmt.Sprintf("|/trn %s,0,%s", c.config.Bot.Username, login.Assertion))
	err = c.ws.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		return fmt.Errorf("websocket writing /trn error: %w", err)
	}

	for _, room := range c.config.Bot.Rooms {
		if err = c.ws.WriteMessage(websocket.TextMessage, []byte("|/j "+room)); err != nil {
			return fmt.Errorf("error trying to join to room `%s` at loign: %w", room, err)
		}
	}

	return nil
}

func (c *Client) initChat(msg []string) error {
	var room *Room
	var users []*User

	if msg[0] == "title" {
		id := toID(msg[1])
		room = NewRoom(c, id, msg[1])
	}
	msg = msg[2:]

	if msg[0] == "users" {
		userlist := strings.Split(msg[1], ",")[1:]
		users = make([]*User, 0)

		for i := range userlist {
			rank := RankTyp(userlist[i][0])
			name := userlist[i][1:]
			id := toID(name)

			u := NewUser(c, id, name)
			u.Rank = rank
			users = append(users, u)
		}
	}

	room.Users = users
	c.Rooms = append(c.Rooms, room)
	return nil
}

func (c *Client) handleChatMessage(m *Message) {
	if strings.HasPrefix(m.Content, "--ping") {
		res := fmt.Sprintf("|/msgroom %s, Pong!", m.RoomID)
		err := c.ws.WriteMessage(websocket.TextMessage, []byte(res))
		if err != nil {
			panic(err)
		}
	}
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
