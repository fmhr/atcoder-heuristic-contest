package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"
)

var ATCODER bool        // AtCoder環境かどうか
var startTime time.Time // 開始時刻
var frand *rand.Rand    // 固定用乱数生成機

func main() {
	if os.Getenv("ATCODER") == "1" {
		ATCODER = true
		log.Println("on AtCoder")
		log.SetOutput(io.Discard)
	}
	log.SetFlags(log.Lshortfile)
	frand = rand.New(rand.NewSource(1))
	startTime = time.Now()
	in := readInput()
	ans := beamSearch(in)
	for _, a := range ans {
		fmt.Println(a)
	}
	//log.Printf("in: %+v\n", in)
}

func beamSearch(in In) (ans []Action) {
	root := &Node{parent: nil}
	// 初期状態
	initialState := newState(in)
	initialState.showGrid()
	initialState.act = root
	log.Println("initialState.eval()", initialState.eval())
	// ビーム幅
	beamWidth := 1
	states := make([]*State, 0, beamWidth)
	states = append(states, initialState)
	nextStates := make([]*State, 0, beamWidth)
	for i := 0; i < 10000; i++ {
		// ビーム幅分の状態を生成
		for j := range states {
			actions := states[j].generateAction()
			for _, action := range actions {
				newState := states[j].Clone()
				if newState.Do(action) {
					newState.act = &Node{act: action, parent: states[j].act}
					nextStates = append(nextStates, newState)
				} else {
					//log.Println("Do failed", action)
				}
			}
		}
		//log.Println(i, len(nextStates))
		sort.Slice(nextStates, func(i, j int) bool {
			return nextStates[i].eval() > nextStates[j].eval()
		})
		states = make([]*State, minInt(beamWidth, len(nextStates)))
		copy(states, nextStates)
		nextStates = make([]*State, 0)
		log.Println(i, states[0].score, states[0].eval(), states[0].act.act)
		states[0].showGrid()
	}
	lastAct := states[0].act
	for lastAct != nil {
		ans = append(ans, lastAct.act)
		lastAct = lastAct.parent
	}
	return ans
}

func index(y, x int) int {
	return y*GridSize + x
}

type State struct {
	grid  [GridSize * GridSize]byte
	pos   Pos
	score int
	act   *Node
}

func (s State) eval() int {

	var minStoneDist int
	// なにもないとき
	// もっとも近い鉱石までの距離
	minStoneDist = 100
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			if isStone(s.grid[index(i, j)]) {
				dist := abs(i-s.pos.y) + abs(j-s.pos.x)
				minStoneDist = minInt(minStoneDist, dist)
			}
		}
	}
	var minHoleDist int
	// もっとも近い穴までの距離
	if minStoneDist == 0 {
		minHoleDist = 100
		for i := 0; i < 20; i++ {
			for j := 0; j < 20; j++ {
				if isHole(s.grid[index(i, j)]) {
					dist := minInt(abs(i-s.pos.y), abs(j-s.pos.x))
					minHoleDist = minInt(minHoleDist, dist)
				}
			}
		}
	}
	bonus := 0
	if minStoneDist == 0 {
		bonus = 10
	}
	return s.score*1000 - minStoneDist - minHoleDist + bonus
}

func (s State) showGrid() {
	log.Println("Pos:", s.pos, "score:", s.score, "eval", s.eval(), "have", string(s.grid[index(s.pos.y, s.pos.x)]))
	for i := 0; i < 20; i++ {
		log.Println(i, string(s.grid[i*20:i*20+20]))
	}
}

func (s State) Clone() *State {
	newState := &State{}
	copy(newState.grid[:], s.grid[:])
	newState.pos = s.pos
	newState.score = s.score
	return newState
}

func newState(in In) *State {
	s := &State{}
	copy(s.grid[:], in.grid[:])
	for i := 0; i < 20*20; i++ {
		if in.grid[i] == 'A' {
			s.pos = Pos{i / 20, i % 20}
		}
	}
	log.Println("start pos", s.pos)
	return s
}

func (s *State) Do(a Action) bool {
	// 移動は全てグリッドの中
	if a.dict == Up && s.pos.y == 0 {
		return false
	}
	if a.dict == Down && s.pos.y == 19 {
		return false
	}
	if a.dict == Left && s.pos.x == 0 {
		return false
	}
	if a.dict == Right && s.pos.x == 19 {
		return false
	}
	if a.act == Carry && s.grid[index(s.pos.y, s.pos.x)] == '.' {
		return false
	}
	if a.act == Carry && isHole(s.grid[index(s.pos.y, s.pos.x)]) {
		return false
	}
	if a.act == Roll && s.grid[index(s.pos.y, s.pos.x)] == '.' {
		return false
	}
	if a.act == Roll && isHole(s.grid[index(s.pos.y, s.pos.x)]) {
		return false
	}
	// 移動先に鉱石か岩があったら運べない、転がせない
	if a.act == Carry || a.act == Roll {
		nextPos := Pos{s.pos.y + dy[a.dict], s.pos.x + dx[a.dict]}
		if isRock(s.grid[index(nextPos.y, nextPos.x)]) || isHole(s.grid[index(nextPos.y, nextPos.x)]) {
			return false
		}
	}
	switch a.act {
	case Move:
		// 単純な移動
		s.pos.y += dy[a.dict]
		s.pos.x += dx[a.dict]
	case Carry:
		// 持って運ぶ　行き先は穴か空き地
		nextPos := Pos{s.pos.y + dy[a.dict], s.pos.x + dx[a.dict]}
		if isHole(s.grid[index(nextPos.y, nextPos.x)]) {
			// 行き先が穴ならなんでも落ちる
			if isStone(s.grid[index(s.pos.y, s.pos.x)]) {
				if s.grid[index(s.pos.y, s.pos.x)] == s.grid[index(nextPos.y, nextPos.x)]+32 {
					// 同じアルファベットの大文字と小文字
					s.score++
					log.Println("鉱石を穴に落とした")
				} else {
					log.Println("鉱石を違う穴に落とした")
				}
			} else if isRock(s.grid[index(s.pos.y, s.pos.x)]) {
				// 岩を落とした
				log.Println("岩を落とした")
			}
			// 穴に落としたら元のマスは空になる
			s.grid[index(s.pos.y, s.pos.x)] = '.'
		} else if s.grid[index(nextPos.y, nextPos.x)] == '.' {
			s.grid[index(nextPos.y, nextPos.x)] = s.grid[index(s.pos.y, s.pos.x)]
			s.grid[index(s.pos.y, s.pos.x)] = '.'
			s.pos = nextPos
		}
	case Roll:
		// 転がす 現在位置は動かない
		nextPos := s.pos
		prev := s.pos
		nextPos.y += dy[a.dict]
		nextPos.x += dx[a.dict]
		// 進めるだけ進む
		for {
			prev = nextPos
			nextPos.y += dy[a.dict]
			nextPos.x += dx[a.dict]
			if !(nextPos.y >= 0 && nextPos.y < 20 && nextPos.x >= 0 && nextPos.x < 20) {
				nextPos = prev
				break
			}
			if s.grid[index(nextPos.y, nextPos.x)] != '.' {
				break
			}
		}
		log.Println("next", nextPos, "prev", prev, string(s.grid[index(nextPos.y, nextPos.x)]))
		if isHole(s.grid[index(nextPos.y, nextPos.x)]) {
			// 穴で止まる
			if isStone(s.grid[index(s.pos.y, s.pos.x)]) {
				if s.grid[index(s.pos.y, s.pos.x)] == s.grid[index(nextPos.y, nextPos.x)]+32 {
					log.Println("鉱石を穴に落とした")
					s.score++
				} else {
					log.Println(s.grid[index(s.pos.y, s.pos.x)], s.grid[index(nextPos.y, nextPos.x)])
					log.Println("鉱石を違う穴に落とした")
				}
			}
			if isRock(s.grid[index(s.pos.y, s.pos.x)]) {
				// 岩を落とした
				log.Println("岩を穴に落とした")
			}
		} else {
			s.grid[index(prev.y, prev.x)] = s.grid[index(s.pos.y, s.pos.x)]
		}
		s.grid[index(s.pos.y, s.pos.x)] = '.'
	}
	return true
}

var dy = []int{0, -1, 1, 0, 0}
var dx = []int{0, 0, 0, -1, 1}

func (s State) generateAction() (actions []Action) {
	actions = make([]Action, 0, 12)
	// 全ての条件で使える行動
	actions = append(actions, Action{act: Move, dict: Up})
	actions = append(actions, Action{act: Move, dict: Down})
	actions = append(actions, Action{act: Move, dict: Left})
	actions = append(actions, Action{act: Move, dict: Right})
	if s.grid[index(s.pos.y, s.pos.x)] == '.' || isHole(s.grid[index(s.pos.y, s.pos.x)]) {
		// 現在地が、空き地または穴の時は運ぶ、転がすはできない
		return actions
	}
	// a-zの鉱山, または岩が存在する時
	actions = append(actions, Action{act: Carry, dict: Up})
	actions = append(actions, Action{act: Carry, dict: Down})
	actions = append(actions, Action{act: Carry, dict: Left})
	actions = append(actions, Action{act: Carry, dict: Right})
	actions = append(actions, Action{act: Roll, dict: Up})
	actions = append(actions, Action{act: Roll, dict: Down})
	actions = append(actions, Action{act: Roll, dict: Left})
	actions = append(actions, Action{act: Roll, dict: Right})
	return actions
}

type Pos struct {
	y, x int
}

type Act int

const (
	Move  Act = 1
	Carry Act = 2
	Roll  Act = 3
)

var acts = []Act{Move, Carry, Roll}

type Direction int

const (
	Up    Direction = 1
	Down  Direction = 2
	Left  Direction = 3
	Right Direction = 4
)

var directions = []Direction{Up, Down, Left, Right}

type Action struct {
	act  Act
	dict Direction
}

func (a Action) String() string {
	switch a.dict {
	case Up:
		return fmt.Sprintf("%d U", a.act)
	case Down:
		return fmt.Sprintf("%d D", a.act)
	case Left:
		return fmt.Sprintf("%d L", a.act)
	case Right:
		return fmt.Sprintf("%d R", a.act)
	}
	return ""
}

type Node struct {
	act    Action
	parent *Node
}

const GridSize = 20

func isHole(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func isStone(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func isRock(c byte) bool {
	return c == '@'
}

type In struct {
	N, M int
	grid [GridSize * GridSize]byte
}

func readInput() In {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	var in In
	_, err := fmt.Fscan(reader, &in.N, &in.M)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < in.N; i++ {
		var line string
		_, err := fmt.Fscan(reader, &line)
		if err != nil {
			log.Fatal(err)
		}
		for j, c := range line {
			in.grid[i*GridSize+j] = byte(c)
		}
	}
	return in
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
