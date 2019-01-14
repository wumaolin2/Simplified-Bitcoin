package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"fmt"
	"./base58"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"crypto/elliptic"
	"strings"
)

//- 交易输入（TXInput）
//
//指明交易发起人可支付资金的来源，包含：
//
//- 引用utxo所在交易的ID（知道在哪个房间）
//- 所消费utxo在output中的索引（具体位置）
//- 解锁脚本（签名，公钥）
//
//- 交易输出（TXOutput）
//
//包含资金接收方的相关信息,包含：
//
//- 接收金额（数字）
//- 锁定脚本（对方公钥的哈希，这个哈希可以通过地址反推出来，所以转账时知道地址即可！）
//
//-交易ID
//
//一般是交易结构的哈希值（参考block的哈希做法）

//定义交易结构
//定义input
//定义output
//设置交易ID

type TXInput struct {
	TXID []byte //交易id

	Index int64 //output的索引
	//Address string //解锁脚本，先使用地址来模拟

	Signature []byte //交易签名

	PubKey []byte //公钥本身，不是公钥哈希
}

type TXOutput struct {
	Value float64 //转账金额
	//Address string  //锁定脚本

	PubKeyHash []byte //是公钥的哈希，不是公钥本身
}

//给定转账地址，得到这个地址的公钥哈希，完成对output的锁定
func (output *TXOutput) Lock(address string) {

	//address -> public key hash
	//25字节
	decodeInfo := base58.Decode(address)

	pubKeyHash := decodeInfo[1:len(decodeInfo)-4]

	output.PubKeyHash = pubKeyHash
}

func NewTXOutput(value float64, address string) TXOutput {
	output := TXOutput{Value: value}

	output.Lock(address)

	return output
}

type Transaction struct {
	TXid      []byte     //交易id
	TXInputs  []TXInput  //所有的inputs
	TXOutputs []TXOutput //所有的outputs
}

func (tx *Transaction) SetTXID() {

	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(tx)

	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(buffer.Bytes())
	tx.TXid = hash[:]
}

//挖矿奖励
const reward = 12.5

//实现挖矿挖矿交易，
//特点：只有输出，没有有效的输入(不需要引用id，不需要索引，不需要签名)

//把挖矿的人传递进来，因为有奖励
func NewCoinbaseTx(miner string, data string) *Transaction {

	//我们在后面的程序中，需要识别一个交易是否为coinbase，所以我们需要设置一些特殊的值，用于判断
	inputs := []TXInput{TXInput{nil, -1, nil, []byte(data)}}
	//outputs := []TXOutput{TXOutput{12.5, miner}}

	output := NewTXOutput(reward, miner)
	outputs := []TXOutput{output}

	tx := Transaction{nil, inputs, outputs}
	tx.SetTXID()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	//特点：1. 只有一个input 2. 引用的id是nil 3. 引用的索引是-1
	inputs := tx.TXInputs
	if len(inputs) == 1 && inputs[0].TXID == nil && inputs[0].Index == -1 {
		return true
	}

	return false
}

//内部逻辑：
//

func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	//1. 打开钱包
	ws := NewWallets()

	//获取秘钥对
	wallet := ws.WalletsMap[from]

	if wallet == nil {
		fmt.Printf("%s 的私钥不存在，交易创建失败!\n", from)
		return nil
	}

	//2. 获取公钥，私钥
	privateKey := wallet.PrivateKey //目前使用不到，步骤三签名时使用
	publickKey := wallet.PublicKey

	pubKeyHash := HashPubKey(wallet.PublicKey)

	utxos := make(map[string][]int64) //标识能用的utxo
	var resValue float64              //这些utxo存储的金额
	//假如李四转赵六4，返回的信息为:
	//utxos[0x333] = int64{0, 1}
	//resValue : 5

	//1. 遍历账本，找到属于付款人的合适的金额，把这个outputs找到
	utxos, resValue = bc.FindNeedUtxos(pubKeyHash, amount)

	//2. 如果找到钱不足以转账，创建交易失败。
	if resValue < amount {
		fmt.Printf("余额不足，交易失败!\n")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//3. 将outputs转成inputs
	for txid /*0x333*/ , indexes := range utxos {
		for _, i /*0, 1*/ := range indexes {
			input := TXInput{[]byte(txid), i, nil, publickKey}
			inputs = append(inputs, input)
		}
	}

	//4. 创建输出，创建一个属于收款人的output
	//output := TXOutput{amount, to}
	output := NewTXOutput(amount, to)
	outputs = append(outputs, output)

	//5. 如果有找零，创建属于付款人output
	if resValue > amount {
		//output1 := TXOutput{resValue - amount, from}
		output1 := NewTXOutput(resValue-amount, from)
		outputs = append(outputs, output1)
	}

	//创建交易
	tx := Transaction{nil, inputs, outputs}

	//6. 设置交易id
	tx.SetTXID()
	//把查找引用交易的环节放到BlockChain中去，同时在BlockChain进行调用签名

	//我们付款人再创建交易时，已经得到了所有引用的output的详细信息。
	//但是我们不去使用，因为在矿工校验的时候，矿工是没有这部分信息的，矿工需要遍历账本找到所有引用交易
	//我们为了统一操作，所以再次查询一次，进行签名。

	bc.SignTransaction(&tx, privateKey)

	//7. 返回交易结构
	return &tx
}

//第一个参数时私钥，
//第二个参数时这个交易的input所引用的所有的交易
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	fmt.Printf("对交易进行签名...\n")

	//校验的时候，如果是挖矿交易，直接返回true
	if tx.IsCoinbase() {
		return
	}

	//1. 拷贝一份交易txCopy，
	// >做相应裁剪：把每一个input的Sig和pubkey设置为nil
	// > output不做改变
	txCopy := tx.TrimmedCopy()

	//2. 遍历txCopy.inputs，
	// > 把这个input所引用的output的公钥哈希拿过来，赋值给pubkey

	for i, input := range txCopy.TXInputs {
		//找到引用的交易
		preTX := prevTXs[string(input.TXID)]
		output := preTX.TXOutputs[input.Index]

		//for循环迭代出来的数据是一个副本，对这个input进行修改，不会影响到原始数据
		//所以我们这里需要使用下标方式修改

		//input.PubKey = output.PubKeyHash
		txCopy.TXInputs[i].PubKey = output.PubKeyHash

		//签名要对数据的hash进行签名
		//我们的数据都在交易中，我们要求交易的哈希
		//Transaction的SetTXID函数就是对交易的哈希
		//所以我们可以使用交易id作为我们的签名的内容

		//3. 生成要签名的数据（哈希）
		txCopy.SetTXID()
		signData := txCopy.TXid

		//清理,原理同上
		//input.PubKey = nil
		txCopy.TXInputs[i].PubKey = nil

		fmt.Printf("要签名的数据， signData: %x\n", signData)

		//4. 对数据进行签名r, s
		r, s, err := ecdsa.Sign(rand.Reader, privKey, signData)

		if err != nil {
			fmt.Printf("交易签名失败, err : %v\n", err)
			//return false
		}

		//5. 拼接r,s为字节流
		signature := append(r.Bytes(), s.Bytes()...)

		//6. 赋值给原始的交易的Signature字段
		tx.TXInputs[i].Signature = signature
	}
	//return true
}

//trim:裁剪
// >做相应裁剪：把每一个input的Sig和pubkey设置为nil
// > output不做改变
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		input1 := TXInput{input.TXID, input.Index, nil, nil}
		inputs = append(inputs, input1)
	}

	outputs = tx.TXOutputs

	tx1 := Transaction{tx.TXid, inputs, outputs}
	return tx1
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	fmt.Printf("对交易进行校验...\n")

	//1. 拷贝修剪的副本
	txCopy := tx.TrimmedCopy()

	//2. 遍历原始交易（注意，不是txCopy)
	for i, input := range tx.TXInputs {

		//3. 遍历原始交易的input所引用的前交易prevTX
		prevTX := prevTXs[string(input.TXID)]
		output := prevTX.TXOutputs[input.Index]

		//4. 找到output的公钥哈希，赋值给txCopy对应的input
		txCopy.TXInputs[i].PubKey = output.PubKeyHash

		//5. 还原签名的数据
		txCopy.SetTXID()

		//清理动作，重要！！！
		txCopy.TXInputs[i].PubKey = nil

		verifyData := txCopy.TXid
		fmt.Printf("verifyData : %x\n", verifyData)

		//6. 校验
		//还原签名为r,s
		signature := input.Signature

		//公钥字节流
		pubKeyBytes := input.PubKey

		r := big.Int{}
		s := big.Int{}

		rData := signature[: len(signature)/2]
		sData := signature[len(signature)/2:]

		r.SetBytes(rData)
		s.SetBytes(sData)

		//type PublicKey struct {
		//	elliptic.Curve
		//	X, Y *big.Int
		//}

		//还原公钥为curve，X，Y
		x := big.Int{}
		y := big.Int{}

		xData := pubKeyBytes[: len(pubKeyBytes)/2]
		yData := pubKeyBytes[len(pubKeyBytes)/2:]

		x.SetBytes(xData)
		y.SetBytes(yData)

		curve := elliptic.P256()

		publicKey := ecdsa.PublicKey{curve, &x, &y}

		//数据，签名，公钥准备完毕，开始校验
		//func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {
		if !ecdsa.Verify(&publicKey, verifyData, &r, &s) {
			return false
		}
	}

	return true
}

func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXid))

	for i, input := range tx.TXInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	//11111, 2222, 3333, 44444, 5555

	//`11111
	 //2222
	 //3333
	 //44444
	 //5555`

	return strings.Join(lines, "\n")
}





