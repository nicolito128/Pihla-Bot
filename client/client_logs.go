package client

import "log"

func (c *Client) Logger() *log.Logger {
	return c.config.Logs
}

func (c *Client) Print(v ...any) {
	c.config.Logs.Print(v...)
}

func (c *Client) Printf(f string, v ...any) {
	c.config.Logs.Printf(f, v...)
}

func (c *Client) Println(v ...any) {
	c.config.Logs.Println(v...)
}
