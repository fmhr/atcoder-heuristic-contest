package main

import (
	"fmt"
	"log"
	"sort"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	log.Println("Hello, World!")
	in := readInput()
	log.Printf("M=%d L=%d W=%d\n", in.M, in.L, in.W)
	solver(in)
}

type Point struct {
	Y, X int
}

type City struct {
	Point
	ID   int
	area [4]int
}

func solver(in Input) {
	var cities [N]City
	for i := 0; i < N; i++ {
		cities[i].ID = i
		cities[i].Y = (in.lxrxlyry[i*4+2] + in.lxrxlyry[i*4+3]) / 2
		cities[i].X = (in.lxrxlyry[i*4+0] + in.lxrxlyry[i*4+1]) / 2
		cities[i].area[0] = in.lxrxlyry[i*4+0]
		cities[i].area[1] = in.lxrxlyry[i*4+1]
		cities[i].area[2] = in.lxrxlyry[i*4+2]
		cities[i].area[3] = in.lxrxlyry[i*4+3]
	}
	sortedCity := cities[:]
	sort.Slice(sortedCity, func(i, j int) bool {
		if sortedCity[i].Y == sortedCity[j].Y {
			return sortedCity[i].X < sortedCity[j].X
		}
		return sortedCity[i].Y < sortedCity[j].Y
	})
	gropes := make([][]int, in.M)
	index := 0
	for i := 0; i < in.M; i++ {
		gropes[i] = make([]int, in.G[i])
		for j := 0; j < in.G[i]; j++ {
			gropes[i][j] = sortedCity[index].ID
			index++
		}
		log.Println(i, in.G[i], gropes[i])
	}
	// クエリの終了
	fmt.Println("!")

}

const (
	N = 800 // 都市の個数
	Q = 400 // クエリの個数
)

type Input struct {
	M        int        // 都市のグループの数 1<= M <= 400
	L        int        // クエリの都市の最大数 1<= L <= 15
	W        int        //　二次元座標の最大値 500 <= W <= 2500
	G        [400]int   // 各グループの都市の数 1<= G[i] <= N(800) i= 0..M-1
	lxrxlyry [N * 4]int // 各都市の座標 0 <= lxrxlyry[i] <= W
	// lxrxlyry[i] = (lx, rx, ly, ry) i=0..N-1
}

// 固定入力はとばす
func readInput() (in Input) {
	var n, q int
	fmt.Scan(&n, &in.M, &q, &in.L, &in.W)
	for i := 0; i < in.M; i++ {
		fmt.Scan(&in.G[i])
	}
	for i := 0; i < N*4; i++ {
		fmt.Scan(&in.lxrxlyry[i])
	}
	return in
}

type UnionFind struct {
	parent []int // 親ノードのインデックス
}

func NewUnionFind(n int) *UnionFind {
	uf := &UnionFind{
		parent: make([]int, n),
	}
	for i := 0; i < n; i++ {
		uf.parent[i] = i
	}
	return uf
}
func (uf *UnionFind) Find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.Find(uf.parent[x])
	}
	return uf.parent[x]
}
func (uf *UnionFind) Union(x, y int) {
	rootX := uf.Find(x)
	rootY := uf.Find(y)
	if rootX != rootY {
		uf.parent[rootY] = rootX
	}
}
func (uf *UnionFind) Same(x, y int) bool {
	return uf.Find(x) == uf.Find(y)
}

type Edge struct {
	From, To int
	Weight   int
}
type Edges []Edge

func (e Edges) Len() int {
	return len(e)
}
func (e Edges) Less(i, j int) bool {
	return e[i].Weight < e[j].Weight
}
func (e Edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func Kruskal(n int, edges Edges) (int, []Edge) {
	uf := NewUnionFind(n)
	sort.Sort(edges)
	var mst []Edge
	mstWeight := 0
	for _, edge := range edges {
		if !uf.Same(edge.From, edge.To) {
			uf.Union(edge.From, edge.To)
			mst = append(mst, edge)
			mstWeight += edge.Weight
		}
	}
	return mstWeight, mst
}
