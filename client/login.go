package client

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
