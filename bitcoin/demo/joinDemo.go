package main

import (
	"strings"
	"fmt"
	"bytes"
)

func main() {
	strsA := []string{"hello", "world", "itcast"}

	strRes := strings.Join(strsA, "=")
	fmt.Printf("strRes : %s\n", strRes)

	//func Join(s [][]byte, sep []byte) []byte {
	joinRes := bytes.Join([][]byte{[]byte("hello"), []byte("world"), []byte("itcast")}, []byte{})
	fmt.Printf("joinRes: %s\n", joinRes)

}
