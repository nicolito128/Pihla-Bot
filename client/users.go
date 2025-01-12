package client

type RankTyp rune

type User struct {
	client *Client

	ID   string
	Name string
	Rank RankTyp
}

func NewUser(c *Client, id, name string) *User {
	return &User{
		client: c,
		ID:     id,
		Name:   name,
	}
}

func (u *User) Send(message string) error {
	return u.client.SendPrivateMessage(u.ID, message)
}
