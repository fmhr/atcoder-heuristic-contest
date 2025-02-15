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
	EMPTY           = -1
	DO_NOTHING      = -1
	STATION         = 0
	RAIL_HORIZONTAL = 1
	RAIL_VERTICAL   = 2
	RAIL_LEFT_DOWN  = 3
	RAIL_LEFT_UP    = 4
	RAIL_RIGHT_UP   = 5
	RAIL_RIGHT_DOWN = 6
	COST_STATION    = 5000
	COST_RAIL       = 100
)

type Field struct {
	rail [50][50]int
	uf   *UnionFind
}

func NewField(n int) *Field {
	f := new(Field)
	f.uf = NewUnionFind(n * n)
	return f
}

type Pos struct {
	X, Y int16
}

func distance(a, b Pos) int16 {
	return absInt16(a.X-b.X) + absInt16(a.Y-b.Y)
}

type Input struct {
	N      int     // 縦長 N=50
	M      int     // 人数 50<=M<=1600
	K      int     // 初期資金 11000<=K<=20000
	T      int     // ターン数 T=800
	src    []Pos   // 人の初期位置
	dst    []Pos   // 人の目的地
	income []int16 // 人の収入
}

func readInput(re *bufio.Reader) *Input {
	var in Input
	fmt.Fscan(re, &in.N, &in.M, &in.K, &in.T)
	src := make([]Pos, in.M)
	dst := make([]Pos, in.M)
	income := make([]int, in.M)
	for i := 0; i < in.M; i++ {
		fmt.Fscan(re, &src[i].X, &src[i].Y, &dst[i].X, &dst[i].Y)
		income[i] = int(distance(src[i], dst[i]))
		log.Println(src[i], dst[i], income[i])
	}
	log.Printf("N=%v, M=%v, K=%v, T=%v\n", in.N, in.M, in.K, in.T)
	return &in
}

func greedy(in *Input) {
	points := make([][5]int16, in.M)
	// income が小さい順にソート
	for i := 0; i < in.M; i++ {
		points[i] = [5]int16{in.src[i].X, in.src[i].Y, in.dst[i].X, in.dst[i].Y, in.income[i]}
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i][4] < points[j][4]
	})

}

func main() {
	startTime := time.Now()
	log.SetFlags(log.Lshortfile)
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	in := readInput(reader)
	log.Printf("in=%+v\n", in)
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
