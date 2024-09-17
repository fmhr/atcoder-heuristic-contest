package main

import (
	"fmt"
	"log"
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
	x, y int
}

func containsSoda(slice []soda, item soda) bool {
	for _, s := range slice {
		if s.x == item.x && s.y == item.y {
			return true
		}
	}
	return false
}

func readInput() (in Input) {
	_N := 0
	fmt.Scan(&_N)
	for i := 0; i < N; i++ {
		fmt.Scan(&in.sodas[i].x, &in.sodas[i].y)
		in.L = maxInt(in.L, in.sodas[i].x)
		in.L = maxInt(in.L, in.sodas[i].y)
	}
	return in
}

// x = seet, y = carbon とすると、
// x'>=x y'>=y なので、小さいものからつくっていく

func solve(in Input) {
	S := make([]soda, 0, N+1)
	for i := 0; i < N; i++ {
		S = append(S, in.sodas[i])
	}
	ans := make([][4]int, 0)
	for {
		max := int(0)
		maxPos := soda{}
		i_, j_ := -1, -1
		for i := 0; i < len(S); i++ {
			for j := i + 1; j < len(S); j++ {
				x, y := minInt(S[i].x, S[j].x), minInt(S[i].y, S[j].y)
				if max < x+y {
					max = x + y
					maxPos.x, maxPos.y = x, y
					i_, j_ = i, j
				}
			}
		}
		if max > 0 {
			ans = append(ans, [4]int{maxPos.x, maxPos.y, S[i_].x, S[i_].y})
			ans = append(ans, [4]int{maxPos.x, maxPos.y, S[j_].x, S[j_].y})
			S = append(S[:j_], S[j_+1:]...)
			S = append(S[:i_], S[i_+1:]...)
			if !containsSoda(S, maxPos) {
				S = append(S, maxPos)
			}
		} else {
			break
		}
		if len(S) < 10 {
			log.Println(S)
		}
	}
	if len(S) > 0 {
		log.Println(S[0])
		ans = append(ans, [4]int{0, 0, S[0].x, S[0].y})
		S = S[1:]
	}
	log.Println(len(ans))
	fmt.Println(len(ans) - 1)
	for i := len(ans) - 1; i > 0; i-- {
		fmt.Println(ans[i][0], ans[i][1], ans[i][2], ans[i][3])
	}
	return

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
