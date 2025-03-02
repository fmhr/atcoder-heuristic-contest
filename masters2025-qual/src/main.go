package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"
)

var ATCODER bool        // AtCoder環境かどうか
var startTime time.Time // 開始時刻
var frand *rand.Rand    // 固定用乱数生成機

func main() {
	if os.Getenv("ATCODER") == "1" {
		ATCODER = true
		log.Println("on AtCoder")
		log.SetOutput(io.Discard)
	}
	log.SetFlags(log.Lshortfile)
	frand = rand.New(rand.NewSource(1))
	startTime = time.Now()
	in := readInput()
	log.Printf("in: %+v\n", in)

}

type In struct {
	N int
}

func readInput() In {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	var in In
	_, err := fmt.Fscan(reader, &in.N)
	if err != nil {
		log.Fatal(err)
	}
	return in
}
