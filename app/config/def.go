package config

import (
	"io/ioutil"

	"github.com/EpiK-Protocol/go-epik-data/utils/logging"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var log = logging.Log()

type Config struct {
	App App `yaml:"app"`

	Server  Server  `yaml:"server"`
	Storage Storage `yaml:"storage"`
	Chains  []Chain `yaml:"chains"`
	Nebula  Nebula  `yaml:"nebula"`
}

type App struct {
	Name     string `yaml:"name"`
	LogLevel string `yaml:"log_level"`
	LogDir   string `yaml:"log_dir"`
	LogAge   uint64 `yaml:"log_age"`
	Pprof    string `yaml:"pprof"`
	Version  string `yaml:"version"`

	KeyPath string `yaml:"key_path"`
}

type Server struct {
	// Host string
	// Port int64
	// Mode string `yaml:"mode"`

	// Name   string `yaml:"name"`
	// LogDir string `yaml:"log_dir"`
	RemoteHost string `yaml:"remote_host"`
	AppSecret  string `yaml:"app_secret"`
}

type Storage struct {
	DBDir   string `yaml:"db_dir"`
	DataDir string `yaml:"data_dir"`
}

type Chain struct {
	SSHHost string `yaml:"ssh_host"`
	SSHPort uint64 `yaml:"ssh_port"`
	SSHUser string `yaml:"ssh_user"`

	Miner    string `yaml:"miner"`
	RPCHost  string `yaml:"rpc_host"`
	RPCToken string `yaml:"rpc_token"`
}

type Nebula struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"user_name"`
	Password string `yaml:"password"`
}

var DefaultConfig Config

func Load(file string) (*Config, error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bs, &DefaultConfig)
	if err != nil {
		return nil, err
	}
	log.WithFields(logrus.Fields{
		"path": file,
	}).Info("load config.")
	return &DefaultConfig, nil
}
