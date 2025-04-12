package main

import (
	"bufio"
	crand "crypto/rand"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"
)

var ATCODER bool        // AtCoder環境かどうか
var startTime time.Time // 開始時刻
var frand *rand.Rand    // 固定用乱数生成機

var primes [500]uint64

func init() {
	for i := 0; i < 500; i++ {
		prime, _ := crand.Prime(crand.Reader, 64)
		primes[i] = prime.Uint64()
	}
}

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
	log.Printf("time=%v", time.Since(startTime).Milliseconds())
}

func beamSearch(in In) (ans []Action) {
	root := &Node{parent: nil}
	// 初期状態
	initialState := newState(in)
	//log.Printf("a=%d b=%d c=%d\n", initialState.stones[0], initialState.stones[1], initialState.stones[2])
	initialState.act = root
	log.Println("initialState.eval()", initialState.calEval())
	// 初期状態でのallDistanceを計算
	//calcAllDistance(initialState)
	beamWidth := 40 // ビーム幅
	states := make([]*State, 0, beamWidth)
	states = append(states, initialState)
	nextStates := make([]*State, 0, beamWidth)
	exits := make(map[uint64]bool)
	for i := 0; i < 5000; i++ {
		// ビーム幅分の状態を生成
		for j := range states {
			//actions := states[j].generateAction()
			for _, action := range allactions {
				g := states[j].grid[index(states[j].pos.y, states[j].pos.x)]
				if g == '.' && (action.act == Carry || action.act == Roll) {
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
					if !exits[newState.hash] {
						newState.act = &Node{act: action, parent: states[j].act}
						nextStates = append(nextStates, newState)
						exits[newState.hash] = true
					} else {
						//log.Println("exit", newState.hash)
					}
				} else {
					log.Println("Do failed", action)
				}
			}
		}
		//log.Println(i, len(nextStates))
		sort.Slice(nextStates, func(i, j int) bool {
			return nextStates[i].eval > nextStates[j].eval
		})
		states = make([]*State, minInt(beamWidth, len(nextStates)))
		copy(states, nextStates)
		nextStates = make([]*State, 0)
		//log.Println(i, states[0].score, states[0].eval, states[0].act.act, states[0].pos)
		//states[0].outputState()

		// 終了条件
		//fmt.Println(i, len(states))
		//states[0].outputState()
		if states[0].stones[0]+states[0].stones[1]+states[0].stones[2] == 0 {
			break
		}
		//log.Println(states[0].stones)
	}
	lastAct := states[0].act
	for lastAct.parent != nil {
		ans = append(ans, lastAct.act)
		lastAct = lastAct.parent
	}
	for i := 0; i < len(ans)/2; i++ {
		ans[i], ans[len(ans)-1-i] = ans[len(ans)-1-i], ans[i]
	}
	log.Printf("turn=%d\n", len(ans))
	return ans
}

func index(y, x int) int {
	return y*GridSize + x
}

var holes [3]Pos

type State struct {
	grid   [GridSize * GridSize]byte
	stones [3]int
	pos    Pos
	score  int
	act    *Node
	eval   int
	hash   uint64
}

func (s State) makeHash() (hash uint64) {
	hash = primes[0]
	hash ^= uint64(s.pos.y) * primes[1]
	hash ^= uint64(s.pos.x) * primes[2]
	for i := 0; i < GridSize*GridSize; i++ {
		hash ^= uint64(s.grid[i]) * primes[i+3]
	}
	return
}

func (s State) outputState() {
	log.Println(s.stones)
	log.Println(s.pos)
	log.Println(s.score)
	log.Println(s.eval)
	for i := 0; i < 20; i++ {
		log.Println(string(s.grid[i*20 : i*20+20]))
	}
}

// 岩を避けて移動する時の全点間の距離
// Warshall Floyd
func calcAllDistance(s *State) {
	allDistance := make([][]int, 20*20)
	for i := 0; i < 20*20; i++ {
		allDistance[i] = make([]int, 20*20)
		for j := 0; j < 20*20; j++ {
			allDistance[i][j] = math.MaxInt16
		}
	}
	for i := 0; i < GridSize*GridSize; i++ {
		y, x := i/GridSize, i%GridSize
		if isRock(s.grid[i]) {
			continue
		}
		for j := 1; j < 5; j++ {
			ny, nx := y+dy[j], x+dx[j]
			if ny >= 0 && ny < 20 && nx >= 0 && nx < 20 && !isRock(s.grid[ny*GridSize+nx]) {
				allDistance[i][ny*GridSize+nx] = 1
			}
		}
	}
	for i := 0; i < 20*20; i++ {
		allDistance[i][i] = 0
	}
	for k := 0; k < 20*20; k++ {
		for i := 0; i < 20*20; i++ {
			for j := 0; j < 20*20; j++ {
				allDistance[i][j] = minInt(allDistance[i][j], allDistance[i][k]+allDistance[k][j])
			}
		}
	}
}

// distanceFromHole 穴までの距離を計算する
// typ = 'A' 'B' 'C'
// holeの上下左右の空きますには1をいれる
func (s State) distanceFromHole(typ byte) (dist [GridSize * GridSize]int) {
	var start Pos // 開始位置
	for _, p := range holes {
		if s.grid[index(p.y, p.x)] == typ {
			start = p
			break
		}
	}
	// 初期化
	for i := 0; i < GridSize*GridSize; i++ {
		dist[i] = math.MaxInt32
	}
	dist[index(start.y, start.x)] = 0
	q := make([]Pos, 0, 20*20)
	q = append(q, start)
	// 上下左右の空きますには1をいれる
	for d := 1; d < 5; d++ {
		ny, nx := start.y+dy[d], start.x+dx[d]
		for {
			if ny >= 0 && ny < 20 && nx >= 0 && nx < 20 && s.grid[index(ny, nx)] != '@' {
				dist[index(ny, nx)] = 1
				q = append(q, Pos{ny, nx})
				ny, nx = ny+dy[d], nx+dx[d]
			} else {
				break
			}
		}
	}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		for d := 1; d < 5; d++ {
			ny, nx := p.y+dy[d], p.x+dx[d]
			if ny >= 0 && ny < 20 && nx >= 0 && nx < 20 && s.grid[index(ny, nx)] != '@' {
				if dist[index(ny, nx)] > dist[index(p.y, p.x)]+1 {
					dist[index(ny, nx)] = dist[index(p.y, p.x)] + 1
					q = append(q, Pos{ny, nx})
				}
			}
		}
	}
	// デバッグ用
	//	for i := 0; i < 20; i++ {
	//line := ""
	//for j := 0; j < 20; j++ {
	//d := dist[index(i, j)]
	//if d == math.MaxInt {
	//line += " XX"
	//} else {
	//line += fmt.Sprintf("%3d", d)
	//}
	//}
	//log.Println(line)
	//}
	return dist
}

func (s State) calEval() int {
	//if s.stones[0]+s.stones[1]+s.stones[2] == 0 {
	//return s.score * 10000000
	//}

	dist2 := s.distanceFromHole('A') // Aの穴までの距離
	sumDistStoneToHole := 0          // すべての'a'の穴までの距離の合計

	// もっとも近い鉱石までの距離
	nearestStoneFromNow := 1000
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			if isStone(s.grid[index(i, j)]) {
				dist := abs(s.pos.y-i) + abs(s.pos.x-j) // なにも持っていない時は、岩の上も通れるのでマンハッタン距離
				nearestStoneFromNow = minInt(nearestStoneFromNow, dist)
				sumDistStoneToHole += dist2[index(i, j)]                                      // 穴までの手数
				sumDistStoneToHole += distance(holes[s.grid[index(i, j)]-'a'], Pos{i, j}) / 2 // 穴までのマンハッタン距離
			}
		}
	}
	// 石がないときは0にする
	if nearestStoneFromNow == 1000 {
		nearestStoneFromNow = 0
	}
	//if nearestStoneFromNow == 1000 {
	// 鉱石が@に囲まれている
	// 乱数を入れて、周りの岩を操作するようにする
	//log.Println(s.score, minStoneDist, minHoleDist, bonus, nearStone, s.pos)
	//nearestStoneFromNow = abs(nearStone.y-s.pos.y) + abs(nearStone.x-s.pos.x) + rand.Intn(5)
	//s.showGrid()
	//os.Exit(1)
	//}
	//log.Println("minStoneDist", minStoneDist, "minHoleDist", minHoleDist, "bonus", bonus)
	//log.Println("eval", s.score*10000-minStoneDist-minHoleDist+bonus-sumDist2)
	// s.scoreに40(GridSize*2)をかけることで、つぎの鉱石が遠くても穴に落とすようにする
	// 42：w
	return s.score*42 - nearestStoneFromNow - sumDistStoneToHole
}

func (s State) showGrid() {
	log.Println("Pos:", s.pos, "score:", s.score, "eval", s.eval, "have", string(s.grid[index(s.pos.y, s.pos.x)]))
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
	newState.eval = s.eval
	newState.hash = s.hash
	return newState
}

func newState(in In) *State {
	s := &State{}
	copy(s.grid[:], in.grid[:])
	for i := 0; i < 20*20; i++ {
		if in.grid[i] == 'A' {
			s.pos = Pos{i / 20, i % 20}
		}
		if in.grid[i] == 'a' || in.grid[i] == 'b' || in.grid[i] == 'c' {
			s.stones[in.grid[i]-'a']++
		}
		if in.grid[i] == 'A' || in.grid[i] == 'B' || in.grid[i] == 'C' {
			holes[in.grid[i]-'A'] = Pos{i / 20, i % 20}
		}
	}
	log.Println("start pos", s.pos)
	// Holeの上下左右を'a'にする

	//s.showGrid()
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
				s.stones[0+itemToRall-'a']--
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
	s.eval = s.calEval()
	s.hash = s.makeHash()
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

func distance(a, b Pos) int {
	return abs(a.y-b.y) + abs(a.x-b.x)
}

type Act int

const (
	Move  Act = 1
	Carry Act = 2
	Roll  Act = 3
)

type Direction int

const (
	Up    Direction = 1
	Down  Direction = 2
	Left  Direction = 3
	Right Direction = 4
)

var directionStrings = []string{"", "U", "D", "L", "R"}

var directions = []Direction{Up, Down, Left, Right}

type Action struct {
	act  Act
	dict Direction
}

func (a Action) String() string {
	return fmt.Sprintf("%d %s", a.act, directionStrings[a.dict])
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
