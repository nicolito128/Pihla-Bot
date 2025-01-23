package client

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/nicolito128/Pihla-Bot/internal/commands"
	"github.com/nicolito128/Pihla-Bot/pkg/utils"
)

func (c *Client) Parse(data []byte) error {
	if len(data) < 1 {
		return nil
	}

	parts := c.handleRawData(data)

	switch parts[1] {
	case "challstr":
		id, str := parts[2], parts[3]
		if err := c.login(id, str); err != nil {
			return err
		}

	case "init":
		id := utils.ToID(parts[2])
		if id == "chat" {
			var title string
			var userlist []string

			if parts[3] == "title" {
				title = strings.TrimSuffix(parts[4], "\n")
			}

			if parts[5] == "users" {
				userlist = strings.Split(parts[6], ",")[1:]
			}

			if err := c.initChat(title, userlist); err != nil {
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

			user.UpdateProfile(username)
			user.Chatrooms = append(user.Chatrooms, room.ID)

			_, rank, _ := utils.ParseUsername(username)
			room.AddAuth(rank, username)
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

			user.UpdateProfile(username)

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
			user.UpdateProfile(newUsername)
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

			_, rank, _ := utils.ParseUsername(newUsername)
			room.AddAuth(rank, newUsername)
		}
	}

	return nil
}

func (c *Client) initChat(roomTitle string, userlist []string) error {
	room := NewRoom(c, roomTitle)
	for i := range userlist {
		username := userlist[i]
		u := NewUser(c, username)
		u.Chatrooms = append(u.Chatrooms, room.ID)

		c.Users[u.ID] = u
		room.Users = append(room.Users, u.ID)

		name, rank, _ := utils.ParseUsername(username)
		room.AddAuth(rank, name)
	}

	c.Rooms[room.ID] = room
	return nil
}

func (c *Client) handleChatMessage(m *Message) {
	if strings.HasPrefix(m.Content, c.config.Bot.Prefix) && !m.User.Bot {
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

		hasPerm := m.User.HasPermission(m, baseCmd.Permissions)
		if !hasPerm {
			permMsg := "You don't have sufficient permissions. Requires: " + baseCmd.Permissions.String()
			m.Send(permMsg)
		}

		cmd, rest := commands.FindDeeperSubCommand(baseCmd, body)
		if cmd != nil {
			go func() {
				m.Content = strings.Join(rest, " ")
				m.Content = strings.Trim(m.Content, " ")

				if m.FromPM && (!cmd.AllowPM && !baseCmd.AllowPM) {
					m.User.Send("This command is not allowed in PMs.")
					return
				}

				hasPerm := m.User.HasPermission(m, cmd.Permissions)
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

func (c *Client) handleRawData(data []byte) []string {
	d := bytes.Split(data, []byte("|"))

	s := make([]string, 0)
	for i := range d {
		s = append(s, string(d[i]))
	}
	return s
}
