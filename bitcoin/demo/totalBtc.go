package main

import "fmt"

const blockInterval = 21 //区块删减区间，单位是万

func main() {
	total := 0.0   //总量
	reward := 50.0 //最初奖励50个, 每过21万个区块，奖励减半，大约是4年

	for reward > 0 {
		amount := reward * blockInterval //在21万个块中产生的比特币数量
		total = total + amount
		//reward /= 2 //比特币奖励衰减一半
		reward *= 0.5 //除法效率低，使用乘法代替
	}

	fmt.Printf("比特币总量为: %f\n", total)
}
