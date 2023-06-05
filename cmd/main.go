package main

import (
	"log"

	"kaori/internal/app"
	"kaori/internal/config"

	"go.uber.org/zap"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	logger, err := createLogger(config)
	if err != nil {
		log.Fatal(err)
	}

	app.Start(&config, logger)
}

func createLogger(config config.Config) (*zap.SugaredLogger, error) {
	var logger *zap.Logger
	var err error
	if config.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
