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
	-1:              " ", // EMPTY
	STATION:         "◎", // STATION
	RAIL_HORIZONTAL: "─", // RAIL_HORIZONTAL
	RAIL_VERTICAL:   "│", // RAIL_VERTICAL
	RAIL_LEFT_DOWN:  "┐", // RAIL_LEFT_DOWN
	RAIL_LEFT_UP:    "┘", // RAIL_LEFT_UP
	RAIL_RIGHT_UP:   "└", // RAIL_RIGHT_UP
	RAIL_RIGHT_DOWN: "┌", // RAIL_RIGHT_DOWN
	7:               "+", // other
}

var railCost = map[int16]int{
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
	Kind int16
	Y, X int16
}

func (a Action) String() string {
	return fmt.Sprintf("%d %d %d", a.Kind, a.Y, a.X)
}

type Field struct {
	cell [50][50]int16
	uf   *UnionFind
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
	if f.cell[act.Y][act.X] != EMPTY {
		panic("already built")
	}
	if act.Kind < 0 || act.Kind > 6 {
		panic("invalid kind:" + fmt.Sprint(act.Kind))
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

// 2点間の最短経路を返す (a から b へ)
// a, bを含む経路を返す
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
		if p == b {
			break
		}
		// 上
		if p.Y > 0 && f.cell[p.Y-1][p.X] == EMPTY {
			if dist[int(p.Y-1)*50+int(p.X)] > dist[int(p.Y)*50+int(p.X)]+1 {
				dist[int(p.Y-1)*50+int(p.X)] = dist[int(p.Y)*50+int(p.X)] + 1
				que = append(que, Pos{Y: p.Y - 1, X: p.X})
			}
		}
		// 下
		if p.Y < 49 && f.cell[p.Y+1][p.X] == EMPTY {
			if dist[int(p.Y+1)*50+int(p.X)] > dist[int(p.Y)*50+int(p.X)]+1 {
				dist[int(p.Y+1)*50+int(p.X)] = dist[int(p.Y)*50+int(p.X)] + 1
				que = append(que, Pos{Y: p.Y + 1, X: p.X})
			}
		}
		// 左
		if p.X > 0 && f.cell[p.Y][p.X-1] == EMPTY {
			if dist[int(p.Y)*50+int(p.X-1)] > dist[int(p.Y)*50+int(p.X)]+1 {
				dist[int(p.Y)*50+int(p.X-1)] = dist[int(p.Y)*50+int(p.X)] + 1
				que = append(que, Pos{Y: p.Y, X: p.X - 1})
			}
		}
		// 右
		if p.X < 49 && f.cell[p.Y][p.X+1] == EMPTY {
			if dist[int(p.Y)*50+int(p.X+1)] > dist[int(p.Y)*50+int(p.X)]+1 {
				dist[int(p.Y)*50+int(p.X+1)] = dist[int(p.Y)*50+int(p.X)] + 1
				que = append(que, Pos{Y: p.Y, X: p.X + 1})
			}
		}
	}
	if dist[int(b.Y)*50+int(b.X)] == 10000 {
		return nil
	}
	// b から a への経路を復元
	path = append(path, b)
	for path[len(path)-1] != a {
		p := path[len(path)-1]
		// 上
		if p.Y > 0 && f.cell[p.Y-1][p.X] == EMPTY {
			if dist[int(p.Y-1)*50+int(p.X)] == dist[int(p.Y)*50+int(p.X)]-1 {
				path = append(path, Pos{Y: p.Y - 1, X: p.X})
				continue
			}
		}
		// 下
		if p.Y < 49 && f.cell[p.Y+1][p.X] == EMPTY {
			if dist[int(p.Y+1)*50+int(p.X)] == dist[int(p.Y)*50+int(p.X)]-1 {
				path = append(path, Pos{Y: p.Y + 1, X: p.X})
				continue
			}
		}
		// 左
		if p.X > 0 && f.cell[p.Y][p.X-1] == EMPTY {
			if dist[int(p.Y)*50+int(p.X-1)] == dist[int(p.Y)*50+int(p.X)]-1 {
				path = append(path, Pos{Y: p.Y, X: p.X - 1})
				continue
			}
		}
		// 右
		if p.X < 49 && f.cell[p.Y][p.X+1] == EMPTY {
			if dist[int(p.Y)*50+int(p.X+1)] == dist[int(p.Y)*50+int(p.X)]-1 {
				path = append(path, Pos{Y: p.Y, X: p.X + 1})
				continue
			}
		}
		log.Println(path)
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

// レールを建設する
func (s *State) buildRail(act Action) error {
	if s.money < COST_RAIL {
		return ErrNotEnoughMoney
	}
	s.field.build(act)
	s.money -= COST_RAIL
	s.turn++
	s.actions = append(s.actions, act)
	return nil
}

// 駅を建設する
func (s *State) buildStation(y, x int16) error {
	if s.money < COST_STATION {
		return ErrNotEnoughMoney
	}
	act := Action{Kind: STATION, Y: y, X: x}
	s.field.build(act)
	s.money -= COST_STATION
	s.turn++
	s.actions = append(s.actions, act)
	return nil
}

// 何もしない
func (s *State) doNothing() {
	s.turn++
	s.actions = append(s.actions, Action{Kind: DO_NOTHING})
}

type Pos struct {
	Y, X int16
}

func distance(a, b Pos) int16 {
	return absInt16(a.X-b.X) + absInt16(a.Y-b.Y)
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

func greedy(in Input) {
	p := make([][2]int, in.M)
	// income(駅館距離) が小さい順にソート
	for i := 0; i < in.M; i++ {
		p[i] = [2]int{i, int(in.income[i])}
	}
	sort.Slice(p, func(i, j int) bool {
		return p[i][1] < p[j][1]
	})
	// 駅の配置
	state := NewState(&in)
	path := state.field.shortestPath(in.src[p[0][0]], in.dst[p[0][0]])
	if path == nil {
		panic("no path")
	}
	log.Println(path)
	types := state.field.selectRails(path)
	log.Println(types)
	for i := 0; i < len(path); i++ {
		act := Action{Kind: types[i], Y: path[i].Y, X: path[i].X}
		state.buildRail(act)
		log.Println(state)
		fmt.Println(STATION, path[i].Y, path[i].X)
		in.T--
	}
	for ; in.T > 0; in.T-- {
		fmt.Println(DO_NOTHING)
	}
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
	log.Printf("time=%v\n", time.Since(startTime))
}

func absInt16(x int16) int16 {
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
