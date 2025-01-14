package client

import (
	"strings"
	"time"
)

type RankTyp rune

type User struct {
	client *Client

	ID        string
	Name      string
	Rank      RankTyp
	Busy      bool
	LastSeen  time.Time
	Chatrooms []string
}

func NewUser(c *Client, username string) *User {
	id, name, rank, busy := parseUsername(username)
	return &User{
		client:   c,
		ID:       id,
		Name:     name,
		Rank:     rank,
		Busy:     busy,
		LastSeen: time.Now(),
	}
}

func (u *User) Send(message string) error {
	return u.client.SendPrivateMessage(u.ID, message)
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
	id = toID(username)
	rank = RankTyp(username[0])
	busy = strings.HasSuffix(username, "@!")
	if busy {
		name = username[1 : len(username)-2]
	} else {
		name = username[1:]
	}
	return
}
