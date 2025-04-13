package proxy

import (
	"context"
	"craftyreverseproxy/config"
	"craftyreverseproxy/pkg/semaphore"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const tickerCooldown = 1 * time.Second
const awaitTimeout = 5 * time.Minute
const dialTimeout = 1 * time.Second

type (
	Logger interface {
		Debug(format string, args ...any)
		Info(format string, args ...any)
		Fail(format string, args ...any)
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

	startupSemaphore *semaphore.Semaphore
	shutdownTimer    *time.Timer
	mu               *sync.Mutex
}

func NewProxyServer(cfg config.Config, proxyCfg config.ServerType, logger Logger, crafty Crafty) *ProxyServer {
	return &ProxyServer{
		listenPort:       proxyCfg.Listener.Port,
		targetPort:       proxyCfg.ProxyHost.Port,
		protocol:         proxyCfg.Protocol,
		listenAddr:       fmt.Sprintf("%s:%d", proxyCfg.Listener.Addr, proxyCfg.Listener.Port),
		targetAddr:       fmt.Sprintf("%s:%d", proxyCfg.ProxyHost.Addr, proxyCfg.ProxyHost.Port),
		cfg:              cfg,
		logger:           logger,
		crafty:           crafty,
		startupSemaphore: semaphore.New(logger),
		mu:               &sync.Mutex{},
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

	serverConnection, err := ps.connectOrStartServer(ctx)
	if err != nil {
		return err
	}
	defer serverConnection.Close()

	ps.logger.Info("Starting proxy from %s to %s", client.RemoteAddr(), serverConnection.RemoteAddr())

	completed := make(chan struct{})
	go func() {
		defer func() {
			completed <- struct{}{}
			close(completed)
		}()
		_, err := io.Copy(client, serverConnection)
		if err != nil {
			ps.logger.Fail("An error occurred copying from server to client: %v", err)
		}
		ps.logger.Info("Proxying from %s to %s completed", client.RemoteAddr(), serverConnection.RemoteAddr())
	}()

	_, err = io.Copy(serverConnection, client)
	if err != nil {
		ps.logger.Error("Error copying from client to server: %s", err)
	}

	<-completed

	return nil
}

func (ps *ProxyServer) connectOrStartServer(ctx context.Context) (net.Conn, error) {
	serverConnection, err := net.DialTimeout(ps.protocol, ps.targetAddr, dialTimeout)
	if err != nil {
		acquired := ps.startupSemaphore.TryAcquire(ctx)
		if acquired {
			defer ps.startupSemaphore.Release()

			ps.logger.Info("Server is not running. Starting server with port %d", ps.targetPort)
			err := ps.crafty.StartMcServer(ps.targetPort)
			if err != nil {
				return nil, err
			}
		} else {
			ps.logger.Info("Server is starting up. Waiting for it to start...")
		}

		serverConnection, err = ps.awaitForServerStart(ctx, ps.protocol, ps.targetAddr, awaitTimeout, tickerCooldown)
		if err != nil {
			return nil, fmt.Errorf("failed awaiting for server start: %w", err)
		}
	}

	return serverConnection, nil
}

func (ps *ProxyServer) awaitForServerStart(ctx context.Context, protocol, target string, timeout, cooldown time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(cooldown)
	defer ticker.Stop()

	attempt := 1

	ps.logger.Info("Waiting for server :%d to start...", ps.targetPort)
	for {
		select {
		case <-ctx.Done():
			return nil, ErrTimeoutReached
		case <-ticker.C:
			ps.logger.Debug("Attempt %d: connecting to %s (%s)", attempt, target, protocol)
			conn, err := net.DialTimeout(protocol, target, dialTimeout)
			if err != nil {
				ps.logger.Fail("Connection attempt %d failed: %v", attempt, err)
				attempt++
				continue
			}
			ps.logger.Info("Server %s is up! Connected on attempt %d", target, attempt)
			return conn, nil
		}
	}
}

func (ps *ProxyServer) incrementPlayerCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.playerCount++

	if ps.shutdownTimer != nil {
		_ = ps.shutdownTimer.Stop()
		ps.shutdownTimer = nil
	}
}

func (ps *ProxyServer) decrementPlayerCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.playerCount--

	if ps.playerCount == 0 {
		ps.shutdownTimer = time.AfterFunc(ps.cfg.Timeout, func() {
			if ps.isServerEmpty() {
				ps.logger.Info("No players left, shutting down MC server with port %d", ps.targetPort)
				ps.crafty.StopMcServer(ps.targetPort)
			}
		})
	}
}

func (ps *ProxyServer) isServerEmpty() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.playerCount == 0
}
