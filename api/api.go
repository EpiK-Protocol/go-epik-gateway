package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/EpiK-Protocol/go-epik-gateway/app/config"
	"github.com/EpiK-Protocol/go-epik-gateway/service"
	"github.com/EpiK-Protocol/go-epik-gateway/storage"
	"github.com/EpiK-Protocol/go-epik-gateway/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

type App interface {
	Config() config.Config
	Log() *logrus.Logger
	Storage() storage.Storage
	Service() service.IService
}

type API struct {
	conf    config.Config
	storage storage.Storage
	service service.IService

	engine *gin.Engine
}

func NewAPI(app App) (*API, error) {

	log = app.Log()
	engine := gin.New()
	if gin.Mode() == gin.DebugMode {
		engine.Use(gin.Logger())
	}
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	engine.Use(cors.New(config))
	engine.Use(gin.Recovery())
	if gin.Mode() == "debug" {
		engine.Use(ginBodyLogMiddleware)
	}
	api := &API{
		conf:    app.Config(),
		storage: app.Storage(),
		service: app.Service(),
		engine:  engine,
	}
	return api, nil
}

func (a *API) Start(ctx context.Context) error {
	if err := a.setupRouter(); err != nil {
		return err
	}
	go a.engine.Run(":" + utils.ParseString(a.conf.Server.Port))
	return nil
}

func (a *API) Stop(ctx context.Context) error {
	addr := ":" + utils.ParseString(a.conf.Server.Port)
	serve := http.Server{Addr: addr, Handler: a.engine}
	return serve.Shutdown(ctx)
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ginBodyLogMiddleware(c *gin.Context) {

	// fmt.Println("token:", c.Request.Header.Get("token"))
	if c.Request.Method == "POST" {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)
		if c.Request.URL.String() != "upload" {
			fmt.Printf("\033[1;32;40m%s\033[0m\n", string(body))
		}
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
	}
	c.Next()
}

func responseJSON(c *gin.Context, code Code, args ...interface{}) {
	body := make(map[string]interface{})
	body["code"] = code
	for i := 0; i < len(args); i += 2 {

		switch args[i].(type) {
		case string:
			body[args[i].(string)] = args[i+1]
			// break
		}
	}
	c.JSON(http.StatusOK, body)
}
