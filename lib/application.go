package lib

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Application struct {
	ctx    context.Context
	logger *logrus.Entry
}

func NewApplication(log *logrus.Entry) *Application {
	return &Application{
		logger: log,
	}
}

// Initialize: initializes application context
func (a *Application) Initialize() {

	// Init DB connection, etc
}

// Start starting application loop
func (a *Application) Start() error {

	return nil
}

// Stop application handler. Correct place to stop all connections/DB/etc...
func (a *Application) Stop() {

}
