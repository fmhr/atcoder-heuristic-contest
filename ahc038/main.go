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
	if direction == 3 {
		p.Rotate(center, CW)
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
				node.HasTakoyaki = true
				s.s.Unset(node.Y, node.X)
				s.takoyakiPos = deleteItem(s.takoyakiPos, node.Point)
				s.takoyakiInRobot++
				s.takoyakiOnField--
			} else {
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

// closestTakoyaki はpに最も近いたこ焼きの座標を返す
func (s State) closestTakoyaki(p Point) (t Point) {
	minDist := 1000
	for i := 0; i < len(s.takoyakiPos); i++ {
		root := s.nodes[0].Point
		target := s.takoyakiPos[i]
		log.Println(root, target, p)
		root.Y += target.Y - p.Y
		root.X += target.X - p.X
		if inField(root) {
			dist := abs(p.Y-target.Y) + abs(p.X-target.X)
			if dist < minDist {
				log.Println(dist)
				minDist = dist
				t.Y, t.X = target.Y, target.X
				if dist == 0 {
					return
				}
			}
		}
	}
	if minDist == 1000 {
		panic("no takoyaki")
	}
	return t
}

// closestTakoyaki はpに最も近いたこ焼きの座標を返す
func (s State) closestTarget(p Point) (t Point) {
	minDist := 1000
	for i := 0; i < len(s.targetPos); i++ {
		root := s.nodes[0].Point
		target := s.targetPos[i]
		log.Println(root, target, p)
		root.Y += target.Y - p.Y
		root.X += target.X - p.X
		if inField(root) {
			dist := abs(p.Y-target.Y) + abs(p.X-target.X)
			if dist < minDist {
				log.Println(dist)
				minDist = dist
				t.Y, t.X = target.Y, target.X
				if dist == 0 {
					return
				}
			}
		}
	}
	if minDist == 1000 {
		panic("no target")
	}
	return t
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
	if !inField(s.nodes[0].Point) {
		log.Fatal("root is out of field", s.nodes[0].Point)
	}
	miniD = 1000
	var target Point
	// FLIPを使わない理由:
	//  robot全体が動き続けることで、他のノードが行動できる可能性があがる(?)
	log.Printf("%+v %+v\n", s.nodes[0].Point, s.nodes[v].Point)
	n := &s.nodes[v]
	for i := 0; i < 4; i++ {
		root := s.nodes[0].Point
		log.Println(i, "root:", root, "n:", n.Point)
		RotateRobot(i, n, s.nodes[0].Point)
		var t Point
		if !n.HasTakoyaki {
			t = s.closestTakoyaki(n.Point)
		} else {
			t = s.closestTarget(n.Point)
		}
		log.Printf("%+v %+v %+v\n", i, t, n.Point)
		root.Y += t.Y - n.Point.Y
		root.X += t.X - n.Point.X
		log.Printf("%+v\n", root)
		if inField(root) {
			dis := DistancePP(n.Point, t)
			if i == 3 {
				dis++
			}
			if dis < miniD {
				miniD = dis
				target = t
			}
		}
		ReverseRobot(i, n, s.nodes[0].Point)
	}
	log.Println("v:", s.nodes[v].Point, "target:", target, "miniD:", miniD)
	direction = DirectionPP(s.nodes[v].Point, target)
	return
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
	log.Printf("%+v\n", v)
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
	// FLIPの場合は１80度回転
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
	move := s.calcMoveDirection()
	s.MoveRobot(move, &s.nodes[0])
	if !inField(s.nodes[0].Point) {
		log.Fatal("root is out of field", s.nodes[0].Point, move)
	}

	action = append(action, moveAction[move]) // V0 の移動
	// V1 ~
	subAction := make([]byte, V-1)
	takoAction := make([]byte, V)
	takoAction[0] = '.'
	nodeLocked := make([]bool, V)
	nodeLocked[0] = true
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
						inFieldCnt++
						if !nodes[k].HasTakoyaki && s.s.Get(nodes[k].Y, nodes[k].X) {
							// GetTakoyaki
							s.s.Unset(nodes[k].Y, nodes[k].X)
							takoyakiUnsetLog = append(takoyakiUnsetLog, nodes[k].Point)
							takoPoint++
							subP[k] = 'P'
						} else if nodes[k].HasTakoyaki && s.t.Get(nodes[k].Y, nodes[k].X) {
							// ReleaseTakoyaki
							s.t.Unset(nodes[k].Y, nodes[k].X)
							targetUnsetLog = append(targetUnsetLog, nodes[k].Point)
							takoPoint++
							subP[k] = 'P'
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
			//log.Println(bestTakoPoint, bestInFieldCnt)
			//log.Println(bestRotate, string(bestP))
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
		//log.Println(j, takoAction[j], string(takoAction[j]))
		s.moveLeaf(&s.nodes[j], takoAction[j])
	}

	//takoAction[i] = '.'
	//if s.nodes[i].isLeaf() {
	//center := s.nodes[i].parent.Point
	//pi := s.nodes[i].parent.index
	//// 親がロックされている (単体で動かす)
	//if nodeLocked[pi] {
	//var j int
	//var cwOutField, ccwOutField bool
	//for j = 0; j < 4; j++ {
	//var nextPoint Point
	//if j == 3 {
	//// 180度回転
	//nextPoint = s.nodes[i].Point.Rotate(center, CW)
	//nextPoint = nextPoint.Rotate(center, CW)
	//} else {
	//nextPoint = s.nodes[i].Point.Rotate(center, j)
	//}
	//if !inField(nextPoint) {
	//if j == CW {
	//cwOutField = true
	//} else if j == CCW {
	//ccwOutField = true
	//}
	//continue
	//}
	//// catchできる
	//if !s.nodes[i].HasTakoyaki && s.s.Get(nextPoint.Y, nextPoint.X) && !s.t.Get(nextPoint.Y, nextPoint.X) {
	//if j != 3 {
	//takoAction[i] = 'P'
	//s.nodes[i].HasTakoyaki = true
	////log.Println(s.s.Get(nextPoint.Y, nextPoint.X), s.t.Get(nextPoint.Y, nextPoint.X))
	//s.s.Unset(nextPoint.Y, nextPoint.X)
	//s.takoyakiInRobot++
	//s.takoyakiOnField--
	//s.takoyakiPos = deleteItem(s.takoyakiPos, nextPoint)
	//}
	//break
	//}
	//// releaseできる
	//if s.nodes[i].HasTakoyaki && s.t.Get(nextPoint.Y, nextPoint.X) && !s.s.Get(nextPoint.Y, nextPoint.X) {
	//if j != 3 {
	//takoAction[i] = 'P'
	//s.nodes[i].HasTakoyaki = false
	//s.t.Unset(nextPoint.Y, nextPoint.X)
	//s.remainTakoyaki--
	//s.takoyakiInRobot--
	//s.targetPos = deleteItem(s.targetPos, nextPoint)
	//}
	//break
	//}
	//}
	//if j == 3 {
	//// 180度回転
	//if cwOutField {
	//j = CCW
	//} else {
	//j = CW
	//}
	//}
	//if j == 4 {
	//// なにもない
	//if inField(s.nodes[i].Point) && !cwOutField && !ccwOutField {
	//j = 0
	//} else if cwOutField && !ccwOutField {
	//j = CCW
	//} else if !cwOutField && ccwOutField {
	//j = CW
	//} else {
	//j = 0
	//}
	//}
	//move = j // 0:None, 1:CW, 2:CCW
	////center := s.nodes[i].parent.Point
	//RotateRobot(move, &s.nodes[i], center)
	//subAction[i-1] = VAction[move]
	//nodeLocked[i] = true
	//} else {
	//// 親の場所がロックされていない
	//// 親を先に動かす
	//parent := s.nodes[i].parent
	//next := &s.nodes[i]
	//takoAction[i] = '.'
	//takoAction[parent.index] = '.'
	//subAction[i-1] = '.'
	//subAction[parent.index-1] = '.'
	//var pm, m int
	//for pm = 0; pm < 3; pm++ {
	//RotateRobot(pm, parent, parent.parent.Point)
	//for m = 0; m < 3; m++ {
	//RotateRobot(m, next, parent.Point)
	//if inField(next.Point) {
	//if !next.HasTakoyaki && s.s.Get(next.Y, next.X) {
	//takoAction[i] = 'P'
	//next.HasTakoyaki = true
	//s.s.Unset(next.Y, next.X)
	//s.takoyakiInRobot++
	//s.takoyakiOnField--
	//s.takoyakiPos = deleteItem(s.takoyakiPos, next.Point)
	//nodeLocked[i] = true
	//nodeLocked[parent.index] = true
	//subAction[i-1] = VAction[m]
	//subAction[parent.index-1] = VAction[pm]
	////log.Println("catch", next.Point)
	//} else if next.HasTakoyaki && s.t.Get(next.Y, next.X) {
	//takoAction[i] = 'P'
	//next.HasTakoyaki = false
	//s.t.Unset(next.Y, next.X)
	//s.remainTakoyaki--
	//s.takoyakiInRobot--
	//s.targetPos = deleteItem(s.targetPos, next.Point)
	//nodeLocked[i] = true
	//nodeLocked[parent.index] = true
	//subAction[i-1] = VAction[m]
	//subAction[parent.index-1] = VAction[pm]
	////log.Println("release", next.Point)
	//}
	//}
	//if nodeLocked[i] {
	//break
	//}
	//ReverseRobot(m, next, parent.Point)
	//}
	//if nodeLocked[parent.index] {
	//break
	//}
	//ReverseRobot(pm, parent, parent.parent.Point)
	//}
	//if !nodeLocked[i] && !nodeLocked[parent.index] {
	//subAction[i-1] = VAction[0]
	//subAction[parent.index-1] = VAction[0]
	//} else if nodeLocked[i] != nodeLocked[parent.index] {
	//panic("not locked")
	//}
	//}
	//} else {
	//// not leaf
	//// なにもしない
	//}
	action = append(action, subAction...)
	action = append(action, takoAction...)
	action = append(action, '\n')
	//log.Printf("%+v %+v %+v\n", s.nodes[1], s.s.Get(10, 13), s.t.Get(10, 13))
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
		//log.Printf("%+v\n", state.nodes[2])
		//log.Printf("%+v\n", state.nodes[3])
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
		for i := 0; i < 50; i++ {
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
			log.Println(i, state.remainTakoyaki, string(tout[:V]), string(tout[V:len(tout)-1]))
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
