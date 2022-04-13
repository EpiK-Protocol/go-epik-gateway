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

	"github.com/EpiK-Protocol/go-epik-gateway/app/config"
	"github.com/EpiK-Protocol/go-epik-gateway/storage"
	"github.com/EpiK-Protocol/go-epik-gateway/utils"
	"github.com/asaskevich/EventBus"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

var (
	DownloadFilesKey = []byte("task:download")
)

type downloadTask struct {
	conf config.Config

	storage storage.Storage
	bus     EventBus.Bus

	lk    sync.Mutex
	files map[string]*FileRef

	quitChs     map[string]chan bool
	needRefresh bool

	page uint64
}

func newDownloadTask(conf config.Config, st storage.Storage, bus EventBus.Bus) (*downloadTask, error) {
	task := &downloadTask{
		conf:        conf,
		bus:         bus,
		storage:     st,
		files:       nil,
		quitChs:     map[string]chan bool{},
		needRefresh: false,
	}
	task.bus.Subscribe(FileEventNeedDownload, task.handleNeedDownload)
	return task, nil
}

func (t *downloadTask) handleNeedDownload(fileID string) {
	file, err := loadFile(t.storage, fileID)
	if err != nil {
		log.Errorf("failed to load file info:%s", fileID)
		return
	}

	t.lk.Lock()
	defer t.lk.Unlock()
	t.files[fileID] = file

	err = saveDatas(t.storage, DownloadFilesKey, t.files, false)
	if err != nil {
		log.Errorf("failed to save file info:%v", err)
	}
	log.Info("replay file:", fileID)
}

func (t *downloadTask) process(ctx context.Context) error {
	if t.files == nil {
		files, err := loadDatas(t.storage, DownloadFilesKey)
		if err != nil {
			return err
		}
		t.files = files
		log.WithFields(logrus.Fields{
			"count": len(t.files),
		}).Info("load download files.")
	}

	if err := t.fetchDatas(t.needRefresh); err != nil {
		return err
	}

	t.downloadDatas()

	return nil
}

func (t *downloadTask) stop() {
	for _, ch := range t.quitChs {
		ch <- true
	}
}

func (t *downloadTask) fetchDatas(reflesh bool) error {
	url := fmt.Sprintf("%s/sequence/allFileList?status=send&page=%d", t.conf.Server.DownloadUrl, t.page)
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

		file.Index = data.Index
		file.Count = data.Count
		file.Url = data.FileUrl
		file.Expert = data.Expert
		file.FileSize = data.FileSize
		file.CheckSum = data.CheckSum

		dir := t.conf.Storage.DataDir
		path := fmt.Sprintf("%s/%s", dir, file.ID)
		file.Path = path
		file.LocalPath = path

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
		t.page++
		return saveDatas(t.storage, DownloadFilesKey, t.files, false)
	}
	return nil
}

func (t *downloadTask) downloadDatas() {
	for _, file := range t.files {
		if file.Status > FileStatusDownloading {
			continue
		}
		go t.asyncDownload(file)
	}
}

func (t *downloadTask) asyncDownload(file *FileRef) {
	err := t.download(file)
	if err != nil {
		log.WithFields(logrus.Fields{
			"fileRef": file,
			"err":     err,
		}).Error("failed to download data.")
	}
}

func (t *downloadTask) download(file *FileRef) error {
	exist, err := utils.Exists(file.Path)
	if err != nil {
		return err
	}
	needDownload := false
	if exist {
		checkSum, err := getFileMd5(file.Path)
		if err != nil {
			return err
		}
		if len(file.CheckSum) > 0 && checkSum != file.CheckSum {
			needDownload = true
		}
	} else {
		needDownload = true
	}

	if needDownload {
		err = t.fileDownload(file.Url, file.Path)
		if err != nil {
			return err
		}
	}

	checkSum, err := getFileMd5(file.Path)
	if err != nil {
		return err
	}

	if len(file.CheckSum) > 0 && checkSum != file.CheckSum {
		log.WithFields(logrus.Fields{
			"ID":           file.ID,
			"fileChecksum": file.CheckSum,
			"osChecksum":   checkSum,
		}).Error("failed to check checksum.")
		return xerrors.Errorf("failed to check file checksum.")
	}
	if file.Status < FileStatusDownloading {
		file.Status = FileStatusDownloaded
		if err := saveFile(t.storage, file); err != nil {
			log.Errorf("failed to save file:%v", err)
			return err
		}
		log.Info("file downloaded:", file.ID)

		t.bus.Publish(FileEventDownloaded, file.ID)

		t.lk.Lock()
		defer t.lk.Unlock()
		delete(t.files, file.ID)

		delete(t.quitChs, file.ID)

		if err := saveDatas(t.storage, DownloadFilesKey, t.files, false); err != nil {
			log.Errorf("failed to save file:%v", err)
			return err
		}
	}

	return nil
}

func (t *downloadTask) fileDownload(url, path string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
