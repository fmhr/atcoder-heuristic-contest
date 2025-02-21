package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	N = 50
)

const (
	COST_STATION = 5000
	COST_RAIL    = 100
)
const (
	// 0:UP, 1:RIGHT, 2:DOWN, 3:LEFT
	UP    = 0
	RIGHT = 1
	DOWN  = 2
	LEFT  = 3
)

// UP, RIGHT, DOWN, LEFT
var dy = []int16{-1, 0, 1, 0}
var dx = []int16{0, 1, 0, -1}

const (
	EMPTY           int16 = -1
	DO_NOTHING      int16 = -1
	STATION         int16 = 0
	RAIL_HORIZONTAL int16 = 1
	RAIL_VERTICAL   int16 = 2
	RAIL_LEFT_DOWN  int16 = 3
	RAIL_LEFT_UP    int16 = 4
	RAIL_RIGHT_UP   int16 = 5
	RAIL_RIGHT_DOWN int16 = 6
	OTHER           int16 = 7 // テストの障害物として使う
)

// int16ToString は、レールタイプのint16の種類を文字列に変換する
// EMPTY = DO_NOTHING = -1 に注意
func int16ToString(a int16) string {
	switch a {
	case EMPTY:
		return "EMPTY"
	case STATION:
		return "STATION"
	case RAIL_HORIZONTAL:
		return "RAIL_HORIZONTAL"
	case RAIL_VERTICAL:
		return "RAIL_VERTICAL"
	case RAIL_LEFT_DOWN:
		return "RAIL_LEFT_DOWN"
	case RAIL_LEFT_UP:
		return "RAIL_LEFT_UP"
	case RAIL_RIGHT_UP:
		return "RAIL_RIGHT_UP"
	case RAIL_RIGHT_DOWN:
		return "RAIL_RIGHT_DOWN"
	case OTHER:
		return "OTHER"
	}
	return "UNKNOWN"
}

func isRail(kind int16) bool {
	return kind >= RAIL_HORIZONTAL && kind <= RAIL_RIGHT_DOWN
}

// railToString は、[]int16のレールの種類を文字列に変換する
func railToString(rails []int16) string {
	var sb strings.Builder
	for _, rail := range rails {
		sb.WriteString(" ")
		sb.WriteString(railMap[rail])
	}
	return sb.String()
}

var railMap = map[int16]string{
	EMPTY:           ".",
	STATION:         "◎",
	RAIL_HORIZONTAL: "─",
	RAIL_VERTICAL:   "│",
	RAIL_LEFT_DOWN:  "┐",
	RAIL_LEFT_UP:    "┘",
	RAIL_RIGHT_UP:   "└",
	RAIL_RIGHT_DOWN: "┌",
	OTHER:           "#",
}

var buildCost = map[int16]int{
	EMPTY:           0,            // EMPTY
	STATION:         COST_STATION, // STATION
	RAIL_HORIZONTAL: COST_RAIL,
	RAIL_VERTICAL:   COST_RAIL,
	RAIL_LEFT_DOWN:  COST_RAIL,
	RAIL_LEFT_UP:    COST_RAIL,
	RAIL_RIGHT_UP:   COST_RAIL,
	RAIL_RIGHT_DOWN: COST_RAIL,
	OTHER:           0, // other
}

// calBuildCost は、[]actの建設コストを計算する
func calBuildCost(act []int16) (cost int) {
	for _, a := range act {
		if val, ok := buildCost[a]; ok {
			cost += val
		} else {
			log.Printf("calBuildCost: invalid kind:%d\n", a)
			panic("calBuildCost: invalid kind")
		}
	}
	return
}

type Action struct {
	Kind    int16
	Y, X    int16
	comment string
}

func (a Action) String() (str string) {
	if a.Kind == DO_NOTHING {
		str = fmt.Sprintf("%d", a.Kind)
	} else {
		str = fmt.Sprintf("%d %d %d", a.Kind, a.Y, a.X)
	}
	str = a.comment + str + "\n"
	return
}

type Field struct {
	cell     [50][50]int16
	stations []Pos
	uf       *UnionFind
}

// n == 50
// 全マスをノードとして、UFをもつ
func NewField(n int) *Field {
	if n != 50 {
		panic("n need to be 50")
	}
	f := new(Field)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			f.cell[i][j] = EMPTY
		}
	}
	f.uf = NewUnionFind()
	return f
}

func (f *Field) Clone() *Field {
	if f == nil {
		return nil
	}
	newField := &Field{
		cell:     [50][50]int16{},
		stations: make([]Pos, len(f.stations)),
		uf:       f.uf.Clone(),
	}
	copy(newField.stations, f.stations)
	// 2次元配列のコピーを最適化
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			newField.cell[i][j] = f.cell[i][j]
		}
	}
	return newField
}

// typeToString は、posのセルの種類を返す 表示用のレールの記号
func (f Field) typeToString(pos Pos) string {
	return railMap[f.cell[pos.Y][pos.X]]
}

func (f Field) cellString() string {
	str := "view cellString()\n"
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			str += railMap[f.cell[i][j]]
		}
		str += "\n"
	}
	return str
}

func (f *Field) build(act Action) error {
	if act.Kind == DO_NOTHING {
		return nil
	}
	if act.Kind < 0 || act.Kind > 6 {
		//panic("invalid kind:" + fmt.Sprint(act.Kind))
		return fmt.Errorf("invalid kind:%s", fmt.Sprint(act.Kind))
	}
	if f.cell[act.Y][act.X] != EMPTY {
		if !(act.Kind == STATION && f.cell[act.Y][act.X] >= 1 && f.cell[act.Y][act.X] <= 6) {
			// 駅は線路の上に建てることができる
			log.Println(f.cellString())
			log.Printf("try to build: typ:%d Y:%d X:%d but already built %d\n", act.Kind, act.Y, act.X, f.cell[act.Y][act.X])
			return fmt.Errorf("already built")
		}
	}
	if act.Kind == STATION {
		f.stations = append(f.stations, Pos{Y: act.Y, X: act.X})
	}
	f.cell[act.Y][act.X] = act.Kind
	// 連結成分をつなげる
	y, x := act.Y, act.X
	kind := act.Kind
	// 上下左右を確認して、線路が繋がっている場合は連結成分をつなげる
	// 上
	if y > 0 {
		switch kind {
		case STATION, RAIL_VERTICAL, RAIL_LEFT_UP, RAIL_RIGHT_UP:
			switch f.cell[y-1][x] {
			case STATION, RAIL_VERTICAL, RAIL_LEFT_DOWN, RAIL_RIGHT_DOWN:
				f.uf.unite((y*50 + x), ((y-1)*50 + x))
			}
		}
	}
	// 下
	if y < 49 {
		switch kind {
		case STATION, RAIL_VERTICAL, RAIL_LEFT_DOWN, RAIL_RIGHT_DOWN:
			switch f.cell[y+1][x] {
			case STATION, RAIL_VERTICAL, RAIL_LEFT_UP, RAIL_RIGHT_UP:
				f.uf.unite(y*50+x, (y+1)*50+x)
			}
		}
	}
	// 左
	if x > 0 {
		switch kind {
		case STATION, RAIL_HORIZONTAL, RAIL_LEFT_DOWN, RAIL_LEFT_UP:
			switch f.cell[y][x-1] {
			case STATION, RAIL_HORIZONTAL, RAIL_RIGHT_DOWN, RAIL_RIGHT_UP:
				f.uf.unite(y*50+x, y*50+(x-1))
			}
		}
	}
	// 右
	if x < 49 {
		switch kind {
		case STATION, RAIL_HORIZONTAL, RAIL_RIGHT_DOWN, RAIL_RIGHT_UP:
			switch f.cell[y][x+1] {
			case STATION, RAIL_HORIZONTAL, RAIL_LEFT_DOWN, RAIL_LEFT_UP:
				f.uf.unite(y*50+x, y*50+(x+1))
			}
		}
	}
	return nil
}

// collectStationsは、posから距離２以内の駅の位置を返す
func (f Field) collectStations(pos Pos) (stations []Pos) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if absInt(dy)+absInt(dx) > 2 {
				continue
			}
			y, x := pos.Y+int16(dy), pos.X+int16(dx)
			if y >= 0 && y < 50 && x >= 0 && x < 50 && f.cell[y][x] == STATION {
				stations = append(stations, Pos{Y: y, X: x})
			}
		}
	}
	return stations
}

// checkConnect 駅,路線をつかって、a,bがつながっているかを返す
func (f Field) checkConnect(a, b Pos) bool {
	if f.uf.same(a.Y*50+a.X, b.Y*50+b.X) {
		return true
	}

	stations0 := f.collectStations(a)
	if len(stations0) == 0 {
		return false
	}
	stations1 := f.collectStations(b)
	if len(stations1) == 0 {
		return false
	}
	for _, s0 := range stations0 {
		for _, s1 := range stations1 {
			if f.uf.same(s0.Y*50+s0.X, s1.Y*50+s1.X) {
				return true
			}
		}
	}
	return false
}

// 2点間の最短経路を返す (a から b へ)
// bはEMPTYまたはSTATION
func (f *Field) findShortestPath(a, b Pos) (path []Pos) {
	// a から b への最短経路を返す
	// field=EMPTY なら移動可能 それ以外は移動不可
	var dist [2500]int16
	for i := 0; i < 2500; i++ {
		dist[i] = 10000
	}
	dist[int(a.Y)*50+int(a.X)] = 0
	var que []Pos
	que = append(que, a)
	for len(que) > 0 {
		p := que[0]
		que = que[1:]
		//if p == b {
		//break
		//}
		for d := 0; d < 4; d++ {
			y, x := p.Y+dy[d], p.X+dx[d]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if f.cell[y][x] != EMPTY && f.cell[y][x] != STATION {
				continue
			}
			if dist[int(y)*50+int(x)] > dist[int(p.Y)*50+int(p.X)]+1 {
				dist[int(y)*50+int(x)] = dist[int(p.Y)*50+int(p.X)] + 1
				que = append(que, Pos{Y: y, X: x})
			}
		}
		//log.Println(len(que))
	}
	if dist[int(b.Y)*50+int(b.X)] == 10000 {
		//log.Println(gridToString(dist))
		log.Println("can't reach", a, b)
		//log.Println(f.cellString())
		return nil
	}
	//log.Println(f.cellString())
	//log.Println(showGrid(dist))

	// b から a への経路を復元
	path = append(path, b)
	for path[len(path)-1] != a {
		p := path[len(path)-1]
		for d := 0; d < 4; d++ {
			y, x := p.Y+dy[d], p.X+dx[d]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if y == a.Y && x == a.X {
				path = append(path, Pos{Y: y, X: x})
				break
			}
			if f.cell[y][x] != EMPTY && f.cell[y][x] != STATION {
				continue
			}
			if dist[int(y)*50+int(x)] == dist[int(p.Y)*50+int(p.X)]-1 {
				path = append(path, Pos{Y: y, X: x})
				break
			}
		}
	}
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
}

// 2点間の最短経路を返す (a から b へ)
// paht[0]とpath[len(path)-1]は駅
func (f *Field) selectRails(path []Pos) (types []int16) {
	types = make([]int16, len(path))
	types[0] = STATION
	types[len(path)-1] = STATION
	for i := 1; i < len(path)-1; i++ {
		y0, x0 := path[i-1].Y, path[i-1].X
		y1, x1 := path[i].Y, path[i].X
		y2, x2 := path[i+1].Y, path[i+1].X
		if y0 == y1 && y1 == y2 {
			types[i] = RAIL_HORIZONTAL
		} else if x0 == x1 && x1 == x2 {
			types[i] = RAIL_VERTICAL
		} else if (y0 < y1 && x1 < x2) || (y2 < y1 && x1 < x0) {
			types[i] = RAIL_RIGHT_UP
		} else if (y0 < y1 && x1 > x2) || (y2 < y1 && x1 > x0) {
			types[i] = RAIL_LEFT_UP
		} else if (y0 > y1 && x1 < x2) || (y2 > y1 && x1 < x0) {
			types[i] = RAIL_RIGHT_DOWN
		} else if (y0 > y1 && x1 > x2) || (y2 > y1 && x1 > x0) {
			types[i] = RAIL_LEFT_DOWN
		} else {
			panic("invalid path")
		}
	}
	return
}

// countSrcDst は、posから距離２以内のsrc,dstの数を返す ただしすでに駅がある場合はカウントしない
// 駅を想定
func (f Field) countSrcDst(pos Pos, in Input) (srcNum, dstNum int) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if absInt(dy)+absInt(dx) > 2 {
				continue
			}
			y, x := pos.Y+int16(dy), pos.X+int16(dx)
			if y >= 0 && y < 50 && x >= 0 && x < 50 {
			NEXT:
				for i := 0; i < in.M; i++ {
					for _, s := range f.stations {
						if distance(s, Pos{Y: y, X: x}) <= 2 {
							break NEXT
						}
					}
					if in.src[i].Y == y && in.src[i].X == x {
						srcNum++
					}
					if in.dst[i].Y == y && in.dst[i].X == x {
						dstNum++
					}
				}
			}
		}
	}
	return srcNum, dstNum
}

// railが繋がる向きを返す,dy,dxに対応
// 0:UP, 1:RIGHT, 2:DOWN, 3:LEFT
func railDirection(rail int16) []int16 {
	switch rail {
	case RAIL_HORIZONTAL:
		return []int16{1, 3}
	case RAIL_VERTICAL:
		return []int16{0, 2}
	case RAIL_LEFT_DOWN:
		return []int16{3, 2}
	case RAIL_LEFT_UP:
		return []int16{0, 3}
	case RAIL_RIGHT_UP:
		return []int16{0, 1}
	case RAIL_RIGHT_DOWN:
		return []int16{1, 2}
	case STATION, EMPTY:
		return []int16{0, 1, 2, 3}
	}
	return nil
}

// isRailConnected はレールの接続ルールを判定する
func isRailConnected(railType int16, direction int, isStart bool) bool {
	switch direction {
	case UP:
		if isStart {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_UP || railType == RAIL_RIGHT_UP
		} else {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_DOWN || railType == RAIL_RIGHT_DOWN
		}
	case RIGHT:
		if isStart {
			return railType == RAIL_HORIZONTAL || railType == RAIL_RIGHT_DOWN || railType == RAIL_RIGHT_UP
		} else {
			return railType == RAIL_HORIZONTAL || railType == RAIL_LEFT_DOWN || railType == RAIL_LEFT_UP
		}
	case DOWN:
		if isStart {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_DOWN || railType == RAIL_RIGHT_DOWN
		} else {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_UP || railType == RAIL_RIGHT_UP
		}
	case LEFT:
		if isStart {
			return railType == RAIL_HORIZONTAL || railType == RAIL_LEFT_DOWN || railType == RAIL_LEFT_UP
		} else {
			return railType == RAIL_HORIZONTAL || railType == RAIL_RIGHT_DOWN || railType == RAIL_RIGHT_UP
		}
	}
	return false
}

// canMove は、aからbに移動可能かを返す
// dist[2500]を参照してpathを返すときに、セルの種類もチェックする必要がある
// aからbの向きに移動できる && bがaから受けいることができるか
func (f Field) canMove(a, b Pos) bool {
	if distance(a, b) != 1 {
		log.Println("distance", a, b, distance(a, b))
		return false
	}
	if f.cell[b.Y][b.X] == OTHER {
		return false
	}
	// directionはaからbに移動する向き
	direction := UP
	switch {
	case a.Y == b.Y && a.X < b.X:
		direction = RIGHT
	case a.Y == b.Y:
		direction = LEFT
	case a.Y < b.Y:
		direction = DOWN
	}
	// aからbに移動する向きが繋がっているか
	x, y := a.X+dx[direction], a.Y+dy[direction]
	if !(x == b.X && y == b.Y) {
		log.Fatal("canMove: invalid direction")
	}
	// aのレールが繋がっていなかったらfalse
	if isRail(f.cell[a.Y][a.X]) {
		if !isRailConnected(f.cell[a.Y][a.X], direction, true) {
			return false
		}
	}
	// bのレールが繋がっていなかったらfalse
	if isRail(f.cell[b.Y][b.X]) {
		if !isRailConnected(f.cell[b.Y][b.X], direction, false) {
			return false
		}
	}
	return true
}

// 駅a, bが繋がることができるか、できないときnil,できるとき距離を返すpath
// すでにある線路も活用できるようにする
// 繋がらなかった時はnilを返す
// 繋がっているはずにnilになる場合はバグなので、呼び出し元でチェックする
func (f Field) canConnect(a, b Pos) []Pos {
	var dist [2500]int16
	for i := 0; i < 2500; i++ {
		dist[i] = 10000
	}
	dist[int(a.Y)*50+int(a.X)] = 0
	q := []Pos{a}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		if p == b {
			break
		}

		direction := railDirection(f.cell[p.Y][p.X])

		if len(direction) == 0 {
			log.Println(f.cell[p.Y][p.X])
			panic("invalid rail")
		}
		for _, d := range direction {
			y, x := p.Y+dy[d], p.X+dx[d]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if f.cell[y][x] == EMPTY || f.cell[y][x] == STATION {
				if dist[int(y)*50+int(x)] > dist[int(p.Y)*50+int(p.X)]+1 {
					dist[int(y)*50+int(x)] = dist[int(p.Y)*50+int(p.X)] + 1
					q = append(q, Pos{Y: y, X: x})
				}
			}
			if isRail(f.cell[y][x]) {
				connect := false
				switch d {
				case UP:
					if f.cell[y][x] == RAIL_VERTICAL || f.cell[y][x] == RAIL_LEFT_DOWN || f.cell[y][x] == RAIL_RIGHT_DOWN {
						connect = true
					}
				case RIGHT:
					if f.cell[y][x] == RAIL_HORIZONTAL || f.cell[y][x] == RAIL_LEFT_DOWN || f.cell[y][x] == RAIL_LEFT_UP {
						connect = true
					}
				case DOWN:
					if f.cell[y][x] == RAIL_VERTICAL || f.cell[y][x] == RAIL_LEFT_UP || f.cell[y][x] == RAIL_RIGHT_UP {
						connect = true
					}
				case LEFT:
					if f.cell[y][x] == RAIL_HORIZONTAL || f.cell[y][x] == RAIL_RIGHT_DOWN || f.cell[y][x] == RAIL_RIGHT_UP {
						connect = true
					}
				}
				if connect {
					if dist[int(y)*50+int(x)] > dist[int(p.Y)*50+int(p.X)]+1 {
						dist[int(y)*50+int(x)] = dist[int(p.Y)*50+int(p.X)] + 1
						q = append(q, Pos{Y: y, X: x})
					}
				}
			}
		}
	}
	//log.Println(gridToString(dist))
	//log.Println("dist", dist[int(b.Y)*50+int(b.X)])
	if dist[int(b.Y)*50+int(b.X)] == 10000 {
		log.Println("can't reach", a, b)
		return nil
	}
	//log.Println(f.cellString())
	//log.Println(gridToString(dist))
	//log.Println(dist[int(b.Y)*50+int(b.X)], dist[int(a.Y)*50+int(a.X)])
	//log.Println(b, a)
	// b から a への経路を復元
	path := []Pos{b}
MAKEPATH:
	for {
		p := path[len(path)-1]
		for d := 0; d < 4; d++ {
			y, x := p.Y+dy[d], p.X+dx[d]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				// 場外
				continue
			}
			if y == a.Y && x == a.X {
				// 駅に到達
				path = append(path, Pos{Y: y, X: x})
				break MAKEPATH
			}
			if dist[int(y)*50+int(x)] == dist[int(p.Y)*50+int(p.X)]-1 {
				if f.canMove(p, Pos{Y: y, X: x}) {
					path = append(path, Pos{Y: y, X: x})
					break
				}
			}
		}
	}
	//log.Println("len(path)", len(path))
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
}

var ErrNotEnoughMoney = fmt.Errorf("not enough money")

type State struct {
	field     *Field
	money     int
	turn      int
	income    int
	score     int // 最終ターンでの予想スコア
	actions   []Action
	connected []bool // in.Mが接続済みかどうか
}

func (s *State) Clone() *State {
	newActions := make([]Action, len(s.actions))
	copy(newActions, s.actions)
	newConnected := make([]bool, len(s.connected))
	copy(newConnected, s.connected)
	newState := &State{
		field:     s.field.Clone(),
		money:     s.money,
		turn:      s.turn,
		income:    s.income,
		score:     s.score,
		actions:   newActions,
		connected: newConnected,
	}
	return newState
}

func NewState(in *Input) *State {
	s := new(State)
	s.field = NewField(in.N)
	s.money = in.K
	s.connected = make([]bool, in.M)
	return s
}

func (s *State) do(act Action, in Input) error {
	if s.money < buildCost[act.Kind] {
		return ErrNotEnoughMoney
	}
	if act.Kind != DO_NOTHING {
		err := s.field.build(act)
		if err != nil {
			log.Println("acttype:", act.Kind, "pos:", act.Y, act.X)
			log.Println("build error", err)
			return err
		}
		s.money -= buildCost[act.Kind]
		if act.Kind == STATION {
			for i := 0; i < in.M; i++ {
				if !s.connected[i] {
					if s.field.checkConnect(in.src[i], in.dst[i]) {
						s.income += in.income[i]
						s.connected[i] = true
					}
				}
			}
		}
	}
	s.turn++
	s.money += s.income
	s.score = s.money + s.income*(in.T-s.turn)
	act.comment = fmt.Sprintf("#turn=%d, \n#money=%d, \n#income=%d\n #Score=%d\n",
		s.turn, s.money, s.income, s.score)
	s.actions = append(s.actions, act)
	return nil
}

type Pos struct {
	Y, X int16
}

func (p Pos) add(a Pos) Pos {
	return Pos{Y: p.Y + a.Y, X: p.X + a.X}
}

func (p Pos) Clone() Pos {
	return Pos{Y: p.Y, X: p.X}
}

func distance(a, b Pos) int16 {
	return absInt16(a.X-b.X) + absInt16(a.Y-b.Y)
}

type Pair [2]Pos

// uniquePair は、p1, p2の順番を統一してのPairを返す
func uniquePair(p1, p2 Pos) Pair {
	if p1.Y < p2.Y || p1.Y == p2.Y && p1.X < p2.X {
		return Pair{p1, p2}
	}
	return Pair{p2, p1}
}

// stationの周辺
var ddy = [13]int16{0, -1, 0, 1, 0, -1, 1, 1, -1, -2, 0, 2, 0}
var ddx = [13]int16{0, 0, 1, 0, -1, 1, 1, -1, -1, 0, 2, 0, -2}

// shortestPathLimited gridのaからbまでで、movenableの範囲で最短経路を返す
func shortestPathLimited(grid [2500]int16, a, b Pos, movenable int16) []Pos {
	//log.Println(a, b, movenable)
	var dist [2500]int16
	for i := 0; i < 2500; i++ {
		dist[i] = 1000
	}
	dist[a.Y*50+a.X] = 0
	que := make([]Pos, 0, 2500)
	que = append(que, a)
	for len(que) > 0 {
		p := que[0]
		que = que[1:]
		if p == b {
			break
		}
		for i := 0; i < 4; i++ {
			y, x := p.Y+dy[i], p.X+dx[i]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if grid[y*50+x] == movenable || y == b.Y && x == b.X {
				if dist[y*50+x] > dist[p.Y*50+p.X]+1 {
					dist[y*50+x] = dist[p.Y*50+p.X] + 1
					que = append(que, Pos{Y: y, X: x})
				}
			}
		}
	}
	if dist[b.Y*50+b.X] == 1000 {
		return nil
	}
	//log.Println(dist[a.Y*50+a.X], dist[b.Y*50+b.X])
	// b から a への経路を復元
	path := []Pos{b}
	for path[len(path)-1] != a {
		p := path[len(path)-1]
		for i := 0; i < 4; i++ {
			y, x := p.Y+dy[i], p.X+dx[i]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if y == a.Y && x == a.X {
				path = append(path, Pos{Y: y, X: x})
				break
			}
			if (grid[y*50+x] == movenable || grid[y*50+x] == STATION) && dist[y*50+x] == dist[p.Y*50+p.X]-1 {
				if dist[y*50+x] < dist[p.Y*50+p.X] {
					path = append(path, Pos{Y: y, X: x})
					break
				}
			}
		}
	}
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
}

// すべての駅を繋ぐ鉄道を敷設する
// クラスカル法を使っているが、簡易距離と制約によって、無駄なエッジが作られることがある
// .............◎.............
// ......................┌◎..........│....◎─┐│.└───◎◎
// ........◎─┐.....┌◎───◎┘└─◎........│......││.......
// ........│.◎──◎─┐│........│.◎┐..┌◎─◎─◎────◎◎──┐....
// ........│......◎┘.....┌──◎─┘└─◎┘│...│........◎─┐..
// ↓
// .............◎.............
// .............│........┌◎..........│....◎─┐│.└───◎◎
// ........◎─┐..│..┌◎───◎┘└─◎........│......││.......
// ........│.◎──◎─┐│........│.◎┐..┌◎─◎─◎────◎◎──┐....
// ........│.└──┘.
func constructRailway(in Input, stations []Pos) []Edge {
	numStations := len(stations)
	stationIndexMap := make(map[Pos]int)
	for i, s := range stations {
		stationIndexMap[s] = i
	}
	// 決めておいた駅を建設する
	field := NewField(in.N)
	for i := 0; i < numStations; i++ {
		field.build(Action{Kind: STATION, Y: stations[i].Y, X: stations[i].X})
	}
	log.Println(field.cellString())

	// マンハッタン距離を使って,全駅間の暫定距離を求める
	edges := []Edge{}
	for i := 0; i < numStations; i++ {
		for j := i + 1; j < numStations; j++ {
			dist := int(distance(stations[i], stations[j]))
			edges = append(edges, Edge{From: i, To: j, Cost: dist})
		}
	}
	// ここまで前処理
	//////////////////////////
	// MSTを求める
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Cost < edges[j].Cost
	})

	// UnionFindで連結成分を管理
	uf := NewUnionFind()
	// Kruskal法で最小全域木を求める
	mstEdges := []Edge{}
	for _, edge := range edges {
		// すでに連結されている場合はスキップ
		if uf.same(int16(edge.From), int16(edge.To)) {
			continue
		}
		// 連結可能か確認
		path := field.canConnect(stations[edge.From], stations[edge.To])
		// ここではフィールドを使っているので、path=nilの可能性もあり
		if path != nil {
			types := field.selectRails(path)
			// 途中に駅があるか確認
			// サイクルを作らない場合もあるので消さない
			for i := 1; i < len(path)-1; i++ {
				if field.cell[path[i].Y][path[i].X] == STATION {
					continue
				}
				field.build(Action{Kind: types[i], Y: path[i].Y, X: path[i].X})
			}
			uf.unite(int16(edge.From), int16(edge.To))
			edge.Path = path
			edge.Rail = types
			mstEdges = append(mstEdges, edge)
			//log.Println("connect", stations[edge.From], stations[edge.To], "cost=", edge.Cost)
			//log.Println(path)
			//log.Println(field.cellString())
		}
	}
	log.Println(field.cellString())
	//log.Println("Num of Stations", numStations)
	//log.Println("Num of Edges", len(mstEdges))
	//log.Println(field.cellString())
	// これをもとに、純粋なエッジ、駅、を分解する
	// エッジに駅が含まれている時は再度typesを求めて含まれていないようにする
	// 次に、src,dstを繋ぐエッジを探す
	//　このときマンハッタン距離と、上の結果上を使うエッジのさが大きい時は、新しいエッジを作る
	// 線路->駅は建築可能
	//var grid [2500]int16
	//for i := 0; i < 50; i++ {
	//for j := 0; j < 50; j++ {
	//grid[i*50+j] = field.cell[i][j]
	//}
	//}
	//for i := 0; i < 50; i++ {
	//fmt.Println(grid[i*50 : i*50+50])
	//}

	//for i := 0; i < in.M; i++ {
	//src, dst := in.src[i], in.dst[i]
	//dist0 := distance(src, dst)
	//dest1 := shortestPathLimited(grid, src, dst, EMPTY)
	//dist1 := shortestPathLimited(grid, src, dst, EMPTY)
	//log.Println(src, "->", dst, dist0, len(dest1), len(dist1))
	//}
	//log.Println(len(mstEdges))
	//for _, edge := range mstEdges {
	//log.Println(edge.From, edge.To, edge.Cost, len(edge.Path), len(edge.Rail))
	//}
	///////////////////////////////////
	// 全てに駅間を繋ぐエッジを作る
	// TODO:線路の種類がMSTと違うルートがあるので要修正
	// 使わない場所はOTHERにして、線路と駅をコピーする
	field2 := NewField(in.N) // mstEdgesで使われている場所だけを使う
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			field2.cell[i][j] = OTHER
		}
	}
	for _, edge := range mstEdges {
		for _, p := range edge.Path {
			if isRail(field.cell[p.Y][p.X]) || field.cell[p.Y][p.X] == STATION {
				field2.cell[p.Y][p.X] = field.cell[p.Y][p.X]
			}
		}
	}
	log.Println(field2.cellString())
	pathpath := make([][]Pos, 0, numStations)
	for i := 0; i < numStations; i++ {
		for j := i + 1; j < numStations; j++ {
			path := field2.canConnect(stations[i], stations[j])
			// すべての駅は繋がっているはずなので、nilはありえない
			if path == nil {
				log.Fatal("can't connect", i, stations[i], j, stations[j])
			}
			pathpath = append(pathpath, path)
		}
	}
	log.Println("stations", len(stations))
	log.Println("pathpath", len(pathpath))
	///////////////////////////////////
	// src,dstの対応する駅を探して、その間を繋ぐエッジを作る
	count := 0
	unique := make(map[Pair]bool)
	for i := 0; i < in.M; i++ {
		src, dst := in.src[i], in.dst[i]
		statinsSrc := make([]Pos, 0)
		statinsDst := make([]Pos, 0)
		for _, s := range stations {
			if distance(s, src) <= 2 {
				statinsSrc = append(statinsSrc, s)
			}
			if distance(s, dst) <= 2 {
				statinsDst = append(statinsDst, s)
			}

		}
		if len(statinsSrc) == 0 || len(statinsDst) == 0 {
			// すべての家と職場は駅から距離２以内にある
			panic("no station")
		}
		for _, s0 := range statinsSrc {
			for _, s1 := range statinsDst {
				if unique[uniquePair(s0, s1)] {
					continue
				}
				if s0 == s1 {
					continue
				}
				unique[uniquePair(s0, s1)] = true
				path := field2.canConnect(s0, s1)
				if len(path) == 0 {
					log.Fatal("can't connect", s0, s1)
				}
				types := field2.selectRails(path)
				edge := Edge{From: stationIndexMap[s0], To: stationIndexMap[s1], Path: path, Rail: types}
				mstEdges = append(mstEdges, edge)
				count++
			}
		}
	}
	log.Println("count", count, len(unique))
	return mstEdges
}

// choseStationPosition は,駅の場所をあらかじめ決める
// Inputからすべての家と職場の位置を所得して、その全てが駅から距離２以下になるように駅を配置する
func choseStationPosition(in Input) (poss []Pos) {
	var grid [2500]int16
	for i := 0; i < in.M; i++ {
		y, x := in.src[i].Y, in.src[i].X
		grid[int(y)*50+int(x)] += 1
		y, x = in.dst[i].Y, in.dst[i].X
		grid[int(y)*50+int(x)] += 1
	}
	uncoverd := make([]Pos, 0, in.M*2)
	for i := 0; i < in.M; i++ {
		uncoverd = append(uncoverd, in.src[i])
		uncoverd = append(uncoverd, in.dst[i])
	}
	for len(uncoverd) > 0 {
		bestPos := Pos{Y: 0, X: 0}
		bestHit := 0
		for i := 0; i < 50; i++ {
			for j := 0; j < 50; j++ {
				count := 0
				y, x := int16(i), int16(j)
				for _, u := range uncoverd {
					if distance(u, Pos{Y: y, X: x}) <= 2 {
						count++
					}
				}
				if count > bestHit {
					bestHit = count
					bestPos = Pos{Y: y, X: x}
				}
			}
		}
		if bestHit == 0 {
			panic("no station position")
		}
		poss = append(poss, bestPos)
		newUncoverd := make([]Pos, 0, len(uncoverd))
		for _, u := range uncoverd {
			if distance(u, bestPos) > 2 {
				newUncoverd = append(newUncoverd, u)
			}
		}
		uncoverd = newUncoverd
	}

	return poss
}

type bsState struct {
	state       State
	restActions []uint
}

func newBsState(in *Input, numActions int) *bsState {
	new := &bsState{
		state:       *NewState(in),
		restActions: make([]uint, numActions),
	}
	for i := 0; i < numActions; i++ {
		new.restActions[i] = uint(i)
	}
	return new
}

func (s *bsState) Clone() *bsState {
	clonedState := &bsState{
		state:       *s.state.Clone(),
		restActions: make([]uint, len(s.restActions)),
	}
	copy(clonedState.restActions, s.restActions)
	return clonedState
}

type bsAction struct {
	path []Pos
	typ  []int16
}

// すべての駅の場所と、それらをつなぐエッジを行動にする
func beamSearch(in Input) {
	// 駅の位置を選ぶ
	stations := choseStationPosition(in)
	// 駅を繋ぐエッジを求める
	edges := constructRailway(in, stations)

	allAction := make([]bsAction, 0, len(stations)+len(edges))
	for _, s := range stations {
		allAction = append(allAction, bsAction{path: []Pos{s}, typ: []int16{STATION}})
	}
	for _, e := range edges {
		allAction = append(allAction, bsAction{path: e.Path, typ: e.Rail})
		//log.Println(railToString(e.Rail))
	}
	log.Println("actionNum", len(allAction))
	initialState := newBsState(&in, len(allAction))
	beamWidth := 100
	beamStates := make([]bsState, 0, beamWidth)
	beamStates = append(beamStates, *initialState)
	nextStates := make([]bsState, 0, beamWidth)
	bestState := initialState.Clone()
	var loop int
	for len(beamStates) > 0 {
		for i := 0; i < min(beamWidth, len(beamStates)); i++ {
			if beamStates[i].state.turn > 800 {
				if beamStates[i].state.score > bestState.state.score {
					bestState = beamStates[i].Clone()
				}
				continue
			}
			// DO_NOTHINGの場合
			//newState := beamStates[i].Clone()
			//err := newState.state.do(Action{Kind: DO_NOTHING}, in)
			//if err != nil {
			//panic(err)
			//}
			//nextStates = append(nextStates, *newState)
		NEWSTATE:
			for j := 0; j < len(beamStates[i].restActions); j++ {
				act := allAction[beamStates[i].restActions[j]]
				// costの確認
				// actionを精査（駅がすでにあるなら除く)
				if len(act.path) != len(act.typ) {
					panic("invalid action")
				}
				// 不要なactionをのぞいたp,tを作る
				p := make([]Pos, 0, len(act.path))
				t := make([]int16, 0, len(act.typ))
				tmp := beamStates[i].state.field // これはcloneしていないので使わない
				for i := 0; i < len(act.path); i++ {
					if act.typ[i] == tmp.cell[act.path[i].Y][act.path[i].X] {
						continue
					}
					if isRail(act.typ[i]) && tmp.cell[act.path[i].Y][act.path[i].X] == STATION {
						continue
					}
					if isRail(act.typ[i]) && isRail(tmp.cell[act.path[i].Y][act.path[i].X]) {
						// 両方線路で種類が違う時
						break NEWSTATE
					}
					p = append(p, act.path[i])
					t = append(t, act.typ[i])
				}
				costMoney := calBuildCost(t) //純粋なコスト(money)
				if beamStates[i].state.money < costMoney && beamStates[i].state.income == 0 {
					// お金が足りない＋収入がない時はスキップ
					continue
				}
				// DO_NOTHINGで必要な分だけ待つ
				costTime := 0
				if beamStates[i].state.income > 0 {
					costTime = costMoney / beamStates[i].state.income // incomeを考慮したコスト
				}
				////log.Println(len(p), "DoNothing", costTime-len(p))
				//// 残されたターン数で実行できない時
				if beamStates[i].state.turn+costTime > 800 {
					continue
				}
				newState := beamStates[i].Clone()
				for costMoney > newState.state.money {
					err := newState.state.do(Action{Kind: DO_NOTHING}, in)
					if err != nil {
						panic(err)
					}
					if newState.state.turn > 800 {
						log.Println("over time 800:", newState.state.turn)
						panic("over time")
					}
				}
				if newState.state.money < costMoney {
					log.Println("----------------------------")
					log.Println("actions", railToString(t))
					log.Println("costMoney", costMoney, "costTime", costTime)
					log.Println("money", newState.state.money, "income", newState.state.income)
					log.Println("turn", newState.state.turn)
					panic("not enough money")
				}
				for j := 0; j < len(p); j++ {
					err := newState.state.do(Action{Kind: t[j], Y: p[j].Y, X: p[j].X}, in)
					if err != nil {
						log.Println("actions", railToString(t))
						log.Println("action", t[j], p[j])
						panic(err)
					}
				}
				// delete action
				newState.restActions = append(newState.restActions[:j], newState.restActions[j+1:]...)
				nextStates = append(nextStates, *newState)
				if newState.state.score > bestState.state.score {
					bestState = newState.Clone()
				}
			}
		}
		log.Println("nextStates", len(nextStates))
		sort.Slice(nextStates, func(i, j int) bool {
			return nextStates[i].state.score > nextStates[j].state.score
		})
		if len(nextStates) > 0 {
			log.Println("score:", nextStates[0].state.score, nextStates[len(nextStates)-1].state.score)
			log.Println("income:", nextStates[0].state.income, nextStates[len(nextStates)-1].state.income)
			log.Println("0:", nextStates[0].state.score, "last:", nextStates[len(nextStates)-1].state.score)
		}
		log.Println("loop", loop, "beamStates", len(beamStates))
		if len(nextStates) > beamWidth {
			beamStates = nextStates[:beamWidth]
		} else {
			beamStates = nextStates
		}
		nextStates = make([]bsState, 0, beamWidth)
		loop++
	}
	log.Println("bestScore", bestState.state.score, "income:", bestState.state.income, "turn:", bestState.state.turn)
	log.Println(bestState.state.field.cellString())

}

func greedy(in Input) {
	state := NewState(&in)
	bestPos := Pos{Y: 0, X: 0}
	bestCover := 0
	for i := 0; i < in.N; i++ {
		for j := 0; j < in.N; j++ {
			a, b := state.field.countSrcDst(Pos{Y: int16(i), X: int16(j)}, in)
			if a+b > bestCover {
				bestCover = a + b
				bestPos = Pos{Y: int16(i), X: int16(j)}
			}
		}
	}
	//log.Println(bestCover, bestPos)
	state.do(Action{Kind: STATION, Y: bestPos.Y, X: bestPos.X}, in)
	//log.Printf("\n%s", state.field.cellString())
	// 現在あるstationでカバーされている片方だけがカバーされていないケースを探す
	uncoverd_home_workplace := make([]Pos, 0)
	for i := 0; i < in.M; i++ {
		var a_coverd, b_coverd bool
		if state.field.checkConnect(in.src[i], in.dst[i]) {
			continue
		}
		for _, s := range state.field.stations {
			if !a_coverd {
				a := distance(s, in.src[i])
				if a <= 2 {
					a_coverd = true
				}
			}
			if !b_coverd {
				b := distance(s, in.dst[i])
				if b <= 2 {
					b_coverd = true
				}
			}
		}
		if a_coverd && !b_coverd {
			uncoverd_home_workplace = append(uncoverd_home_workplace, in.dst[i])
		}
		if !a_coverd && b_coverd {
			uncoverd_home_workplace = append(uncoverd_home_workplace, in.src[i])
		}
	}
	log.Println(len(uncoverd_home_workplace), "uncoverd_home_workplace=", uncoverd_home_workplace)
	if len(uncoverd_home_workplace) == 0 {
		panic("no uncoverd_home_workplace")
	}
	// 近い順に処理する
	sort.Slice(uncoverd_home_workplace, func(i, j int) bool {
		return distance(bestPos, uncoverd_home_workplace[i]) < distance(bestPos, uncoverd_home_workplace[j])
	})

	// uncoverd_home_workplace [0]を次にstationを置く場所とする
	for i := 0; i < len(uncoverd_home_workplace); i++ {
		nextPos := uncoverd_home_workplace[0]
		uncoverd_home_workplace = uncoverd_home_workplace[1:]
		ss := state.field.collectStations(nextPos)
		if len(ss) > 0 {
			continue
		}
		// 一番近いstationに繋げる (すべての駅は繋がっていると仮定)
		for _, st := range state.field.stations {
			if distance(st, nextPos) < distance(bestPos, nextPos) {
				bestPos = st
			}
		}
		path := state.field.findShortestPath(bestPos, nextPos)
		if len(path) > in.T-state.turn {
			// 建設するためのターンが足りない
			break
		}
		if path == nil {
			// 到達不可能
			continue
		}
		types := state.field.selectRails(path)
		cost := calBuildCost((types[1:]))
		needMoney := cost - state.money
		doNotthingTurn := 0
		if needMoney > 0 {
			if state.income == 0 {
				// 収入がないので待っても無駄
				continue
			}
			// 建設途中にDoNothingが入るターン数 簡易計算
			doNotthingTurn = needMoney / state.income
			if state.turn+doNotthingTurn+len(path) > in.T {
				// 完成するまでに残りターンが足りない
				continue
			}
		}
		// 建設にかかターン数
		needTurn := len(path) + doNotthingTurn
		// 建築完成のターン
		endTurn := state.turn + needTurn
		// 建築完了後の持ち金
		money := state.money + state.income*(endTurn) - cost
		// 建築完了後の収入
		newIncome := state.income + len(path)
		// t==in.Tでの資金
		lastMoney := money + newIncome*(in.T-endTurn)
		log.Println("lastMoney=", lastMoney, "money=", money, "income=", newIncome)
		if lastMoney < state.money {
			// 最終的な資金が減っている
			continue
		}

		// ここから建築orWait
		log.Println("cost=", cost, "path:", bestPos, "->", nextPos)
		// 最初の一箇所だけ建設済み
		for j := 1; j < len(path); {
			act := Action{Kind: types[j], Y: path[j].Y, X: path[j].X}
			err := state.do(act, in)
			if err == ErrNotEnoughMoney {
				// お金が足りない場合は何もしない
				state.do(Action{Kind: DO_NOTHING}, in)
			} else {
				j++
			}
		}
		if state.turn > 500 {
			break
		}
		if state.turn >= in.T {
			break
		}
	}
	log.Println(state.field.cellString())
	var t int
	for i, a := range state.actions {
		fmt.Print(a)
		t++
		if i == in.T-1 {
			break
		}
	}
	log.Printf("turn=%d\n", t)
	for t < 800 {
		fmt.Println(DO_NOTHING)
		t++
	}
	log.Printf("stations=%d\n", len(state.field.stations))
}

type Input struct {
	N      int   // 縦長 N=50
	M      int   // 人数 50<=M<=1600
	K      int   // 初期資金 11000<=K<=20000
	T      int   // ターン数 T=800
	src    []Pos // 人の初期位置
	dst    []Pos // 人の目的地
	income []int // 人の収入
}

func readInput(re *bufio.Reader) *Input {
	var in Input
	fmt.Fscan(re, &in.N, &in.M, &in.K, &in.T)
	src := make([]Pos, in.M)
	dst := make([]Pos, in.M)
	income := make([]int, in.M)
	for i := 0; i < in.M; i++ {
		fmt.Fscan(re, &src[i].Y, &src[i].X, &dst[i].Y, &dst[i].X)
		income[i] = int(distance(src[i], dst[i]))
	}
	log.Printf("readInput: N=%v, M=%v, K=%v, T=%v\n", in.N, in.M, in.K, in.T)
	in.src = src
	in.dst = dst
	in.income = income
	return &in
}

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	startTime := time.Now()
	log.SetFlags(log.Lshortfile)
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	in := readInput(reader)
	_ = in
	greedy(*in)
	//log.Printf("in=%+v\n", in)
	log.Printf("time=%v\n", time.Since(startTime).Milliseconds())
}

func absInt16(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type UnionFind struct {
	par [2500]int16
}

func (uf *UnionFind) Clone() *UnionFind {
	if uf == nil {
		return nil
	}
	newUF := new(UnionFind)
	newUF.par = uf.par
	return newUF
}

// NewUnionFind は、UnionFindを初期化して返す
// 2500固定
func NewUnionFind() *UnionFind {
	uf := new(UnionFind)
	for i := int16(0); i < 2500; i++ {
		uf.par[i] = i
	}
	return uf
}

func (uf *UnionFind) root(a int16) int16 {
	if uf.par[a] == a {
		return a
	}
	uf.par[a] = uf.root(uf.par[a])
	return uf.par[a]
}

func (uf *UnionFind) same(a, b int16) bool {
	return uf.root(int16(a)) == uf.root(int16(b))
}

func (uf *UnionFind) unite(a, b int16) {
	a = uf.root(a)
	b = uf.root(b)
	if a == b {
		return
	}
	uf.par[a] = b
}

func gridToString(grid [2500]int16) (str string) {
	str = "showGrid()\n"
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			if grid[i*50+j] == 10000 {
				str += "## "
			} else {
				str += fmt.Sprintf("%2d ", grid[i*50+j])
			}
		}
		str += "\n"
	}
	return str
}

// MST用
type Edge struct {
	From, To int
	Cost     int
	Path     []Pos
	Rail     []int16
}

type Graph struct {
	NumNodes int
	Edges    []Edge
}
