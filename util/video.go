package util

import (
	"bytes"
	"os/exec"
)

func GetFrame(inPath string, outPath string) error {
	//首先生成 cmd 结构体,该结构体包含了很多信息，如执行命令的参数，命令的标准输入输出等
	command := exec.Command("ffmpeg", "-y", "-i", inPath, "-vframes", "1", "-f", "image2", outPath)
	//给标准输入以及标准错误初始化一个 buffer ，每条命令的输出位置可能是不一样的，
	//比如有的命令会将输出放到 stdout ，有的放到 stderr
	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}
	//执行命令，直到命令结束
	return command.Run()
}
