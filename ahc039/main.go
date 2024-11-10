package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/bits"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

func isInsice(point Point, plygon []Point) bool {
	intersections := 0
	j := len(plygon) - 1
	for i := 0; i < len(plygon); i++ {
		if (plygon[i].Y > point.Y) != (plygon[j].Y > point.Y) &&
			point.X < (plygon[j].X-plygon[i].X)*(point.Y-plygon[i].Y)/(plygon[j].Y-plygon[i].Y)+plygon[i].X {
			intersections++
		}
		j = i
	}
	return intersections%2 == 1
}

func countPointsInside(polygon []Point, points [N]Point) int {
	count := 0
	for _, point := range points {
		if isInsice(point, polygon) {
			count++
		}
	}
	log.Println("count", count)
	return count
}

func claceScore(ans Ans, in Input) int {
	score := 1
	score += countPointsInside(ans, in.mackerels)
	score -= countPointsInside(ans, in.sardines)
	return score
}

type Ans []Point

func (a Ans) output() {
	fmt.Println(len(a))
	for _, p := range a {
		fmt.Println(p.X, p.Y)
	}
}

func solve(in Input) {
	var ans Ans
	ans = append(ans, Point{X: 76431, Y: 45731})
	ans = append(ans, Point{X: 83820, Y: 45731})
	ans = append(ans, Point{X: 83820, Y: 87777})
	ans = append(ans, Point{X: 70545, Y: 87777})
	ans = append(ans, Point{X: 70545, Y: 33678})
	ans = append(ans, Point{X: 53022, Y: 33678})
	ans = append(ans, Point{X: 53022, Y: 44745})
	ans = append(ans, Point{X: 16230, Y: 44745})
	ans = append(ans, Point{X: 16230, Y: 25693})
	ans = append(ans, Point{X: 76431, Y: 25693})
	ans.output()
	score := claceScore(ans, in)
	log.Println("score", score)
}

type Point struct {
	X, Y uint
}

type Input struct {
	mackerels [N]Point
	sardines  [N]Point
}

const N int = 5000

func input() Input {
	var input Input
	var _N int
	fmt.Scan(&_N)
	for i := 0; i < N; i++ {
		fmt.Scan(&input.mackerels[i].X, &input.mackerels[i].Y)
	}
	for i := 0; i < N; i++ {
		fmt.Scan(&input.sardines[i].X, &input.sardines[i].Y)
	}
	return input
}

var startTime time.Time
var timeLimit time.Duration = 2500 * time.Millisecond

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

	rand.Seed(1)
	startTime = time.Now()
	in := input()
	solve(in)

	elapse := time.Since(startTime)
	log.Printf("time=%v\n", elapse.Seconds())
}

// ------------------------------------------------------------------
// util
// bitArrayを管理するためのセット
const uint64Size = 64
const widthBits = 30
const heightBits = 30
const arraySize = (widthBits*heightBits*uint64Size - 1) / uint64Size

type BitArray [arraySize]uint64

func (b *BitArray) Set(y, x int) {
	if y < 0 || y >= heightBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
	index := y*widthBits + x
	b[index/uint64Size] |= 1 << (index % uint64Size)
}

func (b *BitArray) Unset(y, x int) {
	if y < 0 || y >= heightBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
	index := y*widthBits + x
	b[index/uint64Size] &= ^(1 << (index % uint64Size))
}

func (b *BitArray) Get(y, x int) bool {
	if y < 0 || y >= heightBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
	if y < 0 || y >= widthBits || x < 0 || x >= widthBits {
		panic("out of range")
	}
	index := y*widthBits + x
	return b[index/uint64Size]&(1<<(index%uint64Size)) != 0
}

func (b BitArray) PopCount() (count int) {
	for i := 0; i < arraySize; i++ {
		count += bits.OnesCount64(b[i])
	}
	return count
}

func (b BitArray) XorPopCount(a BitArray) (count int) {
	for i := 0; i < arraySize; i++ {
		count += bits.OnesCount64(b[i] ^ a[i])
	}
	return count
}

func (b *BitArray) Reset() {
	for i := 0; i < arraySize; i++ {
		b[i] = 0
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// findMthCombinatin はm番目の組み合わせが、optionsの中でどれかを復元して返す
// options = [1,2.3.4]
func findMthCombinatin(options []int, length, m int) []int {
	n := len(options)
	var result []int

	for i := 0; i < length; i++ {
		index := m % n
		result = append(result, options[index])
		m /= n
	}
	return result
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
