package app

import (
	"craftyreverseproxy/config"
	"craftyreverseproxy/internal/adapters/crafty"
	"fmt"
	"io"
	"log"
	"net"
)

func handleClient(client net.Conn, target string, server config.ServerType, protocol string) {
	incrementPlayerCount(server)
	defer decrementPlayerCount(server)
	serverConnection, err := net.Dial(protocol, target)
	if err != nil {
		println("Server not up and running: " + err.Error() + "\n")
		crafty.StartMcServer(server)
		scheduleStopServerIfEmpty(server)
		serverConnection = crafty.AwaitForServerStart(protocol, target)
		if serverConnection == nil {
			client.Close()
			return
		}
	}

	defer serverConnection.Close()
	defer client.Close()

	go func() {
		_, err := io.Copy(client, serverConnection)
		log.Printf("User connected!\n")
		if err != nil {
			log.Printf("Error copying from server to client: %s", err)
		}
	}()

	_, err = io.Copy(serverConnection, client)
	if err != nil {
		log.Printf("Error copying from client to server: %s", err)
	}
}

func handleMainServer(server config.ServerType) {
	listenAddr := server.ExternalIp + ":" + server.ExternalPort
	targetAddr := server.InternalIp + ":" + server.InternalPort

	listener, err := net.Listen(server.Protocol, listenAddr)
	if err != nil {
		log.Fatalf("Error starting "+server.Protocol+" server: %s\n", err)
	}
	defer func() {
		listener.Close()
		println("Listener closed for external port: " + server.ExternalPort + "\n")
	}()

	fmt.Printf(server.Protocol+" reverse proxy running on %s, forwarding to %s\n", listenAddr, targetAddr)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}

		go handleClient(client, targetAddr, server, server.Protocol)
	}
}

func handleSubServers(subServer config.OthersType, server config.ServerType) {
	listenAddr := subServer.ExternalIp + ":" + subServer.ExternalPort
	targetAddr := subServer.InternalIp + ":" + subServer.InternalPort

	listener, err := net.Listen(subServer.Protocol, listenAddr)
	if err != nil {
		log.Fatalf("Error starting "+server.Protocol+" server: %s", err)
	}
	defer listener.Close()

	fmt.Printf(subServer.Protocol+" reverse proxy running on %s, forwarding to %s\n", listenAddr, targetAddr)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}

		go handleClient(client, targetAddr, server, subServer.Protocol)
	}
}

func handleServer(server config.ServerType) {
	handleMainServer(server)

	for _, subServer := range server.Others {
		handleSubServers(subServer, server)
	}
}
