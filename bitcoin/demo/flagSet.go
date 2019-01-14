package main

import (
	"flag"
	"os"
	"fmt"
)

func Usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("./flagTest person --name Lily --age 20\n")
	fmt.Printf("./flagTest book --pages 300 --new true\n")
}

func main() {
	if len(os.Args) < 2 {
		Usage()
		os.Exit(1)
	}

	//这两句相当于：var CommandLine = NewFlagSet(os.Args[0], ExitOnError)
	//分别创建两个命令集合
	personCmd := flag.NewFlagSet("person", flag.ExitOnError)
	bookCmd := flag.NewFlagSet("book", flag.ExitOnError)

	//下面的语句相当于对命令添加参数
	name := personCmd.String("name", "Jim", "这是测试人的名字的命令")
	age := personCmd.Int("age", 20, "这是测试人的年龄的命令")

	pages := bookCmd.Int("pages", 500, "这是测试书页码的命令")
	new := bookCmd.Bool("new", true, "这是测试书新旧的命令")

	//根据第0个参数进行命令判断
	switch os.Args[1] {
	case "person":
		//调用解析函数
		personCmd.Parse(os.Args[2:])
		fmt.Printf("name : %v\n", *name)
		fmt.Printf("age : %v\n", *age)

	case "book":
		bookCmd.Parse(os.Args[2:])
		fmt.Printf("pages: %v\n", *pages)
		fmt.Printf("new: %v\n", *new)
	default:
		Usage()
	}
}
