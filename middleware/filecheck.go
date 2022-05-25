package middleware

import (
	"bytes"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/controller"
	"github.com/goldenBill/douyin-fighting/global"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
)

// 获取前面结果字节的二进制
func bytesToHexString(src []byte) string {
	res := bytes.Buffer{}
	if src == nil || len(src) <= 0 {
		return ""
	}
	temp := make([]byte, 0)
	for _, v := range src {
		sub := v & 0xFF
		hv := hex.EncodeToString(append(temp, sub))
		if len(hv) < 2 {
			res.WriteString(strconv.FormatInt(int64(0), 10))
		}
		res.WriteString(hv)
	}
	return res.String()
}

// 用文件前面几个字节来判断
// fSrc: 文件字节流（就用前面几个字节）
func GetFileType(fSrc []byte) string {
	var fileType string
	fileCode := bytesToHexString(fSrc)
	//println(fileCode[:40])
	global.GVAR_FILE_TYPE_MAP.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(string)
		if strings.HasPrefix(fileCode, strings.ToLower(k)) ||
			strings.HasPrefix(k, strings.ToLower(fileCode)) {
			fileType = v
			return false
		}
		return true
	})
	return fileType
}

// 定义中间
func FileCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := c.FormFile("data")
		if err != nil {
			// 状态码不确定
			c.JSON(http.StatusOK, controller.Response{
				StatusCode: 1,
				StatusMsg:  err.Error(),
			})
			c.Abort()
			return
		}
		fileSuffix := path.Ext(data.Filename)[1:]
		if _, ok := global.GVAR_WHITELIST_VIDEO[fileSuffix]; ok == false {
			//println(fileSuffix)
			c.JSON(http.StatusForbidden, controller.Response{
				StatusCode: 1,
				StatusMsg:  "Unsupported video type",
			})
			c.Abort()
			return
		}
		f, err := data.Open()
		fSrc, err := ioutil.ReadAll(f)
		fileType := GetFileType(fSrc)
		if fileType == "" {
			println("here\n\n\n")
			c.JSON(http.StatusForbidden, controller.Response{
				StatusCode: 1,
				StatusMsg:  "Unsupported video type",
			})
			c.Abort()
			return
		}
		// 保存文件类型
		c.Set("FileType", fileType)
		// 执行函数
		c.Next()
	}
}
