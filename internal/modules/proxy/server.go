package proxy

import (
	"context"
	"craftyreverseproxy/config"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
)

const tickerCooldown = 2 * time.Second
const awaitTimeout = 5 * time.Minute

type (
	Logger interface {
		Debug(format string, args ...any)
		Info(format string, args ...any)
		Error(format string, args ...any)
	}
	Crafty interface {
		StartMcServer(port int) error
		StopMcServer(port int) error
	}
)

type ProxyServer struct {
	playerCount int32

	listenPort int
	targetPort int
	listenAddr string
	targetAddr string
	protocol   string

	cfg    config.Config
	logger Logger
	crafty Crafty
}

func NewProxyServer(cfg config.Config, proxyCfg config.ServerType, logger Logger, crafty Crafty) *ProxyServer {
	return &ProxyServer{
		listenPort: proxyCfg.Listener.Port,
		targetPort: proxyCfg.ProxyHost.Port,
		protocol:   proxyCfg.Protocol,
		listenAddr: fmt.Sprintf("%s:%d", proxyCfg.Listener.Addr, proxyCfg.Listener.Port),
		targetAddr: fmt.Sprintf("%s:%d", proxyCfg.ProxyHost.Addr, proxyCfg.ProxyHost.Port),
		cfg:        cfg,
		logger:     logger,
		crafty:     crafty,
	}
}

func (ps *ProxyServer) ListenAndProxy(ctx context.Context) error {
	listener, err := net.Listen(ps.protocol, ps.listenAddr)
	if err != nil {
		return fmt.Errorf("%w with protocol %s, err: %w", ErrStartingServer, ps.protocol, err)
	}
	defer func() {
		listener.Close()
		ps.logger.Info("Listener closed for external port: %s", ps.targetPort)
	}()

	ps.logger.Info("%s: reverse proxy running on %s, forwarding to %s", ps.protocol, ps.listenAddr, ps.targetAddr)

	for {
		client, err := listener.Accept()
		if err != nil {
			ps.logger.Error("Failed to accept connection: %v", err)
			continue
		}

		go func() {
			if err := ps.handleClient(ctx, client); err != nil {
				ps.logger.Error("Failed to handle client: %v", err)
			}
		}()
	}
}

func (ps *ProxyServer) handleClient(ctx context.Context, client net.Conn) error {
	defer client.Close()

	ps.incrementPlayerCount()
	defer ps.decrementPlayerCount()

	serverConnection, err := net.Dial(ps.protocol, ps.targetAddr)
	if err != nil {
		ps.logger.Info("Server not up and running: %v", err)
		err := ps.crafty.StartMcServer(ps.targetPort)
		if err != nil {
			return err
		}

		ps.scheduleStopServerIfEmpty()
		serverConnection = ps.awaitForServerStart(ctx, ps.protocol, ps.targetAddr, awaitTimeout, tickerCooldown)
		if serverConnection == nil {
			return fmt.Errorf("failed awaiting for server start: %w", ErrTimeoutReached)
		}
	}

	defer serverConnection.Close()

	ps.logger.Info("Proxying from %s to %s", client.RemoteAddr(), serverConnection.RemoteAddr())

	completed := make(chan struct{})
	go func() {
		_, err := io.Copy(client, serverConnection)
		if err != nil {
			ps.logger.Error("Proxying from server to client failed: %v", err)
		}
		ps.logger.Info("Proxying from %s to %s completed", client.RemoteAddr(), serverConnection.RemoteAddr())
		completed <- struct{}{}
		close(completed)
	}()

	_, err = io.Copy(serverConnection, client)
	if err != nil {
		log.Printf("Error copying from client to server: %s", err)
	}

	<-completed

	return nil
}

func (ps *ProxyServer) awaitForServerStart(ctx context.Context, protocol, target string, timeout, cooldown time.Duration) net.Conn {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(cooldown)
	defer ticker.Stop()

	attempt := 1

	for {
		select {
		case <-ctx.Done():
			ps.logger.Error("Timeout waiting for server %s after %s", target, timeout)
			return nil
		case <-ticker.C:
			ps.logger.Debug("Attempt %d: connecting to %s (%s)", attempt, target, protocol)
			conn, err := net.Dial(protocol, target)
			if err == nil {
				ps.logger.Info("Server %s is up! Connected on attempt %d", target, attempt)
				return conn
			}
			ps.logger.Info("Connection attempt %d failed: %v", attempt, err)
			attempt++
		}
	}
}

func (ps *ProxyServer) scheduleStopServerIfEmpty() {
	if !ps.cfg.AutoShutdown {
		return
	}
	time.AfterFunc(ps.cfg.Timeout, func() {
		if ps.isServerEmpty() {
			ps.crafty.StopMcServer(ps.targetPort)
		}
	})
}

func (ps *ProxyServer) incrementPlayerCount() {
	atomic.AddInt32(&ps.playerCount, 1)
}

func (ps *ProxyServer) decrementPlayerCount() {
	atomic.AddInt32(&ps.playerCount, -1)
}

func (ps *ProxyServer) isServerEmpty() bool {
	return atomic.LoadInt32(&ps.playerCount) == 0
}
