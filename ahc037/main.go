package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"
)

const (
	N = 1000
)

type Input struct {
	sodas [1000]soda
	L     int
}

type soda struct {
	x, y     int
	parent   int
	children []int
	cost     int // parentからのコスト
	created  bool
	required bool
}

func searchMini(u []soda, n int) (p int) {
	if n == 0 {
		log.Println("n=0", u[0])
	}
	if u[n].x == 0 && u[n].y == 0 {
		return -1
	}
	c := u[n]
	miniCost := int(math.MaxInt64)
	for i := 0; i < len(u); i++ {
		if i == n {
			continue
		}
		if c.x >= u[i].x && c.y >= u[i].y {
			cost := (c.x - u[i].x) + (c.y - u[i].y)
			if cost < miniCost {
				miniCost = cost
				p = i
				if n == 0 {
					log.Println(miniCost, p, u[p])
				}
			}
		}
	}
	return
}

type ans struct {
	out  [][4]int
	cost int
}

func (a ans) Score(L int) int {
	return int(math.Round(1000000 * (float64((N * L)) / float64(1+a.cost))))
}

func readInput() (in Input) {
	_N := 0
	fmt.Scan(&_N)
	for i := 0; i < N; i++ {
		fmt.Scan(&in.sodas[i].x, &in.sodas[i].y)
		in.L = maxInt(in.L, in.sodas[i].x)
		in.L = maxInt(in.L, in.sodas[i].y)
		in.sodas[i].required = true
	}
	return in
}

// x = seet, y = carbon とすると、
// x'>=x y'>=y なので、小さいものからつくっていく

func solve(in Input) {
	sort.Slice(in.sodas[:], func(i, j int) bool {
		return in.sodas[i].x+in.sodas[i].y > in.sodas[j].x+in.sodas[j].y
	})

	used := map[[2]int]bool{}
	used[[2]int{0, 0}] = true
	S := make([]soda, 0, N+1)
	S = append(S, soda{x: 0, y: 0, created: true})
	for i := 0; i < N; i++ {
		S = append(S, in.sodas[i])
		used[[2]int{in.sodas[i].x, in.sodas[i].y}] = true
	}

	for i := 0; i < len(S); i++ {
		if !S[i].required {
			continue
		}
		min_cost := int(math.MaxInt32)
		var c soda
		for j := 0; j < len(S); j++ {
			if i == j {
				continue
			}
			a := S[i]
			b := S[j]
			if a.x == b.x || a.y == b.y {
				continue
			}
			x, y := minInt(a.x, b.x), minInt(a.y, b.y)
			// 中間点がa, bのどちらかと同じ座標だったらスキップ
			if (a.x == x && a.y == y) || (b.x == x && b.y == y) {
				continue
			}
			// すでに使われていたらスキップ
			if _, ok := used[[2]int{x, y}]; ok {
				continue
			}
			cost := (a.x - x) + (a.y - y) + (b.x - x) + (b.y - y)
			if min_cost > cost {
				// 追加
				min_cost = cost
				c.x, c.y = x, y
				used[[2]int{x, y}] = true
			}
		}
		if min_cost == int(math.MaxInt32) {
			continue
		}
		//log.Println("add", c.y+c.x)
		S = append(S, c)
		sort.Slice(S, func(i, j int) bool {
			return S[i].x+S[i].y > S[j].x+S[j].y
		})
	}
	log.Println(len(S))
	for i := 0; i < len(S); i++ {
		p := searchMini(S, i)
		S[i].parent = p
		//S[p].children = append(S[p].children, i)
	}

	var a ans
	var createSoda func(i int)
	createSoda = func(i int) {
		if S[i].created || S[i].parent == -1 {
			return
		}
		p := S[S[i].parent]
		if !p.created {
			createSoda(S[i].parent)
		}
		a.out = append(a.out, [4]int{p.x, p.y, S[i].x, S[i].y})
		a.cost += S[i].x - p.x + S[i].y - p.y
		S[i].created = true
	}
	for i := 0; i < len(S); i++ {
		if S[i].required {
			createSoda(i)
		}
	}
	log.Println(len(a.out), a.cost, a.Score(in.L))
	fmt.Println(len(a.out))
	log.Printf("point=%d\n", len(a.out))
	for i := 0; i < len(a.out); i++ {
		fmt.Println(a.out[i][0], a.out[i][1], a.out[i][2], a.out[i][3])
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	startTime := time.Now()
	in := readInput()
	solve(in)
	elapsedTime := time.Since(startTime)
	log.Printf("elapsedT=%v\n", elapsedTime)
}

// utils
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

type Point struct {
	x, y int
}
