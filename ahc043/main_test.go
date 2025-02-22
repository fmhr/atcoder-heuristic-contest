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
	path := f.findShortestPath(a, b)
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
	// go test -timeout 30s -run ^TestConstructRailway$ ahc043 -v
	in, err := readInputFile("tools/in/0000.txt")
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	stationPos := choseStationPosition(*in)
	t.Log("number of station:", len(stationPos))
	edges := constructRailway(*in, stationPos)
	t.Log("stations=", len(stationPos), "edges=", len(edges))
	for i := 0; i < len(edges); i++ {
		if len(edges[i].Rail) != len(edges[i].Path) {
			// テスト失敗
			t.Error("len(edges[i].Rail) != len(edges[i].Path)")
		}
		str := fmt.Sprintf("%d ", len(edges[i].Rail))
		for j := 0; j < len(edges[i].Rail); j++ {
			str += fmt.Sprintf(" %s", railMap[edges[i].Rail[j]])
		}
	}
	var errCount int
	for i := 0; i < len(stationPos); i++ {
		for j := i + 1; j < len(stationPos); j++ {
			if !CanReach(i, j, edges) {
				log.Println("Can't reach", i, j)
				errCount++
			}
		}
	}
	if errCount > 0 {
		t.Error("Can't reach")
	}
}

func TestDebugBeamSearch(t *testing.T) {
	// 線路上に駅を配置することができるのかを確認する
	f, err := readGridFileToFild("test/t0000.txt")
	if err != nil {
		t.Fatalf("failed to read grid: %v", err)
	}
	for _, s := range f.stations {
		t.Log(s)
	}
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			if isRail(f.cell[i][j]) {
				if rand.Intn(10) < 5 {
					err := f.build(Action{Kind: STATION, X: int16(j), Y: int16(i)})
					if err != nil {
						t.Fatalf("failed to build: %v", err)
					}
				}
			}
		}
	}
	t.Log(f.cellString())
}

// ベンチマークの使い方
// go test . -bench . -run ^TestBeamSearch$ -v -cpuprofile cpu.prof
// go tool pprof -http=:8080 cpu.prof

func TestBeamSearch(t *testing.T) {
	in, err := readInputFile("tools/in/0000.txt")
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	beamSearch(*in)
}

func TestChoseStationPosition(t *testing.T) {
	in, err := readInputFile("tools/in/0001.txt")
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	stationPos := choseStationPosition(*in)
	//t.Log(stationPos)
	//var grid [2500]int16
	//for _, p := range stationPos {
	//grid[p.Y*50+p.X] = 1
	//}
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

func TestIsRailConnected(t *testing.T) {
	tests := []struct {
		railType  int16
		direction int
		isStart   bool
		expected  bool
	}{
		// 上方向
		{RAIL_HORIZONTAL, UP, true, false},
		{RAIL_LEFT_DOWN, UP, true, false},
		{RAIL_RIGHT_DOWN, UP, true, false},
		{RAIL_VERTICAL, UP, true, true},
		{RAIL_RIGHT_UP, UP, false, false},
		{RAIL_LEFT_UP, UP, false, false},

		// 下方向
		{RAIL_HORIZONTAL, DOWN, true, false},
		{RAIL_LEFT_UP, DOWN, true, false},
		{RAIL_RIGHT_UP, DOWN, true, false},
		{RAIL_VERTICAL, DOWN, true, true},
		{RAIL_LEFT_DOWN, DOWN, false, false},
		{RAIL_RIGHT_DOWN, DOWN, false, false},

		// 右方向
		{RAIL_VERTICAL, RIGHT, true, false},
		{RAIL_LEFT_DOWN, RIGHT, true, false},
		{RAIL_LEFT_UP, RIGHT, true, false},
		{RAIL_HORIZONTAL, RIGHT, true, true},
		{RAIL_RIGHT_DOWN, RIGHT, false, false},
		{RAIL_RIGHT_UP, RIGHT, false, false},

		// 左方向
		{RAIL_VERTICAL, LEFT, true, false},
		{RAIL_RIGHT_DOWN, LEFT, true, false},
		{RAIL_RIGHT_UP, LEFT, true, false},
		{RAIL_HORIZONTAL, LEFT, true, true},
		{RAIL_LEFT_DOWN, LEFT, false, false},
		{RAIL_LEFT_UP, LEFT, false, false},
	}

	for _, tt := range tests {
		t.Run(int16ToString(tt.railType), func(t *testing.T) {
			result := checkConnec(tt.railType, tt.direction, tt.isStart)
			if result != tt.expected {
				t.Errorf("isRailConnected(%s, %d, %t) = %t; expected %t",
					int16ToString(tt.railType), tt.direction, tt.isStart, result, tt.expected)
			}
		})
	}
}

// テストに必要な補助関数

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

func TestReadGridFile(t *testing.T) {
	grid, err := readGridFile("test/t0000.txt")
	if err != nil {
		t.Fatalf("failed to read grid: %v", err)
	}
	f := NewField(50)
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			a := Action{Kind: grid[i][j], X: int16(j), Y: int16(i)}
			err := f.build(a)
			if err != nil {
				t.Fatalf("failed to build: %v", err)
			}
		}
	}
}

func readGridFileToFild(filename string) (*Field, error) {
	grid, err := readGridFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read grid: %v", err)
	}
	f := NewField(50)
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			a := Action{Kind: grid[i][j], X: int16(j), Y: int16(i)}
			err := f.build(a)
			if err != nil {
				return nil, fmt.Errorf("failed to build: %v", err)
			}
		}
	}
	return f, nil
}

func readGridFile(filename string) ([][]int16, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	grid := make([][]int16, 50)
	scanner := bufio.NewScanner(file)
	for i := 0; i < 50; i++ {
		scanner.Scan()
		line := scanner.Text()
		runes := []rune(line)
		grid[i] = make([]int16, 50)
		for j := 0; j < 50; j++ {
			v, exit := reverseRailMap[string(runes[j])]
			if !exit {
				log.Println("invalid character", runes[j], string(runes[j]))
				return nil, fmt.Errorf("invalid character")
			}
			grid[i][j] = v
		}
	}
	return grid, nil
}

// reverseRailMap は、railMapの逆引き
var reverseRailMap = map[string]int16{
	".": EMPTY,
	"◎": STATION,
	"─": RAIL_HORIZONTAL,
	"│": RAIL_VERTICAL,
	"┐": RAIL_LEFT_DOWN,
	"┘": RAIL_LEFT_UP,
	"└": RAIL_RIGHT_UP,
	"┌": RAIL_RIGHT_DOWN,
	"#": OTHER,
}

// CanReach は、グラフ内でノード a からノード b に到達可能かどうかを判断します。
func CanReach(a, b int, g []Edge) bool {
	visited := make(map[int]bool) // 訪問済みのノードを追跡します
	queue := []int{a}             // 幅優先探索(BFS)で使用するキュー

	visited[a] = true // 開始ノードを訪問済みにします

	for len(queue) > 0 {
		currentNode := queue[0] // キューから最初のノードを取り出します
		queue = queue[1:]       // キューの先頭を削除します

		if currentNode == b {
			return true // 目的ノードに到達しました
		}

		// 現在のノードから到達可能なすべての隣接ノードを見つけます
		for _, edge := range g {
			if edge.From == currentNode {
				neighbor := edge.To
				if !visited[neighbor] { // まだ訪問していない隣接ノードの場合
					visited[neighbor] = true        // 訪問済みにします
					queue = append(queue, neighbor) // キューに追加します
				}
			} else if edge.To == currentNode {
				neighbor := edge.From
				if !visited[neighbor] { // まだ訪問していない隣接ノードの場合
					visited[neighbor] = true        // 訪問済みにします
					queue = append(queue, neighbor) // キューに追加します
				}
			}
		}
	}

	return false // 目的ノードに到達できませんでした
}
