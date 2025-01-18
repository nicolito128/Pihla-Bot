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

func (m *Message) Client() *Client {
	return m.client
}

// Send sends a message to the room or user that sent the message.
//
// If the message was sent in a room, the message will be sent to the room.
// If the message was sent in a private message, the message will be sent to the user.
func (m *Message) Send(content string) {
	if m.PM {
		m.User.Send(content)
	} else {
		m.Room.Send(content)
	}
}
