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

func isIndide(point Point, plygon []Point) bool {
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
		if isIndide(point, polygon) {
			count++
		}
	}
	return count
}

func claceScore(ans Polygon, in Input) int {
	for _, p := range ans {
		if p.X < 0 || p.X > 100000 || p.Y < 0 || p.Y > 100000 {
			return -1000000000
		}
	}
	if isSelfIntersecting(ans) {
		return -1000000000
	}
	score := 1
	score += countPointsInside(ans, in.mackerels)
	score -= countPointsInside(ans, in.sardines)
	return score
}

type Polygon []Point

func (p *Polygon) Copy(src Polygon) {
	*p = make(Polygon, len(src))
	copy(*p, src)
}

func (a Polygon) output() {
	fmt.Println(len(a))
	for _, p := range a {
		fmt.Println(p.X, p.Y)
	}
}

func solve(in Input) {
	var polygon Polygon
	polygon = append(polygon, Point{X: 76431, Y: 45731})
	polygon = append(polygon, Point{X: 83820, Y: 45731})
	polygon = append(polygon, Point{X: 83820, Y: 87777})
	polygon = append(polygon, Point{X: 70545, Y: 87777})
	polygon = append(polygon, Point{X: 70545, Y: 33678})
	polygon = append(polygon, Point{X: 53022, Y: 33678})
	polygon = append(polygon, Point{X: 53022, Y: 44745})
	polygon = append(polygon, Point{X: 16230, Y: 44745})
	polygon = append(polygon, Point{X: 16230, Y: 25693})
	polygon = append(polygon, Point{X: 76431, Y: 25693})
	score := claceScore(polygon, in)

	bestScore := score
	log.Println(score)
	var bestPolygon Polygon
	bestPolygon.Copy(polygon)
	// やきなますぞー
	for loop := 0; loop < 1000; loop++ {
		switter := rand.Intn(4)
		switch switter {
		case 0:
			// 全体を移動させる
			direct := rand.Intn(3)
			distance := rand.Intn(1000)
			for i := 0; i < len(polygon); i++ {
				polygon[i].X += dx[direct] * distance
				polygon[i].Y += dy[direct] * distance
			}

		case 1:
			// 一点を移動させる
			// 頂点を選ぶ
			index := rand.Intn(len(polygon))
			target := polygon[index]
			prevIndex := (index - 1 + len(polygon)) % len(polygon)
			prev := polygon[prevIndex]
			nextIndex := (index + 1) % len(polygon)
			next := polygon[nextIndex]
			if target.X == prev.X && target.Y != prev.Y && target.X != next.X && target.Y == next.Y {
				maxX := maxInt(prev.X, maxInt(next.X, target.X))
				minX := minInt(prev.X, minInt(next.X, target.X))
				maxY := maxInt(prev.Y, maxInt(next.Y, target.Y))
				minY := minInt(prev.Y, minInt(next.Y, target.Y))
				if maxX-minX < 2 || maxY-minY < 2 {
					continue
				}
				newX := rand.Intn(int(maxX-minX)-1) + minX + 1
				newY := rand.Intn(int(maxY-minY)-1) + minY + 1
				polygon[index] = Point{X: newX, Y: newY}
				polygon[prevIndex].X = newX
				polygon[nextIndex].Y = newY
			}
			if target.X != prev.X && target.Y == prev.Y && target.X == next.X && target.Y != next.Y {
				maxX := maxInt(prev.X, maxInt(next.X, target.X))
				minX := minInt(prev.X, minInt(next.X, target.X))
				maxY := maxInt(prev.Y, maxInt(next.Y, target.Y))
				minY := minInt(prev.Y, minInt(next.Y, target.Y))
				if maxX-minX < 2 || maxY-minY < 2 {
					continue
				}
				newX := rand.Intn(int(maxX-minX)-1) + minX + 1
				newY := rand.Intn(int(maxY-minY)-1) + minY + 1
				polygon[index] = Point{X: newX, Y: newY}
				polygon[prevIndex].Y = newY
				polygon[nextIndex].X = newX
			}
		case 2:
			// １辺を移動させる
			index1 := rand.Intn(len(polygon))
			index2 := (index1 + 1) % len(polygon)
			target1 := &polygon[index1]
			target2 := &polygon[index2]
			if target1.X == target2.X {
				target1.X += rand.Intn(1000) - 500
				target2.X = target1.X
			}
			if target1.Y == target2.Y {
				target1.Y += rand.Intn(1000) - 500
				target2.Y = target1.Y
			}
		case 3:
			// 凸を追加する
			index := rand.Intn(len(polygon))
			target := polygon[index]
			next := polygon[(index+1)%len(polygon)]
			var newPoint Point
			if target.X == next.X {
				newPoint = Point{X: target.X, Y: target.Y + next.Y/2}
			} else {
				newPoint = Point{X: target.X + next.X/2, Y: target.Y}
			}
			newPoint2 := Point{X: newPoint.X, Y: newPoint.Y}
			polygon = append(polygon[:index+1], append([]Point{newPoint}, polygon[index+1:]...)...)
			polygon = append(polygon[:index+1], append([]Point{newPoint2}, polygon[index+1:]...)...)
		}

		score := claceScore(polygon, in)
		if score > bestScore {
			bestScore = score
			bestPolygon.Copy(polygon)
			log.Println("update score", score)
			bestPolygon.output()
		} else {
			polygon.Copy(bestPolygon)
		}
	}
	bestPolygon.output()
	log.Println("isSelfIntersectingCount", isSelfIntersectingCount)
}

type Point struct {
	X, Y int
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

	//rand.Seed(1)
	rand.Seed(time.Now().UnixNano())
	startTime = time.Now()
	in := input()
	solve(in)

	elapse := time.Since(startTime)
	log.Printf("time=%v\n", elapse.Seconds())
}

// ------------------------------------------------------------------
// util

var dx = []int{1, 0, -1, 0}
var dy = []int{0, 1, 0, -1}

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

// 線分同士が交差するか判定する関数
func isIntersecting(p1, q1, p2, q2 Point) bool {
	// 線分 p1q1 と p2q2 の方向ベクトルを計算

	// 交差判定 (詳細な説明は後述)
	// 方向ベクトルの外積を計算
	d1 := (q2.X-p2.X)*(p1.Y-p2.Y) - (q2.Y-p2.Y)*(p1.X-p2.X)
	d2 := (q2.X-p2.X)*(q1.Y-p2.Y) - (q2.Y-p2.Y)*(q1.X-p2.X)
	d3 := (q1.X-p1.X)*(p2.Y-p1.Y) - (q1.Y-p1.Y)*(p2.X-p1.X)
	d4 := (q1.X-p1.X)*(q2.Y-p1.Y) - (q1.Y-p1.Y)*(q2.X-p1.X)

	// 交差判定
	return d1*d2 < 0 && d3*d4 < 0
}

var isSelfIntersectingCount [2]int

// 多角形の自己交差を検出する関数
func isSelfIntersecting(polygon Polygon) bool {
	n := len(polygon)
	for i := 0; i < n-1; i++ {
		for j := i + 2; j < n; j++ {
			if isIntersecting(polygon[i], polygon[i+1], polygon[j%n], polygon[((j%n)+1%n)%n]) {
				isSelfIntersectingCount[1]++
				return true
			}
		}
	}
	isSelfIntersectingCount[0]++
	return false
}
