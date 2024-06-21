package main

import (
	"fmt"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	solver()
}

func solver() {
	var N, M int
	fmt.Scan(&N, &M)
	// N個のアイテム
	// M個の山
	log.Println(N, M)
	b := make([][]int, M)
	var state [20][]*Box
	for i := 0; i < M; i++ {
		bj := make([]int, N/M)
		for j := 0; j < N/M; j++ {
			fmt.Scan(&bj[j])
			boxs[bj[j]] = Box{bj[j], i, j}
			state[i] = append(state[i], &boxs[bj[j]])
		}
		b[i] = make([]int, len(bj))
		b[i] = bj
	}
	log.Println(b)
	score := 0
	cnt := 0
	for i := 1; i < 201; i++ {
		for {
			if onTop(boxs[i], &state) {
				delete(i, 0, &state)
				cnt++
				log.Println("delete", i, 0)
				fmt.Println(i, 0)
				break
			}
			// スタックの上から見ていき、単純増加が崩れたところまでを移動させる
			// 上を繰り返して目的の箱が一番上に来るようにする
			stackNum := boxs[i].stackNum
			for j := len(state[stackNum]) - 1; j > boxs[i].stackIndex; j-- {
				// 短調増加が崩れる条件
				if state[stackNum][j].index > state[stackNum][j-1].index {
					// jまでの箱を移動させる
					boxj := state[stackNum][j]
					next := selectNext(*boxj, &state)
					score += move(boxj.index, next, &state)
					cnt++
					log.Println("target", i, "move", boxj, stackNum, "->", next)
					fmt.Println(boxj.index, next+1)
					break
				}
			}
			str := ""
			for k := 0; k < len(state[stackNum]); k++ {
				str += fmt.Sprintf("%v ", state[stackNum][k])
			}
		}
	}
	log.Println(cnt, score)
}

var boxs [201]Box

type Box struct {
	index      int
	stackNum   int
	stackIndex int
}

func move(u, v int, state *[20][]*Box) (num int) {
	// 箱u(とその上の箱全て）を山vの上にに移動する
	// 箱uがどの山の何番目にあるか
	size := len(state[v])
	b := boxs[u]
	state[v] = append(state[v], state[b.stackNum][b.stackIndex:]...)
	state[b.stackNum] = state[b.stackNum][:b.stackIndex]
	for i := 0; i < len(state[v]); i++ {
		boxIndex := state[v][i].index
		boxs[boxIndex].stackNum = v
		boxs[boxIndex].stackIndex = i
	}
	return len(state[v]) - size
}

func delete(u, v int, state *[20][]*Box) {
	if v != 0 {
		log.Fatal("v != 0")
	}
	if boxs[u].stackIndex != len(state[boxs[u].stackNum])-1 {
		log.Fatal("u is not top")
	}
	state[boxs[u].stackNum] = state[boxs[u].stackNum][:boxs[u].stackIndex]
	boxs[u].stackNum = -1
	boxs[u].stackIndex = -1
}

func onTop(u Box, state *[20][]*Box) bool {
	if u.stackIndex == len(state[u.stackNum])-1 {
		return true
	}
	return false
}

func selectNext(box Box, state *[20][]*Box) (next int) {
	current := box.stackNum
	minmax := 0
	next = -1
	for i := 0; i < 10; i++ {
		if i == current {
			continue
		}
		if len(state[i]) == 0 {
			return i
		}
		minimum := 10000
		for j := 0; j < len(state[i]); j++ {
			if state[i][j].index < minimum {
				minimum = state[i][j].index
			}
		}
		if minimum > minmax {
			minmax = minimum
			next = i
		}
	}
	return
}

// 1. 山の上には小さい数字がある方がいい
// 2. ある箱の下に、その箱より小さい数字がいくつあるか（マイナス要素）
// 3. 箱の中で、数字はソートされていた方がいい
// 4. ある箱の中で、数字が連続している方が良い状態と言える
// 5. ハノイの塔　ハイパーキューブにおけるハミルトン閉路
