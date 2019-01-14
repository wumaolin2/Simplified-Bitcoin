package main

import (
	"os"
	"fmt"
)

func main() {
	cmds := os.Args

	for i, cmd := range cmds {
		fmt.Printf("cmd[%d] : %s\n", i, cmd)
	}
}
