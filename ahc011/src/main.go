package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
)

func main() {
	log.SetFlags(log.Lshortfile)
	loardInput()
	solver()
}

var TileGraph [16]string = [16]string{"@", "┥", "┸", "┛", "┝", "━", "┗", "┻", "┰", "┓", "┃", "┫", "┏", "┳", "┣", "╋"}
var N, T int
var startTile [10]string

func loardInput() {
	fmt.Scan(&N, &T)
	for i := 0; i < N; i++ {
		fmt.Scan(&startTile[i])
	}
}

type Point struct {
	x, y int
}

type Tile struct {
	tile           [10][10]uint
	stackTile      [16]uint
	emptySquare    Point
	operationCount int
	operation      []uint64
}

func (t Tile) WriteTile() (str string) {
	str += "\n"
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			str += TileGraph[t.tile[i][j]]
		}
		str += "\n"
	}
	return
}

func (t *Tile) Swap(a, b Point) error {
	if a.x < 0 || a.x >= N || a.y < 0 || a.y >= N {
		return fmt.Errorf("target is outside of tiles")
	}
	if b.x < 0 || b.x >= N || b.y < 0 || b.y >= N {
		return fmt.Errorf("target is outside of tiles")
	}
	t.tile[a.y][a.x], t.tile[b.y][b.x] = t.tile[b.y][b.x], t.tile[a.y][a.x]
	if t.tile[a.y][a.x] == 0 {
		t.emptySquare = a
	}
	if t.tile[b.y][b.x] == 0 {
		t.emptySquare = b
	}
	return nil
}

const (
	Upward = iota + 1
	Downward
	Leftward
	Rightward
)

var UDLR [5]string = [5]string{"", "U", "D", "L", "R"}

func (t Tile) Output() {
	var str string
	for i := 0; i < len(t.operation); i++ {
		str += UDLR[t.operation[i]]
	}
	fmt.Println(str)
}

func (t *Tile) Move(operation uint64) error {
	if !(operation > 0 && operation < 5) {
		log.Fatalf("operation need 1~4, got %d", operation)
	}
	var target Point
	switch operation {
	case Upward:
		target.x = t.emptySquare.x
		target.y = t.emptySquare.y - 1
	case Downward:
		target.x = t.emptySquare.x
		target.y = t.emptySquare.y + 1
	case Leftward:
		target.x = t.emptySquare.x - 1
		target.y = t.emptySquare.y
	case Rightward:
		target.x = t.emptySquare.x + 1
		target.y = t.emptySquare.y
	}
	err := t.Swap(target, t.emptySquare)
	if err != nil {
		return err
	}
	t.operationCount++
	t.operation = append(t.operation, operation)
	return nil
}

func (t Tile) Score() int {
	uf := NewUnionFind(N * N)
	tree := make([]bool, N*N)
	for i := 0; i < N*N; i++ {
		tree[i] = true
	}
	var tiles [10][10]uint
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			tiles[i][j] = t.tile[i][j]
		}
	}
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if i+1 < N && tiles[i][j]&8 != 0 && tiles[i+1][j]&2 != 0 {
				a := uf.Find(i*N + j)
				b := uf.Find((i+1)*N + j)
				if a == b {
					tree[a] = false
				} else {
					t := tree[a] && tree[b]
					uf.Unit(a, b)
					tree[uf.Find(a)] = t
				}
			}
			if j+1 < N && tiles[i][j]&4 != 0 && tiles[i][j+1]&1 != 0 {
				a := uf.Find(i*N + j)
				b := uf.Find(i*N + j + 1)
				if a == b {
					tree[a] = false
				} else {
					t := tree[a] && tree[b]
					uf.Unit(a, b)
					tree[uf.Find(a)] = t
				}
			}
		}
	}
	maxTree := 0
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if tiles[i][j] != 0 && tree[uf.Find(i*N+j)] {
				if maxTree == 0 || uf.Size(maxTree) < uf.Size(i*N+j) {
					maxTree = i*N + j
				}
			}
		}
	}
	var bs [10][10]bool
	if maxTree != 0 {
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				bs[i][j] = uf.Same(maxTree, i*N+j)
			}
		}
	}
	var score int
	if uf.Size(maxTree) == N*N-1 {
		score = int(500000.0 * (1.0 + float64(T-t.operationCount)/float64(T)))
	} else {
		score = int(500000.0 * float64(uf.Size(maxTree)) / float64(N*N-1))
	}
	log.Printf("ScoreParamater N=%d maxTreeSize=%d\n", N, maxTree)
	return score
}

func moveEmptyToPosition(s Tile) Tile {
	// emptyTileを(N-1,N-1)に移動させる
	for s.emptySquare.x != N-1 {
		err := s.Move(Rightward)
		if err != nil {
			log.Fatal(err)
		}
	}
	for s.emptySquare.y != N-1 {
		err := s.Move(Downward)
		if err != nil {
			log.Fatal(err)
		}
	}
	return s
}

////---------------------------------------------------------------
// コピーコストを抑えるために最小のデータ構造にする
// N は最大10 合計100マス
// 1マス4bit 64bitには16マス
// [7]uint64で112マス 64*7 = 448bit
// uint8を1マスで使った場合 = 800bit

type LightTile struct {
	tiles [100]uint8
	used  [100]bool // swap時に更新する
	next  [100]bool // swap時に更新する
}

func (lt *LightTile) Swap(newPoint, swappedPoint int) {
	lt.tiles[newPoint], lt.tiles[swappedPoint] = lt.tiles[newPoint], lt.tiles[swappedPoint]
	lt.used[newPoint] = true
	lt.next[newPoint] = false
	if lt.tiles[newPoint]&1 != 0 && newPoint%N != 0 && !lt.used[newPoint-1] {
		lt.next[newPoint] = true
	}
	if lt.tiles[newPoint]&2 != 0 && newPoint/N != 0 && !lt.used[newPoint-N] {
		lt.next[newPoint-N] = true
	}
	if lt.tiles[newPoint]&4 != 0 && newPoint%N != N-1 && !lt.used[newPoint+1] {
		lt.next[newPoint+1] = true
	}
	if lt.tiles[newPoint]&8 != 0 && newPoint/N != N-1 && !lt.used[newPoint+N] {
		lt.next[newPoint+N] = true
	}
}

func solver() {
	var s Tile
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if startTile[i][j] == '0' {
				s.emptySquare.y = i
				s.emptySquare.x = j
			}
			n, err := strconv.ParseUint(string(startTile[i][j]), 16, 10)
			if err != nil {
				log.Fatal(err)
			}
			s.tile[i][j] = uint(n)
		}
	}
	//s = random(s)
	//beamSearch(s)
	s = moveEmptyToPosition(s)
	log.Println(s.Score())
	s.Output()
	log.Print(s.WriteTile())
}

func random(s Tile) Tile {
	for i := 0; i < 100; i++ {
		op := rand.Uint64()%4 + 1
		err := s.Move(op)
		if err != nil {
			//log.Println(err)
		}
		//log.Println(i, s.emptySquare, UDLR[op])
	}
	return s
}

func beamSearch(s Tile) {
	// 全てのタイルを取り出して、隅から埋めていく
	// それが構築可能か条件をチェックする
	// NO -> 再beamsearch
	// YES -> 最小手順を探す
}

func swapBeamSearchSolver(t Tile) {
	// 構築可能条件を満たしたswapで木の構築を目指す
	//  空白マスを固定した上で、マスからの距離（の奇遇）が同じ？
	//  generaterは右下にマスを置く
	//   最初に空白マスを右下に置く
	//    偶数回のスワップで木を作る
	//    開きマスはN＊N-1 最後のひとマスは不動なのでひとマスひとスワップで木を完成させる
	// beamSearch
	// 　木を構成する手を全て生成する
	// 最小手順探索
	var lt LightTile
	// copy tile
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			lt.tiles[i*N+j] = uint8(t.tile[i][j])
		}
	}
	lt.used[(N-1)*N+(N-1)] = true

}

func generateNextState(t LightTile) []uint {
	// 1.次のマスの列挙(構築中のツリーの枝の伸びた先のマス)
	// 2.それぞれのマスに対してスワップ可能なマスの列強
}

func necessaryCondition(t LightTile, pos int) (uint8, uint8) {
	//　タイルは任意に伸びていくのでどのタイルが確定しているのかを知っておかないといけない
	//  確定ずみのタイルとは一致していないといけない
	//  確定ずみのタイルで閉路を作ってはいけない
	var a uint8
	var b uint8
	// タイルが端にあるとき
	// 左を見る
	if pos%N == 0 {
		a |= 1
		//  ***1
	} else if t.used[pos-1] {
		a |= 1
		if t.tiles[pos-1]&4 != 0 {
			b |= 1
		}
	}
	// 右を見る
	if pos%N == N-1 {
		a |= 4
	} else if t.used[pos+1] {
		a |= 4
		if t.tiles[pos+1]&1 != 0 {
			b |= 4
		}
	}
	// 上を見る
	if pos/N == 0 {
		a |= 2
	} else if t.used[pos-N] {
		a |= 2
		if t.tiles[pos-N]&8 != 0 {
			b |= 2
		}
	}
	// 下を見る
	if pos/N == N-1 {
		a |= 8
	} else if t.used[pos+N] {
		a |= 8
		if t.tiles[pos+N]&2 != 0 {
			b |= 8
		}
	}
	log.Println("TODO: 複数のつながりがあった場合閉路をうむ")
	log.Println(strconv.FormatUint(uint64(a), 2), strconv.FormatUint(uint64(b), 2))
	return a, b
}

//////////////////////////////////////////////

type UnionFind struct {
	root []int
	size []int
	link [][]int
}

func NewUnionFind(size int) *UnionFind {
	var uf UnionFind
	uf.root = make([]int, size)
	uf.size = make([]int, size)
	uf.link = make([][]int, size)
	for i := 0; i < size; i++ {
		uf.link[i] = make([]int, 1)
		uf.link[i][0] = i
	}
	for i := 0; i < size; i++ {
		uf.root[i] = i
		uf.size[i] = 1
	}
	return &uf
}

func (uf *UnionFind) Find(a int) int {
	for uf.root[a] != a {
		uf.root[a] = uf.root[uf.root[a]]
		a = uf.root[a]
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
		uf.root[a] = b
		uf.size[b] += uf.size[a]
		uf.link[b] = append(uf.link[b], uf.link[a]...)
	}
}

func (uf *UnionFind) Same(a, b int) bool {
	return uf.Find(a) == uf.Find(b)
}

func (uf *UnionFind) Size(a int) int {
	return uf.size[uf.Find(a)]
}
