package app

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/EpiK-Protocol/go-epik-gateway/api"
	"github.com/EpiK-Protocol/go-epik-gateway/app/config"
	"github.com/EpiK-Protocol/go-epik-gateway/service"
	"github.com/EpiK-Protocol/go-epik-gateway/storage"
	"github.com/EpiK-Protocol/go-epik-gateway/task"
	"github.com/EpiK-Protocol/go-epik-gateway/utils/logging"

	"github.com/asaskevich/EventBus"
)

type IApp interface {
	Config() config.Config
	Storage() storage.Storage
	Service() service.IService
}

var log *logrus.Logger

// App manages life cycle of services.
type App struct {
	context context.Context

	config config.Config

	log *logrus.Logger

	storage storage.Storage

	eventBus EventBus.Bus

	api *api.API

	task task.Task

	service service.IService

	lock sync.RWMutex

	running bool
}

// New returns a new app.
func New(config config.Config) (*App, error) { // init random seed.
	rand.Seed(time.Now().UTC().UnixNano())

	log = logging.Log()

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
		log:      logging.Log(),
		storage:  st,
		eventBus: bus,
		task:     task,
	}

	service, err := service.NewService(a)
	if err != nil {
		return nil, err
	}
	a.service = service

	api, err := api.NewAPI(a)
	if err != nil {
		return nil, err
	}
	a.api = api
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

	if err := a.api.Start(a.context); err != nil {
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

	if err := a.api.Stop(a.context); err != nil {
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

// Log returns Log reference.
func (n *App) Log() *logrus.Logger {
	return n.log
}

// Storage returns storage reference.
func (n *App) Storage() storage.Storage {
	return n.storage
}

// Service returns service reference.
func (n *App) Service() service.IService {
	return n.service
}
