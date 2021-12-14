package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (a *API) setGraphAPI() {
	data := a.engine.Group("graph")
	data.POST("query", a.GraphQuery)
}

func (a *API) GraphQuery(ctx *gin.Context) {
	// req := &struct {
	// 	Sql string `json:"sql"`
	// }{}
	// if err := ctx.ShouldBindJSON(req); err != nil {
	// 	responseJSON(ctx, serverError(err))
	// 	return
	// }

	sql := ctx.PostForm("sql")

	log.WithFields(logrus.Fields{
		"sql": sql,
	}).Debug("query")

	data, err := a.service.Nebula().Query(sql)
	if err != nil {
		responseJSON(ctx, serverError(err))
		return
	}
	responseJSON(ctx, errOK, "data", data)
}
