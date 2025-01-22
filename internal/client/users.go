package client

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/nicolito128/Pihla-Bot/internal/commands"
	"github.com/nicolito128/Pihla-Bot/pkg/utils"
)

type RankTyp rune

type User struct {
	client *Client

	ID        string
	Name      string
	Rank      RankTyp
	Busy      bool
	IsBot     bool
	LastSeen  time.Time
	Chatrooms []string
	Alts      []string
}

func NewUser(c *Client, username string) *User {
	id, name, rank, busy := parseUsername(username)
	return &User{
		client:   c,
		ID:       id,
		Name:     name,
		Rank:     rank,
		Busy:     busy,
		IsBot:    rank == '*',
		LastSeen: time.Now(),
		Alts:     make([]string, 0),
	}
}

func NewUserByID(c *Client, userId string) *User {
	return &User{
		client:   c,
		ID:       userId,
		Name:     userId,
		LastSeen: time.Now(),
		Alts:     make([]string, 0),
	}
}

func (u *User) Send(message string) error {
	return u.client.SendPrivateMessage(u.ID, message)
}

func (u *User) HasPermission(p commands.Permission) bool {
	if slices.Contains(u.client.config.Bot.Admins, u.ID) {
		return true
	}

	if p.String() == "none" {
		return true
	}

	switch p.String() {
	case "#":
		return u.Rank == '#'
	case "@":
		return u.Rank == '#' || u.Rank == '@'
	case "%":
		return u.Rank == '#' || u.Rank == '@' || u.Rank == '%'
	case "+":
		return u.Rank == '#' || u.Rank == '@' || u.Rank == '%' || u.Rank == '+'
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

func (u *User) updateProfile(username string) {
	id, name, rank, busy := parseUsername(username)
	u.ID = id
	u.Name = name
	u.Rank = rank
	u.Busy = busy
	u.LastSeen = time.Now()
}

func parseUsername(username string) (id string, name string, rank RankTyp, busy bool) {
	id = utils.ToID(username)
	rank = RankTyp(username[0])
	busy = strings.HasSuffix(username, "@!")
	if busy {
		name = username[1 : len(username)-2]
	} else {
		name = username[1:]
	}
	return
}
