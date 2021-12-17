package main

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/EpiK-Protocol/go-epik-data/app"
	"github.com/EpiK-Protocol/go-epik-data/app/config"
	"github.com/EpiK-Protocol/go-epik-data/utils/logging"
)

var (
	version    string
	commit     string
	branch     string
	compileAt  string
	configPath string
)

func main() {

	app := cli.NewApp()
	app.Action = action
	app.Name = "app"
	app.Version = fmt.Sprintf("%s, branch %s, commit %s", version, branch, commit)
	timestamp, _ := strconv.ParseInt(compileAt, 10, 64)
	app.Compiled = time.Unix(timestamp, 0)
	app.Usage = "the command line interface"
	app.Copyright = ""

	app.Flags = append(app.Flags, &ConfigFlag)

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}

func action(ctx *cli.Context) error {
	n, err := makeNode(ctx)
	if err != nil {
		panic(err)
	}

	select {
	case <-runNode(ctx, n):
		return nil
	}
}

func makeNode(ctx *cli.Context) (*app.App, error) {
	conf, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	conf.App.Version = version
	logging.Init(conf.App.LogDir, conf.App.Name, conf.App.LogLevel, conf.App.LogAge)

	// load config from cli args

	n, err := app.New(*conf)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func runNode(ctx *cli.Context, a *app.App) chan bool {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if err := a.Start(); err != nil {
		logging.Log().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to start app.")
	}

	quitCh := make(chan bool, 1)

	go func() {
		select {
		case <-c:
			if err := a.Stop(); err != nil {
				logging.Log().WithFields(logrus.Fields{
					"err": err,
				}).Fatal("Failed to stop app.")
			}
			quitCh <- true

		}
	}()

	return quitCh
}

// FatalF fatal format err
func FatalF(format string, args ...interface{}) {
	err := fmt.Sprintf(format, args...)
	fmt.Println(err)
	os.Exit(1)
}
