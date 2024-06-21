package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

func main() {
	rand.NewSource(1)
	log.SetFlags(log.Lshortfile)
	startTime := time.Now()
	readinput()
	timeDistance := time.Since(startTime)
	log.Println(timeDistance)
}

var N int
var start [2]int

const BLOCk byte = 0x10
const IMMOVABLE byte = 0x20

func readinput() {
	fmt.Scan(&N)
	var line string
	fmt.Scan(&start[0], &start[1])
	var board [50][50]byte
	for i := 0; i < N; i++ {
		fmt.Scan(&line)
		for j := 0; j < N; j++ {
			switch line[j] {
			case '#':
				board[i][j] = 0x10
				board[i][j] |= 0x20 // can't remove floag
			case '.':
				board[i][j] = 0x00
			}
		}
	}
	simulatedAnnealing(start, board)
	//stp, _ := simulate(start, board)
	//log.Println("step=", stp)
	//str := strBoard(board)
	//log.Println("\n", str)
}

type State struct {
	pos   [2]int
	board [50][50]byte
}

func (state State) score() int {
	stp, _ := simulate(state.pos, state.board)
	return stp
}

func (state State) getNeighbor() State {
	var neighbor State
	neighbor.pos = state.pos
	neighbor.board = state.board
	// 1 ブロックをランダムに置く
	// 2 ブロックをランダムに取り除く ブロックが多いほど取り除きやすい
	// ブロックをランダムに移動する 1と２の同時
	for {
		y := rand.Intn(N)
		x := rand.Intn(N)
		if start[0] == y && start[1] == x {
			// start position
			continue
		} else if neighbor.board[y][x]&BLOCk != 0 {
			// block
			if neighbor.board[y][x]&IMMOVABLE != 0 {
				// can't remove
				continue
			} else {
				neighbor.board[y][x] ^= BLOCk // remove block
				break
			}
		} else {
			// not block
			neighbor.board[y][x] |= BLOCk // add block
			break
		}
	}
	return neighbor
}

// スコアの差がtempの根の時に
var startTemp float64 = 0.5
var endTemp float64 = 0.1
var R int64 = 20

func temp(t int, T int) float64 {
	//return startTemp + (endTemp-startTemp)*float64(t)/float64(T)
	return startTemp + (endTemp-startTemp)*(float64(T-t)/float64(T))
}

func prob(nextScore, currentScore int, t int, T int) float64 {
	return math.Exp(float64(nextScore-currentScore) / temp(t, T))
}

func forceNext(nextScore, currentScore int, t int, T int) bool {
	tmp := startTemp + (endTemp-startTemp)*(float64(t)/float64(T))
	diff := nextScore - currentScore
	probab := math.Exp(float64(diff) * math.Pow(0.1, tmp))
	if probab > rand.Float64() {
		//log.Println("true", probab, t, diff)
		return true
	} else {
		//log.Println("false", probab)
		return false
	}
	//return probab > rand.Float64()
	//return prob(nextScore, currentScore, t, T) > rand.Float64()
}

var startTime time.Time

func simulatedAnnealing(start [2]int, board [50][50]byte) {

	var timeLimit time.Duration = 1900 * time.Millisecond
	startTime = time.Now()
	var t time.Duration

	var currentState State
	currentState.pos = start
	currentState.board = board

	var timeout bool
	go func() {
		time.Sleep(timeLimit)
		timeout = true
	}()

	timeoutCh := time.After(timeLimit)
	//ticker := time.NewTicker(1 * time.Millisecond)
	//defer ticker.Stop()

	// rand.Seed(time.Now().UnixNano())
	var loopCnt int
OuterLoop:
	for !timeout {
		loopCnt++
		t = time.Since(startTime)
		if t > timeLimit {
			break
		}
		select {
		case <-timeoutCh:
			break OuterLoop
		default:
			newState := currentState.getNeighbor()
			newScore := newState.score()
			currentScore := currentState.score()
			if newScore >= currentScore || forceNext(newScore, currentScore, int(t), int(timeLimit)) {
				currentState = newState
			}

		}
	}
	stp, brd := simulate(currentState.pos, currentState.board)
	log.Println("step=", stp)
	log.Println(strBoard(brd))
	output(brd)
}

func output(b [50][50]byte) {
	ans := make([][2]byte, 0)
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if b[i][j]&BLOCk != 0 && b[i][j]&IMMOVABLE == 0 {
				ans = append(ans, [2]byte{byte(i), byte(j)})
			}
		}
	}
	fmt.Println(len(ans))
	for i := 0; i < len(ans); i++ {
		fmt.Println(ans[i][0], ans[i][1])
	}
}

// 右回り
var dy [4]int = [4]int{0, 1, 0, -1}
var dx [4]int = [4]int{1, 0, -1, 0}

func simulate(pos [2]int, board [50][50]byte) (int, [50][50]byte) {
	var dir int
	var step int
	for {
		for i := 0; i < 4; i++ {
			d := (dir + i) % 4
			nexty := pos[0] + dy[d]
			nextx := pos[1] + dx[d]
			if !(nexty < 0 || nexty >= N || nextx < 0 || nextx >= N) {
				// on board
				if board[nexty][nextx]&BLOCk == 0 {
					// not block
					// 無限ループ判定 正常終了
					if board[nexty][nextx]&convertDir(d) != 0 {
						return step, board
					}
					board[nexty][nextx] |= convertDir(d)
					pos[0] = nexty
					pos[1] = nextx
					dir = d
					break
				}
			}
		}
		step++
		if step == 1 && pos[0] == start[0] && pos[1] == start[1] {
			return step, board
		}
	}
}

func convertDir(dir int) byte {
	switch dir {
	case 0:
		return 0x01
	case 1:
		return 0x02
	case 2:
		return 0x04
	case 3:
		return 0x08
	}
	return 0
}

func strBoard(board [50][50]byte) string {
	var str string
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if board[i][j]&0x10 != 0 {
				str += "#"
			} else if board[i][j] == 0 {
				str += "."
			} else {
				str += string(board[i][j] + '0')
			}
		}
		str += "\n"
	}
	return str
}
