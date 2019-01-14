package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"crypto/sha256"
	"fmt"
	"math/big"
)

func main() {
	//1. 创建私钥
	//2. 创建公钥
	//3. 私钥对数据进行签名(对数据的哈希值进行签名)
	//4. 使用数据，签名，公钥进行校验

	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)

	if err != nil {
		os.Exit(1)
	}

	pubKey := privateKey.PublicKey

	data := "hello"

	dataHash := sha256.Sum256([]byte(data))

	//func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, dataHash[:])
	//一般传输过程中，会把r，s拼成字节流再传输

	fmt.Printf("r : %x, len(r): %d\n", r.Bytes(), len(r.Bytes()))
	fmt.Printf("s : %x, len(s): %d\n", s.Bytes(), len(s.Bytes()))
	signature := append(r.Bytes(), s.Bytes()...)

	if err != nil {
		os.Exit(1)
	}

	//传输中....: 数据， 签名signature，公钥

	//在接收端从中把r和s切出来
	var r1 big.Int
	var s1 big.Int
	r1Data := signature[: len(signature)/2]
	s1Data := signature[len(signature)/2:]

	r1.SetBytes(r1Data)
	s1.SetBytes(s1Data)

	fmt.Printf("r1 : %x, len(r1): %d\n", r1.Bytes(), len(r1.Bytes()))
	fmt.Printf("s1 : %x, len(s1): %d\n", s1.Bytes(), len(s1.Bytes()))

	//func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {
	res := ecdsa.Verify(&pubKey, dataHash[:], &r1, &s1)
	fmt.Printf("res : %v\n", res)
}
