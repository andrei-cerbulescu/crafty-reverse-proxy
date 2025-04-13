package app

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"sync"

	"github.com/andrei-cerbulescu/crafty-reverse-proxy/config"
	"github.com/andrei-cerbulescu/crafty-reverse-proxy/internal/adapters/crafty"
	"github.com/andrei-cerbulescu/crafty-reverse-proxy/internal/modules/proxy"
	"github.com/andrei-cerbulescu/crafty-reverse-proxy/pkg/logger"
)

type App struct {
	cfg    config.Config
	logger *logger.Logger
	crafty *crafty.Crafty
}

func NewApp(cfg config.Config, logger *logger.Logger, crafty *crafty.Crafty) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
		crafty: crafty,
	}
}

func (app *App) Run(ctx context.Context) {
	var wg sync.WaitGroup

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for _, address := range app.cfg.Addresses {
		wg.Add(1)
		go func(serverConfig config.ServerType) {
			defer wg.Done()
			server := proxy.NewProxyServer(app.cfg, serverConfig, app.logger, app.crafty)
			if err := server.ListenAndProxy(ctx); err != nil {
				log.Fatal(err)
			}
		}(address)
	}

	wg.Wait()
}
