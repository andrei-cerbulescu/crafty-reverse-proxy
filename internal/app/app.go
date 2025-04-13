package app

import (
	"craftyreverseproxy/config"
	"crypto/tls"
	"net/http"
	"sync"
)

type App struct {
	cfg       config.Config
	playerMap map[string]int
}

func NewApp(cfg config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (app *App) Run() {
	var wg sync.WaitGroup

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for _, address := range app.cfg.Addresses {
		wg.Add(1)
		go func(address config.ServerType) {
			defer wg.Done()
			app.handleServer(address)
		}(address)
	}
	wg.Wait()
}
