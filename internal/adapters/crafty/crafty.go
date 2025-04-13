package crafty

import (
	"bytes"
	"craftyreverseproxy/config"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AwaitForServerStart(protocol string, target string) net.Conn {
	for i := 0; i < 25; i++ {
		conn, err := net.Dial(protocol, target)
		if err == nil {
			return conn
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

func getBearer() string {
	loginBody := LoginPayload{
		Username: config.GetConfig().Username,
		Password: config.GetConfig().Password,
	}
	jsonData, _ := json.Marshal(loginBody)
	resp, err := http.Post(config.GetConfig().ApiUrl+"/api/v2/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic("Could not connect to the server\n")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Could not read response body\n")
	}

	var response LoginResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic("Could not decode JSON\n")
	}

	return "Bearer " + response.Data.Token
}

func getServers(bearer string) ServerList {
	client := &http.Client{}

	serversListReq, _ := http.NewRequest("GET", config.GetConfig().ApiUrl+"/api/v2/servers", nil)
	serversListReq.Header.Add("Authorization", bearer)

	serverListRes, err := client.Do(serversListReq)

	if err != nil {
		panic("Error getting servers: " + err.Error() + "\n")
	}

	defer serverListRes.Body.Close()

	serversListBody, err := io.ReadAll(serverListRes.Body)
	if err != nil {
		panic("Error reading response body: " + err.Error() + "\n")
	}
	var serverList ServerList
	err = json.Unmarshal(serversListBody, &serverList)
	if err != nil {
		panic("Error decoding JSON: " + err.Error() + "\n")
	}

	return serverList
}

func startMcServerCall(server Server, bearer string) {
	client := &http.Client{}
	startServerUrl := config.GetConfig().ApiUrl + "/api/v2/servers/" + server.ServerId + "/action/start_server"
	startServerReq, _ := http.NewRequest("POST", startServerUrl, nil)
	startServerReq.Header.Add("Authorization", bearer)
	_, err := client.Do(startServerReq)

	if err != nil {
		panic("Error getting servers: " + err.Error() + "\n")
	}
}

func stopMcServerCall(server Server, bearer string) {
	client := &http.Client{}

	startServerUrl := config.GetConfig().ApiUrl + "/api/v2/servers/" + server.ServerId + "/action/stop_server"
	startServerReq, _ := http.NewRequest("POST", startServerUrl, nil)
	startServerReq.Header.Add("Authorization", bearer)
	_, err := client.Do(startServerReq)

	if err != nil {
		panic("Error getting servers: " + err.Error() + "\n")
	}
}

func StartMcServer(server config.ServerType) {
	internalPort := server.InternalPort

	bearer := getBearer()

	serverList := getServers(bearer)

	comparator := func(s Server) bool { return strings.Compare(strconv.Itoa(s.Port), internalPort) == 0 }
	filteredServer := filter(serverList.Data, comparator)[0]

	startMcServerCall(filteredServer, bearer)
}

func StopMcServer(port int) {
	bearer := getBearer()

	var serverList = getServers(bearer)

	comparator := func(s Server) bool { return s.Port == port }
	server := filter(serverList.Data, comparator)[0]

	stopMcServerCall(server, bearer)
}
