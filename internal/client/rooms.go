package client

import (
	"slices"
	"strings"

	"github.com/nicolito128/Pihla-Bot/pkg/utils"
)

type Room struct {
	client *Client

	ID    string
	Title string
	Users []string
	Auth  map[rune]string
}

func NewRoom(c *Client, title string) *Room {
	return &Room{
		client: c,
		ID:     utils.ToID(title),
		Title:  title,
		Users:  make([]string, 0),
		Auth: map[rune]string{
			'~': "",
			'#': "",
			'@': "",
			'%': "",
			'*': "",
			'^': "",
			'+': "",
			' ': "",
		},
	}
}

func (r *Room) Send(message string) error {
	return r.client.SendRoomMessage(r.ID, message)
}

func (r *Room) HasUser(username string) bool {
	return slices.Contains(r.Users, utils.ToID(username))
}

func (r *Room) HasAuth(rank rune, username string) bool {
	return strings.Contains(r.Auth[rank], utils.ToID(username))
}

func (r *Room) AddAuth(rank rune, username string) {
	userid := utils.ToID(username)
	if !r.HasAuth(rank, username) {
		r.Auth[rank] += " " + userid
		r.Auth[rank] = strings.Trim(r.Auth[rank], " ")
	}
}

func (r *Room) FindUser(username string) (user *User, ok bool) {
	userid := utils.ToID(username)
	if !r.HasUser(userid) {
		return
	}

	user, ok = r.client.Users[userid]
	return
}
