package main

import (
	"flag"
	"os"
	"shortener/internal/app"
	"shortener/internal/config"
	"shortener/pkg/logger"
)

func main() {
	var pathToConfig string
	flag.StringVar(&pathToConfig, "config", pathToConfig, "path to config file")
	flag.Parse()

	cfg := config.MustLoadConfig(pathToConfig)

	log := logger.New(cfg.Env)

	log.Info("Start app")
	if err := app.Run(cfg, log); err != nil {
		log.Error("Failed to run app", "error", err)
		os.Exit(1)
	}
}
