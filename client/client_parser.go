package client

import (
	"bytes"
	"encoding/json"
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
	"github.com/nicolito128/Pihla-Bot/commands"
)

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

	case "chat", "c":
		roomId := toID(strings.TrimPrefix(parts[0], ">"))

		room, existsRoom := c.Rooms[roomId]
		if existsRoom {
			username, content := parts[2], parts[3]

			user, existsUser := room.Users[toID(username)]
			if existsUser {
				msg := &Message{
					client:  c,
					Room:    room,
					User:    user,
					Content: content,
					PM:      false,
				}

				c.handleChatMessage(msg)
			}
		}

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
		username := parts[2]
		user, ok := c.Users[toID(username)]
		if !ok {
			user = NewUser(c, username)
			c.Users[toID(username)] = user
		}

		roomId := toID(strings.TrimPrefix(parts[0], ">"))
		room, ok := c.Rooms[roomId]
		if ok {
			_, hasUser := room.Users[user.ID]
			if !hasUser {
				room.Users[user.ID] = user
			}

			user.updateProfile(username)
			user.Chatrooms = append(user.Chatrooms, roomId)
		}

	case "leave", "l", "L":
		username := parts[2]
		user, ok := c.Users[toID(username)]
		if !ok {
			user = NewUser(c, username)
			c.Users[toID(username)] = user
		}

		roomId := toID(strings.TrimPrefix(parts[0], ">"))
		room, ok := c.Rooms[roomId]
		if ok {
			_, hasUser := room.Users[user.ID]
			if hasUser {
				delete(room.Users, user.ID)
			}

			user.updateProfile(username)
			ci := slices.Index(user.Chatrooms, room.ID)
			if ci != -1 {
				user.Chatrooms = append(user.Chatrooms[:ci], user.Chatrooms[ci+1:]...)
			}
		}

	case "name", "n", "N":
		newUsername := parts[2]
		newUser, ok := c.Users[toID(newUsername)]
		if !ok {
			newUser = NewUser(c, newUsername)
			c.Users[newUser.ID] = newUser
		} else {
			newUser.updateProfile(newUsername)
		}

		oldUserId := parts[3]
		oldUser, ok := c.Users[oldUserId]
		if !ok {
			oldUser = NewUserByID(c, oldUserId)
			c.Users[oldUser.ID] = oldUser
		}

		newUser.AddAlt(oldUser.ID)
		oldUser.AddAlt(newUser.ID)
	}

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

			c.Users[u.ID] = u
			room.Users[u.ID] = u
		}
	}

	c.Rooms[room.ID] = room
	return nil
}

func (c *Client) handleChatMessage(m *Message) {
	if strings.HasPrefix(m.Content, c.config.Bot.Prefix) && !m.User.IsBot {
		parts := strings.Split(m.Content, " ")
		cmdName := strings.Trim(parts[0][len(c.config.Bot.Prefix):], " ")
		body := parts[1:]

		baseCmd, ok := c.chatCommands[cmdName]
		if !ok {
			return
		}

		cmd, rest := commands.FindDeeperSubCommand(baseCmd, body)
		if cmd != nil {
			m.Content = strings.Join(rest, " ")
			m.Content = strings.Trim(m.Content, " ")

			err := cmd.Handler(m)
			if err != nil {
				m.Room.Send(err.Error())
			}
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
