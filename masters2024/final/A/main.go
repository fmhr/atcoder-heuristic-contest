package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

// ./a.out -cpuprofile cpu.prof < tools/in/0000.txt > out.txt
// ./a.out -cpuprofile cpu.prof -memprofile mem.prof < tools/in/0000.txt > out.txt
// go tool pprof -http=localhost:8888 main cpu.prof
// go tool pprof -http=localhost:8888 main mem.prof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

// StartCPUProfile は、CPUプロファイルを開始する
func StartCPUProfile() func() {
	if *cpuprofile == "" {
		return func() {}
	}
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		log.Fatal("could not start CPU profile: ", err)
	}

	return func() {
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			log.Fatal("could not stop CPU profile: ", err)
		}
	}
}

// writeMemProfile は、メモリプロファイルを書き込む
func writeMemProfile() {
	if *memprofile == "" {
		return
	}
	f, err := os.Create(*memprofile)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}

func flagCheck() {
	flag.Parse()
	if _, atcoder := os.LookupEnv("ATCODER"); atcoder {
		log.SetOutput(io.Discard)
		return
	}

	log.SetFlags(log.Lshortfile)
	//runtime.GOMAXPROCS(1) // 並列処理を抑制
	//debug.SetGCPercent(2000) // GCを抑制 2000% に設定
	//debug.SetGCPercent(-1) // GCを停止
	//rand.Seed(1) // 乱数のシードを固定することで、デバッグ時に再現性を持たせる
}

func main() {
	flagCheck()
	if *cpuprofile != "" {
		stopCPUProfile := StartCPUProfile()
		defer stopCPUProfile()
	}
	if *memprofile != "" {
		defer writeMemProfile()
	}
	// --- start
	startTime := time.Now()
	solver()
	elapseTime := time.Since(startTime)
	log.Printf("time=%f", float64(elapseTime)/float64(time.Second))
}

type Point struct {
	y, x int
}

func (p Point) distance(q Point) float64 {
	return math.Sqrt(float64((p.y-q.y)*(p.y-q.y) + (p.x-q.x)*(p.x-q.x)))
}

type Vector struct {
	y, x float64
}

type Wall struct {
	y1, x1, y2, x2 int
}

type Input struct {
	N, M           int
	epsilon, delta float64
	start          Point
	checkPoints    []Point
	walls          []Wall
}

func read() (in Input) {
	fmt.Scan(&in.N, &in.M, &in.epsilon, &in.delta)
	fmt.Scan(&in.start.y, &in.start.x)
	for i := 0; i < in.N; i++ {
		var p Point
		fmt.Scan(&p.y, &p.x)
		in.checkPoints = append(in.checkPoints, p)
	}
	for i := 0; i < in.M; i++ {
		var w Wall
		fmt.Scan(&w.y1, &w.x1, &w.y2, &w.x2)
		in.walls = append(in.walls, w)
	}
	log.Printf("epsilon=%f, delta=%f", in.epsilon, in.delta)
	return in
}

type State struct {
	p       Point
	v       Vector
	t       uint64
	visited []bool
}

func (s State) nextPoint(in Input) Point {
	minDistance := math.MaxFloat64
	minPoint := Point{}
	for i := 0; i < in.N; i++ {
		if !s.visited[i] {
			d := s.p.distance(in.checkPoints[i])
			if d < minDistance {
				minDistance = d
				minPoint = in.checkPoints[i]
			}
		}
	}
	return minPoint
}

func solver() {
	in := read()
	log.Println(in)
	var state State
	state.visited = make([]bool, in.N)
	state.p.y = in.start.y
	state.p.x = in.start.x
	var crash int
	var hit int
	var power int = 100
	var preCrach int
	var prepreCrach int
	var preY, preX int
	for i := 0; i < 5000; i++ {
		var hits []int
		target := state.nextPoint(in)
		if state.p.y == target.y && state.p.x == target.x {
			state.p.y += rand.Intn(2000) - 1000
			state.p.x += rand.Intn(2000) - 1000
		}
		//dis := state.p.distance(target)
		//log.Println("now", state.p, "speed", state.v)
		//log.Println("target", target, "distanc", dis, "power", power)
		ay, ax := CalculateAcceleration(state.p, state.v, target, power)
		fmt.Println("A", int(ay), int(ax))
		//log.Println("A", int(ay), int(ax))
		fmt.Scan(&crash, &hit)
		//log.Println(crash, hit)
		for j := 0; j < hit; j++ {
			var q int
			fmt.Scan(&q)
			hits = append(hits, q)
		}
		if crash == 1 {
			power = 500
			state.v.y = 0
			state.v.x = 0
			if preCrach == 1 {
				var yd, xd int
				if state.p.y < 0 {
					yd = state.p.y - 100000
				} else {
					yd = 100000 - state.p.y
				}
				if state.p.x < 0 {
					xd = state.p.x - 100000
				} else {
					xd = 100000 - state.p.x
				}
				if yd < xd {
					state.p.y = 5000
					state.p.x = rand.Intn(200000) - 100000
				} else {
					state.p.x = 5000
					state.p.y = rand.Intn(200000) - 100000
				}
			} else if prepreCrach == 0 {
				state.p.y = state.p.y + rand.Intn(400) - 200
				state.p.x = state.p.x + rand.Intn(400) - 200
			} else {
				state.p.y = preX
				state.p.x = preY
			}
		} else {
			state.v.y += ay
			state.v.x += ax
			state.p.y += int(state.v.y)
			state.p.x += int(state.v.x)
		}
		if hit > 0 {
			power = 100
			log.Println("hit----------------------")
		}
		if crash == 0 && hit == 0 {
			state.p.y = state.p.y + rand.Intn(400) - 200
			state.p.x = state.p.x + rand.Intn(400) - 200
		}
		// visitedの更新
		for j := 0; j < len(hits); j++ {
			state.visited[hits[j]] = true
		}
		state.t++
		prepreCrach = preCrach
		preCrach = crash
		//log.Println(i, state)
		//log.Println(i, "visited", state.visited)
	}
	log.Println(state.visited)
}
func CalculateAcceleration(p Point, v Vector, target Point, power int) (ay, ax float64) {
	dy := float64(target.y - p.y)
	dx := float64(target.x - p.x)

	desiredVy := dy - v.y
	desiredVx := dx - v.x

	distance := math.Sqrt(desiredVy*desiredVy + desiredVx*desiredVx)
	unitY := dy / distance
	unitX := dx / distance

	ay = unitY * float64(power)
	ax = unitX * float64(power)

	if ay+v.y > dy {
		ay = dy
	}
	if ax+v.x > dx {
		ax = dx
	}

	for ay*ay+ax*ax >= 500*500 {
		ay *= 0.9
		ax *= 0.9
	}
	return
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
