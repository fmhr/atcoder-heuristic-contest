package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestShortestPaht(t *testing.T) {
	rand.Seed(0)
	f := NewField(50)
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			f.cell[i][j] = EMPTY
			if rand.Intn(100) < 5 {
				f.cell[i][j] = OTHER
			}
		}
	}
	a := Pos{Y: int16(rand.Intn(50)), X: int16(rand.Intn(50))}
	b := Pos{Y: int16(rand.Intn(50)), X: int16(rand.Intn(50))}
	f.cell[a.Y][a.X] = STATION
	f.cell[b.Y][b.X] = STATION
	//t.Log(f.cellString())
	path := f.shortestPath(a, b)
	t.Log(path)
	if path == nil {
		t.Error("no path")
		return
	}
	for i := 0; i < 50; i++ {
		str := ""
		for j := 0; j < 50; j++ {
			str += railMap[f.cell[i][j]] + " "
		}
	}
	rtn := f.selectRails(path)
	for i := 0; i < len(rtn); i++ {
		f.cell[path[i].Y][path[i].X] = rtn[i]
	}
	t.Log(f.cellString())
}

func TestConstructRailway(t *testing.T) {
	in, err := readInputFile("tools/in/0013.txt")
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	stationPos := choseStationPosition(*in)
	t.Log("number of station:", len(stationPos))
	p := constructRailway(*in, stationPos)
	t.Log(p)
}

func TestChoseStationPosition(t *testing.T) {
	in, err := readInputFile("tools/in/0001.txt")
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	stationPos := choseStationPosition(*in)
	//t.Log(stationPos)
	var grid [2500]int16
	for _, p := range stationPos {
		grid[p.Y*50+p.X] = 1
	}
	//t.Log("Grid result:" + gridToString(grid))
	t.Log("number of station:", len(stationPos))
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
	//t.Log("Grid result:" + gridToString(grid))
}
