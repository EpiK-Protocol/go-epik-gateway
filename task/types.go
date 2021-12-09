package task

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"

	"github.com/EpiK-Protocol/go-epik/chain/actors"
	"github.com/EpiK-Protocol/go-epik/chain/actors/builtin/expert"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
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
	ID       string `json:"id,omitempty"`
	Index    int64  `json:"index,omitempty"`
	Count    int64  `json:"count,omitempty"`
	FileSize int64  `json:"file_size,omitempty"` //文件大小
	CheckSum string `json:"check_sum,omitempty"` //文件md5

	Expert string `json:"expert,omitempty"`

	// file saved path on node machine
	Path string `json:"path,omitempty"`

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

type TxMsg struct {
	V int64  `json:"v,omitempty"`
	T string `json:"t,omitempty"`
	S string `json:"s,omitempty"`
	M TxData `json:"m,omitempty"`
}

type TxData struct {
	To     string `json:"to,omitempty"`
	Value  string `json:"value,omitempty"`
	Method int64  `json:"method,omitempty"`
	Params []byte `json:"params,omitempty"`
}

func (f *FileRef) TxMsg() ([]byte, error) {

	expertParams, err := actors.SerializeParams(&expert.BatchImportDataParams{
		Datas: []expert.ImportDataParams{
			{
				RootID:    f.RootCID,
				PieceID:   f.PieceCID,
				PieceSize: abi.PaddedPieceSize(f.PieceSize),
			},
		},
	})
	if err != nil {
		return nil, xerrors.Errorf("serializing params failed: %w", err)
	}

	params := TxData{
		To:     f.Expert,
		Value:  "0",
		Method: 5,
		Params: expertParams,
	}
	msg := &TxMsg{
		V: 1,
		T: "deal",
		S: "bls",
		M: params,
	}

	mbytes, aerr := json.Marshal(msg)
	if aerr != nil {
		return nil, aerr
	}
	return mbytes, nil
}

type Code struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// 获取文件的md5码
func getFileMd5(path string) (string, error) {
	// 文件全路径名
	pFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer pFile.Close()
	md5h := md5.New()
	io.Copy(md5h, pFile)

	return hex.EncodeToString(md5h.Sum(nil)), nil
}
