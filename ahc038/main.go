package main

import (
	"bytes"
	"fmt"
	"log"
	"math/bits"
	"math/rand"
	"time"
)

// rootの移動
// var direction = []string{".", "U", "D", "L", "R"}
var dy = []int{0, -1, 1, 0, 0}
var dx = []int{0, 0, 0, -1, 1}

const (
	None = iota
	Up
	Down
	Left
	Right
)

var V0Action = []byte{'.', 'U', 'D', 'L', 'R'}

type Point struct {
	Y, X int
}

const (
	CW  = 1 // "clockwise" は時計回り "R"
	CCW = 2 // "counterclockwise" は反時計回り "L"
	P   = 4 // grabs or releases a takoyaki "P"
)

var VAction = []byte{'.', 'R', 'L'}
var VAction2 = []byte{'.', '?', '?', '?', 'P'}

// rotate は中心を中心にdirection方向に回転する
func (p Point) Rotate(center Point, direction int) (np Point) {
	if direction == None {
		return p
	}
	translatedX := p.X - center.X
	translatedY := p.Y - center.Y
	var rotatedX, rotatedY int
	if direction == CW {
		rotatedX = -translatedY
		rotatedY = translatedX
	} else if direction == CCW {
		rotatedX = translatedY
		rotatedY = -translatedX
	} else {
		panic("invalid direction")
	}
	np.X = center.X + rotatedX
	np.Y = center.Y + rotatedY
	return np
}

type Node struct {
	index  int
	length int
	Point
	HasTakoyaki bool
	parent      *Node
	children    []*Node
}

func (n Node) isLeaf() bool {
	return len(n.children) == 0
}

func viewField(f BitArray) {
	var line [30][30]byte
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if f.Get(i, j) {
				line[i][j] = '1'
			} else {
				line[i][j] = '0'
			}
		}
	}
	for i := 0; i < N; i++ {
		log.Println(string(line[i][:N]))
	}
}

type State struct {
	startPos        Point
	nodes           [15]Node
	s               BitArray
	t               BitArray
	remainTakoyaki  int
	takoyakiOnField int
	takoyakiInRobot int
}

// closestTakoyaki はpに最も近いたこ焼きの座標を返す
func (s State) closestTakoyaki(p Point) (t Point) {
	minDist := 1000
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if s.s.Get(i, j) && !s.t.Get(i, j) {
				dist := abs(p.Y-i) + abs(p.X-j)
				if dist < minDist {
					minDist = dist
					t = Point{i, j}
				}
			}
		}
	}
	return t
}

func (s State) firstOutput() []byte {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%d\n", V))
	for i := 1; i < V; i++ {
		out.WriteString(fmt.Sprintf("%d %d\n", s.nodes[i].parent.index, s.nodes[i].length))
	}
	out.WriteString(fmt.Sprintf("%d %d\n", s.nodes[0].Point.X, s.nodes[0].Point.Y))
	return out.Bytes()
}

func newState() (s State) {
	for i := 0; i < 15; i++ {
		s.nodes[i].index = -1
	}
	return
}

// MoveRobot はdirection方向にnodeを移動する
// 最初はrootから呼び出す
func (s *State) MoveRobot(direction int, node *Node) {
	if node == nil {
		return
	}
	for i := 0; i < len(node.children); i++ {
		s.MoveRobot(direction, node.children[i])
	}
	node.Y += dy[direction]
	node.X += dx[direction]
}

func (s *State) RotateRobot(direction int, node *Node, center Point) {
	if node == nil {
		log.Fatal("node is nil")
	}
	if direction == None {
		return
	}
	if direction != CW && direction != CCW {
		log.Fatal("invalid direction")
	}
	for i := 0; i < len(node.children); i++ {
		s.RotateRobot(direction, node.children[i], center)
	}
	node.Point = node.Point.Rotate(center, direction)
}

// 状態
// フィールドにたこ焼きがある
//   たこ焼きを持っている
//    持っているたこ焼きを置きにいく
//      この時持っているロボットがi=1とは限らない
//   たこ焼きを持っていない
//    たこ焼きを取りに行く
//       どのアームで撮りに行くのが最適かわからない
// フィールドにたこ焼きがない
//   たこ焼きを持っている
//     たこ焼きを置きに行く
//   たこ焼きを持っていない
//     終了

// 状態評価
//  アームが４方向すべべての方向にあるとして考える？
//  どうすれば、たこ焼きを最短で取りに行けるか？

func turnSolver(s *State) []byte {
	action := make([]byte, 0, 2*V)
	// V0の移動
Reset:
	move := rand.Intn(5)
	v0Point := s.nodes[0].Point
	v0Point.Y += dy[move]
	v0Point.X += dx[move]
	if v0Point.Y < 0 || v0Point.Y >= N || v0Point.X < 0 || v0Point.X >= N {
		goto Reset
	}
	s.MoveRobot(move, &s.nodes[0])
	action = append(action, V0Action[move]) // V0 の移動
	// V1 ~
	for i := 1; i < V; i++ {
		if s.nodes[i].isLeaf() {
			center := s.nodes[i].parent.Point
			var j int
			for j = 0; j < 3; j++ {
				nextPoint := s.nodes[i].Point.Rotate(center, j)
				if !inField(nextPoint) {
					continue
				}
				// releaseできる
				if s.nodes[i].HasTakoyaki && s.t.Get(nextPoint.Y, nextPoint.X) && !s.s.Get(nextPoint.Y, nextPoint.X) {
					break
				}
				// catchできる
				if !s.nodes[i].HasTakoyaki && s.s.Get(nextPoint.Y, nextPoint.X) && !s.t.Get(nextPoint.Y, nextPoint.X) {
					break
				}
			}
			if j == 3 {
				j = 1
			}
			move = j // 0:None, 1:CW, 2:CCW
			//center := s.nodes[i].parent.Point
			s.RotateRobot(move, &s.nodes[i], center)
			action = append(action, VAction[move]) // (V-1)回転
		}
	}
	// たこ焼きをつかむor離す どちらもできるときはする
	for i := 0; i < V; i++ {
		// node is joint, V0もここ
		if !s.nodes[i].isLeaf() {
			action = append(action, '.')
			continue
		}
		// node is out of field
		if !inField(s.nodes[i].Point) {
			action = append(action, '.')
			continue
		}
		//node is leaf
		if !s.nodes[i].HasTakoyaki {
			if s.s.Get(s.nodes[i].Y, s.nodes[i].X) && !s.t.Get(s.nodes[i].Y, s.nodes[i].X) {
				//log.Println("catch takoyaki", i, s.nodes[i].Point)
				// たこ焼きをつかむ
				s.nodes[i].HasTakoyaki = true
				s.s.Unset(s.nodes[i].Y, s.nodes[i].X)
				action = append(action, 'P')
				s.takoyakiInRobot++
				s.takoyakiOnField--
			} else {
				// なにもできない
				action = append(action, '.')
			}
		} else {
			if s.t.Get(s.nodes[i].Y, s.nodes[i].X) && !s.s.Get(s.nodes[i].Y, s.nodes[i].X) {
				// たこ焼きを離す
				s.nodes[i].HasTakoyaki = false
				s.t.Unset(s.nodes[i].Y, s.nodes[i].X)
				s.remainTakoyaki--
				action = append(action, 'P')
				s.takoyakiInRobot--
			} else {
				// なにもできない
				action = append(action, '.')
			}
		}
	}
	//log.Println(len(action), string(action))
	action = append(action, '\n')
	return action
}

func solver(in Input) {
	iterations := 0
	var minOut []byte
	for elapsed := time.Since(startTime); elapsed < timeLimit; elapsed = time.Since(startTime) {
		iterations++
		state := newState()
		for i := 0; i < in.N; i++ {
			for j := 0; j < in.N; j++ {
				if in.s[i][j] == '1' { // 1: たこ焼きあり
					state.s.Set(i, j)
				}
				if in.t[i][j] == '1' {
					state.t.Set(i, j)
				}
			}
		}
		// 初期化
		state.startPos.Y = N / 2
		state.startPos.X = N / 2
		state.remainTakoyaki = M
		state.takoyakiOnField = M
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				if state.s.Get(i, j) && state.t.Get(i, j) {
					state.remainTakoyaki--
					state.takoyakiOnField--
				}
			}
		}
		for i := 0; i < V; i++ {
			state.nodes[i].index = i
			state.nodes[i].length = rand.Intn(N)/2 + 1
			state.nodes[i].HasTakoyaki = false
			if i == 0 {
				state.nodes[i].Point = state.startPos
			} else {
				state.nodes[i].parent = &state.nodes[0]
				p := state.nodes[i].parent
				p.children = append(p.children, &state.nodes[i])
				state.nodes[i].Point.Y = state.nodes[i].parent.Point.Y
				state.nodes[i].Point.X = state.nodes[i].parent.Point.X + state.nodes[i].length
			}
		}
		//	for i := 0; i < V; i++ {
		//log.Printf("%+v\n", state.nodes[i])
		//}
		// 初期出力
		out := state.firstOutput()
		// シミュレーション
		//pre := state.remainTakoyaki
		//preTurn := 0
		for i := 0; i < 50000; i++ {
			tout := turnSolver(&state)
			out = append(out, tout...)
			if state.remainTakoyaki == 0 {
				log.Printf("finish turn=%d\n", i)
				break
			}
			//if pre != state.remainTakoyaki {
			//log.Printf("%d remain:%d(%d %d) turn:%d\n", i, state.remainTakoyaki, state.takoyakiOnField, state.takoyakiInRobot, i-preTurn)
			//pre = state.remainTakoyaki
			//}
			if minOut != nil && len(out) > len(minOut) {
				break
			}
		}
		if minOut == nil || len(out) < len(minOut) {
			minOut = out
		}
	}
	fmt.Print(string(minOut))
	log.Println(len(minOut))
	log.Printf("iter=%d\n", iterations)
}

var N, M, V int

func inField(p Point) bool {
	return 0 <= p.Y && p.Y < N && 0 <= p.X && p.X < N
}

type Input struct {
	N, M, V int
	s       [30]string
	t       [30]string
}

func readInput() Input {
	var input Input
	fmt.Scan(&input.N, &input.M, &input.V)
	for i := 0; i < input.N; i++ {
		fmt.Scan(&input.s[i])
	}
	for i := 0; i < input.N; i++ {
		fmt.Scan(&input.t[i])
	}
	N = input.N
	M = input.M
	V = input.V
	return input
}

var startTime time.Time
var timeLimit time.Duration = 2500 * time.Millisecond

func main() {
	log.SetFlags(log.Lshortfile)
	rand.Seed(1)
	startTime = time.Now()
	in := readInput()
	log.Printf("N=%d, M=%d, V=%d\n", in.N, in.M, in.V)
	solver(in)
	elapse := time.Since(startTime)
	log.Printf("time=%v\n", elapse.Seconds())
}

// ------------------------------------------------------------------
// util
// bitArrayを管理するためのセット
const widthBits = 30
const arraySize = 30 * 30
const uint64Size = 64

type BitArray [arraySize]uint64

func (b *BitArray) Set(y, x int) {
	index := y*widthBits + x
	b[index/uint64Size] |= 1 << (index % uint64Size)
}

func (b *BitArray) Unset(y, x int) {
	index := y*widthBits + x
	b[index/uint64Size] &= ^(1 << (index % uint64Size))
}

func (b *BitArray) Get(y, x int) bool {
	index := y*widthBits + x
	return b[index/uint64Size]&(1<<(index%uint64Size)) != 0
}

func (b BitArray) PopCount() (count int) {
	for i := 0; i < arraySize; i++ {
		count += bits.OnesCount64(b[i])
	}
	return count
}

func (b BitArray) XorPopCount(a BitArray) (count int) {
	for i := 0; i < arraySize; i++ {
		count += bits.OnesCount64(b[i] ^ a[i])
	}
	return count
}

func (b *BitArray) Reset() {
	for i := 0; i < arraySize; i++ {
		b[i] = 0
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
