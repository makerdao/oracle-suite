package main

import (
	"log"
	"makerdao/gofer/lib"

	"github.com/micro/go-micro/v2/config/source/env"
	"github.com/micro/go-micro/v2/config/source/file"

	"github.com/micro/go-micro/v2/config"
	"github.com/sirupsen/logrus"
)

func main() {
	err := config.Load(
		// base config from file
		file.NewSource(
			file.WithPath("./config/config.json"),
		),
		// base config from env
		env.NewSource(),
	)
	if err != nil {
		log.Fatalln(err)
	}

	logger := initLogger()

	logger.Info("Initializing application...")
	app := lib.NewApplication(logger)
	app.Initialize()

	logger.Info("Application initialized. Starting...")
	if err := app.Start(); err != nil {
		logger.Error(err)
	}
	logger.Info("Shutting down.")

	app.Stop()
	logger.Info("Application stopped.")
}

func initLogger() *logrus.Entry {
	level, err := logrus.ParseLevel(config.Get("logger", "level").String("debug"))
	if err != nil {
		log.Fatalln(err)
	}
	logrus.SetLevel(level)
	logger := logrus.WithField("app", config.Get("env").String("dev"))
	logger.Info("Logger initialized with level: ", level)

	return logger
}
