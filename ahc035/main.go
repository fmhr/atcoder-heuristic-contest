package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
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
	seeds := input()
	solver(seeds)
}

const (
	N int = 6
	M int = 15
	T int = 10
	// 2N(N-1)=60
)

func input() [60]Seed {
	var N, M, T int
	var seeds [60]Seed
	fmt.Scan(&N, &M, &T)
	for i := 0; i < 60; i++ {
		for j := 0; j < M; j++ {
			fmt.Scan(&seeds[i][j])
		}
	}
	return seeds
}

func loadSeed() (ns [60]Seed) {
	for i := 0; i < 60; i++ {
		for j := 0; j < M; j++ {
			fmt.Scan(&ns[i][j])
		}
	}
	return ns
}

type Seed [M]int

type State struct {
	seeds [60]Seed
	turn  int
}

func (s State) Score() (score int) {
	for i := 0; i < M; i++ {
		score += maxV(s.seeds[i])
	}
	return
}

func (s *State) generate(grid [N][N]int) {
	var newSeeds [60]Seed
	cnt := 0
	// 左右方向の２ペア
	for i := 0; i < N; i++ {
		for j := 0; j < N-1; j++ {
			var new Seed
			s1, s2 := grid[i][j], grid[i][j+1]
			for k := 0; k < M; k++ {
				if rand.Intn(2) == 0 {
					new[k] = s.seeds[s1][k]
				} else {
					new[k] = s.seeds[s2][k]
				}
			}
			newSeeds[cnt] = new
			cnt++
		}
	}
	// 上下方向の２ペア
	for i := 0; i < N-1; i++ {
		for j := 0; j < N; j++ {
			var new Seed
			s1, s2 := grid[i][j], grid[i+1][j]
			for k := 0; k < M; k++ {
				if rand.Intn(2) == 0 {
					new[k] = s.seeds[s1][k]
				} else {
					new[k] = s.seeds[s2][k]
				}
			}
			newSeeds[cnt] = new
			cnt++
		}
	}
	//log.Println(cnt, "== 60")
	s.seeds = newSeeds
	s.turn++
}

func solver(s [60]Seed) {
	var now State
	now.seeds = s
	for t := 0; t < T; t++ {
		bestGrid := monteCarloSolver(now)
		gridOutput(bestGrid)
		now.seeds = loadSeed()
		now.turn++
	}
}

const (
	SIMULATIONS = 20 // モンテカルロシミュレーションの回数
	CANDIDATES  = 20 // 候補となるgridの数
	MAXSTEP     = 5
)

func monteCarloSolver(initialState State) (bestGrid [N][N]int) {
	bestScore := 0
	for i := 0; i < CANDIDATES; i++ {
		nowState := initialState
		testGrid := randomGenerateGrid() // 最初のターンのグリッドは決め打ち
		score := 0
		for j := 0; j < SIMULATIONS; j++ {
			score += monteCarloSimuration(nowState, testGrid)
		}
		if bestScore < score {
			bestScore = score
			bestGrid = testGrid
			log.Println(nowState.turn, bestScore)
		}
	}
	return bestGrid
}

// Tまたは、s.time+5までをみる
func monteCarloSimuration(s State, firstGrid [N][N]int) (score int) {
	stopTime := s.turn + MAXSTEP
	s.generate(firstGrid)
	for s.turn < T || s.turn < stopTime {
		grid := randomGenerateGrid()
		s.generate(grid)
	}
	return s.Score()
}

func randomGenerateGrid() (grid [N][N]int) {
	numbers := make([]int, 60)
	for i := range numbers {
		numbers[i] = i
	}
	for i := len(numbers) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}

	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			grid[i][j] = numbers[i*N+j]
		}
	}
	return
}

func gridOutput(grid [N][N]int) {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			fmt.Printf("%v ", grid[i][j])
		}
		fmt.Println("")
	}
}

func maxV(s Seed) (rtn int) {
	rtn = s[0]
	for i := 1; i < M; i++ {
		if rtn < s[i] {
			rtn = s[i]
		}
	}
	return
}
