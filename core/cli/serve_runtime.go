package cli

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/contrib/websocket"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/initializer"
	"go.uber.org/zap"
)

// runServeLifecycle runs server listen loop and handles graceful shutdown signals.
func runServeLifecycle(runtime *initializer.Runtime, module *corehttp.Module, address string, listen ServeListenFunc) error {
	if runtime == nil || runtime.Logger == nil {
		return fmt.Errorf("runtime logger is required")
	}
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if listen == nil {
		listen = defaultListen
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	return runServeLifecycleWithSignals(runtime, module, address, listen, signals)
}

// runServeLifecycleWithSignals runs serve lifecycle using an injected signal channel.
func runServeLifecycleWithSignals(runtime *initializer.Runtime, module *corehttp.Module, address string, listen ServeListenFunc, signals <-chan os.Signal) error {
	listenErrors := make(chan error, 1)
	go func() { listenErrors <- listen(module, address) }()
	serverErr := error(nil)
	disposeErr := error(nil)
	interrupted := false
	select {
	case serverErr = <-listenErrors:
	case <-signals:
		interrupted = true
		runtime.Logger.Info("shutdown signal received")
		disposeErr = module.Dispose()
		if disposeErr != nil {
			runtime.Logger.Warn("http shutdown failed", zap.Error(disposeErr))
		}
		serverErr = <-listenErrors
	}
	if !interrupted {
		disposeErr = module.Dispose()
	}
	cleanupErr := cleanupServeRuntime(runtime)
	if interrupted {
		if serverErr != nil {
			runtime.Logger.Warn("http listen loop exited during shutdown", zap.Error(serverErr))
		}
		return errors.Join(disposeErr, cleanupErr)
	}
	return errors.Join(serverErr, disposeErr, cleanupErr)
}

// cleanupServeRuntime releases runtime-owned resources.
func cleanupServeRuntime(runtime *initializer.Runtime) error {
	redisErr := error(nil)
	if runtime.Redis != nil {
		redisErr = runtime.Redis.Close()
	}
	postgresErr := error(nil)
	if runtime.PostgreSQL != nil {
		sqlDatabase, dbErr := runtime.PostgreSQL.DB()
		if dbErr != nil {
			postgresErr = dbErr
		} else {
			postgresErr = sqlDatabase.Close()
		}
	}
	syncErr := runtime.Logger.Sync()
	if isIgnorableSyncError(syncErr) {
		syncErr = nil
	}
	if redisErr != nil || postgresErr != nil {
		return errors.Join(redisErr, postgresErr)
	}
	return syncErr
}

// isIgnorableSyncError reports logger sync errors that can be safely ignored.
func isIgnorableSyncError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EINVAL) {
		return true
	}
	if errors.Is(err, syscall.EBADF) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "invalid argument") || strings.Contains(message, "bad file descriptor")
}

// NewEchoWebSocketHandler returns an echo handler with debug packet telemetry.
func NewEchoWebSocketHandler(logger *zap.Logger) corehttp.WebSocketHandler {
	return func(connection *websocket.Conn) {
		remoteAddress := ""
		if connection.RemoteAddr() != nil {
			remoteAddress = connection.RemoteAddr().String()
		}
		for {
			messageType, payload, err := connection.ReadMessage()
			if err != nil {
				logger.Debug("websocket connection disposed", zap.String("remote", remoteAddress), zap.Error(err))
				return
			}
			logger.Debug("websocket packet received", zap.String("remote", remoteAddress), zap.Int("message_type", messageType), zap.Int("size", len(payload)))
			if err := connection.WriteMessage(messageType, payload); err != nil {
				logger.Debug("websocket connection disposed", zap.String("remote", remoteAddress), zap.Error(err))
				return
			}
			logger.Debug("websocket packet sent", zap.String("remote", remoteAddress), zap.Int("message_type", messageType), zap.Int("size", len(payload)))
		}
	}
}
