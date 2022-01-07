package service

import (
	"github.com/EpiK-Protocol/go-epik-gateway/app/config"
	"github.com/EpiK-Protocol/go-epik-gateway/storage"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

type App interface {
	Config() config.Config
	Storage() storage.Storage
	Log() *logrus.Logger
}

type IService interface {
	Nebula() *NebulasService
}

type Servive struct {
	nebula *NebulasService
}

func NewService(app App) (IService, error) {
	log = app.Log()
	nebula, err := NewNebula(app)
	if err != nil {
		return nil, err
	}
	return &Servive{
		nebula: nebula,
	}, nil
}

func (s *Servive) Nebula() *NebulasService {
	return s.nebula
}
