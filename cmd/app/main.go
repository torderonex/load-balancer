package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/torderonex/load-balancer/internal/balancer"
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/limiter"
	"github.com/torderonex/load-balancer/internal/storage"
	"github.com/torderonex/load-balancer/pkg/server"
	"github.com/torderonex/load-balancer/pkg/sl"
)

func main() {
	//load config
	config := config.MustLoad()
	fmt.Println(config)
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
	//http server start
	server := server.New(strconv.Itoa(config.HttpServer.Port), balancer, config.HttpServer.ReadTimeout)
	server.Run()
}
