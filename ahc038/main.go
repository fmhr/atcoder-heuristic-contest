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

var moveAction = []byte{'.', 'U', 'R', 'D', 'L'}

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
	if direction == FLIP {
		p = p.Rotate(center, CW)
		return p.Rotate(center, CW)
	}
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

func (s *State) moveLeaf(node *Node, m byte) {
	if !(m == 'P' || m == '.') {
		panic("invalid move")
	}
	if node.isLeaf() {
		if m == 'P' {
			if !node.HasTakoyaki {
				// Catch
				node.HasTakoyaki = true
				s.s.Unset(node.Y, node.X)
				s.takoyakiPos = deleteItem(s.takoyakiPos, node.Point)
				s.takoyakiInRobot++
				s.takoyakiOnField--
			} else {
				// Put
				node.HasTakoyaki = false
				s.t.Unset(node.Y, node.X)
				s.targetPos = deleteItem(s.targetPos, node.Point)
				s.takoyakiInRobot--
				s.remainTakoyaki--
			}
		}
	}
	if !node.isLeaf() && m == 'P' {
		panic("invalid move")
	}
}

func (src State) Clone() (clone State) {
	clone.startPos = src.startPos
	clone.nodes = src.nodes
	clone.s = src.s
	clone.t = src.t
	clone.remainTakoyaki = src.remainTakoyaki
	clone.takoyakiOnField = src.takoyakiOnField
	clone.takoyakiInRobot = src.takoyakiInRobot

	clone.relatevePositions = [15][]Point{}
	for i, ps := range src.relatevePositions {
		if ps != nil {
			clone.relatevePositions[i] = make([]Point, len(ps))
			copy(clone.relatevePositions[i], ps)
		}
	}

	clone.takoyakiPos = make([]Point, len(src.takoyakiPos))
	copy(clone.takoyakiPos, src.takoyakiPos)

	clone.targetPos = make([]Point, len(src.targetPos))
	copy(clone.targetPos, src.targetPos)

	return clone
}

func (s *State) ResetState() {
	s.startPos = Point{0, 0}
	for i := 0; i < 15; i++ {
		s.nodes[i] = Node{}
	}
	s.s.Reset()
	s.t.Reset()
	s.remainTakoyaki = 0
	s.takoyakiOnField = 0
	s.takoyakiInRobot = 0
	for i := 0; i < 15; i++ {
		s.relatevePositions[i] = nil
	}
	s.takoyakiPos = nil
	s.targetPos = nil
}

func (s State) infoLength() {
	length := make([]int, V)
	for i := 0; i < V; i++ {
		length[i] = s.nodes[i].length
	}
}

func (s State) closestPoint(p Point, pp []Point) (t Point) {
	minDist := 1000
	log.Println("root", s.nodes[0].Point, "p", p, s.nodes[1].HasTakoyaki)
	for i := 0; i < len(pp); i++ {
		root := s.nodes[0].Point
		target := s.targetPos[i]
		root.Y += target.Y - p.Y
		root.X += target.X - p.X
		if inField(root) {
			dist := abs(p.Y-target.Y) + abs(p.X-target.X)
			if dist < minDist {
				minDist = dist
				t.Y, t.X = target.Y, target.X
				log.Println(i, "New closest target:", t, "root", root, "dis", dist) // 追加
			}
		}
	}
	if minDist == 1000 {
		panic("no target")
	}
	log.Println("closest target:", t)
	return
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

func (s State) closetTakoyakiRenge(v int, target *Point) (direction, miniD int) {
	if !inField(s.nodes[0].Point) {
		log.Fatal("root is out of field", s.nodes[0].Point)
	}
	nextPos := s.nodes[v].Point // 回転後の位置 これをもとに移動方向を決める
	// targetが範囲外（初期化)または、たこ焼きも目的地もない場合、次のターゲットを探す
	if !inField(s.nodes[v].Point) || !(s.s.Get(target.Y, target.X) || s.t.Get(target.Y, target.X)) {
		miniD = 1000
		// FLIPを使わない理由:
		//  robot全体が動き続けることで、他のノードが行動できる可能性があがる(?)
		n := &s.nodes[v]
		for i := 0; i < 4; i++ {
			root := s.nodes[0].Point
			RotateRobot(i, n, s.nodes[0].Point)
			var t Point
			if !n.HasTakoyaki {
				t = s.closestPoint(n.Point, s.takoyakiPos)
			} else {
				t = s.closestPoint(n.Point, s.targetPos)
			}
			root.Y += t.Y - n.Point.Y
			root.X += t.X - n.Point.X
			if inField(root) {
				dis := DistancePP(n.Point, t)
				if i == 3 {
					dis++
				}
				if dis < miniD {
					miniD = dis
					*target = t
					nextPos = s.nodes[v].Point
				}
			}
			ReverseRobot(i, n, s.nodes[0].Point)
		}
		log.Println("target update:", *target)
		log.Println("node[0]", s.nodes[0].Point, "node[v]", s.nodes[v].Point, "nextPos", nextPos, "target", *target)
	} else {
		log.Println("target keep:", *target)
	}
	direction = DirectionPP(nextPos, *target)
	return
}

// calcMoveDirection は最適な移動方向を計算する
// v1がなにももっていないとき
// v1 の位置から最も近いたこ焼きの位置最小にする
// v1がたこ焼きを持っているとき
// v1の位置から最も近い設定位置を最小にする
// vはターゲットを探す指先
func (s State) calcMoveDirection(target *Point) (direction, v int) {
	log.Println("target", *target)
	v = 1
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
		direct, length := s.closetTakoyakiRenge(v, target)
		if miniD > length {
			miniD = length
			direction = direct
		}
		v++
		if miniD == 0 {
			break
		}
	}
	if miniD == 1000 {
		panic("no target")
	}
	return direction, v
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

// RotateRobot はcenterを中心にdirection方向にnodeを回転する
func RotateRobot(direction int, node *Node, center Point) {
	if node == nil {
		log.Fatal("node is nil")
	}
	if direction == None {
		return
	}
	for i := 0; i < len(node.children); i++ {
		//RotateRobot(direction, node.children[i], center)
		node.children[i].Point = node.children[i].Point.Rotate(center, direction)
	}
	node.Point = node.Point.Rotate(center, direction)
	switch direction {
	case CW:
		node.direction = (node.direction+1-1)%4 + 1
	case CCW:
		node.direction = (node.direction+3-1)%4 + 1
	case FLIP:
		node.direction = (node.direction+1-1)%4 + 1
		node.direction = (node.direction+1-1)%4 + 1
	}
}

// RotateRobot2 はdirection	が byteできた時に対応する
func RotateRobot2(direction byte, node *Node, center Point) {
	dir := 0
	switch direction {
	case 'R':
		dir = CW
	case 'L':
		dir = CCW
	case '.':
		dir = None
	}
	RotateRobot(dir, node, center)
}

// ReverseRobot はcenterを中心に-direction方向にnodeを回転する
func ReverseRobot(direction int, node *Node, center Point) {
	switch direction {
	case CW:
		direction = CCW
	case CCW:
		direction = CW
	}
	// FLIPの場合はそのまま１80度回転
	RotateRobot(direction, node, center)
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

func turnSolver(s *State, target *Point) []byte {
	action := make([]byte, 0, 2*V)
	// V0の移動
	move, v := s.calcMoveDirection(target)
	_ = v
	s.MoveRobot(move, &s.nodes[0])
	if !inField(s.nodes[0].Point) {
		log.Println("root:", s.nodes[0].Point, "[0]:", s.nodes[1].Point, "target:", *target)
		log.Fatal("root is out of field", s.nodes[0].Point, move)
	}

	action = append(action, moveAction[move]) // V0 の移動
	// V1 ~
	subAction := make([]byte, V-1)
	takoAction := make([]byte, V)
	takoAction[0] = '.'
	nodeLocked := make([]bool, V)
	nodeLocked[0] = true
	log.Printf("%+v %+v\n", s.nodes[0].Point, s.nodes[1])
	for i := 1; i < V; i++ {
		if s.nodes[i].parent == &s.nodes[0] {
			nodes := make([]*Node, 0, 4)
			sub := make([]*Node, 0, 4)
			sub = append(sub, &s.nodes[i])
			for len(sub) > 0 {
				n := sub[0]
				sub = sub[1:]
				sub = append(sub, n.children...)
				nodes = append(nodes, n)
			}
			// 今回操作するノードの確定
			// log.Println(nodes)
			bestRotate := make([]int, len(nodes))
			bestP := make([]byte, len(nodes))
			bestTakoPoint := -1
			bestInFieldCnt := -1
			bestUnsetTakoyaki := make([]Point, 0)
			bestUnsetTarget := make([]Point, 0)
			// 回転の組み合わせ
			totalCombinations := 1
			for j := 0; j < len(nodes); j++ {
				totalCombinations *= 3
			}
			for j := 0; j < totalCombinations; j++ {
				takoyakiUnsetLog := make([]Point, 0)
				targetUnsetLog := make([]Point, 0)
				//subRotate := make([]byte, len(nodes))
				subP := make([]byte, len(nodes))
				comb := make([]int, len(nodes))
				num := j
				for k := 0; k < len(nodes); k++ {
					comb[k] = num % 3
					num /= 3
				}
				for k := 0; k < len(nodes); k++ {
					RotateRobot(comb[k], nodes[k], nodes[k].parent.Point)
				}
				// ここで評価
				takoPoint := 0
				inFieldCnt := 0
				for k := 0; k < len(nodes); k++ {
					// 先端かつ、フィールド内
					if nodes[k].isLeaf() && inField(nodes[k].Point) {
						if i == 1 {
							log.Println("i=1", nodes[k].Point, nodes[k].HasTakoyaki, s.s.Get(nodes[k].Y, nodes[k].X), s.t.Get(nodes[k].Y, nodes[k].X))
						}
						inFieldCnt++
						if !nodes[k].HasTakoyaki && s.s.Get(nodes[k].Y, nodes[k].X) {
							// GetTakoyaki
							s.s.Unset(nodes[k].Y, nodes[k].X)
							takoyakiUnsetLog = append(takoyakiUnsetLog, nodes[k].Point)
							takoPoint++
							subP[k] = 'P'
							log.Println("GetTakoyaki", nodes[k].index, nodes[k].Point)
						} else if nodes[k].HasTakoyaki && s.t.Get(nodes[k].Y, nodes[k].X) {
							// ReleaseTakoyaki
							s.t.Unset(nodes[k].Y, nodes[k].X)
							targetUnsetLog = append(targetUnsetLog, nodes[k].Point)
							takoPoint++
							subP[k] = 'P'
							log.Println("ReleaseTakoyaki", nodes[k].index, nodes[k].Point)
						}
					}
				}
				//log.Println(comb, takoPoint, inFieldCnt)
				// Undo
				for k := 0; k < len(nodes); k++ {
					ReverseRobot(comb[k], nodes[k], nodes[k].parent.Point)
				}
				for k := 0; k < len(takoyakiUnsetLog); k++ {
					s.s.Set(takoyakiUnsetLog[k].Y, takoyakiUnsetLog[k].X)
				}
				for k := 0; k < len(targetUnsetLog); k++ {
					s.t.Set(targetUnsetLog[k].Y, targetUnsetLog[k].X)
				}
				// Update
				if takoPoint > bestTakoPoint || (takoPoint == bestTakoPoint && inFieldCnt > bestInFieldCnt) {
					bestTakoPoint = takoPoint
					bestInFieldCnt = inFieldCnt
					copy(bestRotate, comb)
					copy(bestP, subP)
					bestUnsetTakoyaki = make([]Point, len(takoyakiUnsetLog))
					copy(bestUnsetTakoyaki, takoyakiUnsetLog)
					bestUnsetTarget = make([]Point, len(targetUnsetLog))
					copy(bestUnsetTarget, targetUnsetLog)
				}
			}
			// Update best to true
			for j := 0; j < len(nodes); j++ {
				if bestP[j] == 0 {
					bestP[j] = '.'
				}
			}
			for j := 0; j < len(nodes); j++ {
				subAction[nodes[j].index-1] = VAction[bestRotate[j]]
				takoAction[nodes[j].index] = bestP[j]
			}
			for j := 0; j < len(bestUnsetTakoyaki); j++ {
				s.takoyakiPos = deleteItem(s.takoyakiPos, bestUnsetTakoyaki[j])
				s.s.Unset(bestUnsetTakoyaki[j].Y, bestUnsetTakoyaki[j].X)
			}
			for j := 0; j < len(bestUnsetTarget); j++ {
				s.targetPos = deleteItem(s.targetPos, bestUnsetTarget[j])
				s.t.Unset(bestUnsetTarget[j].Y, bestUnsetTarget[j].X)
			}
		}
	}
	// 適応
	// V0の移動は適応済み
	for j := 0; j < V-1; j++ {
		RotateRobot2(subAction[j], &s.nodes[j+1], s.nodes[j+1].parent.Point)
	}
	for j := 0; j < V; j++ {
		s.moveLeaf(&s.nodes[j], takoAction[j])
	}

	action = append(action, subAction...)
	action = append(action, takoAction...)
	action = append(action, '\n')
	return action
}

func solver(in Input) {
	takoyakiPos := make([]Point, 0, 45)
	targetPos := make([]Point, 0, 45)
	for i := 0; i < in.N; i++ {
		for j := 0; j < in.N; j++ {
			if in.s[i][j] == '1' {
				takoyakiPos = append(takoyakiPos, Point{i, j})
			}
			if in.t[i][j] == '1' {
				targetPos = append(targetPos, Point{i, j})
			}
		}
	}
	takoyakiMean := meanPoints(takoyakiPos)
	targetMean := meanPoints(targetPos)
	log.Println(takoyakiMean, targetMean)
	var meanPoints [2]Point = [2]Point{{takoyakiMean.Y, targetMean.X}, {takoyakiMean.Y, targetMean.X}}
	_ = meanPoints
	iterations := 0
	var minOut []byte
	for elapsed := time.Since(startTime); elapsed < timeLimit; elapsed = time.Since(startTime) {
		iterations++
		//	if iterations == 2000 {
		//		break
		//	}
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
		log.Println(state.s.Get(8, 12), state.t.Get(8, 12))
		//viewField(state.s)
		//log.Println("----")
		//viewField(state.t)
		////log.Println(state.takoyakiPos)

		// 初期化
		//state.startPos = Point{0, 0} // デバッグ用
		state.startPos.Y = rand.Intn(N)
		state.startPos.X = rand.Intn(N)
		//state.startPos = meanPoints[iterations%2]
		//state.startPos.Y += rand.Intn(10) - 5
		//state.startPos.X += rand.Intn(10) - 5
		//state.startPos.Y = min(max(0, state.startPos.Y), N-1)
		//state.startPos.X = min(max(0, state.startPos.X), N-1)
		for i := 0; i < V; i++ {
			state.nodes[i].index = i
			if i != 0 {
				state.nodes[i].length = rand.Intn(N/2) + N/6
			}
			state.nodes[i].HasTakoyaki = false
			if i == 0 {
				state.nodes[i].Point = state.startPos
			} else {
				if i == 2 || i == 3 {
					state.nodes[i].length = state.nodes[i].length * 2 / 3
				}
				if i == 3 {
					state.nodes[i].parent = &state.nodes[2]
				} else {
					state.nodes[i].parent = &state.nodes[0] // root
				}
				p := state.nodes[i].parent
				p.children = append(p.children, &state.nodes[i])
				state.nodes[i].Point.Y = state.nodes[i].parent.Point.Y
				state.nodes[i].Point.X = state.nodes[i].parent.Point.X + state.nodes[i].length
				state.nodes[i].direction = Right // 親から見て右に位置する
			}
		}
		state.calcRelatevePosition()
		// 初期出力
		out := state.firstOutput()
		// シミュレーション
		target := &Point{-1, -1}
		for i := 0; i < 50; i++ {
			tout := turnSolver(&state, target)
			out = append(out, tout...)
			if state.remainTakoyaki == 0 {
				log.Printf("finish turn=%d\n", i)
				state.infoLength()
				break
			}
			if minOut != nil && len(out) > len(minOut) {
				break
			}
			log.Println(i, state.remainTakoyaki, string(tout[:V]), string(tout[V:len(tout)-1]), target, DistancePP(state.nodes[1].Point, *target))
		}
		if minOut == nil || len(out) < len(minOut) {
			minOut = out
		}
		break // 1回だけ デバッグ
	}
	fmt.Print(string(minOut))
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
const uint64Size = 64
const widthBits = 30
const heightBits = 30
const arraySize = (widthBits*heightBits*uint64Size - 1) / uint64Size

type BitArray [arraySize]uint64

func (b *BitArray) Set(y, x int) {
	if y < 0 || y >= heightBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
	index := y*widthBits + x
	b[index/uint64Size] |= 1 << (index % uint64Size)
}

func (b *BitArray) Unset(y, x int) {
	if y < 0 || y >= heightBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
	index := y*widthBits + x
	b[index/uint64Size] &= ^(1 << (index % uint64Size))
}

func (b *BitArray) Get(y, x int) bool {
	if y < 0 || y >= heightBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
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
