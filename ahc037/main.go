package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	//"slices" // "golang.org/x/exp/slices"
	"golang.org/x/exp/slices"
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

func solve(in Input) {
	S := newSetSoda()
	for _, s := range in.sodas {
		S.append(s)
	}
	ans := make([][4]int, 0, 2000)
	for {
		max := int(0)
		maxPos := soda{}
		i_, j_ := -1, -1
		for i := range S.s {
			for j := i + 1; j < len(S.s); j++ {
				if i == j {
					continue
				}
				x, y := minInt(S.s[i].x, S.s[j].x), minInt(S.s[i].y, S.s[j].y)
				if max < x+y {
					max = x + y
					maxPos.x, maxPos.y = x, y
					i_, j_ = i, j
				}
			}
		}
		if max > 0 {
			ans = append(ans, [4]int{maxPos.x, maxPos.y, S.s[i_].x, S.s[i_].y})
			ans = append(ans, [4]int{maxPos.x, maxPos.y, S.s[j_].x, S.s[j_].y})
			S.deleteIndex(j_) // i<j なので、jを先に削除しないとjがずれる
			S.deleteIndex(i_)
			S.append(maxPos)
		} else {
			break
		}
	}
	for _, k := range S.s {
		ans = append(ans, [4]int{0, 0, k.x, k.y})
	}
	fmt.Printf("%d\n", len(ans))
	//slices.Reverse(ans)
	ReverseSlice(ans)
	for _, a := range ans {
		fmt.Printf("%d %d %d %d\n", a[0], a[1], a[2], a[3])
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		log.SetOutput(io.Discard)
	}
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	in := readInput()
	startTime := time.Now()
	solve(in)
	elapsedTime := time.Since(startTime)
	log.Printf("time=%v\n", elapsedTime)

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

func ReverseSlice(a [][4]int) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

type setSoda struct {
	s []soda
}

func newSetSoda() *setSoda {
	return &setSoda{
		s: make([]soda, 0, 1000),
	}
}

func (s *setSoda) append(x soda) {
	if slices.Contains(s.s, x) {
		return
	}
	s.s = append(s.s, x)
}

func (s *setSoda) delete(x soda) {
	for i := range s.s {
		if s.s[i] == x {
			s.s = append(s.s[:i], s.s[i+1:]...)
			return
		}
	}
}

func (s *setSoda) deleteIndex(i int) {
	s.s = append(s.s[:i], s.s[i+1:]...)
}
