package util

import (
	"bytes"
	"fmt"
	"os/exec"
)

func GetFrame(inPath string, outPath string) {
	//首先生成 cmd 结构体,该结构体包含了很多信息，如执行命令的参数，命令的标准输入输出等
	fmt.Println(inPath, outPath)
	command := exec.Command("ffmpeg", "-y", "-i", inPath, "-vframes", "1", "-f", "image2", outPath)
	//command := exec.Command("bash", "-c", "ls")
	//给标准输入以及标准错误初始化一个 buffer ，每条命令的输出位置可能是不一样的，
	//比如有的命令会将输出放到 stdout ，有的放到 stderr
	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}
	//执行命令，直到命令结束
	err := command.Run()
	if err != nil {
		//打印程序中的错误以及命令行标准错误中的输出
		fmt.Println(err)
		fmt.Println(command.Stderr.(*bytes.Buffer).String())
		return
	}
	//打印命令行的标准输出
	fmt.Println(command.Stdout.(*bytes.Buffer).String())
}
