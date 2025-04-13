package main

import (
	"craftyreverseproxy/config"
	"craftyreverseproxy/internal/app"
	"log"
)

const configPath = "config.yaml"

func main() {
	cfg := config.NewConfig()
	err := cfg.Load(configPath)
	if err != nil {
		log.Fatal("Failed to start app, err:", err)
	}

	reverseProxyApp := app.NewApp(cfg)
	reverseProxyApp.Run()
}
