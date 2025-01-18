package client

import "github.com/nicolito128/Pihla-Bot/utils"

type Room struct {
	client *Client

	ID    string
	Title string
	Users map[string]*User
}

func NewRoom(c *Client, title string) *Room {
	return &Room{
		client: c,
		ID:     utils.ToID(title),
		Title:  title,
		Users:  make(map[string]*User),
	}
}

func (r *Room) Send(message string) error {
	return r.client.SendRoomMessage(r.ID, message)
}
