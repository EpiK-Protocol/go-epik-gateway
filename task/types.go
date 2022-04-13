package task

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"

	"github.com/ipfs/go-cid"
)

type Status uint64

const (
	FileStatusNew Status = iota

	FileStatusDownloading

	FileStatusDownloaded

	FileStatusImporting

	FileStatusImported

	FileStatusRegistering

	FileStatusRegistered

	FileStatusNeedStorage

	FileStatusStoraging

	FileStatusStoraged

	FileStatusReplaied
)

const (
	FileEventNeedDownload = "file:download"
	FileEventDownloaded   = "file:downloaded"
)

// Taskinterface
type Task interface {
	// Start tasks
	Start(context.Context) error

	// Stop tasks
	Stop(context.Context) error
}

type FileRef struct {
	ID    string `json:"id,omitempty"`
	Index int64  `json:"index,omitempty"`

	Count    int64  `json:"count,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	CheckSum string `json:"check_sum,omitempty"`

	Expert string `json:"expert,omitempty"`

	// file saved path on node machine
	Path      string `json:"path,omitempty"`
	LocalPath string `json:"local_path,omitempty"`

	// file info
	RootCID   cid.Cid `json:"root_cid,omitempty"`
	PieceCID  cid.Cid `json:"piece_cid,omitempty"`
	PieceSize uint64  `json:"piece_size,omitempty"`

	// file download url
	Url string `json:"url,omitempty"`

	Status Status `json:"status,omitempty"`
}

func (f *FileRef) Unmarshal(bytes []byte) error {
	if err := json.Unmarshal(bytes, f); err != nil {
		return err
	}
	return nil
}

func (f *FileRef) Marshal() ([]byte, error) {
	return json.Marshal(f)
}

type Code struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// returns file md5 checksum
func getFileMd5(path string) (string, error) {
	pFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer pFile.Close()
	md5h := md5.New()
	io.Copy(md5h, pFile)

	return hex.EncodeToString(md5h.Sum(nil)), nil
}

type ResponseCode struct {
	Code    int64
	Message string
}

type ListResponse struct {
	Code            ResponseCode
	List            []ListData
	Callback        string `json:"callback"`
	OnchainCallback string `json:"onchain_callback"`
}

type ListData struct {
	Id       string `json:"id"`
	Expert   string `json:"expert"`
	Index    int64  `json:"index"`
	FileName string `json:"file_name"`
	FileUrl  string `json:"file_url"`
	Status   string `json:"status"`
	Count    int64  `json:"count"`
	FileSize int64  `json:"file_size"` //文件大小
	CheckSum string `json:"check_sum"` //文件md5
}
