package task

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/EpiK-Protocol/go-epik-gateway/app/config"
	"github.com/EpiK-Protocol/go-epik-gateway/epik/api"
	"github.com/EpiK-Protocol/go-epik-gateway/epik/api/client"
	"github.com/EpiK-Protocol/go-epik-gateway/storage"
	"github.com/EpiK-Protocol/go-epik-gateway/utils"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	"github.com/asaskevich/EventBus"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/ipfs/go-cid"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

var (
	RetrieveFilesKey = []byte("task:retrieve")
)

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
		experts:      conf.Server.Experts,
		quitChs:      make(map[string]chan bool),
		isProcessing: false,
		page:         make(map[string]uint64),
	}

	task.bus.Subscribe(FileEventNeedDownload, task.handleNeedDownload)

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
		if err := t.fetchDatas(ctx, false, expert); err != nil {
			log.WithFields(logrus.Fields{
				"count": len(t.files),
			}).Error("failed to fetch retrieve data.")
		}
	}

	for _, file := range t.files {
		if file.Status >= FileStatusDownloaded {
			continue
		}

		exist, err := utils.Exists(file.LocalPath)
		if err != nil {
			return err
		}
		if exist {
			return t.updateFileStatus(file)
		}

		chain := t.conf.Chains[0]

		conf := utils.SSHConfig{
			IP:             chain.SSHHost,
			Port:           chain.SSHPort,
			UserName:       chain.SSHUser,
			Password:       "",
			PrivateKeyPath: t.conf.App.KeyPath,
		}

		// TEST
		// file.Index = 1
		// file.Path = "/root/data/d4ae9e27-0b65-4e92-8d17-2a601f8e6511"
		checkCmd := fmt.Sprintf("mkdir -p %s;test -f %s", t.conf.Storage.DataDir, file.Path)
		if _, err := utils.SSHRun(conf, checkCmd); err != nil {
			log.WithFields(logrus.Fields{
				"id":    file.ID,
				"error": err,
			}).Warnf("remote file not found.")

			if err := t.exportFile(ctx, chain, file); err != nil {
				log.WithFields(logrus.Fields{
					"id":      file.ID,
					"pieceID": file.PieceCID,
					"rootID":  file.RootCID,
					"error":   err,
				}).Warn("failed to export data.")
				err = t.retrieveFile(conf, chain, file)
				if err != nil {
					log.WithFields(logrus.Fields{
						"id":      file.ID,
						"pieceID": file.PieceCID,
						"rootID":  file.RootCID,
						"error":   err,
					}).Error("failed to retrieve data.")
					continue
				}
			}
		}
		if err := t.downloadFile(conf, file); err != nil {
			return err
		}
	}
	return nil
}

func (t *retrieveTask) fetchDatas(ctx context.Context, reflesh bool, expertStr string) error {

	expert, err := address.NewFromString(expertStr)
	if err != nil {
		return err
	}
	chain := t.conf.Chains[0]
	client, closer, err := getFullAPI(ctx, chain)
	if err != nil {
		return err
	}
	defer closer()

	infos, err := client.StateExpertDatas(ctx, expert, nil, false, types.EmptyTSK)
	if err != nil {
		return err
	}
	// log.WithFields(logrus.Fields{
	// 	"url":   url,
	// 	"count": len(respData.List),
	// }).Debug("fetch download files.")
	if len(infos) == 0 {
		return nil
	}

	log.WithFields(logrus.Fields{
		"expert": expertStr,
		"count":  len(infos),
	}).Info("fetch download files.")

	t.lk.Lock()
	defer t.lk.Unlock()
	listChanged := false
	for _, info := range infos {
		file, err := loadFile(t.storage, info.PieceID)
		if err != nil {
			if err == storage.ErrKeyNotFound {
				file = &FileRef{
					ID:     info.PieceID,
					Status: FileStatusNew,
				}
			} else {
				return err
			}
		}

		if reflesh && file.Status >= FileStatusDownloaded {
			file.Status = FileStatusNew
		}

		pieceID, err := cid.Parse(info.PieceID)
		if err != nil {
			return err
		}

		rootID, err := cid.Parse(info.RootID)
		if err != nil {
			return err
		}

		file.Expert = expertStr
		file.PieceCID = pieceID
		file.RootCID = rootID
		file.FileSize = int64(info.PieceSize)
		file.Path = fmt.Sprintf("%s/%s", t.conf.Storage.DataDir, file.PieceCID)
		file.LocalPath = file.Path

		if file.Status < FileStatusDownloaded {
			listChanged = true
			t.files[file.ID] = file
			if err := saveFile(t.storage, file); err != nil {
				return err
			}

			log.WithFields(logrus.Fields{
				"resp": info,
				"file": file,
			}).Info("add download files.")
		}
	}

	if listChanged {
		return saveDatas(t.storage, RetrieveFilesKey, t.files, false)
	}
	return nil
}

func (t *retrieveTask) exportFile(ctx context.Context, chain config.Chain, file *FileRef) error {
	client, closer, err := getFullAPI(ctx, chain)
	if err != nil {
		return err
	}
	defer closer()

	data, err := client.ClientDealPieceCID(ctx, file.RootCID)
	if err != nil {
		return err
	}
	if data.PieceCID != file.PieceCID {
		return fmt.Errorf("failed to parse file pieceID:%s", data.PieceCID)
	}

	return client.ClientExport(ctx, api.ExportRef{Root: file.RootCID}, api.FileRef{Path: file.Path})
}

func (t *retrieveTask) retrieveFile(conf utils.SSHConfig, chain config.Chain, file *FileRef) error {
	cmd := fmt.Sprintf("epik client retrieve --pieceCid=%s --miner=%s %s %s", file.PieceCID, chain.Miner, file.RootCID, file.Path)
	log.WithFields(logrus.Fields{
		"cmd":     cmd,
		"pieceID": file.PieceCID,
		"rootID":  file.RootCID,
	}).Debug("retrieve file.")
	_, err := utils.SSHRun(conf, cmd)
	if err != nil {
		return fmt.Errorf("failed to retrieve data. shell:%s, err:%v", cmd, err)
	}

	return nil
}

func (t *retrieveTask) downloadFile(conf utils.SSHConfig, file *FileRef) error {
	exist, err := utils.Exists(file.LocalPath)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	err = utils.SCPFileFromRemote(conf, file.Path, file.LocalPath)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":        file.ID,
			"path":      file.Path,
			"localPath": file.LocalPath,
			"error":     err,
		}).Error("replay copy file failed.")
		return err
	}
	return t.updateFileStatus(file)
}

func (t *retrieveTask) updateFileStatus(file *FileRef) error {
	index, err := parseFileIndex(file.LocalPath)
	if err != nil {
		return err
	}
	file.Index = int64(index)
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

func parseFileIndex(file string) (int, error) {
	fi, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	content, _, err := br.ReadLine()
	if err != nil {
		return 0, err
	}
	headers := strings.Split(string(content), ",")
	// domains := strings.Split(headers[0], ":")
	// domain = strings.TrimSpace(domains[1])
	indexs := strings.Split(headers[1], ":")
	i, err := strconv.Atoi(indexs[1])
	if err != nil {
		return 0, err
	}
	return i, nil
}

func getFullAPI(ctx context.Context, chain config.Chain) (api.FullNode, jsonrpc.ClientCloser, error) {
	ainfo := api.APIInfo{
		Addr:  chain.RPCHost,
		Token: []byte(chain.RPCToken),
	}
	addr, err := ainfo.DialArgs()
	if err != nil {
		return nil, nil, xerrors.Errorf("could not get DialArgs: %w", err)
	}
	return client.NewFullNodeRPC(ctx, addr, ainfo.AuthHeader())
}
