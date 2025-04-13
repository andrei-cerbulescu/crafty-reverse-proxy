package crafty

import (
	"bytes"
	"craftyreverseproxy/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Crafty struct {
	apiUrl   string
	username string
	password string
	client   *http.Client
}

func NewCrafty(cfg config.Config) *Crafty {
	return &Crafty{
		apiUrl:   cfg.ApiUrl,
		username: cfg.Username,
		password: cfg.Password,
		client:   &http.Client{},
	}
}

func (c *Crafty) StartMcServer(port int) error {
	bearer, err := c.getBearer()
	if err != nil {
		return fmt.Errorf("%w, %v", ErrAuthorizationFailed, err)
	}

	serverList, err := c.getServers(bearer)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToGetServers, err)
	}

	for _, server := range serverList.Data {
		if server.Port == port {
			c.sendStartServerRequest(server, bearer)
			return nil
		}
	}

	return ErrNoSuchServer
}

func (c *Crafty) StopMcServer(port int) error {
	bearer, err := c.getBearer()
	if err != nil {
		return fmt.Errorf("%w, %v", ErrAuthorizationFailed, err)
	}

	serverList, err := c.getServers(bearer)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToGetServers, err)
	}

	for _, server := range serverList.Data {
		if server.Port == port {
			c.sendStopServerRequest(server, bearer)
			return nil
		}
	}

	return ErrNoSuchServer
}

func (c *Crafty) sendStartServerRequest(server Server, bearer string) error {
	startServerUrl := c.apiUrl + "/api/v2/servers/" + server.ServerId + "/action/start_server"
	request, err := http.NewRequest(http.MethodPost, startServerUrl, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", bearer)
	_, err = c.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w, id %s, port %d: %v", ErrFailedToStartServer, server.ServerId, server.Port, err)
	}

	return nil
}

func (c *Crafty) sendStopServerRequest(server Server, bearer string) error {
	stopServerUrl := c.apiUrl + "/api/v2/servers/" + server.ServerId + "/action/stop_server"
	request, err := http.NewRequest(http.MethodPost, stopServerUrl, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", bearer)
	_, err = c.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w, id %s, port %d: %v", ErrFailedToStopServer, server.ServerId, server.Port, err)
	}

	return nil
}

func (c *Crafty) getBearer() (string, error) {
	loginBody := LoginPayload{
		Username: c.username,
		Password: c.password,
	}

	jsonData, err := json.Marshal(loginBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.apiUrl+"/api/v2/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrHTTPRequestFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFailedToReadBody, err)
	}

	var response LoginResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", response.Data.Token), nil
}

func (c *Crafty) getServers(bearer string) (ServerList, error) {
	request, _ := http.NewRequest(http.MethodGet, c.apiUrl+"/api/v2/servers", nil)
	request.Header.Add("Authorization", bearer)

	response, err := c.client.Do(request)
	if err != nil {
		return ServerList{}, fmt.Errorf("%w: %v", ErrHTTPRequestFailed, err)
	}
	defer response.Body.Close()

	serversListBody, err := io.ReadAll(response.Body)
	if err != nil {
		return ServerList{}, fmt.Errorf("%w: %v", ErrFailedToReadBody, err)
	}

	var serverList ServerList
	err = json.Unmarshal(serversListBody, &serverList)
	if err != nil {
		return ServerList{}, err
	}

	return serverList, nil
}
