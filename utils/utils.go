package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

//ParseInt 转int
func ParseInt(i interface{}) (val int) {
	if value, ok := i.(float64); ok {
		return int(value)
	}
	val64, _ := strconv.ParseInt(fmt.Sprintf("%v", i), 10, 64)
	val = int(val64)
	return val
}

//ParseInt64 转int64
func ParseInt64(i interface{}) (val int64) {
	if value, ok := i.(float64); ok {
		return int64(value)
	}
	val, _ = strconv.ParseInt(fmt.Sprintf("%v", i), 10, 64)
	return val
}

//ParseFloat64 转float64
func ParseFloat64(i interface{}) (val float64) {
	val64, _ := strconv.ParseFloat(fmt.Sprintf("%v", i), 64)
	return val64
}

//ParseString 转string
func ParseString(i interface{}) string {
	if i == nil {
		return ""
	}
	val := fmt.Sprintf("%v", i)
	return val
}

//Round float精度
func Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

//StructToMap Struct转成Map
func StructToMap(obj interface{}) map[string]interface{} {
	if obj == nil {
		return nil
	}

	objMap := make(map[string]interface{})
	var v reflect.Value
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		if reflect.ValueOf(obj).IsNil() {
			return nil
		}
		v = reflect.ValueOf(obj).Elem()
	} else {
		v = reflect.ValueOf(obj)
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		jsonname := t.Field(i).Tag.Get("json")
		if jsonname == "" || jsonname == "-" || jsonname == "_" {
			continue
		}
		objMap[jsonname] = v.Field(i).Interface()
	}
	return objMap
}

//MapToStruct 转换
func MapToStruct(_map map[string]interface{}, obj interface{}) {

	var value reflect.Value
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		value = reflect.ValueOf(obj).Elem()
	} else {
		panic("obj need pointer")
	}
	t := value.Type()
	for k, v := range _map {
		for i := 0; i < t.NumField(); i++ {
			jsonname := t.Field(i).Tag.Get("json")
			if jsonname == "" || jsonname == "-" || jsonname == "_" {
				continue
			}
			if jsonname == k {
				value.Set(reflect.ValueOf(v))
			}
		}
	}
}

//SHA1 SHA1
func SHA1(str string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(str)))
}

//SHA256 SHA256
func SHA256(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(str)))
}

//HMACSHA256 Hmac
func HMACSHA256(plain string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(plain))
	return hex.EncodeToString(h.Sum(nil))
	// return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

//MD5 MD5
func MD5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

//BASE64EncodeString base64 encode string
func BASE64EncodeString(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

//BASE64DecodeString base64 decode string
func BASE64DecodeString(str string) string {
	result, _ := base64.StdEncoding.DecodeString(str)
	return string(result)
}

//UUIDNewV4 生成UUID
func UUIDNewV4() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		log.Panicln(err)
	}
	return id
}

//GenOrderNO 根据用户生成订单号
func GenOrderNO(uid int64, projectCode string) (no string) {
	rand := RandomString(CHARCHARACTER, 6)
	no = fmt.Sprintf("%s%s%06d%s", projectCode, time.Now().Format("20060102150405"), uid, rand)
	return no
}

const (
	CHARNUMBER    = "0123456789"
	CHARCHARACTER = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

//RandomString 随机字符
func RandomString(chars string, length int) string {
	rand.Seed(time.Now().Unix())
	str := []byte("")
	for i := 0; i < length; i++ {
		str = append(str, chars[rand.Intn(len(chars))])
	}
	return string(str)
}

//TimeHexNano ...
func TimeHexNano(time int64) (hex string) {
	return fmt.Sprintf("%x", math.MaxInt64-time)
}

//TimeHex ...
func TimeHex(time time.Time) (hex string) {
	return fmt.Sprintf("%x", math.MaxInt64-time.UnixNano())
}

//UnGzip ...
func UnGzip(data []byte) (result []byte, err error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return
	}
	return ioutil.ReadAll(reader)
}

//FormatPhoneNum 格式化手机号码
func FormatPhoneNum(src string, zone string) (phone string, err error) {
	switch zone {
	case "cn-zh":
		switch {

		case len(phone) == 11:
			_, err := strconv.ParseInt(src, 10, 64)
			if err != nil {
				return src, fmt.Errorf("Phone Number Error")
			}
			return fmt.Sprintf("+86%s", src), nil
		case len(phone) == 13:
			if strings.Index(src, "86") != 0 {
				return src, fmt.Errorf("Phone Number Error")
			}
			_, err := strconv.ParseInt(src, 10, 64)
			if err != nil {
				return src, fmt.Errorf("Phone Number Error")
			}
			return fmt.Sprintf("+%s", src), nil
		case len(phone) == 14:
			if strings.Index(src, "+86") != 0 {
				return src, fmt.Errorf("Phone Number Error")
			}
			_, err := strconv.ParseInt(strings.TrimLeft(src, "+"), 10, 64)
			if err != nil {
				return src, fmt.Errorf("Phone Number Error")
			}
			return src, nil
		default:
			return src, fmt.Errorf("Phone Number Error")
		}
	case "cn-hk":
	case "cn-tw":
	default:
		return src, fmt.Errorf("Zone Not Support")
	}
	return
}

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
