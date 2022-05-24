package util

import (
	"os"
	"path"
	"strings"
)

func GetFileName(fullFilename string) string {
	filenameWithSuffix := path.Base(fullFilename)
	fileSuffix := path.Ext(filenameWithSuffix)
	filenameOnly := strings.TrimSuffix(filenameWithSuffix, fileSuffix)
	return filenameOnly
}

func CheckPathAndCreate(path string) {
	_, err := os.Stat(path)
	if err == nil {
		return
	} else if os.IsNotExist(err) {
		// 创建文件夹
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			panic("create dir error! err: " + err.Error())
		}
	} else {
		panic("get dir error! err: " + err.Error())
	}
}
