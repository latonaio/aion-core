package common

import (
	"fmt"
	"path"
	"strconv"
	"time"
)

var (
	location = time.FixedZone("Asia/Tokyo", 9*60*60)
)

func RemoveExt(filePath string) string {
	fileName := path.Base(filePath)
	return fileName[0 : len(fileName)-len(path.Ext(fileName))]
}

func GetFileNameDatetime() string {
	now := time.Now().In(location)
	nano := int(now.UnixNano() / int64(time.Millisecond) % 1000)
	return now.Format("20060102150405") + fmt.Sprintf("%03d", nano)
}

func GetIsoDatetime() string {
	return time.Now().In(location).Format("2006-01-02T15:04:05.000+09:00")
}

func GetMsDataPath(dataPath string, serviceName string, processNumber int) string {
	return path.Join(dataPath, serviceName+"_"+strconv.Itoa(processNumber))
}
