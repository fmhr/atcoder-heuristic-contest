package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
)

var sc = bufio.NewScanner(os.Stdin)
var buff []byte

func nextString() string {
	sc.Scan()
	return sc.Text()
}

func nextFloat64() float64 {
	sc.Scan()
	f, err := strconv.ParseFloat(sc.Text(), 64)
	if err != nil {
		panic(err)
	}
	return f
}

func nextInt() int {
	sc.Scan()
	i, err := strconv.Atoi(sc.Text())
	if err != nil {
		panic(err)
	}
	return i
}

func nextInts(n int) (r []int) {
	r = make([]int, n)
	for i := 0; i < n; i++ {
		r[i] = nextInt()
	}
	return r
}

var dy = []int{0, 1, 0, -1}
var dx = []int{1, 0, -1, 0}
var MAX = math.MaxInt64

func maxInt(a ...int) int {
	r := a[0]
	for i := 0; i < len(a); i++ {
		if r < a[i] {
			r = a[i]
		}
	}
	return r
}
func minInt(a ...int) int {
	r := a[0]
	for i := 0; i < len(a); i++ {
		if r > a[i] {
			r = a[i]
		}
	}
	return r
}
func sum(a []int) (r int) {
	for i := range a {
		r += a[i]
	}
	return r
}
func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func init() {
	sc.Split(bufio.ScanWords)
	sc.Buffer(buff, bufio.MaxScanTokenSize*1024)
	log.SetFlags(log.Lshortfile)
}

func main() {
	solver()
}

const N int = 400
const M int = 1995

const R int = 100

//全てのオフィスが専用回線によって連結となるようにしたい。

type Vertice struct {
	x, y int
}

type Edge struct {
	index int
	u, v  int
	l     float64 // 正確な距離[d,3d] ユークリッド距離dで初期化
	rnd   float64 // ランダムで生成する距離[d,3d]
	cnt   int
}

func sortByRnd(edges []Edge) {
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].rnd < edges[j].rnd
	})
}

func sortByCnt(edges []Edge) {
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].cnt > edges[j].cnt
	})
}

func (e *Edge) SetRand() {
	min := e.l * 1.37
	max := e.l * 2.87
	e.rnd = math.Round(rand.Float64()*(max-min) + min)
}

func distance(a, b Vertice) (d float64) {
	return math.Round(math.Sqrt(float64((a.x-b.x)*(a.x-b.x) + (a.y-b.y)*(a.y-b.y))))
}

func solver() {
	var vertices [N]Vertice
	for i := 0; i < N; i++ {
		vertices[i].x = nextInt()
		vertices[i].y = nextInt()
	}
	var edges [M]Edge
	for i := 0; i < M; i++ {
		edges[i].u = nextInt()
		edges[i].v = nextInt()
		edges[i].index = i
		u := vertices[edges[i].u]
		v := vertices[edges[i].v]
		d := distance(u, v)
		edges[i].l = d
	}
	var rm [R][M]Edge
	for i := 0; i < R; i++ {
		for e := 0; e < M; e++ {
			rm[i][e] = edges[e]
			rm[i][e].SetRand()
		}
		e := rm[i][:]
		sortByRnd(e)
	}
	// MST pre cost
	//plan := kruskal4(edges, sortByD2)
	//plan := randomTrialN(edges, 150)
	var plan [M]int
	now := makeUnionFind(N)
	cost := 0.0
	// -----------------------------------------------
	// edge iの実距離(di<=li<=3di)
	for i := 0; i < M; i++ {
		li := nextFloat64()
		edges[i].l = li
		// すでに辺の両端がべつの辺によってつながっているときは0
		if !now.Same(edges[i].u, edges[i].v) {
			s, t := now.Find(edges[i].u), now.Find(edges[i].v)
			avg := 0.0
			for j := 0; j < R; j++ {
				uf2 := now.Clone()
				for k := 0; k < M; k++ {
					e := rm[j][k]
					if e.index > i && !uf2.Same(e.u, e.v) {
						uf2.Unit(e.u, e.v)
						if uf2.Same(s, t) {
							avg += e.rnd
							break
						}
					}
				}
				if !uf2.Same(s, t) {
					avg = 1000000000
					break
				}
			}
			if avg*0.9 >= edges[i].l*float64(R) {
				plan[i] = 1
			}

		}
		if plan[i] == 1 {
			now.Unit(edges[i].u, edges[i].v)
			cost += edges[i].l
			fmt.Println("1")
		} else {
			fmt.Println("0")
		}
	}
	// log.Println(cnt1, cnt2)
	// log.Println(plan)
	_, B := kruskal2(edges) //
	A := calcDistance(plan, edges)
	log.Println(A, B)
	log.Printf("Score=%.0f\n", math.Round(100000000*B/A))
	log.Printf("Cost=%.0f\n", cost)
}

func calcDistance(plan [M]int, edges [M]Edge) (sum float64) {
	for i := 0; i < M; i++ {
		if plan[i] == 1 {
			sum += edges[i].l
		}
	}
	return sum
}

// ランダム生成したグラフでMSTを繰り返し試行して、ある辺が何度使われたを数える
// 辺の長さの代わりに使用回数でsortしたkruslal法を使う

func randomTrialN(edges [M]Edge, n int) [M]int {
	for i := 0; i < n; i++ {
		for j := 0; j < M; j++ {
			edges[j].SetRand()
		}
		used := kruskal4(edges, sortByRnd)
		for j := 0; j < M; j++ {
			edges[j].cnt += used[j]
		}
	}
	return kruskal4(edges, sortByCnt)
}

func kruskal4(e [M]Edge, srt func([]Edge)) (used [M]int) {
	edges := e[:]
	srt(edges)
	uf := makeUnionFind(N)
	for i := 0; i < M; i++ {
		e := edges[i]
		if !uf.Same(e.u, e.v) {
			uf.Unit(e.u, e.v)
			used[e.index] = 1
		}
	}
	return
}

// 途中から構成する
func kruskal3(e []Edge, uf UnionFind, sum float64) ([M]int, [M]float64, bool) {
	edges := make([]Edge, len(e))
	copy(edges, e)
	var used [M]int
	var dist [M]float64
	var success bool
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].l < edges[j].l
	})
	for i := 0; i < len(edges); i++ {
		e := edges[i]
		if !uf.Same(e.u, e.v) {
			uf.Unit(e.u, e.v)
			sum += e.l
			used[e.index] = 1
		}
		dist[e.index] = sum
	}
	success = uf.size[uf.par[21]] == N
	return used, dist, success
}

// lを使った最適解の距離
func kruskal2(e [M]Edge) (used [M]int, res float64) {
	edges := e[:]
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].l < edges[j].l
	})
	uf := makeUnionFind(N)
	for i := 0; i < M; i++ {
		e := edges[i]
		if !uf.Same(e.u, e.v) {
			uf.Unit(e.u, e.v)
			res += e.l
			used[e.index] = 1
		}
	}
	return
}

type UnionFind struct {
	par  [N]int
	size [N]int
}

func (uf UnionFind) Clone() UnionFind {
	var rtn UnionFind
	rtn.par = uf.par
	rtn.size = uf.size
	return rtn
}

func makeUnionFind(count int) (uf UnionFind) {
	for i := 0; i < count; i++ {
		uf.par[i] = i
		uf.size[i] = 1
	}
	return uf
}

func (uf *UnionFind) Find(a int) int {
	for uf.par[a] != a {
		uf.par[a] = uf.par[uf.par[a]]
		a = uf.par[a]
	}
	return a
}

func (uf *UnionFind) Unit(a, b int) {
	a = uf.Find(a)
	b = uf.Find(b)
	if a != b {
		if uf.size[a] > uf.size[b] {
			a, b = b, a
		}
		uf.par[a] = b
		uf.size[b] += uf.size[a]
	}
}

func (uf *UnionFind) Same(a, b int) bool {
	return uf.Find(a) == uf.Find(b)
}
