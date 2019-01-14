package main

import (
	"./bolt"
	"log"
	"fmt"
	"os"
	"bytes"
	"./base58"
	"crypto/ecdsa"
)

//使用bolt进行改写，需要两个字段：
//1. bolt数据库的句柄
//2. 最后一个区块的哈希值
type BlockChain struct {
	db   *bolt.DB //句柄
	tail []byte   //最后一个区块的哈希值
}

const blockChainName = "blockChain.db"
const blockBucketName = "blockBucket"
const lastHashKey = "lastHashKey"

func CreateBlockChain(miner string) *BlockChain {

	if IsFileExist(blockChainName) {
		fmt.Printf("区块链已经存在，不需要重复创建!\n")
		return nil
	}

	//功能分析：
	//1. 获得数据库的句柄，打开数据库，读写数据

	db, err := bolt.Open(blockChainName, 0600, nil)
	//向数据库中写入数据
	//从数据库中读取数据

	if err != nil {
		log.Panic(err)
	}

	//defer db.Close()

	var tail []byte

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blockBucketName))

		if err != nil {
			log.Panic(err)
		}

		//抽屉准备完毕，开始添加创世块
		//创世块中只有一个挖矿交易，只有Coinbase
		coinbase := NewCoinbaseTx(miner, genesisInfo)
		genesisBlock := NewBlock([]*Transaction{coinbase}, []byte{})

		b.Put(genesisBlock.Hash, genesisBlock.Serialize() /*将区块序列化，转成字节流*/)
		b.Put([]byte(lastHashKey), genesisBlock.Hash)

		//为了测试，我们把写入的数据读取出来，如果没问题，注释掉这段代码
		//blockInfo := b.Get(genesisBlock.Hash)
		//block := Deserialize(blockInfo)
		//fmt.Printf("解码后的block数据:%s\n", block)

		tail = genesisBlock.Hash

		return nil
	})

	return &BlockChain{db, tail}
}

//返回区块链实例
func NewBlockChain() *BlockChain {

	if !IsFileExist(blockChainName) {
		fmt.Printf("区块链不存在，请先创建!\n")
		return nil
	}

	//功能分析：
	//1. 获得数据库的句柄，打开数据库，读写数据

	db, err := bolt.Open(blockChainName, 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	//defer db.Close()

	var tail []byte

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))

		if b == nil {
			fmt.Printf("区块链bucket为空，请检查!\n")
			os.Exit(1)
		}

		tail = b.Get([]byte(lastHashKey))

		return nil
	})

	return &BlockChain{db, tail}
}

//添加区块
func (bc *BlockChain) AddBlock(txs []*Transaction) {
	//矿工得到交易时，第一时间对交易进行验证
	//矿工如果不验证，即使挖矿成功，广播区块后，其他的验证矿工，仍然会校验每一笔交易

	validTXs := []*Transaction{}

	for _, tx := range txs {
		if bc.VerifyTransaction(tx) {
			fmt.Printf("--- 该交易有效: %x\n", tx.TXid)
			validTXs = append(validTXs, tx)
		} else {
			fmt.Printf("发现无效的交易: %x\n", tx.TXid)
		}
	}

	//1. 创建一个区块
	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))

		if b == nil {
			fmt.Printf("bucket不存在，请检查!\n")
			os.Exit(1)
		}

		block := NewBlock(validTXs, bc.tail)
		b.Put(block.Hash, block.Serialize() /*将区块序列化，转成字节流*/)
		b.Put([]byte(lastHashKey), block.Hash)

		bc.tail = block.Hash

		return nil
	})
}

//定义一个区块链的迭代器，包含db，current
type BlockChainIterator struct {
	db      *bolt.DB
	current []byte //当前所指向区块的哈希值
}

//创建迭代器，使用bc进行初始化

func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{bc.db, bc.tail}
}

func (it *BlockChainIterator) Next() *Block {

	var block Block

	it.db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockBucketName))
		if b == nil {
			fmt.Printf("bucket不存在，请检查!\n")
			os.Exit(1)
		}

		//真正的读取数据
		blockInfo /*block的字节流*/ := b.Get(it.current)
		block = *Deserialize(blockInfo)

		it.current = block.PrevBlockHash

		return nil
	})

	return &block
}

//我们想把FindMyUtoxs和FindNeedUTXO进行整合
//
//1. FindMyUtoxs： 找到所有utxo（只要output就可以了）
//2. FindNeedUTXO：找到需要的utxo（要output的定位）

//我们可以定义一个结构，同时包含output已经定位信息
type UTXOInfo struct {
	TXID   []byte   //交易id
	Index  int64    //output的索引值
	Output TXOutput //output本身
}

//实现思路：
func (bc *BlockChain) FindMyUtoxs(pubKeyHash []byte) []UTXOInfo {
	fmt.Printf("FindMyUtoxs\n")
	//var UTXOs []TXOutput //返回的结构
	var UTXOInfos []UTXOInfo //新的返回结构

	it := bc.NewIterator()

	//这是标识已经消耗过的utxo的结构，key是交易id，value是这个id里面的output索引的数组
	spentUTXOs := make(map[string][]int64)

	//1. 遍历账本
	for {

		block := it.Next()

		//2. 遍历交易
		for _, tx := range block.Transactions {
			//遍历交易输入:inputs

			if tx.IsCoinbase() == false {
				//如果不是coinbase，说明是普通交易，才有必要进行遍历
				for _, input := range tx.TXInputs {

					//判断当前被使用input是否为目标地址所有
					if bytes.Equal(HashPubKey(input.PubKey), pubKeyHash) {

						fmt.Printf("找到了消耗过的output! index : %d\n", input.Index)
						key := string(input.TXID)
						spentUTXOs[key] = append(spentUTXOs[key], input.Index)
						//spentUTXOs[0x222] = []int64{0}
						//spentUTXOs[0x333] = []int64{0}  //中间状态
						//spentUTXOs[0x333] = []int64{0, 1}
					}
				}
			}

			key := string(tx.TXid)
			indexes /*[]int64{0,1}*/ := spentUTXOs[key]

		OUTPUT:
		//3. 遍历output
			for i, output := range tx.TXOutputs {

				if len(indexes) != 0 {
					fmt.Printf("当前这笔交易中有被消耗过的output!\n")
					for _, j /*0, 1*/ := range indexes {
						if int64(i) == j {
							fmt.Printf("i == j, 当前的output已经被消耗过了，跳过不统计!\n")
							continue OUTPUT
						}
					}
				}

				//4. 找到属于我的所有output
				if bytes.Equal(pubKeyHash, output.PubKeyHash) {
					//fmt.Printf("找到了属于 %s 的output, i : %d\n", address, i)
					//UTXOs = append(UTXOs, output)
					utxoinfo := UTXOInfo{tx.TXid, int64(i), output}
					UTXOInfos = append(UTXOInfos, utxoinfo)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			fmt.Printf("遍历区块链结束!\n")
			break
		}
	}

	return UTXOInfos
}

func (bc *BlockChain) GetBalance(address string) {

	//这个过程，不要打开钱包，因为有可能查看余额的人不是地址本人
	decodeInfo := base58.Decode(address)

	pubKeyHash := decodeInfo[1:len(decodeInfo)-4]

	utxoinfos := bc.FindMyUtoxs(pubKeyHash)

	var total = 0.0
	//所有的output都在utxoinfos内部
	//获取余额时，遍历utxoinfos获取output即可
	for _, utxoinfo := range utxoinfos {
		total += utxoinfo.Output.Value //10, 3, 1
	}

	fmt.Printf("%s 的余额为: %f\n", address, total)
}

//1. 遍历账本，找到属于付款人的合适的金额，把这个outputs找到
//utxos, resValue = bc.FindNeedUtxos(from, amount)
func (bc *BlockChain) FindNeedUtxos(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {

	needUtxos := make(map[string][]int64) //标识能用的utxo, //返回的结构
	var resValue float64                  //统计的金额

	//复用FindMyUtxo函数，这个函数已经包含了所有信息
	utxoinfos := bc.FindMyUtoxs(pubKeyHash)

	for _, utxoinfo := range utxoinfos {
		key := string(utxoinfo.TXID)

		needUtxos[key] = append(needUtxos[key], int64(utxoinfo.Index))
		resValue += utxoinfo.Output.Value

		//2. 判断一下金额是否足够
		if resValue >= amount {
			//a. 足够， 直接返回
			break
		}
	}
	return needUtxos, resValue
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	//1. 遍历账本找到所有应用交易

	prevTXs := make(map[string]Transaction)

	//遍历tx的inputs，通过id去查找所引用的交易
	for _, input := range tx.TXInputs {
		prevTx := bc.FindTransaction(input.TXID)

		if prevTx == nil {
			fmt.Printf("没有找到交易: %x\n", input.TXID)
		} else {
			//把找到的引用交易保存起来
			//0x222
			//0x333
			prevTXs[string(input.TXID)] = *prevTx
		}
	}

	tx.Sign(privateKey, prevTXs)
}

//矿工校验流程
//1. 找到交易input所引用的所有的交易prevTXs
//2. 对交易进行校验
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {

	//校验的时候，如果是挖矿交易，直接返回true
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	//遍历tx的inputs，通过id去查找所引用的交易
	for _, input := range tx.TXInputs {
		prevTx := bc.FindTransaction(input.TXID)

		if prevTx == nil {
			fmt.Printf("没有找到交易: %x\n", input.TXID)
		} else {
			//把找到的引用交易保存起来
			//0x222
			//0x333
			prevTXs[string(input.TXID)] = *prevTx
		}
	}

	return tx.Verify(prevTXs)
}

func (bc *BlockChain) FindTransaction(txid []byte) *Transaction {

	//遍历区块链的交易
	//通过对比id来识别

	it := bc.NewIterator()

	for {
		block := it.Next()

		for _, tx := range block.Transactions {

			//如果找到相同id交易，直接返回交易即可
			if bytes.Equal(tx.TXid, txid) {
				fmt.Printf("找到了所引用交易: %x\n", tx.TXid)
				return tx
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil
}
