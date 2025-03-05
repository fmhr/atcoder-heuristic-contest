package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func TestDistanceFromHole(t *testing.T) {
	in := readFile("../tools/in/0000.txt")
	s := newState(in)
	s.showGrid()
	dist := s.distanceFromHole('A')
	if dist[0] != 4 {
		t.Errorf("dist[0] = %d, want %d", dist[0], 4)
	}
	if dist[GridSize*GridSize-1] != 21 {
		t.Errorf("dist[0] = %d, want %d", dist[GridSize*GridSize-1], 21)
	}
}

// src/ディレクトリで実行
// go test -benchmem -run='^$' -bench '^BenchmarkBeamSearch$' -cpuprofile cpu.prof
// go tool pprof -http=:8080 cpu.prof
func BenchmarkBeamSearch(b *testing.B) {
	ATCODER = true
	log.SetOutput(io.Discard)
	in := readFile("../tools/in/0000.txt")
	for i := 0; i < b.N; i++ {
		beamSearch(in)
	}
}

func TestSolver(t *testing.T) {
	in := readFile("../tools/inA/0000.txt")
	_ = in
	out := beamSearch(in)
	_ = out
	//log.Printf("out: %+v\n", out)
}

// readFile テスト用に直接ファイルを読み込む
func readFile(filename string) (in In) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scaner := bufio.NewScanner(file)
	scaner.Scan()
	line := scaner.Text()
	_, err = fmt.Sscan(line, &in.N, &in.M)
	if err != nil {
		log.Fatal(err)
	}
	// 1行ずつ読み込む
	for i := 0; i < in.N; i++ {
		scaner.Scan()
		line := scaner.Text()
		for j := 0; j < in.N; j++ {
			in.grid[i*GridSize+j] = line[j]
		}
	}
	return
}

// 下の形式のファイルを読み込む
// [2 0 0]
// {18 16}
// 38
// 379961
// ...................A
// ............@.......
// ........@...........
// ................@...
// .................@@.
// ...........@@.......
// .................@..
// ...@.....@..........
// ............@.@.@@..
// ......@..@...@.@....
// ............@.......
// @.....@..@........@.
// .......@@.........@.
// ........@..@........
// .....@..@....@..@...
// ....................
// .........@..........
// @......a............
// ...@.a.@.@..........
// ....................
func readState(filename string) (s State) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scaner := bufio.NewScanner(file)
	scaner.Scan()
	fmt.Sscanf(scaner.Text(), "[%d %d %d]", &s.stones[0], &s.stones[1], &s.stones[2])
	scaner.Scan()
	fmt.Sscanf(scaner.Text(), "{%d %d}", &s.pos.y, &s.pos.x)
	scaner.Scan()
	fmt.Sscanf(scaner.Text(), "%d", &s.score)
	scaner.Scan()
	fmt.Sscanf(scaner.Text(), "%d", &s.eval)

	for i := 0; i < GridSize; i++ {
		scaner.Scan()
		line := scaner.Text()
		for j := 0; j < GridSize; j++ {
			s.grid[i*GridSize+j] = line[j]
		}
	}

	s.showGrid()
	return s
}

func TestEval(t *testing.T) {
	s := readState("test/eval_test.txt")
	e := s.calEval()
	log.Printf("eval: %d neweval:%d\n", s.eval, e)
}
