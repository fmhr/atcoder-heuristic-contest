package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

// rootの移動
// var direction = []string{".", "U", "D", "L", "R"}
var dy = []int{0, -1, 0, 1, 0}
var dx = []int{0, 0, 1, 0, -1}

const (
	None = iota
	Up
	Right
	Down
	Left
)

var DirectionDict []string = []string{"None", "Up", "Right", "Down", "Left"}

var V0Action = []byte{'.', 'U', 'R', 'D', 'L'}

type Point struct {
	Y, X int
}

func meanPoints(ps []Point) Point {
	var sumY, sumX int
	for i := range ps {
		sumY += ps[i].Y
		sumX += ps[i].X
	}
	return Point{sumY / len(ps), sumX / len(ps)}
}

func DistancePP(p1, p2 Point) int {
	return abs(p1.Y-p2.Y) + abs(p1.X-p2.X)
}

// DirectionPP はp1からp2への方向を返す
// 0:None, 1:Up, 2:Down, 3:Left, 4:Right の優先度をもつ
func DirectionPP(p1, p2 Point) int {
	if p1.X == p2.X && p1.Y == p2.Y {
		return None
	}
	if p1.Y == p2.Y {
		if p1.X < p2.X {
			return Right
		}
		return Left
	}
	if p1.Y < p2.Y {
		return Down
	}
	return Up
}

// deleteIndex はaのi番目の要素を削除する
// 順番を考慮しない
func deleteIndex(a []Point, i int) []Point {
	a[i] = a[len(a)-1]
	return a[:len(a)-1]
}

func deleteItem(a []Point, item Point) []Point {
	for i, v := range a {
		if v.Y == item.Y && v.X == item.X {
			return deleteIndex(a, i)
		}
	}
	return a
}

const (
	CW   = 1 // "clockwise" は時計回り "R"
	CCW  = 2 // "counterclockwise" は反時計回り "L"
	FLIP = 3 // "flip" は180度回転 "RR", "LL"
	P    = 4 // grabs or releases a takoyaki "P"
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
		panic("invalid direction in rotate(0, 1, 2) but got " + fmt.Sprint(direction))
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
	direction   int
}

func (n Node) isLeaf() bool {
	return len(n.children) == 0
}

func (n Node) root() *Node {
	if n.parent == nil {
		return &n
	}
	return n.parent.root()
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
	startPos          Point
	nodes             [15]Node
	s                 BitArray
	t                 BitArray
	remainTakoyaki    int
	takoyakiOnField   int
	takoyakiInRobot   int
	relatevePositions [15][]Point // この値をrootからの相対位置
	takoyakiPos       []Point
	targetPos         []Point
}

func (s State) infoLength() {
	length := make([]int, V)
	for i := 0; i < V; i++ {
		length[i] = s.nodes[i].length
	}
	//log.Println(length)
}

// closestTakoyaki はpに最も近いたこ焼きの座標を返す
func (s State) closestTakoyaki(p Point) (t Point) {
	minDist := 1000
	for ti := range s.takoyakiPos {
		i, j := s.takoyakiPos[ti].Y, s.takoyakiPos[ti].X
		dist := abs(p.Y-i) + abs(p.X-j)
		if dist < minDist {
			minDist = dist
			t = Point{i, j}
		}
	}
	return t
}

// closestTakoyaki はpに最も近いたこ焼きの座標を返す
func (s State) closestTarget(p Point) (t Point) {
	minDist := 1000
	for ti := range s.targetPos {
		i, j := s.targetPos[ti].Y, s.targetPos[ti].X
		dist := abs(p.Y-i) + abs(p.X-j)
		if dist < minDist {
			minDist = dist
			t = Point{i, j}
		}
	}
	return t
}

// countMatchingTakoyaki はpのx座標またはy座標が一致するたこ焼きの数を返す
// 一致するたこ焼きがない場合、最小移動回数を返す
func (s State) countMatchingTakoyaki(p Point) (count int) {
	minMove := 0
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if s.s.Get(i, j) && !s.t.Get(i, j) { // たこ焼きがあるがターゲットではない
				if p.Y == i || p.X == j {
					count++
				} else {
					m := min(abs(p.Y-i), abs(p.X-j)) // どちらかの座標が一致するまでの最小移動回数
					minMove = min(minMove, m)
				}
			}
		}
	}
	if count == 0 {
		return minMove
	}
	return count
}

// countMatchingTakoyakiTarget はpのx座標またはy座標が一致するターゲットの数を返す
// 一致するターゲットがない場合、最小移動回数を返す
func (s State) countMatchingTakoyakiTarget(p Point) (count int) {
	minMove := 0
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if !s.s.Get(i, j) && s.t.Get(i, j) { // ターゲットかつたこ焼きがない
				if p.Y == i || p.X == j {
					count++
				} else {
					m := min(abs(p.Y-i), abs(p.X-j))
					minMove = min(minMove, m)
				}
			}
		}
	}
	if count == 0 {
		return minMove
	}
	return count
}

// ロボットアームの指先が取りうる位置を計算する
func (s *State) calcRelatevePosition() {
	for i := 0; i < V; i++ {
		if s.nodes[i].parent == nil { // root
			s.relatevePositions[i] = append(s.relatevePositions[i], Point{0, 0})
			continue
		}
		for _, center := range s.relatevePositions[s.nodes[i].parent.index] {
			for d := 1; d < 5; d++ {
				var nextPoint Point
				nextPoint.Y = center.Y + dy[d]*s.nodes[i].length
				nextPoint.X = center.X + dx[d]*s.nodes[i].length
				s.relatevePositions[i] = append(s.relatevePositions[i], nextPoint)
			}
		}
	}
}

func (s State) closetTakoyakiRenge(v int) (direction, miniD int) {
	var p2 Point
	miniD = 1000
	for _, p := range s.relatevePositions[v] {
		root := s.nodes[0].Point
		p2.Y = root.Y + p.Y
		p2.X = root.X + p.X
		var t Point
		if s.nodes[v].HasTakoyaki {
			t = s.closestTarget(p2)
		} else {
			t = s.closestTakoyaki(p2)
		}
		// このとき、rootが範囲外にあってはいけない
		dy := t.Y - p2.Y
		dx := t.X - p2.X
		root.Y += dy
		root.X += dx
		if !inField(root) {
			continue
		}
		d := DistancePP(p2, t)
		if d < miniD {
			direction = DirectionPP(p2, t)
			miniD = d
		}
	}
	return direction, miniD
}

// calcMoveDirection は最適な移動方向を計算する
// v1がなにももっていないとき
// v1 の位置から最も近いたこ焼きの位置最小にする
// v1がたこ焼きを持っているとき
// v1の位置から最も近い設定位置を最小にする
func (s State) calcMoveDirection() (direction int) {
	v := 1
	// フィールドにたこ焼きがすでにない、たこ焼きを持っている指先がv1以外の時
	if (s.takoyakiOnField == 0 && !s.nodes[v].HasTakoyaki) || !s.nodes[v].isLeaf() {
		for !s.nodes[v].HasTakoyaki {
			v++
		}
	}
	miniD := 1000
	v2 := v + 1 // TODO ここをVまでのばせるように高速化する
	v2 = min(v2, V)
	for v < v2 {
		direct, length := s.closetTakoyakiRenge(v)
		if miniD > length {
			miniD = length
			direction = direct
		}
		v++
		if miniD == 0 {
			break
		}
	}
	//log.Println(s.relatevePositions[1])
	if miniD == 1000 {
		return None
	}
	return direction
}

func (s State) firstOutput() []byte {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%d\n", V))
	for i := 1; i < V; i++ {
		out.WriteString(fmt.Sprintf("%d %d\n", s.nodes[i].parent.index, s.nodes[i].length))
	}
	out.WriteString(fmt.Sprintf("%d %d\n", s.nodes[0].Point.Y, s.nodes[0].Point.X))
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

// rootの位置に評価をつける
//  x,またはyの位置が一致しているたこ焼きの数
//  １つも一致したいない場合、何回移動すれば一致するか

func (s State) moveRandom() int {
Reset:
	move := rand.Intn(5)
	v0Point := s.nodes[0].Point
	v0Point.Y += dy[move]
	v0Point.X += dx[move]
	if v0Point.Y < 0 || v0Point.Y >= N || v0Point.X < 0 || v0Point.X >= N {
		goto Reset
	}
	return move
}

func turnSolver(s *State) []byte {
	action := make([]byte, 0, 2*V)
	// V0の移動
	//move := s.moveRandom()
	move := s.calcMoveDirection()
	s.MoveRobot(move, &s.nodes[0])
	action = append(action, V0Action[move]) // V0 の移動
	// V1 ~
	subAction := make([]byte, V-1)
	takoAction := make([]byte, V)
	takoAction[0] = '.'
	for i := 1; i < V; i++ {
		takoAction[i] = '.'
		if s.nodes[i].isLeaf() {
			center := s.nodes[i].parent.Point
			var j int
			var cwOutField, ccwOutField bool
			for j = 0; j < 4; j++ {
				var nextPoint Point
				if j == 3 {
					// 180度回転
					nextPoint = s.nodes[i].Point.Rotate(center, CW)
					nextPoint = nextPoint.Rotate(center, CW)
				} else {
					nextPoint = s.nodes[i].Point.Rotate(center, j)
				}
				if !inField(nextPoint) {
					if j == CW {
						cwOutField = true
					} else if j == CCW {
						ccwOutField = true
					}
					continue
				}
				// releaseできる
				if s.nodes[i].HasTakoyaki && s.t.Get(nextPoint.Y, nextPoint.X) && !s.s.Get(nextPoint.Y, nextPoint.X) {
					takoAction[i] = 'P'
					break
				}
				// catchできる
				if !s.nodes[i].HasTakoyaki && s.s.Get(nextPoint.Y, nextPoint.X) && !s.t.Get(nextPoint.Y, nextPoint.X) {
					takoAction[i] = 'P'
					break
				}
			}
			if j == 3 {
				// 180度回転
				if cwOutField {
					j = CCW
				} else {
					j = CW
				}
			}
			if j == 4 {
				// なにもない
				if inField(s.nodes[i].Point) && !cwOutField && !ccwOutField {
					j = 0
				} else if cwOutField && !ccwOutField {
					j = CCW
				} else if !cwOutField && ccwOutField {
					j = CW
				} else {
					j = 0
				}
			}
			move = j // 0:None, 1:CW, 2:CCW
			//center := s.nodes[i].parent.Point
			s.RotateRobot(move, &s.nodes[i], center)
			subAction[i-1] = VAction[move]
		} else {
			// not leaf
			subAction[i-1] = '.'
		}
	}
	action = append(action, subAction...)
	subAction2 := make([]byte, V)
	// たこ焼きをつかむor離す どちらもできるときはする
	for i := 0; i < V; i++ {
		subAction2[i] = '.'
		// node is joint, V0もここ
		if !s.nodes[i].isLeaf() {
			continue
		}
		// node is out of field
		if !inField(s.nodes[i].Point) {
			continue
		}
		//node is leaf
		if !s.nodes[i].HasTakoyaki {
			if s.s.Get(s.nodes[i].Y, s.nodes[i].X) && !s.t.Get(s.nodes[i].Y, s.nodes[i].X) {
				//log.Println("catch takoyaki", i, s.nodes[i].Point)
				// たこ焼きをつかむ
				s.nodes[i].HasTakoyaki = true
				s.s.Unset(s.nodes[i].Y, s.nodes[i].X)
				subAction2[i] = 'P'
				s.takoyakiInRobot++
				s.takoyakiOnField--
				s.takoyakiPos = deleteItem(s.takoyakiPos, s.nodes[i].Point)
			} else {
				// なにもできない
			}
		} else {
			if inField(s.nodes[i].Point) && s.t.Get(s.nodes[i].Y, s.nodes[i].X) && !s.s.Get(s.nodes[i].Y, s.nodes[i].X) {
				// たこ焼きを離す
				s.nodes[i].HasTakoyaki = false
				s.t.Unset(s.nodes[i].Y, s.nodes[i].X)
				s.remainTakoyaki--
				subAction2[i] = 'P'
				s.takoyakiInRobot--
				s.targetPos = deleteItem(s.targetPos, s.nodes[i].Point)
			} else {
				// なにもできない
			}
		}
	}
	action = append(action, subAction2...)
	//log.Println(len(action), string(action))
	action = append(action, '\n')
	//log.Println(string(action))
	//log.Printf("%+v\n", s.nodes[0])
	//log.Printf("%+v\n", s.nodes[1])
	return action
}

func solver(in Input) {
	iterations := 0
	var minOut []byte
	for elapsed := time.Since(startTime); elapsed < timeLimit; elapsed = time.Since(startTime) {
		iterations++
		if iterations == 2000 {
			break
		}
		state := newState()
		cnt := 0
		for i := 0; i < in.N; i++ {
			for j := 0; j < in.N; j++ {
				if in.s[i][j] == '1' && in.t[i][j] == '1' {
					continue
				} else if in.s[i][j] == '1' { // 1: たこ焼きあり
					state.s.Set(i, j)
					state.takoyakiPos = append(state.takoyakiPos, Point{i, j})
					state.remainTakoyaki++
					state.takoyakiOnField++
					cnt++
				} else if in.t[i][j] == '1' {
					state.t.Set(i, j)
					state.targetPos = append(state.targetPos, Point{i, j})
				}
			}
		}
		// 初期化
		state.startPos.Y = rand.Intn(N)
		state.startPos.X = rand.Intn(N)
		for i := 0; i < V; i++ {
			state.nodes[i].index = i
			if i != 0 {
				state.nodes[i].length = rand.Intn(N/2) + N/6
			}
			state.nodes[i].HasTakoyaki = false
			if i == 0 {
				state.nodes[i].Point = state.startPos
			} else {
				state.nodes[i].parent = &state.nodes[0] // root
				p := state.nodes[i].parent
				p.children = append(p.children, &state.nodes[i])
				state.nodes[i].Point.Y = state.nodes[i].parent.Point.Y
				state.nodes[i].Point.X = state.nodes[i].parent.Point.X + state.nodes[i].length
			}
		}
		//log.Println(state.nodes[0].Point)
		state.calcRelatevePosition()
		//for i := 0; i < V; i++ {
		//	log.Printf("%d %d %+v\n", i, state.nodes[i].length, state.relatevePositions[i])
		//}
		//os.Exit(0)
		//for i := 0; i < V; i++ {
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
				state.infoLength()
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
	turn := len(strings.Split(string(minOut), "\n")) - V - 1 - 1
	log.Printf("turn=%d\n", turn)
	log.Printf("per=%f\n", float64(turn)*math.Sqrt(float64(V))/float64(trueM))
}

var N, M, V int
var trueM int

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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		log.SetOutput(io.Discard)
	}
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	rand.Seed(1)
	startTime = time.Now()
	in := readInput()
	trueM = in.M
	for i := 0; i < in.N; i++ {
		for j := 0; j < in.N; j++ {
			if in.s[i][j] == '1' && in.t[i][j] == '1' {
				trueM--
			}
		}
	}

	log.Printf("N=%d, M=%d, trueM=%d V=%d\n", in.N, in.M, trueM, in.V)
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
	if y < 0 || y >= widthBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
