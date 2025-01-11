package client

type Room struct {
	client *Client

	ID    string
	Title string
	Users []*User
}

func NewRoom(c *Client, id, title string) *Room {
	return &Room{
		client: c,
		ID:     id,
		Title:  title,
		Users:  make([]*User, 0),
	}
}
