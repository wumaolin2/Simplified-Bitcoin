package main

import (
	"math/big"
	"crypto/sha256"
	"fmt"
	"bytes"
)

//定义一个工作量证明的结构ProofOfWork
//
//a. block
//
//b. 目标值

//2. 提供创建POW的函数
//
//- NewProofOfWork(参数)
//
//3. 提供计算不断计算hash的哈数
//
//- Run()
//
//4. 提供一个校验函数
//
//- IsValid()

type ProofOfWork struct {
	block *Block

	//来存储哈希值，它内置一些方法Cmp:比较方法
	// SetBytes : 把bytes转成big.int类型 []byte("0x00000919011eeb8fbdf0c476d8510b8e1e632eba7b584ac04c11ad20cbbdd394")
	// SetString : 把string转成big.int类型 "0x00000919011eeb8fbdf0c476d8510b8e1e632eba7b584ac04c11ad20cbbdd394"
	target *big.Int //系统提供的，是固定的
}

const Bits = 16

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}

	//写难度值，难度值应该是推导出来的，但是我们为了简化，把难度值先写成固定的，一切完成之后，再去推导
	// 0000100000000000000000000000000000000000000000000000000000000000

	//固定难度值
	//16制格式的字符串
	//targetStr := "0001000000000000000000000000000000000000000000000000000000000000"
	//var bigIntTmp big.Int
	//bigIntTmp.SetString(targetStr, 16)
	//
	//pow.target = &bigIntTmp

	//程序推导难度值, 推导前导为3个难度值
	// 0001000000000000000000000000000000000000000000000000000000000000
	//初始化
	//  0000000000000000000000000000000000000000000000000000000000000001
	//向左移动, 256位
	//1 0000000000000000000000000000000000000000000000000000000000000000
	//向右移动, 四次，一个16进制位代表4个2进制（f:1111）
	//向右移动16位
	//0 0001000000000000000000000000000000000000000000000000000000000000

	bigIntTmp := big.NewInt(1)
	//bigIntTmp.Lsh(bigIntTmp, 256)
	//bigIntTmp.Rsh(bigIntTmp, 16)

	bigIntTmp.Lsh(bigIntTmp, 256-Bits)

	pow.target = bigIntTmp

	return &pow
}

//这是pow的运算函数，为了获取挖矿的随机数，同时返回区块的哈希值
func (pow *ProofOfWork) Run() ([]byte, uint64) {
	//1. 获取block数据
	//2. 拼接nonce
	//3. sha256
	//4. 与难度值比较
	//a. 哈希值大于难度值，nonce++
	//b. 哈希值小于难度值，挖矿成功,退出

	var nonce uint64

	//block := pow.block

	var hash [32]byte

	for ; ; {

		fmt.Printf("%x\r", hash)

		//data := block + nonce
		hash = sha256.Sum256(pow.prepareData(nonce))

		//将hash（数组类型）转成big.int, 然后与pow.target进行比较, 需要引入局部变量
		var bigIntTmp big.Int
		bigIntTmp.SetBytes(hash[:])

		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		//
		//func (x *Int) Cmp(y *Int) (r int) {
		//   x              y
		if bigIntTmp.Cmp(pow.target) == -1 {
			//此时x < y ， 挖矿成功！
			fmt.Printf("挖矿成功！nonce: %d, 哈希值为: %x\n", nonce, hash)
			break
		} else {
			nonce ++
		}
	}

	return hash[:], nonce
}

func (pow *ProofOfWork) prepareData(nonce uint64) []byte {
	block := pow.block

	tmp := [][]byte{
		uintToByte(block.Version),
		block.PrevBlockHash,
		block.MerKleRoot,
		uintToByte(block.TimeStamp),
		uintToByte(block.Difficulity),
		uintToByte(nonce),
	}



	data := bytes.Join(tmp, []byte{})
	return data
}

func (pow *ProofOfWork) IsValid() bool {
	//在校验的时候，block的数据是完整的，我们要做的是校验一下，Hash，block数据，和Nonce是否满足难度值要求

	//获取block数据
	//拼接nonce
	//做sha256
	//比较

	//block := pow.block
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)

	var tmp big.Int
	tmp.SetBytes(hash[:])

	//if tmp.Cmp(pow.target) == -1 {
	//	return true
	//}
	// return false

	return tmp.Cmp(pow.target) == -1
}
