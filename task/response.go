package task

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/spf13/cast"
)

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
	FileSize int64  `json:"file_size"` //file size
	CheckSum string `json:"check_sum"` //file md5
}

func SignPostParam(secret string, param map[string]interface{}) string {
	if len(param) == 0 {
		return ""
	}
	var keys []string
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	waitSignString := ""
	for _, k := range keys {
		waitSignString += k + "=" + cast.ToString(param[k]) + "&"
	}
	waitSignString += "secret=" + secret
	h := hmac.New(sha256.New, []byte(waitSignString))
	sign := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return sign
}

func CheckSignPostparam(secret string, signedParam map[string]interface{}) bool {
	if len(signedParam) == 0 {
		return false
	}
	if sign, ok := signedParam["sign"]; !ok || sign == nil {
		return false
	}
	sign := cast.ToString(signedParam["sign"])
	var keys []string
	for k := range signedParam {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	waitSignString := ""
	for _, k := range keys {
		waitSignString += k + "=" + cast.ToString(signedParam[k]) + "&"
	}
	waitSignString += "secret=" + secret
	h := hmac.New(sha256.New, []byte(waitSignString))
	chkSign := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return chkSign == sign
}
