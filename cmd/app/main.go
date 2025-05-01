package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/torderonex/load-balancer/internal/balancer"
	"github.com/torderonex/load-balancer/internal/config"
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

	//rate limiter init

	//balancer init
	balancer := balancer.NewBalancer(config.Balancer.Strategy)
	balancer.AddBackends(config.Balancer.Backends)
	//health check goroutine start

	//ratelimit token goroutine start

	//http server start
	server := server.New(strconv.Itoa(config.HttpServer.Port), balancer, config.HttpServer.ReadTimeout)
	server.Run()
}
