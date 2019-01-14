package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"./base58"
	"bytes"
)

//1. 创建一个结构WalletKeyPair秘钥对，保存公钥和私钥
//2. 给这个结构提供一个方法GetAddress：私钥->公钥->地址

type WalletKeyPair struct {
	PrivateKey *ecdsa.PrivateKey

	//	type PublicKey struct {
	//	elliptic.Curve
	//	X, Y *big.Int
	//}

	//我们可以将公钥的X，Y进行字节流拼接后传输，这样在对端再进行切割还原，好处是可以方便后面的编码
	PublicKey []byte
}

func NewWalletKeyPair() *WalletKeyPair {

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		log.Panic(err)
	}

	publicKeyRaw := privateKey.PublicKey

	publicKey := append(publicKeyRaw.X.Bytes(), publicKeyRaw.Y.Bytes()...)
	return &WalletKeyPair{PrivateKey: privateKey, PublicKey: publicKey}
}

func (w *WalletKeyPair) GetAddress() string {
	publicHash := HashPubKey(w.PublicKey)

	version := 0x00

	//21字节的数据
	payload := append([]byte{byte(version)}, publicHash...)

	checksum := CheckSum(payload)

	//25字节
	payload = append(payload, checksum...)

	address := base58.Encode(payload)

	return address
}

func IsValidAddress(address string) bool {
	//1. 将输入的地址进行解码得到25字节
	//2. 取出前21个字节， 运行CheckSum函数， 得到checksum1
	//3. 取出后4个字节， 德奥checksum2
	//4. 比较checksum1与checksum2，如果相同则地址有效，反之无效

	decodeInfo := base58.Decode(address)

	if len(decodeInfo) != 25 {
		return false
	}

	payload := decodeInfo[0:len(decodeInfo)-4]
	//自己求出来的校验码
	checksum1 := CheckSum(payload)

	//解出来的校验码
	checksum2 := decodeInfo[len(decodeInfo)-4:]

	return bytes.Equal(checksum1, checksum2)
}

func HashPubKey(pubKey []byte) []byte {
	hash := sha256.Sum256(pubKey)

	//创建一个hash160对象
	//向hash160中write数据
	//做哈希运算

	rip160Haher := ripemd160.New()
	_, err := rip160Haher.Write(hash[:])

	if err != nil {
		log.Panic(err)
	}

	//Sum函数会把我们的结果与Sum参数append到一起，然后返回，我们传入nil,防止数据污染
	publicHash := rip160Haher.Sum(nil)

	return publicHash
}

func CheckSum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])

	//4字节校验码
	checksum := second[0:4]

	return checksum
}
