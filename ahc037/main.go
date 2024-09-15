package main

import (
	"fmt"
	"log"
	"sort"
)

const (
	N = 1000
)

type Input struct {
	sodas [1000]soda
}

type soda struct {
	x, y    int
	parent  int
	created bool
}

func searchMini(u []soda, n int) (p int) {
	s := u[n]
	miniCost := (s.x + s.y) * 2
	for i := 0; i < len(u); i++ {
		if i == n {
			continue
		}
		if s.x >= u[i].x && s.y >= u[i].y {
			cost := s.x - u[i].x + s.y - u[i].y
			if cost < miniCost {
				miniCost = cost
				p = i
			}
		}
	}
	return
}

type ans struct {
	out  [][4]int
	cost int
}

func readInput() (in Input) {
	_N := 0
	fmt.Scan(&_N)
	for i := 0; i < N; i++ {
		fmt.Scan(&in.sodas[i].x, &in.sodas[i].y)
	}
	return in
}

// x = seet, y = carbon とすると、
// x'>=x y'>=y なので、小さいものからつくっていく

func solve(in Input) {
	sort.Slice(in.sodas[:], func(i, j int) bool {
		return in.sodas[i].x < in.sodas[j].x
	})
	S := make([]soda, 0, N+1)
	S = append(S, soda{x: 0, y: 0, created: true})
	for i := 0; i < N; i++ {
		S = append(S, in.sodas[i])
	}

	for i := 1; i < len(S); i++ {
		if S[i].created {
			continue
		}
		S[i].parent = searchMini(S, i)
	}
	var a ans
	var createSoda func(i int)
	createSoda = func(i int) {
		if S[i].created {
			return
		}
		p := S[S[i].parent]
		if p.created {
			//fmt.Println(p.x, p.y, S[i].x, S[i].y)
			S[i].created = true
			a.out = append(a.out, [4]int{p.x, p.y, S[i].x, S[i].y})
			a.cost += S[i].x - p.x + S[i].y - p.y
		} else {
			createSoda(S[i].parent)
		}
	}
	for i := 1; i < N+1; i++ {
		createSoda(i)
	}

	fmt.Println(len(a.out))
	for i := 0; i < len(a.out); i++ {
		fmt.Println(a.out[i][0], a.out[i][1], a.out[i][2], a.out[i][3])
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	in := readInput()
	solve(in)
}

// utils
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
