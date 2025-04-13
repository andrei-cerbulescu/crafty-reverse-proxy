package crafty

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
	Port     int    `json:"server_port"`
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
