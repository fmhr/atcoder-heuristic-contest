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
	_N    int
	sodas [1000]soda
}

type soda struct {
	x, y   int
	parent *soda
}

func searchMini(u []soda, s soda) (p *soda) {
	miniCost := s.x + s.y + 1
	for i := 0; i < len(u); i++ {
		if s.x >= u[i].x && s.y >= u[i].y {
			cost := u[i].x + u[i].y
			if cost < miniCost {
				miniCost = cost
				p = &u[i]
			}
		}
	}
	return
}

func readInput() (in Input) {
	fmt.Scan(&in._N)
	for i := 0; i < in._N; i++ {
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
	for i := 0; i < in._N; i++ {
		p := searchMini(in.sodas[:i], in.sodas[i])
		in.sodas[i].parent = p
		log.Println(in.sodas[i], p)
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
