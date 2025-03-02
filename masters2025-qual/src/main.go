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
	ans := solver(in)
	fmt.Println(ans)
	log.Printf("in: %+v\n", in)
}

type State struct {
	grid  [GridSize * GridSize]byte
	pos   Pos
	score int
}

func (s *State) Do(a Action) {
	if s.grid[s.pos.y][s.pos.x] == '.' {
		s.score++
	}
}

type Pos struct {
	y, x int
}

type Act int

const (
	Move  Act = 1
	Carry Act = 2
	Roll  Act = 3
)

var acts = []Act{Move, Carry, Roll}

type Direction int

const (
	Up    Direction = 1
	Down  Direction = 2
	Left  Direction = 3
	Right Direction = 4
)

var directions = []Direction{Up, Down, Left, Right}

type Action struct {
	act  Act
	dict Direction
}

const GridSize = 20

func solver(in In) (ans string) {
	for i := 0; i < in.N; i++ {
		log.Println(string(in.grid[i*GridSize : i*GridSize+in.N]))
	}
	return ans
}

type In struct {
	N, M int
	grid [GridSize * GridSize]byte
}

func readInput() In {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	var in In
	_, err := fmt.Fscan(reader, &in.N, &in.M)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < in.N; i++ {
		var line string
		_, err := fmt.Fscan(reader, &line)
		if err != nil {
			log.Fatal(err)
		}
		for j, c := range line {
			in.grid[i*GridSize+j] = byte(c)
		}
	}
	return in
}
