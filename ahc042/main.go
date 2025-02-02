package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"
)

var ATCODER int

func init() {
	flag.Parse()
	if os.Getenv("ATCODER") == "1" {
		ATCODER = 1
		log.SetOutput(io.Discard)
	}
}

type Input struct {
	n int
	c [20][20]byte
}

func input(re *bufio.Reader) (in Input) {
	fmt.Fscan(re, &in.n)
	for i := 0; i < in.n; i++ {
		var c []byte
		fmt.Fscan(re, &c)
		for j := 0; j < 20; j++ {
			in.c[i][j] = c[j]
		}
	}
	return
}

type State struct {
	state   [20][20]byte
	t, x, y int
	eval    int
	act     *ActionNode
}

func (s State) clone() *State {
	newState := &State{}
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			newState.state[i][j] = s.state[i][j]
		}
	}
	newState.t = s.t
	newState.x = s.x
	newState.y = s.y
	newState.eval = s.eval
	newState.act = s.act
	return newState
}

func NewState(in Input) *State {
	s := &State{}
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			s.state[i][j] = in.c[i][j]
		}
	}
	s.t = 0
	s.x = 40
	s.y = 40
	return s
}

func (s State) generateActions() []Action {
	oniExitRow := make([]bool, 20)
	oniExitColumn := make([]bool, 20)
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			if s.state[i][j] == oni {
				oniExitRow[i] = true
				oniExitColumn[j] = true
			}
		}
	}
	actions := make([]Action, 0)
	for i := 0; i < 20; i++ {
		if oniExitRow[i] {
			actions = append(actions, Action{Up, uint8(i)})
			actions = append(actions, Action{Down, uint8(i)})
		}
		if oniExitColumn[i] {
			actions = append(actions, Action{Left, uint8(i)})
			actions = append(actions, Action{Right, uint8(i)})
		}
	}
	return actions
}

type Action struct {
	act, target uint8
}

type ActionNode struct {
	act    Action
	parent *ActionNode
}

const (
	Left  = 1
	Right = 2
	Up    = 3
	Down  = 4
)

func actionStr(a uint8) string {
	switch a {
	case Left:
		return "L"
	case Right:
		return "R"
	case Up:
		return "U"
	case Down:
		return "D"
	}
	return "Unknown"
}

const (
	oni  = 'x'
	huku = 'o'
)

func (s *State) move(act Action) bool {
	switch act.act {
	case Left:
		if s.state[act.target][0] == oni {
			s.x--
		} else if s.state[act.target][0] == huku {
			return false
		}
		for i := 1; i < 20; i++ {
			s.state[act.target][i-1] = s.state[act.target][i]
		}
		s.state[act.target][19] = '.'
	case Right:
		if s.state[act.target][19] == oni {
			s.x--
		} else if s.state[act.target][19] == huku {
			return false
		}
		for i := 18; i >= 0; i-- {
			s.state[act.target][i+1] = s.state[act.target][i]
		}
		s.state[act.target][0] = '.'
	case Up:
		if s.state[0][act.target] == oni {
			s.x--
		} else if s.state[0][act.target] == huku {
			return false
		}
		for i := 1; i < 20; i++ {
			s.state[i-1][act.target] = s.state[i][act.target]
		}
		s.state[19][act.target] = '.'
	case Down:
		if s.state[19][act.target] == oni {
			s.x--
		} else if s.state[19][act.target] == huku {
			return false
		}
		for i := 18; i >= 0; i-- {
			s.state[i+1][act.target] = s.state[i][act.target]
		}
		s.state[0][act.target] = '.'
	}
	s.t++
	return true
}

// 評価関数
// 鬼の数が少ないほど良い
// 同じ行,列に複数の鬼がいると嬉しい
// 鬼が福に囲まれていると嬉しくない
// 鬼が外に血が付くと嬉しい
func (s State) calcEval() int {
	oniCount := s.x
	// 同じ列にいる鬼の数を行列ごとにカウント
	onis := 0
	for i := 0; i < 20; i++ {
		oniInColumn := 0
		for j := 0; j < 20; j++ {
			if s.state[i][j] == oni {
				oniInColumn++
			}
		}
		onis += oniInColumn * oniInColumn
	}
	for i := 0; i < 20; i++ {
		oniInRow := 0
		for j := 0; j < 20; j++ {
			if s.state[j][i] == oni {
				oniInRow++
			}
		}
		onis += oniInRow * oniInRow
	}
	// 鬼が外に出るまでの距離
	oniDistanceSum := 0
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			if s.state[i][j] == oni {
				// 前後左右で最も近い壁までの距離
				minDistance := minInt(minInt(i+1, 19-i+1), minInt(j+1, 19-j+1))
				oniDistanceSum += minDistance
			}
		}
	}
	//log.Printf("oni:%d onis:%d oniDis:%d\n", oniCount, onis, oniDistanceSum)
	return -oniCount*1000 + onis + -oniDistanceSum
}

// 問題のヒントを実装する
func hint(s State) {
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			if s.state[i][j] == oni {
				minimumStep := 100000
				minimumAction := make([]Action, 0)
				log.Println("turn:", s.t, "oni:", s.x, "fuku:", s.y, "eval:", s.calcEval())
				//log.Printf("鬼 %d %d\n", i, j)
				// LRUDを選択する
				// Left 左に福がいたらだめ
				hukuHit := false
				for k := j - 1; k >= 0 && s.x > 0; k-- {
					if s.state[i][k] == huku {
						hukuHit = true
						break
					}
				}
				if !hukuHit {
					num := j + 1 // 移動する回数
					if num < minimumStep {
						minimumStep = num
						minimumAction = make([]Action, 0)
						for k := 0; k < num; k++ {
							minimumAction = append(minimumAction, Action{Left, uint8(i)})
						}
						for k := 0; k < num && s.x > 0; k++ {
							minimumAction = append(minimumAction, Action{Right, uint8(i)})
						}
					}
				}
				// Right 右に福がいたらだめ
				hukuHit = false
				for k := j + 1; k < 20; k++ {
					if s.state[i][k] == huku {
						hukuHit = true
						break
					}
				}
				if !hukuHit {
					num := 20 - j // 移動する回数
					if num < minimumStep {
						minimumStep = num
						minimumAction = make([]Action, 0)
						for k := 0; k < num; k++ {
							minimumAction = append(minimumAction, Action{Right, uint8(i)})
						}
						for k := 0; k < num && s.x > 0; k++ {
							minimumAction = append(minimumAction, Action{Left, uint8(i)})
						}
					}
				}
				// Up 上に福がいたらだめ
				hukuHit = false
				for k := i - 1; k >= 0; k-- {
					if s.state[k][j] == huku {
						hukuHit = true
						break
					}
				}
				if !hukuHit {
					num := i + 1 // 移動する回数
					if num < minimumStep {
						minimumStep = num
						minimumAction = make([]Action, 0)
						for k := 0; k < num; k++ {
							minimumAction = append(minimumAction, Action{Up, uint8(j)})
						}
						for k := 0; k < num && s.x > 0; k++ {
							minimumAction = append(minimumAction, Action{Down, uint8(j)})
						}
					}
				}
				// Down 下に福がいたらだめ
				hukuHit = false
				for k := i + 1; k < 20; k++ {
					if s.state[k][j] == huku {
						hukuHit = true
						break
					}
				}
				if !hukuHit {
					num := 20 - i // 移動する回数
					if num < minimumStep {
						minimumStep = num
						minimumAction = make([]Action, 0)
						for k := 0; k < num; k++ {
							minimumAction = append(minimumAction, Action{Down, uint8(j)})
						}
						for k := 0; k < num && s.x > 0; k++ {
							minimumAction = append(minimumAction, Action{Up, uint8(j)})
						}
					}
				}
				if minimumStep < 100000 {
					for _, act := range minimumAction {
						s.move(act)
						fmt.Printf("%s %d\n", actionStr(act.act), act.target)
					}
				} else {
					log.Println("error oni", i, j)
					for k := 0; k < 20; k++ {
						log.Println(string(s.state[k][:]))
					}
					panic("error")
				}
			}
		}
	}
	log.Println("T=", s.t)
}

func beamSearch(in Input) {
	beamWidth := 10
	s := NewState(in)
	states := make([]State, 0)
	states = append(states, *s)
	nexts := make([]State, 0)
	loop := 0
	for loop < 100 {
		loop++
		for _, state := range states {
			actions := state.generateActions()
			if len(actions) == 0 {
				break
			}
			for _, act := range actions {
				newState := state.clone()
				ok := newState.move(act)
				if ok {
					newState.act = &ActionNode{act, state.act}
					newState.eval = newState.calcEval()
					nexts = append(nexts, *newState)
				}
			}
			sort.Slice(nexts, func(i, j int) bool {
				return nexts[i].eval > nexts[j].eval
			})
			if len(nexts) > beamWidth {
				nexts = nexts[:beamWidth]
			}
			states = make([]State, len(nexts))
			copy(states, nexts)
		}
		log.Println(states[0].t, states[0].x, states[0].eval, states[0].act)
	}
}

var startTime time.Time
var limitTime = 1900

func main() {
	startTime = time.Now()
	log.SetFlags(log.Lshortfile)
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	in := input(reader)
	//s := NewState(in)
	//hint(*s)
	beamSearch(in)
	log.Printf("time=%v\n", time.Since(startTime))
}

func getTimeMs() int {
	rtu := int(time.Since(startTime) / time.Millisecond)
	return rtu
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
