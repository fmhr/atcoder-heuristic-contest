package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
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

type ActionLog struct {
	Action
	num int
}

func revearseAction(act Action) Action {
	switch act.act {
	case Left:
		return Action{Right, act.target}
	case Right:
		return Action{Left, act.target}
	case Up:
		return Action{Down, act.target}
	case Down:
		return Action{Up, act.target}
	}
	return Action{}
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

// みつからないとき、actは空のスライスを返す
func findOniMove(s State, i, j int) (act Action, cnt int) {
	minimumStep := 100000
	var minimumAction Action
	//log.Println("turn:", s.t, "oni:", s.x, "fuku:", s.y, "eval:", s.calcEval())
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
			minimumAction = Action{Left, uint8(i)}
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
			minimumAction = Action{Right, uint8(i)}
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
			minimumAction = Action{Up, uint8(j)}
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
			minimumAction = Action{Down, uint8(j)}
		}
	}
	return minimumAction, minimumStep
}

// 問題のヒントを実装する
func hint(s State) {
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			if s.state[i][j] == oni {
				minimumAction, cnt := findOniMove(s, i, j)
				for i := 0; i < cnt; i++ {
					s.move(minimumAction)
					fmt.Printf("%s %d\n", actionStr(minimumAction.act), minimumAction.target)
				}
				for i := 0; i < cnt; i++ {
					s.move(revearseAction(minimumAction))
					fmt.Printf("%s %d\n", actionStr(revearseAction(minimumAction).act), revearseAction(minimumAction).target)
				}
			}
		}
	}
	log.Printf("T=%d\n", s.t)
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

type Pos struct {
	x, y int
}

// 外に出す鬼をランダムで選ぶ
func randomOniMove(s State) (actionLogs []ActionLog, allStep int, success bool) {
	loop := 0
	for ; loop < 50; loop++ {
		oniPos := make([]Pos, 0, 20)
		for i := 0; i < 20; i++ {
			for j := 0; j < 20; j++ {
				if s.state[i][j] == oni {
					oniPos = append(oniPos, Pos{y: i, x: j})
				}
			}
		}
		if len(oniPos) == 0 {
			success = true
			break
		}
		target := rand.Intn(len(oniPos))
		act, num := findOniMove(s, oniPos[target].y, oniPos[target].x)
		//num = rand.Intn(num) + 1
		if num < 100000 {
			for i := 0; i < num; i++ {
				s.move(act)
			}
			actionLogs = append(actionLogs, ActionLog{act, num})
			allStep += num
		}
		if len(oniPos) == 1 && num == 100000 {
			if s.move(Action{Left, uint8(oniPos[0].y)}) {
				actionLogs = append(actionLogs, ActionLog{Action{Left, uint8(oniPos[0].y)}, 1})
				allStep++
			} else if s.move(Action{Up, uint8(oniPos[0].x)}) {
				actionLogs = append(actionLogs, ActionLog{Action{Up, uint8(oniPos[0].x)}, 1})
				allStep++
			} else if s.move(Action{Down, uint8(oniPos[0].x)}) {
				actionLogs = append(actionLogs, ActionLog{Action{Down, uint8(oniPos[0].x)}, 1})
				allStep++
			} else if s.move(Action{Right, uint8(oniPos[0].y)}) {
				actionLogs = append(actionLogs, ActionLog{Action{Right, uint8(oniPos[0].y)}, 1})
				allStep++

			} else {
				return actionLogs, allStep, false
			}
			//log.Println("LAST", oniPos[0].x, oniPos[0].y)
			//log.Println("action", actionStr(actionLogs[len(actionLogs)-1].act), actionLogs[len(actionLogs)-1].target)
			//for i := 0; i < 20; i++ {
			//log.Printf("%s\n", string(s.state[i][:]))
			//}
		}
	}
	return actionLogs, allStep, success
}

func randomSearch(s State) {
	bestLog := make([]ActionLog, 0)
	bestStep := 1600
	var i int
	okCnt := 0
	ngCnt := 0
	for i = 0; ; i++ {
		actLogs, step, ok := randomOniMove(s)
		if ok && step < bestStep {
			bestStep = step
			bestLog = actLogs
			bestLog = make([]ActionLog, len(actLogs))
			copy(bestLog, actLogs)
			log.Println("step:", step)
			okCnt++
		} else {
			ngCnt++
		}
		if i%100 == 0 {
			since := getTimeMs()
			if since > limitTime {
				break
			}
		}
	}
	t := 0
	for _, actLog := range bestLog {
		for i := 0; i < actLog.num; i++ {
			fmt.Printf("%s %d\n", actionStr(actLog.act), actLog.target)
			t++
			if t > 1500 {
				return
			}
		}
	}
	log.Printf("loop=%d ok=%d ng=%d\n", i, okCnt, ngCnt)
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
	s := NewState(in)
	//hint(*s)
	//beamSearch(in)
	randomSearch(*s)
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
