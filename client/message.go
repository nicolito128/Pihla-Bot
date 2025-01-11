package client

import "time"

type Message struct {
	client *Client

	RoomID    string
	UserID    string
	Username  string
	Content   string
	Timestamp time.Time
}
