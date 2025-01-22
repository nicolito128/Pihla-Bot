package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nicolito128/Pihla-Bot/commands"
	"github.com/nicolito128/Pihla-Bot/utils"
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
		id := utils.ToID(parts[2])
		if id == "chat" {
			if err := c.initChat(parts[3:]); err != nil {
				return err
			}
		}

	case "c:":
		roomId := utils.ToID(strings.TrimPrefix(parts[0], ">"))
		room := c.Rooms[roomId]

		i, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return fmt.Errorf("error trying to parse timestamp: %w", err)
		}

		tstamp, username, content := time.Unix(i, 0), parts[3], parts[4]
		user := c.Users[utils.ToID(username)]
		msg := &Message{
			client:    c,
			Room:      room,
			User:      user,
			Timestamp: tstamp,
			Content:   content,
			FromPM:    false,
		}

		c.handleChatMessage(msg)

	case "chat", "c":
		roomId := utils.ToID(strings.TrimPrefix(parts[0], ">"))

		room, existsRoom := c.Rooms[roomId]
		if existsRoom {
			username, content := parts[2], parts[3]
			user, existsUser := c.Users[utils.ToID(username)]

			if existsUser {
				msg := &Message{
					client:  c,
					Room:    room,
					User:    user,
					Content: content,
					FromPM:  false,
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
			FromPM:  true,
		}

		c.handleChatMessage(msg)

	// |join/j/J|USER
	case "join", "j", "J":
		username := parts[2]
		userid := utils.ToID(username)
		user, ok := c.Users[userid]
		if !ok {
			user = NewUser(c, username)
			c.Users[userid] = user
		}

		roomId := utils.ToID(strings.TrimPrefix(parts[0], ">"))
		room, ok := c.Rooms[roomId]
		if ok {
			if !slices.Contains(room.Users, userid) {
				room.Users = append(room.Users, userid)
			}

			user.updateProfile(username)
			user.Chatrooms = append(user.Chatrooms, room.ID)
		}

	// |leave/l/L|USER
	case "leave", "l", "L":
		username := parts[2]
		userid := utils.ToID(username)
		user, ok := c.Users[userid]
		if !ok {
			user = NewUser(c, username)
			c.Users[userid] = user
		}

		roomId := utils.ToID(strings.TrimPrefix(parts[0], ">"))
		room, ok := c.Rooms[roomId]
		if ok {
			ind := slices.Index(room.Users, userid)
			if ind != -1 {
				room.Users = append(room.Users[0:ind], room.Users[ind+1:]...)
			}

			user.updateProfile(username)

			ind = slices.Index(user.Chatrooms, room.ID)
			if ind != -1 {
				user.Chatrooms = append(user.Chatrooms[0:ind], user.Chatrooms[ind+1:]...)
			}
		}

	// >ROOMID
	// |name/n/N|USER|OLDID
	case "name", "n", "N":
		newUsername := parts[2]
		oldUsername := parts[3]
		userid := utils.ToID(newUsername)
		oldUserid := utils.ToID(oldUsername)

		user, ok := c.Users[oldUserid]
		if ok {
			user.updateProfile(newUsername)
			c.Users[userid] = user
		} else {
			user = NewUser(c, newUsername)
			c.Users[userid] = user
		}
		user.AddAlt(oldUserid)

		roomId := utils.ToID(strings.TrimPrefix(parts[0], ">"))
		room, ok := c.Rooms[roomId]
		if ok {
			ind := slices.Index(room.Users, oldUserid)
			if ind != -1 {
				room.Users[ind] = userid
			} else {
				room.Users = append(room.Users, userid)
			}

			ind = slices.Index(user.Chatrooms, room.ID)
			if ind == -1 {
				user.Chatrooms = append(user.Chatrooms, room.ID)
			}
		}
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
	q.Set("name", utils.ToID(c.config.Bot.Username))
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

	d := []byte(fmt.Sprintf("|/trn %s,0,%s", c.config.Bot.Username, login.Assertion))
	err = c.ws.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		return fmt.Errorf("websocket writing /trn error: %w", err)
	}

	for _, room := range c.config.Bot.Rooms {
		if err = c.ws.WriteMessage(websocket.TextMessage, []byte("|/j "+utils.ToID(room))); err != nil {
			return fmt.Errorf("error trying to join to room `%s` at loign: %w", room, err)
		}
	}

	return nil
}

func (c *Client) initChat(msg []string) error {
	var room *Room

	if msg[0] == "title" {
		title := strings.TrimSuffix(msg[1], "\n")
		room = NewRoom(c, title)
	}
	msg = msg[2:]

	if msg[0] == "users" {
		userlist := strings.Split(msg[1], ",")[1:]

		for i := range userlist {
			username := userlist[i]
			u := NewUser(c, username)
			u.Chatrooms = append(u.Chatrooms, room.ID)

			c.Users[u.ID] = u
			room.Users = append(room.Users, u.ID)
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

		if m.FromPM && !baseCmd.AllowPM {
			m.User.Send("This command is not allowed in PMs.")
			return
		}

		hasPerm := m.User.HasPermission(baseCmd.Permissions)
		if !hasPerm {
			permMsg := "You don't have sufficient permissions. Requires: " + baseCmd.Permissions.String()
			m.Send(permMsg)
		}

		cmd, rest := commands.FindDeeperSubCommand(baseCmd, body)
		if cmd != nil {
			go func() {
				m.Content = strings.Join(rest, " ")
				m.Content = strings.Trim(m.Content, " ")

				hasPerm := m.User.HasPermission(cmd.Permissions)
				if !hasPerm {
					permMsg := "You don't have sufficient permissions. Requires: " + cmd.Permissions.String()
					m.Send(permMsg)
				}

				err := cmd.Handler(m)
				if err != nil {
					m.Send(err.Error())
				}
			}()
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
