package client

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/nicolito128/Pihla-Bot/internal/commands"
	"github.com/nicolito128/Pihla-Bot/pkg/utils"
)

type User struct {
	client *Client

	ID        string
	Name      string
	Busy      bool
	Bot       bool
	LastSeen  time.Time
	Chatrooms []string
	Alts      []string
}

func NewUser(c *Client, name string) *User {
	name, rank, busy := utils.ParseUsername(name)
	return &User{
		client:   c,
		ID:       utils.ToID(name),
		Name:     name,
		Busy:     busy,
		Bot:      rank == '*',
		LastSeen: time.Now(),
		Alts:     make([]string, 0),
	}
}

func (u *User) Send(message string) error {
	return u.client.SendPrivateMessage(u.ID, message)
}

func (u *User) HasPermission(msg *Message, p commands.Permission) bool {
	if slices.Contains(u.client.config.Bot.Admins, u.ID) {
		return true
	}

	if p.Rune() == ' ' {
		return true
	}

	if msg.Room != nil {
		room, ok := u.FindRoom(msg.Room.Title)
		if !msg.FromPM && ok {
			userlist, ok := room.Auth[p.Rune()]
			if ok && strings.Contains(userlist, u.ID) {
				return true
			}
		}
	}

	return false
}
func (u *User) InRoom(room string) bool {
	return slices.Contains(u.Chatrooms, utils.ToID(room))
}

func (u *User) FindRoom(room string) (r *Room, ok bool) {
	roomid := utils.ToID(room)
	if !u.InRoom(roomid) {
		return
	}

	r, ok = u.client.Rooms[roomid]
	return
}

func (u *User) HasAlt(userid string) bool {
	for _, alt := range u.Alts {
		if alt == userid {
			return true
		}
	}

	return false
}

func (u *User) AddAlt(userid string) error {
	if u.HasAlt(userid) {
		return errors.New("the user already has this alt registered")
	}

	u.Alts = append(u.Alts, userid)
	return nil
}

func (u *User) UpdateProfile(username string) {
	name, rank, busy := utils.ParseUsername(username)
	u.ID = utils.ToID(name)
	u.Name = name
	u.Busy = busy
	u.Bot = rank == '*'
	u.LastSeen = time.Now()
}
