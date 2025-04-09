package main

import (
	"strconv"
	"time"
)

var playerMap *map[string]int

func indexFromServer(server ServerType) string {
	return server.InternalIp + ":" + server.InternalPort
}

func getPlayerMap() *map[string]int {
	if playerMap == nil {
		intermediaryPlayerMap := make(map[string]int)
		for _, elem := range getConfig().Addresses {
			intermediaryPlayerMap[indexFromServer(elem)] = 0
		}

		playerMap = &intermediaryPlayerMap
	}

	return playerMap
}

func decrementPlayerCount(server ServerType) {
	(*getPlayerMap())[indexFromServer(server)]--
	if isServerEmpty(server) {
		scheduleStopServerIfEmpty(server)
	}
}

func incrementPlayerCount(server ServerType) {
	(*getPlayerMap())[indexFromServer(server)]++
}

func isServerEmpty(server ServerType) bool {
	return (*getPlayerMap())[indexFromServer(server)] == 0
}

func scheduleStopServerIfEmpty(server ServerType) {
	if(getConfig().AutoShutdown == false){
		return
	}
	time.AfterFunc(getConfig().Timeout*time.Minute, func (){
		stopServerIfEmpty(server)
})
}

func stopServerIfEmpty(server ServerType){
	internalPort :=server.InternalPort
	if isServerEmpty(server){
		port, err :=strconv.Atoi(internalPort)
		if(err!=nil){
			println("Expected number port but got: " + internalPort+"\n"+err.Error()+"\n")		}
		stopMcServer(port)
	}
}
