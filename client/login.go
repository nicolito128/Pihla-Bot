package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/nicolito128/Pihla-Bot/utils"
)

type Login struct {
	ActionSuccess bool        `json:"actionsuccess"`
	Assertion     string      `json:"assertion"`
	CurUser       CurrentUser `json:"curuser"`
}

type CurrentUser struct {
	LoggedIn bool   `json:"loggedin"`
	Username string `json:"username"`
	UserID   string `json:"userid"`
}

func (c *Client) login(id, str string) error {
	u, err := url.Parse(c.config.ActionURL)
	if err != nil {
		return fmt.Errorf("invalid url parsing for client action url: %w", err)
	}

	q := u.Query()
	q.Set("act", "login")
	q.Set("name", utils.ToID(c.config.Bot.Username))
	q.Set("pass", c.config.Bot.Password)
	q.Set("challengekeyid", id)
	q.Set("challstr", str)

	u.RawQuery = q.Encode()
	res, err := http.Post(u.String(), "application/x-www-form-urlencoded; encoding=UTF-8", nil)
	if err != nil {
		return fmt.Errorf("post request error when login: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading body of login response request: %w", err)
	}

	var login Login
	if err = json.Unmarshal(body[1:], &login); err != nil {
		return fmt.Errorf("json unmarshal of login session error: %w", err)
	}

	d := []byte(fmt.Sprintf("|/trn %s,0,%s", c.config.Bot.Username, login.Assertion))
	err = c.ws.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		return fmt.Errorf("websocket writing /trn error: %w", err)
	}

	for _, room := range c.config.Bot.Rooms {
		if err = c.ws.WriteMessage(websocket.TextMessage, []byte("|/j "+utils.ToID(room))); err != nil {
			return fmt.Errorf("error trying to join to room `%s` at loign: %w", room, err)
		}
	}

	return nil
}
