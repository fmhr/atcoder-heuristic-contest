package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

type Input struct {
	N, T, sgm int
	w, h      []int
}

func input() Input {
	var n, t, sgm int
	fmt.Scan(&n, &t, &sgm)
	w := make([]int, n)
	h := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&w[i], &h[i])
	}
	return Input{n, t, sgm, w, h}
}

type answer struct {
	p, r int
	d    string
	b    int
}

func solver(in Input) {
	var measured_w, measured_h int
	bestAnses := make([]answer, in.N)

	var bestScore int = math.MaxInt64
	for t := 0; t < in.T-1; t++ {
		fmt.Println(in.N)
		anses := make([]answer, in.N)
		for i := 0; i < in.N; i++ {
			p := i            // 長方形の番号
			r := rand.Intn(2) // 1:90度回転
			d := "U"          // U：下から上に配置 L:右から左に配置
			if rand.Intn(2) == 0 {
				d = "L"
			}
			b := -1
			if i > 0 {
				b = rand.Intn(i) - 1
			}
			fmt.Println(p, r, d, b)
			anses[i] = answer{p, r, d, b}
		}
		fmt.Scan(&measured_w, &measured_h)
		if measured_w+measured_h < bestScore {
			bestScore = measured_w + measured_h
			bestAnses = anses
			log.Println(t, measured_w+measured_h)
		}
	}
	fmt.Println(in.N)
	for i := 0; i < in.N; i++ {
		fmt.Println(bestAnses[i].p, bestAnses[i].r, bestAnses[i].d, bestAnses[i].b)
		log.Println(bestAnses[i].p, bestAnses[i].r, bestAnses[i].d, bestAnses[i].b)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	in := input()
	solver(in)
}
