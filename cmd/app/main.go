package main

import (
	"log/slog"

	"github.com/torderonex/load-balancer/internal/app"
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/pkg/sl"
)

func main() {
	//load config
	config := config.MustLoad()
	//init logger
	slog.SetDefault(sl.Setup(config.Logger.Level))

	app := app.MustNew(config)
	app.Start()
}
