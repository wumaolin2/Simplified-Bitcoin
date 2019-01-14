package main

import (
	//"crypto/sha256"
	"time"
	//"bytes"
	"bytes"
	"encoding/gob"
	"log"
	//"fmt"
	"crypto/sha256"
)

//1. 定义结构（区块头的字段比正常的少）
//>1. 前区块哈希
//>2. 当前区块哈希
//>3. 数据

//2. 创建区块
//3. 生成哈希
//4. 引入区块链
//5. 添加区块
//6. 重构代码

const genesisInfo = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Block struct {
	Version uint64 //区块版本号

	PrevBlockHash []byte //前区块哈希

	MerKleRoot []byte //先填写为空，后续v4的时候使用

	TimeStamp uint64 //从1970.1.1至今的秒数

	Difficulity uint64 //挖矿的难度值, v2时使用

	Nonce uint64 //随机数，挖矿找的就是它!

	//Data []byte //数据，目前使用字节流，v4开始使用交易代替
	Transactions []*Transaction

	Hash []byte //当前区块哈希, 区块中本不存在的字段，为了方便我们添加进来
}

//模拟梅克尔根，做一个简单的处理
func (block *Block) HashTransactions() {
	//我们的交易的id就是交易的哈希值，所以我们可以将交易id拼接起来，整体做一次个哈希运算，作为MerKleRoot

	var hashes []byte

	for _, tx := range block.Transactions {
		txid /*[]byte*/ := tx.TXid
		hashes = append(hashes, txid...)
	}

	hash := sha256.Sum256(hashes)
	block.MerKleRoot = hash[:]
}

//创建区块，对Block的每一个字段填充数据即可
func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := Block{
		Version: 00,

		PrevBlockHash: prevBlockHash,

		MerKleRoot: []byte{},

		TimeStamp: uint64(time.Now().Unix()),

		Difficulity: Bits, //随便写的，v2再调整

		//Nonce: 10, //同Difficulty

		//Data: []byte(data),
		Transactions: txs,

		Hash: []byte{}, //先填充为空，后续会填充数据
	}

	block.HashTransactions()

	//block.SetHash()
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return &block
}

//序列化, 将区块转换成字节流
func (block *Block) Serialize() []byte {

	var buffer bytes.Buffer

	//定义编码器
	encoder := gob.NewEncoder(&buffer)

	//编码器对结构进行编码，一定要进行校验
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

func Deserialize(data []byte) *Block {

	//fmt.Printf("解码传入的数据: %x\n", data)

	var block Block

	//创建解码器
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)

	if err != nil {
		log.Panic(err)
	}

	return &block
}
