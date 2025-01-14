package client

func (c *Client) Print(v ...any) {
	c.logs.Print(v...)
}

func (c *Client) Printf(f string, v ...any) {
	c.logs.Printf(f, v...)
}

func (c *Client) Println(v ...any) {
	c.logs.Println(v...)
}
