package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func (a *API) setGraphAPI() {
	data := a.engine.Group("graph")
	data.POST("query", a.GraphQuery)
	data.POST("export", a.GraphExport)
}

func (a *API) GraphQuery(ctx *gin.Context) {
	req := &struct {
		Sql string `json:"sql"`
	}{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		responseJSON(ctx, serverError(err))
		return
	}

	sql := req.Sql
	// sql := ctx.Param("sql")

	// log.WithFields(logrus.Fields{
	// 	"sql": sql,
	// }).Debug("query")

	data, err := a.service.Nebula().Query(sql)
	if err != nil {
		responseJSON(ctx, serverError(err))
		return
	}
	responseJSON(ctx, errOK, "data", data)
}

func (a *API) GraphExport(ctx *gin.Context) {
	req := &struct {
		Space string `json:"space"`
		Path  string `json:"path"`
	}{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		responseJSON(ctx, serverError(err))
		return
	}

	basePath := req.Path

	// log.WithFields(logrus.Fields{
	// 	"sql": sql,
	// }).Debug("query")

	space := req.Space
	tagSql := fmt.Sprintf("USE %s;SHOW TAGS;", space)

	results, err := a.service.Nebula().Query(tagSql)
	if err != nil {
		responseJSON(ctx, serverError(err))
		return
	}

	tags := []string{}
	for _, rdata := range results {
		for _, data := range rdata.Data {
			for _, row := range data.Row {
				tags = append(tags, row.(string))
			}
		}
	}

	ids := []string{}
	for _, tag := range tags {
		sql := fmt.Sprintf("USE %s;MATCH (v:%s) RETURN v;", space, tag)
		results, err := a.service.Nebula().Query(sql)
		if err != nil {
			responseJSON(ctx, serverError(err))
			return
		}
		for _, rdata := range results {
			for _, data := range rdata.Data {
				for _, dmeta := range data.Meta {
					meta := dmeta.(map[string]interface{})
					id := meta["id"].(string)
					ids = append(ids, id)
				}
			}
		}
	}

	// rdfs := []string{}
	vertexs := []string{"id,attributes"}
	edges := []string{"type,src,dst,rank,name,attributes"}
	for _, vertex := range ids {
		sql := fmt.Sprintf("USE %s;GET SUBGRAPH WITH PROP 1 STEPS FROM '%s';", space, vertex)
		results, err := a.service.Nebula().Query(sql)
		if err != nil {
			responseJSON(ctx, serverError(err))
			return
		}
		for _, rdata := range results {
			for _, data := range rdata.Data {
				for idx, dmeta := range data.Meta {
					// fmt.Println("metaxxxxx:", dmeta)
					ameta := dmeta.([]interface{})
					for iidex, imeta := range ameta {
						meta := imeta.(map[string]interface{})
						mtype := meta["type"].(string)
						if mtype == "vertex" {
							id := meta["id"].(string)
							irow := data.Row[idx].([]interface{})
							drow := irow[iidex].(map[string]interface{})
							attribute, err := json.Marshal(drow)
							if err != nil {
								responseJSON(ctx, serverError(err))
								return
							}
							line := id + "," + string(attribute)
							vertexs = append(vertexs, line)
							// for key, val := range drow {
							// 	line := id + "," + key + "," + val.(string)
							// 	vertexs = append(vertexs, line)
							// 	rdfs = append(rdfs, line)
							// }
						} else if mtype == "edge" {
							id := meta["id"].(map[string]interface{})
							types := id["type"].(float64)
							src := id["src"].(string)
							dst := id["dst"].(string)
							name := id["name"].(string)
							rank := id["ranking"].(float64)
							irow := data.Row[idx].([]interface{})
							drow := irow[iidex].(map[string]interface{})
							attribute, err := json.Marshal(drow)
							if err != nil {
								responseJSON(ctx, serverError(err))
								return
							}
							line := fmt.Sprintf("%.0f,%s,%s,%.0f,%s,%s", types, src, dst, rank, name, string(attribute))
							edges = append(edges, line)
							// rdfs = append(rdfs, line)
						}
					}
				}
			}
		}
	}

	vertexPath := fmt.Sprintf("%s/%s_vertex.csv", basePath, req.Space)
	go WriteFile(vertexPath, vertexs, 0666)
	edgePath := fmt.Sprintf("%s/%s_edge.csv", basePath, req.Space)
	go WriteFile(edgePath, edges, 0666)

	responseJSON(ctx, errOK)
}

func WriteFile(filename string, list []string, perm os.FileMode) error {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer fd.Close()

	w := bufio.NewWriter(fd)
	for _, v := range list {
		_, err = w.WriteString(v + "\n")
		if err != nil {

			return err
		}
	}

	w.Flush()
	fd.Sync()
	return nil
}
