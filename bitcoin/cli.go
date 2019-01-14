package main

import (
	"os"
	"fmt"
	"strconv"
)

//使用命令行分析
//
//1. 所有的支配动作交给命令行来做
//2. 主函数只需要调用命令行结构即可
//3. 根据输入的不同命令，命令行做相应动作
//1. addBlock
//2. printChain
//
//
//
//CLI : command line的缩写
//
//type CLI struct {
//	 bc *BlockChain
//}
//
//
//

const Usage = `
	./blockchain createBlockChain 地址 "创建区块链"
	./blockchain printChain          打印区块链
	./blockchain getBalance 地址      获取地址的余额
	./blockchain send FROM TO AMOUNT MINER DATA "转账命令"
	./blockchain createWallet "创建钱包"
	./blockchain listAddresses "打印所有的钱包地址"
	./blockchain printTx "打印所有交易"
`

type CLI struct {
	//bc *BlockChain //CLI中不需要保存区块链实例了，所有名字在自己调用之前，自己获取区块链实例
}

//给CLI提供一个方法，进行命令解析，从而执行调度
func (cli *CLI) Run() {

	cmds := os.Args

	if len(cmds) < 2 {
		fmt.Printf(Usage)
		os.Exit(1)
	}

	switch cmds[1] {
	case "createBlockChain":
		if len(cmds) != 3 {
			fmt.Printf(Usage)
			os.Exit(1)
		}

		fmt.Printf("创建区块链命令被调用!\n")

		addr := cmds[2]
		cli.CreateBlockChain(addr)

	case "printChain":
		fmt.Printf("打印区块链命令被调用\n")
		cli.PrintChain()

	case "getBalance":
		fmt.Printf("获取余额命令被调用\n")
		cli.GetBalance(cmds[2])

	case "send":
		fmt.Printf("转账命令被调用\n")
		//./blockchain send FROM TO AMOUNT MINER DATA "转账命令"
		if len(cmds) != 7 {
			fmt.Printf("send命令发现无效参数，请检查!\n")
			fmt.Printf(Usage)
			os.Exit(1)
		}

		from := cmds[2]
		to := cmds[3]
		amount, _ := strconv.ParseFloat(cmds[4], 64)
		miner := cmds[5]
		data := cmds[6]
		cli.Send(from, to, amount, miner, data)
	case "createWallet":
		fmt.Printf("创建钱包命令被调用\n")
		cli.CreateWallet()

	case "listAddresses":
		fmt.Printf("打印钱包地址命令被调用\n")
		cli.ListAddresses()

	case "printTx":
		fmt.Printf("打印交易命令被调用\n")
		cli.PrintTx()

	default:
		fmt.Printf("无效的命令，请检查\n")
		fmt.Printf(Usage)
	}
	//添加区块的时候： bc.addBlock(data), data 通过os.Args拿回来
	//打印区块链时候：遍历区块链，不需要外部输入数据
}
