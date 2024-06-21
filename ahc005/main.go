package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

var (
	x uint32 = 123456789
	y uint32 = 362436069
	z uint32 = 521288629
	w uint32 = 88675123
	t uint32
)

func xorshift() uint32 {
	t = x ^ (x << 11)
	x = y
	y = z
	z = w
	w = w ^ (w >> 19) ^ (t ^ (t >> 8))
	return w
}

func xorshiftn(n int) int {
	return int(xorshift()) % n
}

var dy []int = []int{1, 0, -1, 0}
var dx []int = []int{0, 1, 0, -1}

var sc = bufio.NewScanner(os.Stdin)
var buff []byte

func nextInt() int {
	sc.Scan()
	i, err := strconv.Atoi(sc.Text())
	if err != nil {
		panic(err)
	}
	return i
}
func nextString() string {
	sc.Scan()
	return sc.Text()
}

func init() {
	sc.Split(bufio.ScanWords)
	sc.Buffer(buff, bufio.MaxScanTokenSize*1024)
	log.SetFlags(log.Lshortfile)
}

// timer
var StartTime time.Time

// https://golang.org/pkg/runtime/pprof/
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

var local = flag.Bool("local", false, "if local")

func main() {
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

	// ... rest of the program ...
	StartTime = time.Now()
	input()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

type Point struct {
	y, x int
}

var grid [70]string

type Stage struct {
	pos        Point
	cost       int
	reachedNum int
	reached    [70][70]bool
	root       []byte
}

func newStage(start Point) Stage {
	var s Stage
	s.pos = start
	return s
}

func (s Stage) Duplicate() (r Stage) {
	r.pos = s.pos
	r.cost = s.cost
	r.reachedNum = s.reachedNum
	r.reached = s.reached
	r.root = make([]byte, len(s.root))
	for i := 0; i < len(s.root); i++ {
		r.root[i] = s.root[i]
	}
	return
}

var startP Point

func input() {
	N := nextInt()
	_ = N
	startP.y = nextInt()
	startP.x = nextInt()
	for i := 0; i < N; i++ {
		grid[i] = nextString()
	}
}

func solver() {
	beamSearch()
}

func beamSearch() {
	stages := make([]Stage, 0)
	stage := newStage(startP)
	stages = append(stages, stage)
	//goalse := make([]Stage, 0)
	for {
		// ４方向のどちらにいくか どこまでいくか
		for i := 0; i < len(stages); i++ {
			now := stages[i].Duplicate()
			for d := 0; d < 4; d++ {

			}
		}
	}
}
