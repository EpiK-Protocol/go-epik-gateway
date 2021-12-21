package main

import (
	"github.com/urfave/cli/v2"
)

var (
	// ConfigFlag config file path
	ConfigFlag = cli.StringFlag{
		Name:        "config, c",
		Usage:       "load configuration from `FILE`",
		Value:       "conf/config.yml",
		Destination: &configPath,
	}
)
