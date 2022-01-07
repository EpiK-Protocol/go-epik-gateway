package service

import (
	"encoding/json"

	"github.com/EpiK-Protocol/go-epik-gateway/app/config"
	"github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v2"
	"golang.org/x/xerrors"
)

var (
	nebLog = nebula.DefaultLogger{}
)

type ResultData struct {
	Row  []interface{} `json:"row"`
	Meta []interface{} `json:"meta"`
}

type Result struct {
	Columns     []string     `json:"columns"`
	Data        []ResultData `json:"data"`
	LatencyInUs int          `json:"latencyInUs"`
	SpaceName   string       `json:"spaceName"`
	PlanDesc    struct {
		PlanNodeDescs []struct {
			Name        string `json:"name"`
			ID          int    `json:"id"`
			OutputVar   string `json:"outputVar"`
			Description struct {
				Key string `json:"key"`
			} `json:"description"`
			Profiles []struct {
				Rows              int `json:"rows"`
				ExecDurationInUs  int `json:"execDurationInUs"`
				TotalDurationInUs int `json:"totalDurationInUs"`
				OtherStats        struct {
				} `json:"otherStats"`
			} `json:"profiles"`
			BranchInfo struct {
				IsDoBranch      bool `json:"isDoBranch"`
				ConditionNodeID int  `json:"conditionNodeId"`
			} `json:"branchInfo"`
			Dependencies []interface{} `json:"dependencies"`
		} `json:"planNodeDescs"`
		NodeIndexMap struct {
		} `json:"nodeIndexMap"`
		Format           string `json:"format"`
		OptimizeTimeInUs int    `json:"optimize_time_in_us"`
	} `json:"planDesc "`
	Comment string `json:"comment "`
}

type ResultError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Struct used for storing the parsed object
type ResultSet struct {
	Results []Result      `json:"results"`
	Errors  []ResultError `json:"errors"`
}

type NebulasService struct {
	conf config.Config

	pool *nebula.ConnectionPool
}

func NewNebula(app App) (*NebulasService, error) {
	return &NebulasService{
		conf: app.Config(),
		pool: nil,
	}, nil
}

func (s *NebulasService) generatePool() error {
	if s.pool == nil {
		host := nebula.HostAddress{Host: s.conf.Nebula.Address, Port: s.conf.Nebula.Port}
		hostList := []nebula.HostAddress{host}
		poolConf := nebula.GetDefaultConf()
		pool, err := nebula.NewConnectionPool(hostList, poolConf, nebLog)
		if err != nil {
			return err
		}
		s.pool = pool
	}
	return nil
}

func (s *NebulasService) Query(sql string) ([]Result, error) {
	if err := s.generatePool(); err != nil {
		return nil, err
	}
	session, err := s.pool.GetSession(s.conf.Nebula.UserName, s.conf.Nebula.Password)
	if err != nil {
		return nil, err
	}
	defer session.Release()

	resultSet, err := session.ExecuteJson(sql)
	if err != nil {
		return nil, err
	}

	var jsonObj ResultSet
	// Parse JSON
	json.Unmarshal(resultSet, &jsonObj)

	for _, resultErr := range jsonObj.Errors {
		if resultErr.Code != 0 {
			return nil, xerrors.Errorf("nebula execute error sql:%s, code:%d, message:%s", sql, resultErr.Code, resultErr.Message)
		}
	}

	log.WithFields(logrus.Fields{
		"sql":    sql,
		"result": jsonObj,
	}).Info("nebula query")

	return jsonObj.Results, nil
}
