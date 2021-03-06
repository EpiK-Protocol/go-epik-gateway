package task

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EpiK-Protocol/go-epik-gateway/app/config"
	"github.com/EpiK-Protocol/go-epik-gateway/storage"
	"github.com/asaskevich/EventBus"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

var (
	ReplayFilesKey = []byte("task:replay")
	nebLog         = nebula.DefaultLogger{}
	ReservedFields = []string{"GO", "AS", "TO", "OR", "AND", "XOR", "USE", "SET", "FROM", "WHERE", "MATCH", "INSERT", "YIELD", "RETURN", "DESCRIBE", "DESC", "VERTEX", "VERTICES", "EDGE", "EDGES", "UPDATE", "UPSERT", "WHEN", "DELETE", "FIND", "LOOKUP", "ALTER", "STEPS", "STEP", "OVER", "UPTO", "REVERSELY", "INDEX", "INDEXES", "REBUILD", "BOOL", "INT8", "INT16", "INT32", "INT64", "INT", "FLOAT", "DOUBLE", "STRING", "FIXED_STRING", "TIMESTAMP", "DATE", "TIME", "DATETIME", "TAG", "TAGS", "UNION", "INTERSECT", "MINUS", "NO", "OVERWRITE", "SHOW", "ADD", "CREATE", "DROP", "REMOVE", "IF", "NOT", "EXISTS", "WITH", "CHANGE", "GRANT", "REVOKE", "ON", "BY", "IN", "NOT_IN", "DOWNLOAD", "GET", "OF", "ORDER", "INGEST", "COMPACT", "FLUSH", "SUBMIT", "ASC", "ASCENDING", "DESCENDING", "DISTINCT", "FETCH", "PROP", "BALANCE", "STOP", "LIMIT", "OFFSET", "IS", "NULL", "RECOVER", "EXPLAIN", "PROFILE", "FORMAT", "CASE"}
)

type WriteRecord struct {
	Domain  string
	Index   int64
	Line    int64
	History map[int64]string
}

type replayTask struct {
	conf    config.Config
	storage storage.Storage
	bus     EventBus.Bus

	lk      sync.Mutex
	files   map[string]*FileRef
	records map[string]*WriteRecord

	nebulasPool *nebula.ConnectionPool

	quitChs      map[string]chan bool
	isProcessing bool
}

func newReplayTask(conf config.Config, st storage.Storage, bus EventBus.Bus) (*replayTask, error) {

	task := &replayTask{
		conf:         conf,
		storage:      st,
		bus:          bus,
		files:        nil,
		records:      map[string]*WriteRecord{},
		quitChs:      make(map[string]chan bool),
		isProcessing: false,
	}

	task.bus.Subscribe(FileEventDownloaded, task.handleStoraged)

	return task, nil
}

func (t *replayTask) handleStoraged(fileID string) {
	file, err := loadFile(t.storage, fileID)
	if err != nil {
		log.Errorf("failed to load file info:%s", fileID)
		return
	}

	t.lk.Lock()
	defer t.lk.Unlock()
	t.files[fileID] = file

	err = saveDatas(t.storage, ReplayFilesKey, t.files, false)
	if err != nil {
		log.Errorf("failed to save file info:%v", err)
	}
	log.Info("replay file:", fileID)
}

func (t *replayTask) onProcessing() bool {
	return t.isProcessing
}

func (t *replayTask) deleteExpert(expert string) {
	saveDatas(t.storage, ReplayFilesKey, t.files, false)
	t.deleteRecord(expert)
}

func (t *replayTask) process(ctx context.Context) error {
	if t.onProcessing() {
		return nil
	}

	t.isProcessing = true
	defer func() {
		t.isProcessing = false
	}()

	if t.files == nil {
		files, err := loadDatas(t.storage, ReplayFilesKey)
		if err != nil {
			return err
		}
		t.files = files

		log.WithFields(logrus.Fields{
			"count": len(files),
		}).Info("load replay data.")
	}

	if err := t.handleReplaies(ctx); err != nil {
		return err
	}

	return nil
}

func (t *replayTask) stop() {
	for _, ch := range t.quitChs {
		ch <- true
	}
	if t.nebulasPool != nil {
		t.nebulasPool.Close()
	}
}

func (t *replayTask) handleReplaies(ctx context.Context) error {
	for _, file := range t.files {
		log.WithFields(logrus.Fields{
			"id":    file.ID,
			"index": file.Index,
			"count": file.Count,
		}).Debug("start file replay.")
		if file.Status < FileStatusDownloaded {
			log.WithFields(logrus.Fields{
				"id":     file.ID,
				"status": file.Status,
			}).Error("file not download for replay.")
			t.bus.Publish(FileEventNeedDownload, file.ID)
			continue
		}
		if err := t.replayFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (t *replayTask) replayFile(file *FileRef) error {
	// log.Debug("parse file.")
	record, ok := t.records[file.Expert]
	if !ok {
		index := file.Index
		data, err := t.loadRecord(file.Expert)
		log.WithFields(logrus.Fields{
			"fileRef": file,
			"record":  data,
		}).Info("load record.")
		if err == storage.ErrKeyNotFound {
			data = &WriteRecord{
				Index:   1,
				Line:    0,
				History: map[int64]string{},
			}
		} else if err != nil {
			return err
		}
		data.History[index] = file.ID
		t.records[file.Expert] = data
		record = data
	} else {
		record.History[file.Index] = file.ID
	}
	if err := t.saveRecord(file.Expert, record); err != nil {
		return err
	}
	// TEST
	// if record.Index == 1 {
	// 	record.Line = 0
	// }
	// record.Index = 1
	// record.Line = 0

	fileID, ok := record.History[record.Index]
	if !ok {
		log.WithFields(logrus.Fields{
			"expert":  file.Expert,
			"index":   record.Index,
			"history": record.History,
		}).Warn("nebula index file not found.")
		return nil
	}
	file, err := loadFile(t.storage, fileID)
	if err != nil {
		return err
	}
	// update record
	line, err := t.readFileAndWrite(file, record)
	if line == 0 && err == nil {
		record.Index++
		record.Line = 0
		log.WithFields(logrus.Fields{
			"expert": file.Expert,
			"index":  record.Index,
		}).Info("Update record for next file.")
	}
	if err != nil {
		log.WithFields(logrus.Fields{
			"fileRef": file,
			"record":  record,
			"error":   err,
		}).Error("write nebula failed.")
		return err
	}

	return t.saveRecord(file.Expert, record)
}

func RecordKey(expert string) []byte {
	return []byte("task:replay:record:" + expert)
}

func (t *replayTask) deleteRecord(expert string) error {
	return t.storage.Del(RecordKey(expert))
}

func (t *replayTask) saveRecord(expert string, record *WriteRecord) error {
	bytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return t.storage.Put(RecordKey(expert), bytes)
}

func (t *replayTask) loadRecord(expert string) (*WriteRecord, error) {
	bytes, err := t.storage.Get(RecordKey(expert))
	if err != nil {
		return nil, err
	}
	var data WriteRecord
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (t *replayTask) readFileAndWrite(file *FileRef, record *WriteRecord) (int64, error) {
	line := int64(0)
	osfile, err := os.Open(file.Path)
	if err != nil {
		return 0, err
	}
	defer osfile.Close()
	scanner := bufio.NewScanner(osfile)
	scanner.Buffer([]byte{}, bufio.MaxScanTokenSize*100)
	domain := ""
	for scanner.Scan() {
		line++
		content := scanner.Text() // or
		//content := scanner.Bytes()
		log.WithFields(logrus.Fields{
			"id":           file.ID,
			"index":        record.Index,
			"content-line": line,
			"content":      content,
		}).Debug("scan nebula file.")
		// index := 0
		if line == 1 {
			headers := strings.Split(content, ",")
			domains := strings.Split(headers[0], ":")
			domain = strings.TrimSpace(domains[1])
			// indexs := strings.Split(headers[1], ":")
			// i, err := strconv.Atoi(indexs[1])
			// if err != nil {
			// 	return 0, err
			// }
			// index = i
			log.WithFields(logrus.Fields{
				"id":     file.ID,
				"domain": domain,
				"expert": file.Expert,
			}).Info("expert nebula header.")
			// if record.Index == 1 {
			// 	if err := t.writeToNebulaSql(line, domain, ""); err != nil {
			// 		return line - 1, err
			// 	}
			// }
			record.Domain = domain
		} else {
			// if record.Index > index {
			// 	continue
			// }
			if line <= record.Line {
				continue
			}
			if strings.Contains(strings.ToUpper(content), "CREATE SPACE") {
				contents := strings.Split(content, " ")
				space := strings.TrimSpace(contents[5])
				log.WithFields(logrus.Fields{
					"id":      file.ID,
					"space":   space,
					"domain":  domain,
					"expert":  file.Expert,
					"content": content,
				}).Info("expert nebula space.")
				if space != domain {
					domain = space
					record.Domain = domain
				}
			}
			if len(domain) == 0 {
				domain = t.conf.Server.ExpertSpaces[file.Expert]
				record.Domain = domain
				if len(domain) == 0 {
					return line - 1, fmt.Errorf("failed to find domain. expert:%s, index:%d", file.Expert, record.Index)
				}
			}
			if err := t.writeToNebulaSql(line, domain, content); err != nil {
				return line - 1, err
			}
			record.Line = line
			if err := t.saveRecord(file.Expert, record); err != nil {
				return line, err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return line, err
	}
	return 0, nil
}

func (t *replayTask) dropSpace(space string) error {
	sql := fmt.Sprintf("DROP SPACE IF EXISTS %s;", space)
	return t.writeToNebulaSql(0, space, sql)
}

func (t *replayTask) NebulaPool() (*nebula.ConnectionPool, error) {
	if t.nebulasPool == nil {
		host := nebula.HostAddress{Host: t.conf.Nebula.Address, Port: t.conf.Nebula.Port}
		hostList := []nebula.HostAddress{host}
		poolConf := nebula.GetDefaultConf()
		pool, err := nebula.NewConnectionPool(hostList, poolConf, nebLog)
		if err != nil {
			return nil, err
		}
		t.nebulasPool = pool
	}
	return t.nebulasPool, nil
}

func (t *replayTask) writeToNebulaSql(line int64, space string, content string) error {
	_, err := t.NebulaPool()
	if err != nil {
		return err
	}

	session, err := t.nebulasPool.GetSession(t.conf.Nebula.UserName, t.conf.Nebula.Password)
	if err != nil {
		return err
	}
	defer session.Release()

	{
		// createSchema := "CREATE SPACE IF NOT EXISTS basic_example_space(vid_type=FIXED_STRING(20)); " +
		// 	"USE basic_example_space;" +
		// 	"CREATE TAG IF NOT EXISTS person(name string, age int);" +
		// 	"CREATE EDGE IF NOT EXISTS like(likeness double)"
		sql := ""
		// if line == 0 {
		// 	sql = fmt.Sprintf("CREATE SPACE IF NOT EXISTS %s(vid_type=FIXED_STRING(64));", space)
		// } else
		{
			useSql := ""
			if !strings.Contains(strings.ToUpper(content), "CREATE SPACE") {
				useSql = fmt.Sprintf("USE %s;", space)
			}
			sql = useSql + content
		}
		if strings.Contains(strings.ToUpper(sql), "CREATE TAG") {
			strs := strings.Split(content, "(")
			strs = strings.Split(strs[0], " ")
			tag := strs[len(strs)-1]
			wrap := tag
			sql, wrap = replaceReservedFields(sql, tag)
			sql += fmt.Sprintf("CREATE TAG INDEX IF NOT EXISTS i_%s_value on %s(value(16));", tag, wrap)
		} else if strings.Contains(strings.ToUpper(sql), "CREATE EDGE") {
			strs := strings.Split(content, "(")
			strs = strings.Split(strs[0], " ")
			edge := strs[len(strs)-1]
			wrap := edge
			sql, wrap = replaceReservedFields(sql, edge)
			sql += fmt.Sprintf("CREATE EDGE INDEX IF NOT EXISTS i_%s_name on %s(name(16));", edge, wrap)
		} else {
			strs := strings.Split(content, "(")
			strs = strings.Split(strs[0], " ")
			sfield := strs[len(strs)-1]
			sql, _ = replaceReservedFields(sql, sfield)
		}
		// sql = fmt.Sprintf("DROP SPACE IF EXISTS %s;", space)
		resultSet, err := session.Execute(sql)
		if err != nil {
			return err
		}
		if !resultSet.IsSucceed() {
			return xerrors.Errorf("nebula execute error line:%d, sql:%s, code:%d, message:%s", line, sql, resultSet.GetErrorCode(), resultSet.GetErrorMsg())
		}
		if strings.Contains(strings.ToUpper(sql), "CREATE") {
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}

func replaceReservedFields(sql string, sfield string) (string, string) {
	for _, field := range ReservedFields {
		if field == strings.ToUpper(sfield) {
			upSql := strings.ToUpper(sql)
			fields := make(map[int]string)
			for {
				i := strings.Index(upSql, strings.ToUpper(sfield))
				if i == -1 {
					break
				}
				start := len(sql) - len(upSql) + i
				fields[start] = sql[start : start+len(field)]
				upSql = upSql[i+len(sfield):]
			}
			sfields := []string{}
			for i, f := range fields {
				if i > 0 && (sql[i-1:i] == " " || sql[i-1:i] == "`" || sql[i-1:i] == "\"" || sql[i-1:i] == "'") {
					sfields = append(sfields, f)
				}
			}
			wrapSql := sql
			for _, sf := range sfields {
				wrap := fmt.Sprintf("`%s`", sf)
				wrapSql = strings.ReplaceAll(wrapSql, sf, wrap)
			}

			log.WithFields(logrus.Fields{
				"fields":  sfields,
				"sql":     sql,
				"wrapSql": wrapSql,
			}).Warn("replace for reserved fields.")
			// panic(sql)
			return wrapSql, fmt.Sprintf("`%s`", sfield)
		}
	}
	return sql, sfield
}

// func (t *replayTask) createTagIndex(domain string) error {
// 	pool, err := t.NebulaPool()
// 	if err != nil {
// 		return err
// 	}
// 	session, err := pool.GetSession(t.conf.Nebula.UserName, t.conf.Nebula.Password)
// 	if err != nil {
// 		return err
// 	}
// 	defer session.Release()

// 	sql := fmt.Sprintf("USE %s, show tags;")
// 	return nil
// }
