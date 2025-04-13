package app

import (
	"craftyreverseproxy/config"
	"craftyreverseproxy/internal/adapters/crafty"
	"strconv"
	"time"
)

var playerMap *map[string]int

func indexFromServer(server config.ServerType) string {
	return server.InternalIp + ":" + server.InternalPort
}

func getPlayerMap() *map[string]int {
	if playerMap == nil {
		intermediaryPlayerMap := make(map[string]int)
		for _, elem := range config.GetConfig().Addresses {
			intermediaryPlayerMap[indexFromServer(elem)] = 0
		}

		playerMap = &intermediaryPlayerMap
	}

	return playerMap
}

func decrementPlayerCount(server config.ServerType) {
	(*getPlayerMap())[indexFromServer(server)]--
	if isServerEmpty(server) {
		scheduleStopServerIfEmpty(server)
	}
}

func incrementPlayerCount(server config.ServerType) {
	(*getPlayerMap())[indexFromServer(server)]++
}

func isServerEmpty(server config.ServerType) bool {
	return (*getPlayerMap())[indexFromServer(server)] == 0
}

func scheduleStopServerIfEmpty(server config.ServerType) {
	if !config.GetConfig().AutoShutdown {
		return
	}
	time.AfterFunc(time.Duration(config.GetConfig().Timeout)*time.Minute, func() {
		stopServerIfEmpty(server)
	})
}

func stopServerIfEmpty(server config.ServerType) {
	internalPort := server.InternalPort
	if isServerEmpty(server) {
		port, err := strconv.Atoi(internalPort)
		if err != nil {
			println("Expected number port but got: " + internalPort + "\n" + err.Error() + "\n")
		}
		crafty.StopMcServer(port)
	}
}
