package main

import (
	"flag"
	"fmt"
)

func main() {

	name := flag.String("name", "Jim", "这是测试人的名字的命令")
	age := flag.Int("age", 20, "这是测试人的年龄的命令")

	pages := flag.Int("pages", 500, "这是测试书页码的命令")
	new := flag.Bool("new", true, "这是测试书新旧的命令")



	fmt.Printf("name : %v\n", *name)
	fmt.Printf("age : %v\n", *age)
	fmt.Printf("pages : %v\n", *pages)
	fmt.Printf("new : %v\n", *new)
	fmt.Printf("-----------\n")

	flag.Parse()

	fmt.Printf("name : %v\n", *name)
	fmt.Printf("age : %v\n", *age)
	fmt.Printf("pages : %v\n", *pages)
	fmt.Printf("new : %v\n", *new)
}
