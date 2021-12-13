package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/EpiK-Protocol/go-epik-data/app/config"
	"github.com/EpiK-Protocol/go-epik-data/storage"
	"github.com/EpiK-Protocol/go-epik-data/utils"
	byteutils "github.com/EpiK-Protocol/go-epik-data/utils/bytesutils"
	"github.com/asaskevich/EventBus"
	"github.com/sirupsen/logrus"
)

var (
	RetrieveFilesKey = []byte("task:retrieve")
)

func RetrievePageKey(expert string) []byte {
	key := fmt.Sprintf("task:retrieve:page:%s", expert)
	return []byte(key)
}

type retrieveTask struct {
	conf    config.Config
	storage storage.Storage
	bus     EventBus.Bus

	lk      sync.Mutex
	files   map[string]*FileRef
	experts []string

	page map[string]uint64

	isProcessing bool
	quitChs      map[string]chan bool
}

func newRetrieveTask(conf config.Config, st storage.Storage, bus EventBus.Bus) (*retrieveTask, error) {

	err := os.MkdirAll(conf.Storage.DataDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	task := &retrieveTask{
		conf:         conf,
		storage:      st,
		bus:          bus,
		files:        nil,
		experts:      []string{"f01005"},
		quitChs:      make(map[string]chan bool),
		isProcessing: false,
		page:         make(map[string]uint64),
	}

	task.bus.Subscribe(FileEventDownloaded, task.handleNeedDownload)
	return task, nil
}

func (t *retrieveTask) handleNeedDownload(fileID string) {
	file, err := loadFile(t.storage, fileID)
	if err != nil {
		log.Errorf("failed to load file info:%s", fileID)
		return
	}

	t.lk.Lock()
	defer t.lk.Unlock()
	t.files[fileID] = file

	err = saveDatas(t.storage, RetrieveFilesKey, t.files, false)
	if err != nil {
		log.Errorf("failed to save file info:%v", err)
	}
	log.Info("replay file:", fileID)
}

func (t *retrieveTask) onProcessing() bool {
	return t.isProcessing
}

func (t *retrieveTask) process(ctx context.Context) error {
	if t.onProcessing() {
		return nil
	}

	t.isProcessing = true
	defer func() {
		t.isProcessing = false
	}()

	if t.files == nil {
		for _, expert := range t.experts {
			// val, err := t.storage.Get(RetrievePageKey(expert))
			// if err != nil && err != storage.ErrKeyNotFound {
			// 	return err
			// }
			// if err == storage.ErrKeyNotFound {
			// 	t.page[expert] = 0
			// } else {
			// 	t.page[expert] = byteutils.Uint64(val)
			// }
			t.page[expert] = 0
		}

		files, err := loadDatas(t.storage, RetrieveFilesKey)
		if err != nil {
			return err
		}
		t.files = files
		log.WithFields(logrus.Fields{
			"count": len(files),
		}).Info("load import data.")
	}

	if err := t.retrieveData(ctx); err != nil {
		return err
	}

	return nil
}

func (t *retrieveTask) retrieveData(ctx context.Context) error {
	for _, expert := range t.experts {
		reflesh := false
		if err := t.fetchDatas(reflesh, expert); err != nil {
			log.WithFields(logrus.Fields{
				"count": len(t.files),
			}).Error("failed to fetch retrieve data.")
		}
	}

	for _, file := range t.files {
		if file.Status >= FileStatusDownloaded {
			continue
		}
		// chain := t.conf.Chains[0]

		// conf := utils.SSHConfig{
		// 	IP:             chain.SSHHost,
		// 	Port:           chain.SSHPort,
		// 	UserName:       chain.SSHUser,
		// 	Password:       "",
		// 	PrivateKeyPath: t.conf.App.KeyPath,
		// }

		// TEST
		// file.Index = 1
		// file.Path = "/root/data/d4ae9e27-0b65-4e92-8d17-2a601f8e6511"
		// checkCmd := fmt.Sprintf("test -f %s", file.Path)
		// if _, err := utils.SSHRun(conf, checkCmd); err != nil {
		// 	log.WithFields(logrus.Fields{
		// 		"id":    file.ID,
		// 		"error": err,
		// 	}).Warnf("check file failed.")
		// 	t.retrieveFile(conf, chain, file)
		// }
		if err := t.downloadFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (t *retrieveTask) retrieveFile(conf utils.SSHConfig, chain config.Chain, file *FileRef) error {
	log.Debug("retrieve file.")
	cmd := fmt.Sprintf("epik client retrieve --pieceCid=%s --miner=%s %s %s", file.PieceCID, chain.Miner, file.RootCID, file.Path)
	_, err := utils.SSHRun(conf, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (t *retrieveTask) downloadFile(file *FileRef) error {
	path := fmt.Sprintf("%s/%s", t.conf.Storage.DataDir, file.ID)
	file.Path = path

	log.WithFields(logrus.Fields{
		"path": file.Path,
	}).Debug("download file.")

	md5, err := getFileMd5(file.Path)
	if err == nil && md5 == file.CheckSum {
		log.WithFields(logrus.Fields{
			"fileRef": file,
		}).Debug("file has downloaded.")
	} else {
		out, err := os.Create(path)
		defer out.Close()

		resp, err := http.Get(file.Url)
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download bad status: %s", resp.Status)
		}

		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}

		md5, err = getFileMd5(file.Path)
		if md5 != file.CheckSum || err != nil {
			log.WithFields(logrus.Fields{
				"fileRef": file,
			}).Error("file download failed.")
			return nil
		}
	}

	file.Status = FileStatusDownloaded
	err = saveFile(t.storage, file)
	if err != nil {
		return err
	}

	log.Info("file downloaded:", file.ID)

	t.bus.Publish(FileEventDownloaded, file.ID)

	return nil
}

func (t *retrieveTask) stop() {
	for _, ch := range t.quitChs {
		ch <- true
	}
}

func (t *retrieveTask) fetchDatas(reflesh bool, expert string) error {
	url := fmt.Sprintf("%s/sequence/allFileList?status=upload&page=%d&expert=%s", t.conf.Server.RemoteHost, t.page[expert], expert)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respData ListResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return err
	}
	// log.WithFields(logrus.Fields{
	// 	"url":   url,
	// 	"count": len(respData.List),
	// }).Debug("fetch download files.")
	if len(respData.List) == 0 {
		return nil
	}

	log.WithFields(logrus.Fields{
		"count": len(respData.List),
	}).Info("fetch download files.")

	t.lk.Lock()
	defer t.lk.Unlock()
	listChanged := false
	for _, data := range respData.List {
		file, err := loadFile(t.storage, data.Id)
		if err != nil {
			if err == storage.ErrKeyNotFound {
				file = &FileRef{
					ID:     data.Id,
					Status: FileStatusNew,
				}
			} else {
				return err
			}
		}

		if reflesh && file.Status >= FileStatusDownloaded {
			file.Status = FileStatusNew
		}

		found := false
		for _, expert := range t.experts {
			if expert == data.Expert {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		file.Index = data.Index
		file.Count = data.Count
		file.Url = data.FileUrl
		file.Expert = data.Expert
		file.FileSize = data.FileSize
		file.CheckSum = data.CheckSum

		if file.Status < FileStatusDownloaded {
			listChanged = true
			t.files[data.Id] = file
			if err := saveFile(t.storage, file); err != nil {
				return err
			}

			log.WithFields(logrus.Fields{
				"resp": data,
				"file": file,
			}).Info("add download files.")
		}
	}

	if listChanged {
		t.page[expert] = t.page[expert] + 1
		t.storage.Put(RetrievePageKey(expert), byteutils.FromUint64(t.page[expert]))
		return saveDatas(t.storage, ReplayFilesKey, t.files, false)
	}
	return nil
}
