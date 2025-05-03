package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/torderonex/load-balancer/internal/balancer"
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/handler"
	"github.com/torderonex/load-balancer/internal/limiter"
	"github.com/torderonex/load-balancer/internal/storage"
	"github.com/torderonex/load-balancer/pkg/server"
	"github.com/torderonex/load-balancer/pkg/sl"
)

type App struct {
	cfg         *config.Config
	proxyServer *server.Server
	restServer  *server.Server
	storage     *storage.Storage
	limiter     limiter.Limiter
	balancer    *balancer.Balancer
}

func MustNew(config *config.Config) *App {
	app := &App{
		cfg: config,
	}

	storage := storage.New(config)
	app.storage = storage

	//rate limiter init
	limiter := limiter.NewTokenBucket(&config.RateLimiter, storage)
	app.limiter = limiter

	//balancer init
	balancer := balancer.NewBalancer(config.Balancer.Strategy, limiter)
	balancer.AddBackends(config.Balancer.Backends)
	app.balancer = balancer

	//rest api server start
	handler := handler.NewHandler(storage)
	restServer := server.New(strconv.Itoa(config.ApiServer.Port), handler.InitRoutes(), config.ApiServer.ReadTimeout)
	app.restServer = restServer

	//balancer server start
	app.proxyServer = server.New(strconv.Itoa(config.ProxyServer.Port), balancer, config.ProxyServer.ReadTimeout)

	return app
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.balancer.StartHealthCheck(ctx, a.cfg.HealthCheck.CheckInterval)

	go a.limiter.StartRefill(ctx, a.cfg.RateLimiter.RefillInterval)

	errChan := make(chan error, 2)

	go func() {
		slog.Info(fmt.Sprintf("Starting REST API server on port %d", a.cfg.ApiServer.Port))
		if err := a.restServer.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("REST API server error", sl.Err(err))
			errChan <- fmt.Errorf("REST API server failed: %w", err)
		}
	}()

	go func() {
		slog.Info(fmt.Sprintf("Starting Proxy server on port %d", a.cfg.ProxyServer.Port))
		if err := a.proxyServer.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Proxy server error", sl.Err(err))
			errChan <- fmt.Errorf("proxy server failed: %w", err)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		slog.Error("Server failed to start or run, initiating shutdown", sl.Err(err))
	case sig := <-shutdownChan:
		slog.Info(fmt.Sprintf("Shutdown signal received, initiating graceful shutdown: %s", sig.String()))
	}

	slog.Info("Starting graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := a.restServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("REST API server graceful shutdown failed", sl.Err(err))
	} else {
		slog.Info("REST API server stopped gracefully")
	}

	if err := a.proxyServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Proxy server graceful shutdown failed", sl.Err(err))
	} else {
		slog.Info("Proxy server stopped gracefully")
	}

	err := a.storage.Close()
	if err != nil {
		slog.Error("Storage close error", sl.Err(err))
	}
	slog.Info("Graceful shutdown completed")
}

func (a *App) Start() {
	a.Run()
}

func (a *App) GetBalancer() *balancer.Balancer {
	return a.balancer
}

func (a *App) GetLimiter() limiter.Limiter {
	return a.limiter
}

func (a *App) GetStorage() *storage.Storage {
	return a.storage
}
