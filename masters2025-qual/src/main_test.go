package main

import (
	"bufio"
	"log"
	"os"
	"testing"
)

func TestSolver(t *testing.T) {
	in := readFile("input.txt")
	_ = in
	//out := solve(in)
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
	for line := scaner.Text(); line != ""; scaner.Scan() {
		log.Printf("line: %s\n", line)
	}

	return in
}
