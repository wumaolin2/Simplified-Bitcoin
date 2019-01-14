package main

import "fmt"

type Test struct {
	str string
}

//给结构添加一个String()

func (test *Test) String() string {
	res := fmt.Sprintf("hello world : %s\n", test.str)
	return res
}

func main() {

	t1 := &Test{"你好"}
	fmt.Printf("%v\n", t1)
}
