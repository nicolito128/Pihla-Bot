package application

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicolito128/Pihla-Bot/internal/client"
)

type Application struct {
	addr string
	ng   *gin.Engine
	ct   *client.Client
	logs *log.Logger
}

func New(addr string, client *client.Client) *Application {
	s := new(Application)
	s.addr = addr
	s.ct = client
	s.ng = gin.Default()
	s.logs = log.Default()

	s.SetupRoutes()

	return s
}

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a != nil && a.ng != nil {
		a.ng.ServeHTTP(w, r)
	}
}

func (a *Application) Run(ctx context.Context) {
	go a.StartClient(ctx)
	go a.ListenAndServe()
}

func (a *Application) ListenAndServe() error {
	a.logs.Println("Listening and serving to http://localhost" + a.addr + "/ - Press CTRL+C to exit")
	return a.ng.Run(a.addr)
}

func (a *Application) StartClient(ctx context.Context) {
	errch := a.ct.Start(ctx)

outer:
	for {
		select {
		case err := <-errch:
			a.logs.Println("Something went wrong, ending the process with the following error:", err)
			break outer

		case <-ctx.Done():
			a.ct.Stop("Application context job is done.")
		}
	}
}
