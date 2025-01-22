package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Application) SetupRoutes() {
	a.ng.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, my PS name is "+a.ct.Config().Bot.Username)
	})
}
