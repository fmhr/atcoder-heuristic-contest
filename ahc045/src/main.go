package main

import (
	"fmt"
	"log"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	log.Println("Hello, World!")
	in := readInput()
	log.Println(in.M, in.L, in.W)
}

const (
	N = 800 // 都市の個数
	Q = 400 // クエリの個数
)

type Input struct {
	M        int        // 都市のグループの数 1<= M <= 400
	L        int        // クエリの都市の最大数 1<= L <= 15
	W        int        //　二次元座標の最大値 500 <= W <= 2500
	G        [400]int   // 各グループの都市の数 1<= G[i] <= N(800) i= 0..M-1
	lxrxlyry [N * 4]int // 各都市の座標 0 <= lxrxlyry[i] <= W
	// lxrxlyry[i] = (lx, rx, ly, ry) i=0..N-1
}

// 固定入力はとばす
func readInput() (in Input) {
	var n, q int
	fmt.Scan(&n, &in.M, &q, &in.L, &in.W)
	for i := 0; i < in.M; i++ {
		fmt.Scan(&in.G[i])
	}
	for i := 0; i < N*4; i++ {
		fmt.Scan(&in.lxrxlyry[i])
	}
	return in
}
