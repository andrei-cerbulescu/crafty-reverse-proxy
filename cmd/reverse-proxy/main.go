package main

import (
	"context"
	"flag"
	"log"

	"github.com/andrei-cerbulescu/crafty-reverse-proxy/config"
	"github.com/andrei-cerbulescu/crafty-reverse-proxy/internal/adapters/crafty"
	"github.com/andrei-cerbulescu/crafty-reverse-proxy/internal/app"
	"github.com/andrei-cerbulescu/crafty-reverse-proxy/pkg/logger"
)

func main() {
	ctx := context.Background()

	configPath := "config/config.yaml"

	flag.StringVar(&configPath, "c", "config/config.yaml", "Path to config file")
	flag.Parse()

	cfg := config.NewConfig()
	err := cfg.Load(configPath)
	if err != nil {
		log.Fatal("Failed to start app, err:", err)
	}

	logger := logger.New(cfg.LogLevel)

	crafty := crafty.NewCrafty(cfg)
	reverseProxyApp := app.NewApp(cfg, logger, crafty)

	reverseProxyApp.Run(ctx)
}
