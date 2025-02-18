package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestShortestPaht(t *testing.T) {
	f := NewField(50)
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			f.cell[i][j] = EMPTY
			if rand.Intn(100) < 10 {
				f.cell[i][j] = STATION
			}
		}
	}
	a := Pos{Y: 0, X: 0}
	b := Pos{Y: 49, X: 0}
	f.cell[a.Y][a.X] = EMPTY
	f.cell[b.Y][b.X] = EMPTY
	path := f.shortestPath(a, b)
	//log.Println(path)
	if path == nil {
		log.Println("no path")
		return
	}
	for _, p := range path {
		f.cell[p.Y][p.X] = 7
	}
	for i := 0; i < 50; i++ {
		str := ""
		for j := 0; j < 50; j++ {
			str += railMap[f.cell[i][j]] + " "
		}
		//log.Printf("%02d %s\n", i, str)
	}
	rtn := f.selectRails(path)
	//log.Println(rtn)
	for i := 0; i < len(rtn); i++ {
		f.cell[path[i].Y][path[i].X] = rtn[i]
	}
	for i := 0; i < 50; i++ {
		str := ""
		for j := 0; j < 50; j++ {
			str += railMap[f.cell[i][j]] + " "
		}
		//log.Printf("%02d %s\n", i, str)
	}
}

func TestChoseStationPosition(t *testing.T) {
	in, err := readInputFile("tools/in/0000.txt")
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	choseStationPosition(*in)
}

func TestReadInput(t *testing.T) {
	in, err := readInputFile("tools/in/0000.txt")
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	_ = in
	//log.Println(in)
}

// ファイルから入力を読み込む関数
func readInputFile(filename string) (*Input, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	// 最初の行をパース (N, M, K, T)
	header := strings.Fields(scanner.Text())
	if len(header) < 4 {
		return nil, fmt.Errorf("invalid input format")
	}

	in := Input{}
	in.N, _ = strconv.Atoi(header[0])
	in.M, _ = strconv.Atoi(header[1])
	in.K, _ = strconv.Atoi(header[2])
	in.T, _ = strconv.Atoi(header[3])

	in.src = make([]Pos, in.M)
	in.dst = make([]Pos, in.M)
	in.income = make([]int, in.M)

	// M 行の (src.Y, src.X, dst.Y, dst.X) を読み込む
	for i := 0; i < in.M; i++ {
		if !scanner.Scan() {
			return nil, fmt.Errorf("unexpected EOF while reading positions")
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			return nil, fmt.Errorf("invalid position format on line %d", i+2)
		}

		srcY, _ := strconv.Atoi(fields[0])
		srcX, _ := strconv.Atoi(fields[1])
		dstY, _ := strconv.Atoi(fields[2])
		dstX, _ := strconv.Atoi(fields[3])

		in.src[i] = Pos{X: int16(srcX), Y: int16(srcY)}
		in.dst[i] = Pos{X: int16(dstX), Y: int16(dstY)}
		in.income[i] = int(distance(in.src[i], in.dst[i])) // 収入は距離
	}
	//log.Printf("readInput: N=%v, M=%v, K=%v, T=%v\n", in.N, in.M, in.K, in.T)
	return &in, nil
}

func TestGridCalculation(t *testing.T) {
	p := Pos{X: 10, Y: 10}
	var grid [2500]int16
	for i := int16(0); i < int16(len(ddy)); i++ {
		next := p.add(Pos{Y: ddy[i], X: ddx[i]})
		if next.Y < 0 || next.Y >= 50 || next.X < 0 || next.X >= 50 {
			t.Log("out of range", next)
		} else {
			grid[next.Y*50+next.X] = i
		}
	}
	t.Log("Grid result:" + gridToString(grid))
}
