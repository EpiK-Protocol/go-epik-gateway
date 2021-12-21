package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EpiK-Protocol/go-epik-data/app/config"
	"github.com/EpiK-Protocol/go-epik-data/storage"
	"github.com/EpiK-Protocol/go-epik-data/utils/logging"
	"github.com/asaskevich/EventBus"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

type TaskManager struct {
	config  config.Config
	storage storage.Storage

	retrieveTask *retrieveTask
	replayTask   *replayTask

	stop chan bool
}

func NewTask(conf config.Config, st storage.Storage, bus EventBus.Bus) (Task, error) {

	log = logging.Log()
	retrieveTask, err := newRetrieveTask(conf, st, bus)
	if err != nil {
		return nil, err
	}

	replayTask, err := newReplayTask(conf, st, bus)
	if err != nil {
		return nil, err
	}

	return &TaskManager{
		config:  conf,
		storage: st,

		retrieveTask: retrieveTask,
		replayTask:   replayTask,
	}, nil
}

func (t *TaskManager) Start(ctx context.Context) error {
	log.Info("start task.")
	if t.stop != nil {
		return fmt.Errorf("task already started")
	}
	t.stop = make(chan bool, 1)
	go t.process(ctx)
	return nil
}

func (t *TaskManager) process(ctx context.Context) {
	for {
		select {
		case <-t.stop:
			t.stop = nil
			return
		default:
		}

		go func() {
			if err := t.retrieveTask.process(ctx); err != nil {
				log.Errorf("failed to retrieve: %v", err)
			}
		}()

		go func() {
			if err := t.replayTask.process(ctx); err != nil {
				log.Errorf("failed to replay: %v", err)
			}
		}()

		if next := t.niceSleep(3 * 60 * time.Second); !next {
			return
		}
	}
}

func (t *TaskManager) niceSleep(d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-t.stop:
		t.stop = nil
		return false
	}
}

func (t *TaskManager) Stop(ctx context.Context) error {
	log.Info("stop task.")
	t.stop <- true

	t.retrieveTask.stop()
	t.replayTask.stop()

	return nil
}

func loadFileList(st storage.Storage, key []byte) ([]string, error) {
	bytes, err := st.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err == storage.ErrKeyNotFound {
		return []string{}, nil
	}
	ids := []string{}
	if err := json.Unmarshal(bytes, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func loadFile(st storage.Storage, fileID string) (*FileRef, error) {
	d, err := st.Get([]byte(fileID))
	if err != nil {
		return nil, err
	}
	var file FileRef
	if err := json.Unmarshal(d, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func loadDatas(st storage.Storage, key []byte) (map[string]*FileRef, error) {
	files := make(map[string]*FileRef)
	ids, err := loadFileList(st, key)
	if err != nil {
		return nil, err
	}
	for _, i := range ids {
		file, err := loadFile(st, i)
		if err != nil {
			return nil, err
		}
		files[i] = file
	}
	return files, nil
}

func saveFileList(st storage.Storage, key []byte, list []string) error {
	fbytes, err := json.Marshal(list)
	if err != nil {
		return err
	}
	if err := st.Put(key, fbytes); err != nil {
		return err
	}
	return nil
}

func saveFile(st storage.Storage, file *FileRef) error {
	bytes, err := file.Marshal()
	if err != nil {
		return err
	}
	if err := st.Put([]byte(file.ID), bytes); err != nil {
		return err
	}
	return nil
}

func saveDatas(st storage.Storage, key []byte, files map[string]*FileRef, fileSave bool) error {
	fkeys := []string{}
	for fkey, file := range files {
		fkeys = append(fkeys, fkey)

		if fileSave {
			if err := saveFile(st, file); err != nil {
				return err
			}
		}
	}
	saveFileList(st, key, fkeys)
	return nil
}
