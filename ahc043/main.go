package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

const (
	COST_STATION = 5000
	COST_RAIL    = 100
)

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
)

var railMap = map[int16]string{
	-1:              ".", // EMPTY
	STATION:         "◎", // STATION
	RAIL_HORIZONTAL: "─", // RAIL_HORIZONTAL
	RAIL_VERTICAL:   "│", // RAIL_VERTICAL
	RAIL_LEFT_DOWN:  "┐", // RAIL_LEFT_DOWN
	RAIL_LEFT_UP:    "┘", // RAIL_LEFT_UP
	RAIL_RIGHT_UP:   "└", // RAIL_RIGHT_UP
	RAIL_RIGHT_DOWN: "┌", // RAIL_RIGHT_DOWN
	7:               "+", // other
}

var buildCost = map[int16]int{
	-1:              0,            // EMPTY
	STATION:         COST_STATION, // STATION
	RAIL_HORIZONTAL: COST_RAIL,
	RAIL_VERTICAL:   COST_RAIL,
	RAIL_LEFT_DOWN:  COST_RAIL,
	RAIL_LEFT_UP:    COST_RAIL,
	RAIL_RIGHT_UP:   COST_RAIL,
	RAIL_RIGHT_DOWN: COST_RAIL,
	7:               0, // other
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

func (f Field) cellType(pos Pos) string {
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

func NewField(n int) *Field {
	f := new(Field)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			f.cell[i][j] = EMPTY
		}
	}
	f.uf = NewUnionFind(n * n)
	return f
}

func (f *Field) build(act Action) {
	if act.Kind == DO_NOTHING {
		return
	}
	if f.cell[act.Y][act.X] != EMPTY {
		panic("already built")
	}
	if act.Kind < 0 || act.Kind > 6 {
		panic("invalid kind:" + fmt.Sprint(act.Kind))
	}
	if act.Kind == STATION {
		f.stations = append(f.stations, Pos{Y: act.Y, X: act.X})
	}
	f.cell[act.Y][act.X] = act.Kind
	// 連結成分をつなげる
	y, x := act.Y, act.X
	kind := act.Kind
	// 上
	if y > 0 {
		switch kind {
		case STATION, RAIL_VERTICAL, RAIL_LEFT_UP, RAIL_RIGHT_UP:
			switch f.cell[y-1][x] {
			case STATION, RAIL_VERTICAL, RAIL_LEFT_DOWN, RAIL_RIGHT_DOWN:
				f.uf.unite(int(y)*50+int(x), int(y-1)*50+int(x))
			}
		}
	}
	// 下
	if y < 49 {
		switch kind {
		case STATION, RAIL_VERTICAL, RAIL_LEFT_DOWN, RAIL_RIGHT_DOWN:
			switch f.cell[y+1][x] {
			case STATION, RAIL_VERTICAL, RAIL_LEFT_UP, RAIL_RIGHT_UP:
				f.uf.unite(int(y)*50+int(x), int(y+1)*50+int(x))
			}
		}
	}
	// 左
	if x > 0 {
		switch kind {
		case STATION, RAIL_HORIZONTAL, RAIL_LEFT_DOWN, RAIL_LEFT_UP:
			switch f.cell[y][x-1] {
			case STATION, RAIL_HORIZONTAL, RAIL_RIGHT_DOWN, RAIL_RIGHT_UP:
				f.uf.unite(int(y)*50+int(x), int(y)*50+int(x-1))
			}
		}
	}
	// 右
	if x < 49 {
		switch kind {
		case STATION, RAIL_HORIZONTAL, RAIL_RIGHT_DOWN, RAIL_RIGHT_UP:
			switch f.cell[y][x+1] {
			case STATION, RAIL_HORIZONTAL, RAIL_LEFT_DOWN, RAIL_LEFT_UP:
				f.uf.unite(int(y)*50+int(x), int(y)*50+int(x+1))
			}
		}
	}
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

// isConnected 駅,路線をつかって	、a,bがつながっているかを返す
func (f Field) isConnected(a, b Pos) bool {
	stations0 := f.collectStations(a)
	stations1 := f.collectStations(b)
	for _, s0 := range stations0 {
		for _, s1 := range stations1 {
			if f.uf.same(int(s0.Y)*50+int(s0.X), int(s1.Y)*50+int(s1.X)) {
				return true
			}
		}
	}
	return false
}

var dy = []int16{-1, 1, 0, 0}
var dx = []int16{0, 0, -1, 1}

// 2点間の最短経路を返す (a から b へ)
func (f *Field) shortestPath(a, b Pos) (path []Pos) {
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
			if f.cell[y][x] != EMPTY {
				continue
			}
			if dist[int(y)*50+int(x)] > dist[int(p.Y)*50+int(p.X)]+1 {
				dist[int(y)*50+int(x)] = dist[int(p.Y)*50+int(p.X)] + 1
				que = append(que, Pos{Y: y, X: x})
			}
		}
	}
	if dist[int(b.Y)*50+int(b.X)] == 10000 {
		log.Println(showGrid(dist))
		log.Println("can't reach", a, b)
		log.Println(dist[int(a.Y)*50+int(a.X)], dist[int(b.Y)*50+int(b.X)])
		f.cell[a.Y][a.X] = 7
		f.cell[b.Y][b.X] = 7
		log.Println(f.cellString())
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
			if f.cell[y][x] != EMPTY {
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

var ErrNotEnoughMoney = fmt.Errorf("not enough money")

type State struct {
	field   *Field
	money   int
	turn    int
	income  int
	actions []Action
}

func NewState(in *Input) *State {
	s := new(State)
	s.field = NewField(in.N)
	s.money = in.K
	return s
}

func (s *State) do(act Action, in Input) error {
	if s.money < buildCost[act.Kind] {
		return ErrNotEnoughMoney
	}
	if act.Kind != DO_NOTHING {
		s.field.build(act)
		s.money -= buildCost[act.Kind]
		if act.Kind == STATION {
			s.income = 0
			for i := 0; i < in.M; i++ {
				if s.field.isConnected(in.src[i], in.dst[i]) {
					s.income += in.income[i]
				}
			}
		}
	}
	s.turn++
	s.money += s.income
	act.comment = fmt.Sprintf("#turn=%d, \n#money=%d, \n#income=%d\n", s.turn, s.money, s.income)
	s.actions = append(s.actions, act)
	return nil
}

type Pos struct {
	Y, X int16
}

func distance(a, b Pos) int16 {
	return absInt16(a.X-b.X) + absInt16(a.Y-b.Y)
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
		if state.field.isConnected(in.src[i], in.dst[i]) {
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
		log.Println(bestPos, state.field.cellType(bestPos), "->", nextPos, state.field.cellType(nextPos))
		path := state.field.shortestPath(bestPos, nextPos)
		if path == nil {
			log.Printf("no path from %v to %v\n", bestPos, nextPos)
			log.Println(state.field.cell[bestPos.Y][bestPos.X], state.field.cell[nextPos.Y][nextPos.X])
			continue
			//panic("no path")
		}
		types := state.field.selectRails(path)
		// 最初の一箇所だけ建設済み
		for i := 1; i < len(path); {
			act := Action{Kind: types[i], Y: path[i].Y, X: path[i].X}
			err := state.do(act, in)
			if err == ErrNotEnoughMoney {
				// お金が足りない場合は何もしない
				state.do(Action{Kind: DO_NOTHING}, in)
			} else {
				i++
			}
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
		if i == 799 {
			break
		}
	}
	log.Println("turn=", t)
	for t < 800 {
		fmt.Println(DO_NOTHING)
		t++
	}
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

func main() {
	startTime := time.Now()
	log.SetFlags(log.Lshortfile)
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	in := readInput(reader)
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
	par []int
}

func NewUnionFind(n int) *UnionFind {
	uf := new(UnionFind)
	uf.par = make([]int, n)
	for i := 0; i < n; i++ {
		uf.par[i] = i
	}
	return uf
}

func (uf *UnionFind) root(x int) int {
	if uf.par[x] == x {
		return x
	}
	uf.par[x] = uf.root(uf.par[x])
	return uf.par[x]
}

func (uf *UnionFind) same(x, y int) bool {
	return uf.root(x) == uf.root(y)
}

func (uf *UnionFind) unite(x, y int) {
	x = uf.root(x)
	y = uf.root(y)
	if x == y {
		return
	}
	uf.par[x] = y
}

func showGrid(grid [2500]int16) (str string) {
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
