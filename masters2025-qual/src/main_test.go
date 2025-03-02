package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

// go test -benchmem -run=^ -bench '^BenchmarkChokudaiSearch' ahc043 -cpuprofile cpu.prof
// go test -benchmem -run='^$' -bench '^BenchmarkChokudaiSearch$' ahc043 -cpuprofile cpu.prof
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
	// 1行ずつ読み込む
	for scaner.Scan() {
		line := scaner.Text()
		fmt.Sscan(line, &in.N, &in.M)
		for i := 0; i < in.N; i++ {
			scaner.Scan()
			line = scaner.Text()
			fmt.Sscan(line, &in.grid[i])
		}
	}
	return
}
