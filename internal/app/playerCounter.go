package app

import (
	"craftyreverseproxy/config"
	"craftyreverseproxy/internal/adapters/crafty"
	"strconv"
	"time"
)

func indexFromServer(server config.ServerType) string {
	return server.InternalIp + ":" + server.InternalPort
}

func (app *App) getPlayerMap() map[string]int {
	if app.playerMap == nil {
		intermediaryPlayerMap := make(map[string]int)
		for _, elem := range app.cfg.Addresses {
			intermediaryPlayerMap[indexFromServer(elem)] = 0
		}

		app.playerMap = intermediaryPlayerMap
	}

	return app.playerMap
}

func (app *App) decrementPlayerCount(server config.ServerType) {
	app.getPlayerMap()[indexFromServer(server)]--
	if app.isServerEmpty(server) {
		app.scheduleStopServerIfEmpty(server)
	}
}

func (app *App) incrementPlayerCount(server config.ServerType) {
	app.getPlayerMap()[indexFromServer(server)]++
}

func (app *App) isServerEmpty(server config.ServerType) bool {
	return app.getPlayerMap()[indexFromServer(server)] == 0
}

func (app *App) scheduleStopServerIfEmpty(server config.ServerType) {
	if !app.cfg.AutoShutdown {
		return
	}
	time.AfterFunc(time.Duration(app.cfg.Timeout)*time.Minute, func() {
		app.stopServerIfEmpty(server)
	})
}

func (app *App) stopServerIfEmpty(server config.ServerType) {
	internalPort := server.InternalPort
	if app.isServerEmpty(server) {
		port, err := strconv.Atoi(internalPort)
		if err != nil {
			println("Expected number port but got: " + internalPort + "\n" + err.Error() + "\n")
		}
		crafty.StopMcServer(port, app.cfg)
	}
}
