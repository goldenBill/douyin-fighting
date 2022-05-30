package middleware

import (
	"bytes"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/controller"
	"github.com/goldenBill/douyin-fighting/global"
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

// GetFileType 用于判断文件类型
func GetFileType(fSrc []byte) string {
	var fileType string
	fileCode := bytesToHexString(fSrc)
	global.FILE_TYPE_MAP.Range(func(key, value interface{}) bool {
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

// FileCheck 定义中间件
func FileCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := c.FormFile("data")
		if data.Size >= global.MAX_FILE_SIZE {
			// 检验上传文件的大小
			c.JSON(http.StatusForbidden, controller.Response{
				StatusCode: 1,
				StatusMsg:  "Published video should be smaller than 10 MB",
			})
			c.Abort()
			return
		}
		if err != nil {
			// 状态码不确定
			c.JSON(http.StatusOK, controller.Response{
				StatusCode: 1,
				StatusMsg:  err.Error(),
			})
			c.Abort()
			return
		}
		fileSuffix := path.Ext(data.Filename)
		if _, ok := global.WHITELIST_VIDEO[fileSuffix]; ok == false {
			// 文件后缀名不在白名单内
			c.JSON(http.StatusForbidden, controller.Response{
				StatusCode: 1,
				StatusMsg:  "Unsupported video type",
			})
			c.Abort()
			return
		}
		// 通过文件字节流判断文件真实类型
		f, err := data.Open()
		buffer := make([]byte, 30)
		_, err = f.Read(buffer)
		fileType := GetFileType(buffer)

		if fileType == "" {
			// 文件真实类型不在白名单内
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
