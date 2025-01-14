package client

import "time"

type Message struct {
	client *Client

	Room      *Room
	User      *User
	Timestamp time.Time
	Content   string
	PM        bool
}
