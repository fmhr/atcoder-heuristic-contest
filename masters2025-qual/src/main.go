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
	log.Printf("time=%v", time.Since(startTime))
}

func beamSearch(in In) (ans []Action) {
	root := &Node{parent: nil}
	// 初期状態
	initialState := newState(in)
	//initialState.showGrid()
	initialState.act = root
	log.Println("initialState.eval()", initialState.eval())
	//ant := initialState.generateAction()
	//log.Println(ant)

	//panic("stop")
	// ビーム幅
	beamWidth := 1
	states := make([]*State, 0, beamWidth)
	states = append(states, initialState)
	nextStates := make([]*State, 0, beamWidth)
	for i := 0; i < 5000; i++ {
		// ビーム幅分の状態を生成
		for j := range states {
			//actions := states[j].generateAction()
			for _, action := range allactions {
				g := states[j].grid[index(states[j].pos.y, states[j].pos.x)]
				if g == '.' && (action.act == Carry || action.act == Roll) {
					continue
				}
				if isHole(g) && (action.act == Carry || action.act == Roll) {
					continue
				}
				a := action
				s := states[j]
				if a.dict == Up && s.pos.y == 0 {
					continue
				}
				if a.dict == Down && s.pos.y == 19 {
					continue
				}
				if a.dict == Left && s.pos.x == 0 {
					continue
				}
				if a.dict == Right && s.pos.x == 19 {
					continue
				}
				if a.act == Carry && s.grid[index(s.pos.y, s.pos.x)] == '.' {
					continue
				}
				if a.act == Carry && isHole(s.grid[index(s.pos.y, s.pos.x)]) {
					continue
				}
				if a.act == Roll && s.grid[index(s.pos.y, s.pos.x)] == '.' {
					continue
				}
				if a.act == Roll && isHole(s.grid[index(s.pos.y, s.pos.x)]) {
					continue
				}
				// 移動先に鉱石か岩があったら運べない、転がせない
				if a.act == Carry || a.act == Roll {
					nextPos := Pos{s.pos.y + dy[a.dict], s.pos.x + dx[a.dict]}
					if isRock(s.grid[index(nextPos.y, nextPos.x)]) || isStone(s.grid[index(nextPos.y, nextPos.x)]) {
						continue
					}
				}
				newState := states[j].Clone()
				if newState.Do(action) {
					newState.act = &Node{act: action, parent: states[j].act}
					nextStates = append(nextStates, newState)
				} else {
					log.Println("Do failed", action)
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
		//log.Println(i, states[0].score, states[0].eval(), states[0].act.act)
		//states[0].showGrid()
		if states[0].stones[0]+states[0].stones[1]+states[0].stones[2] == 0 {
			log.Println("finish")
			break
		}
		//log.Println(states[0].stones)
	}
	lastAct := states[0].act
	for lastAct != nil {
		ans = append(ans, lastAct.act)
		lastAct = lastAct.parent
	}
	for i := 0; i < len(ans)/2; i++ {
		ans[i], ans[len(ans)-1-i] = ans[len(ans)-1-i], ans[i]
	}
	log.Println(len(ans))
	return ans
}

func index(y, x int) int {
	return y*GridSize + x
}

type State struct {
	grid   [GridSize * GridSize]byte
	stones [3]int
	pos    Pos
	score  int
	act    *Node
}

var distance [GridSize * GridSize]int

func (s State) eval() int {
	if s.stones[0]+s.stones[1]+s.stones[2] == 0 {
		return s.score * 10000000
	}
	// distance from pos
	// @をさけて移動する時の距離
	for i := 0; i < GridSize*GridSize; i++ {
		distance[i] = 100000
	}
	distance[index(s.pos.y, s.pos.x)] = 0
	que := make([]Pos, 0, 20*20)
	que = append(que, s.pos)
	for len(que) > 0 {
		p := que[0]
		que = que[1:]
		for _, d := range directions {
			nextPos := Pos{p.y + dy[d], p.x + dx[d]}
			if !(nextPos.y >= 0 && nextPos.y < 20 && nextPos.x >= 0 && nextPos.x < 20) {
				continue
			}
			if isRock(s.grid[index(nextPos.y, nextPos.x)]) {
				continue
			}
			if distance[index(nextPos.y, nextPos.x)] > distance[index(p.y, p.x)]+1 {
				distance[index(nextPos.y, nextPos.x)] = distance[index(p.y, p.x)] + 1
				que = append(que, nextPos)
			}
		}
	}
	//for i := 0; i < 20; i++ {
	//log.Println(i, distance[i*20:i*20+20])
	//}
	//panic("stop")

	// なにもないとき
	// もっとも近い鉱石までの距離
	minStoneDist := 1000
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			if isStone(s.grid[index(i, j)]) {
				//dist := abs(i-s.pos.y) + abs(j-s.pos.x)
				dist := distance[index(i, j)]
				minStoneDist = minInt(minStoneDist, dist)
			}
		}
	}
	var minHoleDist int
	// もっとも近い穴までの距離
	if minStoneDist == 0 {
		minHoleDist = 1000
		for i := 0; i < 20; i++ {
			for j := 0; j < 20; j++ {
				if isHole(s.grid[index(i, j)]) {
					//dist := minInt(abs(i-s.pos.y), abs(j-s.pos.x))
					dist := distance[index(i, j)]
					minHoleDist = minInt(minHoleDist, dist)
				}
			}
		}
	}
	bonus := 0
	if minStoneDist == 0 {
		bonus = 100
	}
	//log.Println("minStoneDist", minStoneDist, "minHoleDist", minHoleDist, "bonus", bonus)
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
	newState.stones = s.stones
	return newState
}

func newState(in In) *State {
	s := &State{}
	copy(s.grid[:], in.grid[:])
	for i := 0; i < 20*20; i++ {
		if in.grid[i] == 'A' {
			s.pos = Pos{i / 20, i % 20}
		}
		if in.grid[i] == 'a' {
			s.stones[0]++
		}
	}
	log.Println("start pos", s.pos)
	return s
}

func (s *State) Do(a Action) bool {

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
				stone := s.grid[index(s.pos.y, s.pos.x)]
				if s.grid[index(s.pos.y, s.pos.x)] == s.grid[index(nextPos.y, nextPos.x)]+32 {
					//log.Println("鉱石を穴に落とした")
					s.score++
				} else {
					log.Println("鉱石を違う穴に落とした")
				}
				switch stone {
				case 'a':
					s.stones[0]--
				case 'b':
					s.stones[1]--
				case 'c':
					s.stones[2]--
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
		}
		s.pos = nextPos
	case Roll:
		// 転がす 現在位置は動かない
		nextPos := s.pos
		prev := s.pos
		//nextPos.y += dy[a.dict]
		//nextPos.x += dx[a.dict]
		//log.Println("now", s.pos, string(s.grid[index(s.pos.y, s.pos.x)]))
		//log.Println("prev", prev, string(s.grid[index(prev.y, prev.x)]))
		//log.Println("next", nextPos, string(s.grid[index(nextPos.y, nextPos.x)]))
		// 進めるだけ進む
		for {
			prev = nextPos
			nextPos.y += dy[a.dict]
			nextPos.x += dx[a.dict]
			if !(nextPos.y >= 0 && nextPos.y < 20 && nextPos.x >= 0 && nextPos.x < 20) {
				nextPos = prev
				break
			}
			//log.Println(nextPos, string(s.grid[index(nextPos.y, nextPos.x)]))
			if s.grid[index(nextPos.y, nextPos.x)] != '.' {
				break
			}
		}
		itemToRall := s.grid[index(s.pos.y, s.pos.x)]
		s.grid[index(s.pos.y, s.pos.x)] = '.'
		//log.Println("next", nextPos, "prev", prev, string(s.grid[index(nextPos.y, nextPos.x)]))
		if isHole(s.grid[index(nextPos.y, nextPos.x)]) {
			// 穴で止まる
			if isStone(itemToRall) {
				// 石が落ちる
				if itemToRall == s.grid[index(nextPos.y, nextPos.x)]+32 {
					//log.Println("鉱石を穴に落とした")
					s.score++
				} else {
					log.Println(s.grid[index(s.pos.y, s.pos.x)], s.grid[index(nextPos.y, nextPos.x)])
					log.Println("鉱石を違う穴に落とした")
				}
				switch itemToRall {
				case 'a':
					s.stones[0]--
				case 'b':
					s.stones[1]--
				case 'c':
					s.stones[2]--
				}
			}
		} else {
			// 穴ではないので転がってグリッドの上に残る
			if s.grid[index(prev.y, prev.x)] != '.' {
				log.Println("now", s.pos, string(s.grid[index(s.pos.y, s.pos.x)]))
				log.Println("prev", prev, string(s.grid[index(prev.y, prev.x)]))
				log.Println("next", nextPos, string(s.grid[index(nextPos.y, nextPos.x)]))
				panic("stop")
			}
			s.grid[index(prev.y, prev.x)] = itemToRall
		}
	}
	return true
}

var dy = []int{0, -1, 1, 0, 0}
var dx = []int{0, 0, 0, -1, 1}
var allactions = []Action{
	{act: Move, dict: Up},
	{act: Move, dict: Down},
	{act: Move, dict: Left},
	{act: Move, dict: Right},
	{act: Carry, dict: Up},
	{act: Carry, dict: Down},
	{act: Carry, dict: Left},
	{act: Carry, dict: Right},
	{act: Roll, dict: Up},
	{act: Roll, dict: Down},
	{act: Roll, dict: Left},
	{act: Roll, dict: Right},
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
