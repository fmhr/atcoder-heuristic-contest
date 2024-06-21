package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	input := readInput()
	solver(input)
}

var di = [4]int8{0, 1, 0, -1}
var dj = [4]int8{1, 0, -1, 0}

func solver(input Input) (state State) {
	state.color = input.color
	state.score = uint16(input.n) * uint16(input.n)
	log.Println(state.score)
	state.adjacencies = input.adjacencies
	timeLimit := 1900 * time.Millisecond
	loop := 0
	timeCheckInterval := 100
	for {
		// 局所遷移の選択
		selectRand := rand.Intn(100)
		switch {
		case selectRand < 100:
			// １マスを選んで接する別の色に変える
			// TODO カラーが残ってる四角形に縮める
			i := int8(rand.Intn(int(input.n)) + 1)
			j := int8(rand.Intn(int(input.n)) + 1)
			oldColor := state.color[i][j]
			// TODO 最終的には０から他の色に変えることもできるようにする
			if oldColor == 0 {
				continue
			}
			if eliminability(i, j, state.color) {
				randCd := rand.Intn(4)
				var newColor uint8
				for k := 0; k < 4; k++ {
					randCd = (randCd + k) % 4
					newColor = state.color[i+di[randCd]][j+dj[randCd]]
					if newColor != oldColor {
						break
					}
				}
				// 同じ色の場合は変更しない
				if newColor == oldColor {
					continue
				}
				// 変更
				if state.checkAdjacency(i, j, newColor) {
					state.move(i, j, oldColor, newColor)
				}
			}
		case selectRand < 100:
			// 1列or行を選んで削除する
		}
		// check time
		loop++
		if loop%timeCheckInterval == 0 {
			elp := input.timeElapsed()
			if elp > timeLimit {
				break
			}
			if elp < timeLimit+100*time.Millisecond {
				timeCheckInterval = 1
			}
		}
	}
	state.output()
	log.Println("score", state.score, "loop", loop)
	return state
}

// 1マスを選んで接する別の色に変えるとき、同じ色の接続が切れないかを確認する
// 消去可能性
func eliminability(y, x int8, grid [52][52]uint8) bool {
	diy := [9]int8{0, 0, -1, -1, -1, 0, 1, 1, 1}
	dix := [9]int8{0, 1, 1, 0, -1, -1, -1, 0, 1}
	var canDelete int8
	var i int8
	for i = 1; i < 9; i += 2 {
		if grid[y][x] == grid[y+dix[i]][x+diy[i]] {
			canDelete++
		}
	}
	for i := 1; i < 9; i += 2 {
		var b int8 = 1
		for j := 0; j < 3; j++ {
			ij := (i + j)
			if ij == 9 {
				ij = 1
			}
			if grid[y][x] != grid[(y + dix[ij])][x+diy[ij]] {
				b *= 0
				break
			}
		}
		canDelete -= b
	}
	//log.Println("canDelete", canDelete)
	return canDelete == 1
}

// 1マス変更したときに隣接関係が変わらないかを確認する
func (state State) checkAdjacency(y, x int8, newColor uint8) bool {
	oldColor := state.color[y][x]
	if oldColor == newColor {
		log.Fatal("same color")
	}
	// 1マス変更したときに隣接関係が変わらないかを確認する
	// 変化後にできた新しい隣接関係がすでにあるかを確認する
	if newColor != 0 && oldColor != 0 {
		for i := 0; i < 4; i++ {
			c := state.color[y+di[i]][x+dj[i]]
			// 元々の隣接関係がない場合は不受理
			if state.adjacencies[c][newColor] == 0 {
				return false
			}
			state.adjacencies[c][newColor]++
			state.adjacencies[newColor][c]++
		}
		// 減る隣接関係が他の場所で維持されているかを確認する
		for i := 0; i < 4; i++ {
			c := state.color[y+di[i]][x+dj[i]]
			state.adjacencies[c][oldColor]--
			state.adjacencies[oldColor][c]--
			if state.adjacencies[c][oldColor] == 0 {
				return false
			}
		}
	} else if newColor == 0 && oldColor != 0 {
		// 新しい隣接関係はできない
		for i := 0; i < 4; i++ {
			c := state.color[y+di[i]][x+dj[i]]
			// 1しかない場合は隣接関係が消えるので不受理
			if c != 0 {
				state.adjacencies[c][oldColor]--
				state.adjacencies[oldColor][c]--
				if state.adjacencies[c][oldColor] == 0 {
					return false
				}
				if state.adjacencies[c][newColor] == 0 {
					return false
				}
				state.adjacencies[c][newColor]++
				state.adjacencies[newColor][c]++
			}
		}
	} else if oldColor == 0 {
		for i := 0; i < 4; i++ {
			c := state.color[y+di[i]][x+dj[i]]
			// 元々の隣接関係がない場合は不受理
			if state.adjacencies[c][newColor] == 0 {
				return false
			}
			state.adjacencies[c][newColor]++
			state.adjacencies[newColor][c]++
		}
	} else {
		log.Fatal("when some color, not reachable")
	}
	return true
}

func (state *State) move(i, j int8, oldColor uint8, newColor uint8) {
	state.color[i][j] = newColor
	if oldColor != 0 && newColor == 0 {
		state.score--
	}
	for k := 0; k < 4; k++ {
		c := state.color[i+di[k]][j+dj[k]]
		state.adjacencies[c][newColor]++
		state.adjacencies[newColor][c]++
		state.adjacencies[c][oldColor]--
		state.adjacencies[oldColor][c]--
	}
}

type State struct {
	color       [52][52]uint8
	adjacencies [101][101]uint16
	score       uint16
}

//func (state State) output() {
//var i, j uint8
//for i = 1; i < 51; i++ {
//for j = 1; j < 51; j++ {
//fmt.Printf("%2d ", state.color[i][j])
//}
//fmt.Println()
//}
//return
//}

func (state State) output() {
	var i, j uint8
	var outputBuilder strings.Builder
	for i = 1; i < 51; i++ {
		for j = 1; j < 51; j++ {
			outputBuilder.WriteString(fmt.Sprintf("%2d ", state.color[i][j]))
		}
		outputBuilder.WriteString("\n")
	}

	fmt.Print(outputBuilder.String())
}

type Input struct {
	n           uint8
	m           uint8
	color       [52][52]uint8
	adjacencies [101][101]uint16 // カウントすることで隣接関係を管理する
	startTime   time.Time
}

func readInput() (input Input) {
	input.startTime = time.Now()
	var i, j uint8
	fmt.Scan(&input.n, &input.m)
	for i = 1; i < input.n+1; i++ {
		for j = 1; j < input.n+1; j++ {
			fmt.Scan(&input.color[i][j])
		}
	}
	// 隣接関係を記録する
	for i = 0; i < input.n+1; i++ {
		for j = 0; j < input.n+1; j++ {
			a := input.color[i][j]
			b := input.color[i+1][j]
			input.adjacencies[a][b]++
			input.adjacencies[b][a]++
			c := input.color[i][j+1]
			input.adjacencies[a][c]++
			input.adjacencies[c][a]++
		}
	}
	return
}

func (input *Input) timeElapsed() time.Duration {
	return time.Since(input.startTime)
}
