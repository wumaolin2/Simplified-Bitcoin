package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

//这是一个工具函数文件

func uintToByte(num uint64) []byte {
	//使用binary.Write来进行编码
	var buffer bytes.Buffer

	//编码要进行错误检查，一定要做
	err := binary.Write(&buffer, binary.BigEndian, num)

	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

//判断文件是否存在
func IsFileExist(fileName string) bool {
	//使用os.Stat来判断
	//func Stat(name string) (FileInfo, error) {
	_, err := os.Stat(fileName)

	if os.IsNotExist(err) {
		return false
	}

	return true
}
