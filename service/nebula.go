package service

import (
	"github.com/EpiK-Protocol/go-epik-data/app/config"
	"github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v2"
	"golang.org/x/xerrors"
)

var (
	nebLog = nebula.DefaultLogger{}
)

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

func (s *NebulasService) Query(sql string) (interface{}, error) {
	if err := s.generatePool(); err != nil {
		return nil, err
	}
	session, err := s.pool.GetSession(s.conf.Nebula.UserName, s.conf.Nebula.Password)
	if err != nil {
		return nil, err
	}
	defer session.Release()

	resultSet, err := session.Execute(sql)
	if err != nil {
		return nil, err
	}
	if !resultSet.IsSucceed() {
		return nil, xerrors.Errorf("nebula execute error sql:%s, code:%d, message:%s", sql, resultSet.GetErrorCode(), resultSet.GetErrorMsg())
	}

	result := make(map[string]interface{})
	result["columns"] = resultSet.GetColNames()
	result["space"] = resultSet.GetSpaceName()
	result["comment"] = resultSet.GetComment()

	records := make([]string, 0)
	rowSize := resultSet.GetRowSize()
	row := 0
	for row < rowSize {
		// Get a row from resultSet
		record, err := resultSet.GetRowValuesByIndex(row)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		records = append(records, record.String())
		row++
	}
	result["records"] = records

	log.WithFields(logrus.Fields{
		"sql":    sql,
		"result": result,
	}).Info("nebula query")

	// // Extract data from the resultSet
	// {
	// 	// Get all column names from the resultSet
	// 	colNames := resultSet.GetColNames()
	// 	fmt.Printf("column names: %s\n", strings.Join(colNames, ", "))

	// 	resultSet.GetRows()

	// 	// Get a row from resultSet
	// 	record, err := resultSet.GetRowValuesByIndex(0)
	// 	if err != nil {
	// 		log.Error(err.Error())
	// 	}
	// 	// Print whole row
	// 	fmt.Printf("row elements: %s\n", record.String())
	// 	// Get a value in the row by column index
	// 	valueWrapper, err := record.GetValueByIndex(0)
	// 	if err != nil {
	// 		log.Error(err.Error())
	// 	}
	// 	// Get type of the value
	// 	fmt.Printf("valueWrapper type: %s \n", valueWrapper.GetType())
	// 	// Check if valueWrapper is a string type
	// 	if valueWrapper.IsString() {
	// 		// Convert valueWrapper to a string value
	// 		v1Str, err := valueWrapper.AsString()
	// 		if err != nil {
	// 			log.Error(err.Error())
	// 		}
	// 		fmt.Printf("Result of ValueWrapper.AsString(): %s\n", v1Str)
	// 	}
	// 	// Print ValueWrapper using String()
	// 	fmt.Printf("Print using ValueWrapper.String(): %s", valueWrapper.String())
	// }

	return result, nil
}
