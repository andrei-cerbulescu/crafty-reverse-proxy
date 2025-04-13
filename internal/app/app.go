package app

import (
	"context"
	"craftyreverseproxy/config"
	"craftyreverseproxy/internal/adapters/crafty"
	"craftyreverseproxy/internal/modules/proxy"
	"craftyreverseproxy/pkg/logger"
	"crypto/tls"
	"log"
	"net/http"
	"sync"
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
