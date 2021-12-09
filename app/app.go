package app

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/EpiK-Protocol/go-epik-data/app/config"
	"github.com/EpiK-Protocol/go-epik-data/storage"
	"github.com/EpiK-Protocol/go-epik-data/task"
	"github.com/EpiK-Protocol/go-epik-data/utils/logging"

	"github.com/asaskevich/EventBus"
)

var log = logging.Log()

// App manages life cycle of services.
type App struct {
	context context.Context

	config config.Config

	storage storage.Storage

	eventBus EventBus.Bus

	task task.Task

	lock sync.RWMutex

	running bool
}

// New returns a new app.
func New(config config.Config) (*App, error) { // init random seed.
	rand.Seed(time.Now().UTC().UnixNano())

	// storage
	st, err := storage.NewBadgerStorage(config.Storage.DBDir)
	// storage, err = storage.NewMemoryStorage()
	if err != nil {
		return nil, err
	}

	bus := EventBus.New()

	task, err := task.NewTask(config, st, bus)
	if err != nil {
		return nil, err
	}
	a := &App{
		context:  context.TODO(),
		config:   config,
		storage:  st,
		eventBus: bus,
		task:     task,
	}
	return a, nil
}

// Setup setup context
// func (a *App) Setup() {
// 	log.Info("Setuping App...")

// 	log.Info("Setuped App.")
// }

// Start starts the services of the context.
func (a *App) Start() error {
	a.lock.Lock()
	defer a.lock.Unlock()

	log.Info("Starting App...")

	if a.running {
		log.WithFields(logrus.Fields{
			"err": "app is already running",
		}).Fatal("Failed to start app.")
	}
	a.running = true

	if err := a.task.Start(a.context); err != nil {
		return err
	}

	log.Info("Started App.")
	return nil
}

// Stop stops the services of the context.
func (a *App) Stop() error {
	a.lock.Lock()
	defer a.lock.Unlock()

	log.Info("Stopping APP...")

	if err := a.task.Stop(a.context); err != nil {
		return err
	}

	a.running = false

	log.Info("Stopped App.")
	return nil
}

// Config returns context configuration.
func (n *App) Config() config.Config {
	return n.config
}

// Storage returns storage reference.
func (n *App) Storage() storage.Storage {
	return n.storage
}
