package main

import (
	"log/slog"
	"strconv"

	"github.com/torderonex/load-balancer/internal/balancer"
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/handler"
	"github.com/torderonex/load-balancer/internal/limiter"
	"github.com/torderonex/load-balancer/internal/storage"
	"github.com/torderonex/load-balancer/pkg/server"
	"github.com/torderonex/load-balancer/pkg/sl"
)

func main() {
	//load config
	config := config.MustLoad()
	//init logger
	slog.SetDefault(sl.Setup(config.Logger.Level))
	//storage init
	storage := storage.New()

	//rate limiter init
	limiter := limiter.NewTokenBucket(&config.RateLimiter, storage)

	//balancer init
	balancer := balancer.NewBalancer(config.Balancer.Strategy, limiter)
	balancer.AddBackends(config.Balancer.Backends)
	//health check goroutine start
	go balancer.StartHealthCheck(config.HealthCheck.CheckInterval)
	//ratelimit token goroutine start
	go limiter.StartRefill(config.RateLimiter.RefillInterval)

	//rest api server start
	handler := handler.NewHandler(storage)
	restServer := server.New(strconv.Itoa(config.HttpServer.Port), handler.InitRoutes(), config.HttpServer.ReadTimeout)
	go func() {
		restServer.Run()
	}()

	//balancer server start
	server.New(strconv.Itoa(config.HttpServer.Port), balancer, config.HttpServer.ReadTimeout).Run()
}
