package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LoginResponse struct {
	Status string `json:"status"`
	Data   struct {
		Token   string `json:"token"`
		User_id string `json:"user_id"`
	} `json:"data"`
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Server struct {
	ServerId string `json:"server_id"`
	Port     int `json:"server_port"`
}

type ServerList struct {
	Data []Server `json:"data"`
}

type Settings struct {
	Servers struct {
		ProxyPort  int `json:"proxy_port"`
		ServerIp   int `json:"server_ip"`
		ServerPort int `json:"server_port"`
	} `json:"servers"`
}

func awaitForServerStart(protocol string, target string) net.Conn {
	for i := 0; i < 25; i++ {
		conn, err := net.Dial(protocol, target)
		if err == nil {
			return conn
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

func startMcServer(server ServerType) {
	internalPort := server.InternalPort
	loginBody := LoginPayload{
		Username: getConfig().Username,
		Password: getConfig().Password,
	}
	jsonData, _ := json.Marshal(loginBody)
	resp, err := http.Post(getConfig().ApiUrl+"/api/v2/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		println("Error making POST request:", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error reading response body:", err)
		return
	}

	var response LoginResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		println("Error decoding JSON:", err)
		return
	}

	bearer := "Bearer " + response.Data.Token

	client := &http.Client{}

	serversListReq, _ := http.NewRequest("GET", getConfig().ApiUrl+"/api/v2/servers", nil)
	serversListReq.Header.Add("Authorization", bearer)

	serverListRes, err := client.Do(serversListReq)

	if err != nil {
		println("Error getting servers:", err)
		return
	}

	defer serverListRes.Body.Close()

	serversListBody, err := io.ReadAll(serverListRes.Body)
	if err != nil {
		println("Error reading response body:", err)
		return
	}
	var serverList ServerList
	err = json.Unmarshal(serversListBody, &serverList)
	if err != nil {
		println("Error decoding JSON:", err)
		return
	}

	comparator := func(s Server) bool { return strings.Compare(strconv.Itoa(s.Port), internalPort) == 0 }
	filteredServer := filter(serverList.Data, comparator)[0]
	startServerUrl := getConfig().ApiUrl + "/api/v2/servers/" + filteredServer.ServerId + "/action/start_server"
	startServerReq, _ := http.NewRequest("POST", startServerUrl, nil)
	startServerReq.Header.Add("Authorization", bearer)
	_, err = client.Do(startServerReq)

	if err != nil {
		println("Error getting servers:", err)
		return
	}

	scheduleStopServerIfEmpty(server)
}

func stopMcServer(port int) {
	loginBody := LoginPayload{
		Username: getConfig().Username,
		Password: getConfig().Password,
	}
	jsonData, _ := json.Marshal(loginBody)
	resp, err := http.Post(getConfig().ApiUrl+"/api/v2/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		println("Error making POST request:", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error reading response body:", err)
		return
	}

	var response LoginResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		println("Error decoding JSON:", err)
		return
	}

	bearer := "Bearer " + response.Data.Token

	client := &http.Client{}

	serversListReq, _ := http.NewRequest("GET", getConfig().ApiUrl+"/api/v2/servers", nil)
	serversListReq.Header.Add("Authorization", bearer)

	serverListRes, err := client.Do(serversListReq)

	if err != nil {
		println("Error getting servers:", err)
		return
	}

	defer serverListRes.Body.Close()

	serversListBody, err := ioutil.ReadAll(serverListRes.Body)
	if err != nil {
		println("Error reading response body:", err)
		return
	}
	var serverList ServerList
	err = json.Unmarshal(serversListBody, &serverList)
	if err != nil {
		println("Error decoding JSON:", err)
		return
	}

	comparator := func(s Server) bool { return s.Port == port }
	server := filter(serverList.Data, comparator)[0]
	startServerUrl := getConfig().ApiUrl + "/api/v2/servers/" + server.ServerId + "/action/stop_server"
	startServerReq, _ := http.NewRequest("POST", startServerUrl, nil)
	startServerReq.Header.Add("Authorization", bearer)
	_, err = client.Do(startServerReq)

	if err != nil {
		println("Error getting servers:", err)
		return
	}
}
