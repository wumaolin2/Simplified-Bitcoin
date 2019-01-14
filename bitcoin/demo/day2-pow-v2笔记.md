# 一、v2版本思路

## 1. POW介绍

- 定义一个工作量证明的结构ProofOfWork

  a. block

  b. 目标值

## 2. 提供创建POW的函数

- NewProofOfWork(参数)

## 3. 提供计算不断计算hash的哈数

- Run()

## 4. 提供一个校验函数

- IsValid()





# POW定义



```js
type ProofOfWork struct {
	block *Block

	//来存储哈希值，它内置一些方法Cmp:比较方法
	// SetBytes : 把bytes转成big.int类型 []byte("0x00000919011eeb8fbdf0c476d8510b8e1e632eba7b584ac04c11ad20cbbdd394")
	// SetString : 把string转成big.int类型 "0x00000919011eeb8fbdf0c476d8510b8e1e632eba7b584ac04c11ad20cbbdd394"
	target *big.Int //系统提供的，是固定的
}

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}

	//写难度值，难度值应该是推导出来的，但是我们为了简化，把难度值先写成固定的，一切完成之后，再去推导
	// 0000100000000000000000000000000000000000000000000000000000000000

	//16制格式的字符串
	targetStr := "0000100000000000000000000000000000000000000000000000000000000000"
	var bigIntTmp big.Int
	bigIntTmp.SetString(targetStr, 16)

	pow.target = &bigIntTmp

	return &pow
}
```





# Run函数实现

```go

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
		block.Data,
		uintToByte(nonce),
	}

	data := bytes.Join(tmp, []byte{})
	return data
}

```



# 使用pow更新NewBlock

```go
//创建区块，对Block的每一个字段填充数据即可
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := Block{
		Version: 00,

		PrevBlockHash: prevBlockHash,

		MerKleRoot: []byte{},

		TimeStamp: uint64(time.Now().Unix()),

		Difficulity: 10, //随便写的，v2再调整

		Nonce: 10, //同Difficulty

		Data: []byte(data),

		Hash: []byte{}, //先填充为空，后续会填充数据
	}

	//block.SetHash()
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return &block
}
```



# 校验挖矿是否有效

```go
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
```



# 打印block字段

```go
	for i, block := range bc.Blocks {
		fmt.Printf("+++++++++++++++ %d ++++++++++++++\n", i)
		fmt.Printf("Version : %d\n", block.Version)
		fmt.Printf("PrevBlockHash : %x\n", block.PrevBlockHash)
		fmt.Printf("MerKleRoot : %x\n", block.MerKleRoot)

		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("TimeStamp : %s\n", timeFormat)

		fmt.Printf("Difficulity : %d\n", block.Difficulity)
		fmt.Printf("Nonce : %d\n", block.Nonce)
		fmt.Printf("Hash : %x\n", block.Hash)
		fmt.Printf("Data : %s\n", block.Data)

		pow := NewProofOfWork(block)
		fmt.Printf("IsValid: %v\n", pow.IsValid())
	}
```



# 使用Bits推导难度值

```go
const Bits = 20

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

	bigIntTmp.Lsh(bigIntTmp, 256 - Bits)

	pow.target = bigIntTmp

	return &pow
}
```





# bolt数据库图示

![image-20181206113802425](https://ws2.sinaimg.cn/large/006tNbRwly1fxwvgke5m5j31ik0pa0xh.jpg)



boltdemo



```go
package main

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
)

func main() {
	db, err := bolt.Open("test.db", 0600, nil)
	//向数据库中写入数据
	//从数据库中读取数据

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		//所有的操作都在这里

		b1 := tx.Bucket([]byte("bucketName1"))

		if b1 == nil {
			//如果b1为空，说明名字为"buckeName1"这个桶不存在，我们需要创建之
			b1, err = tx.CreateBucket([]byte("bucketName1"))

			if err != nil {
				log.Panic(err)
			}
		}

		//bucket已经创建完成，准备写入数据
		//写数据使用Put，读数据使用Get
		err = b1.Put([]byte("name1"), []byte("Lily"))
		if err != nil {
			fmt.Printf("写入数据失败name1 : Lily!\n")
		}

		err = b1.Put([]byte("name2"), []byte("Jim"))
		if err != nil {
			fmt.Printf("写入数据失败name2 : Jim!\n")
		}

		//读取数据

		name1 := b1.Get([]byte("name1"))
		name2 := b1.Get([]byte("name2"))
		name3 := b1.Get([]byte("name3"))

		fmt.Printf("name1: %s\n", name1)
		fmt.Printf("name2: %s\n", name2)
		fmt.Printf("name3: %s\n", name3)


		return nil
	})

}

```



# 分析bolt存储区块的格式

key一定唯一：

把所有的区块都写到一个bucket中：key-> value :        []byte->[]byte   :  

在bucket存储两种数据：

1. 区块，区块的哈希值作为key，区块的字节流作为value

  block.Hash -> block.toBytes()

1. 最后一个区块的哈希值

 key使用固定的字符串：[]byte("lastHashKey"), value 就是最后一个区块的哈希



==结论==：

添加一个新区块要做两件事情：

- ==添加区块==
- ==更新”lastHashKey“ 这个key对应的值，这个值就是最后一个区块的哈希值，用于新区块的创建添加。==



![](https://ws3.sinaimg.cn/large/006tNbRwly1fv0t7ab5mdj31cl0pt4qp.jpg)







