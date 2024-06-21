package main

import (
	"bytes"
	"container/heap"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"sort"
	"time"
)

var T int          // 期間 100固定
var H, W int       // 土地のサイズ 20*20固定
var i0 int         // 出入り口のy座標
var h [20][20]bool // 縦の障害物
var v [20][20]bool // 横の障害物
var K int          // 作物の種類
var k [][2]int     // 作物の植える時期と収穫時期
var Crops []*Crop  // 作物の種類 // [index-1]でアクセスすること

var FLAG_SOLVER int = 0
var FLAG_TIMEADDITION *int
var FLAG_MAXLENGTH *int
var MaxRegionSize int

func getArguments() {
	maxRegionSize_flag := flag.Int("maxRegionSize", 10, "max region size")
	solver_flag := flag.Int("solver", 4, "solver")
	FLAG_MAXLENGTH = flag.Int("maxLength", 200, "max length")
	FLAG_TIMEADDITION = flag.Int("timeAddition", 0, "time addition")
	flag.Parse()
	if maxRegionSize_flag != nil {
		MaxRegionSize = *maxRegionSize_flag
	}
	FLAG_SOLVER = *solver_flag
}

func zobristHashInit() {
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			zobristTable[i][j] = rand.Uint64()
		}
	}
	for i := range zobristTable2 {
		zobristTable2[i] = rand.Uint64()
	}
}

var startTime time.Time

// ./bin/main -cpuprofile cpuprof < tools/in/0000.txt
// go tool pprof -http=localhost:8888 bin/main cpuprof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	getArguments()
	///////////////////////////////
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
	/////////////////////////////////////

	startTime = time.Now()
	readInput()
	solve()
	t := time.Since(startTime)
	log.Printf("time=%.3f\n", t.Seconds())
}

func readInput() {
	_, err := fmt.Scan(&T, &H, &W, &i0)
	if err != nil{
		log.Fatal(err)
	}
	log.Println(T, H, W, i0)
	var line string
	for i := 0; i < H-1; i++ {
		fmt.Scan(&line)
		for j, b := range line {
			h[i][j] = b == '1'
		}
	}
	for i := 0; i < H; i++ {
		fmt.Scan(&line)
		for j, b := range line {
			v[i][j] = b == '1'
		}
	}
	fmt.Scan(&K)
	//log.Println(T, H, W, i0, K)
	k = make([][2]int, K)
	Crops = make([]*Crop, K)
	for i := 0; i < K; i++ {
		fmt.Scan(&k[i][0], &k[i][1])
		Crops[i] = &Crop{i + 1, k[i][0], k[i][1], 0, 0, 0}
	}
}

func solve() {
	// 出入り口から遠くの場所からエリアを区切る
	//  1. 出入り口からの距離をBFSで計算
	//print(dis)
	var XXX int = FLAG_SOLVER
	switch XXX {
	case 1:
		dis := bfs()
		plotsGrid, plotsPoints := createRegions(dis, MaxRegionSize)
		schedule(plotsGrid, plotsPoints)
		output()
		score()
	case 2:
		dis := bfs()
		plotsGrid, plotsPoints := createRegions(dis, MaxRegionSize)
		startFirstSchedule(plotsGrid, plotsPoints)
		output()
		score()
	case 3:
		dis := bfs()
		plotsGrid, plotsPoints := createRegions(dis, MaxRegionSize)
		beamSearch(plotsPoints, plotsGrid)
	case 4:
		chokudaiSarch()
	}
	//analysis(plotsGrid)
	copyCrops()
	resetCrops()
}

type PlotLite struct {
	capacity    int // 不動
	harvestTime int
	size        int
}

type State struct {
	time          int
	value         int
	score         int
	maxRegionSize int
	cropIndex     [20][20]int
	plotIndex     [20][20]int
	usedCrops     []Crop
	usedCheck     map[int]bool
	plotInfo      []PlotLite
	zhash         uint64
	plotPoints    [][][2]int
}

func cloneState(src State) State {
	dst := src

	dst.usedCrops = make([]Crop, len(src.usedCrops))
	copy(dst.usedCrops, src.usedCrops)
	dst.plotInfo = make([]PlotLite, len(src.plotInfo))
	copy(dst.plotInfo, src.plotInfo)

	dst.usedCheck = make(map[int]bool)
	for key, value := range src.usedCheck {
		dst.usedCheck[key] = value
	}
	dst.plotPoints = make([][][2]int, len(src.plotPoints))
	for i, pts := range src.plotPoints {
		dst.plotPoints[i] = make([][2]int, len(pts))
		copy(dst.plotPoints[i], pts)
	}
	return dst
}

// plots [][][2]int, plotGrid [20][20]int
func newState(maxRegionSize int) (rtn State) {
	dis := bfs()
	plotGrid, plots := createRegions(dis, maxRegionSize)
	dis2 := bfs2(plotGrid)
	for i := 0; i < len(plots); i++ {
		sort.Slice(plots[i], func(j, k int) bool {
			return dis2[plots[i][j][0]][plots[i][j][1]] > dis2[plots[i][k][0]][plots[i][k][1]]
		})
	}

	rtn.time = 1
	rtn.plotIndex = plotGrid
	rtn.plotInfo = make([]PlotLite, len(plots))
	for i := 0; i < len(plots); i++ {
		rtn.plotInfo[i].capacity = len(plots[i])
		rtn.plotInfo[i].harvestTime = 200
	}
	rtn.zhash = rand.Uint64()
	rtn.plotPoints = make([][][2]int, len(plots))
	for i, pts := range plots {
		rtn.plotPoints[i] = make([][2]int, len(pts))
		copy(rtn.plotPoints[i], pts)
	}
	rtn.maxRegionSize = maxRegionSize
	return
}

func (s *State) plant(c Crop, y, x int) error {
	pIndex := s.plotIndex[y][x]
	if s.usedCheck[c.index] {
		return fmt.Errorf("already used %v", c)
	}
	if s.cropIndex[y][x] != 0 {
		return fmt.Errorf("already used cell %+v %+v, %+v", y, x, s.plotInfo[pIndex])
	}
	s.cropIndex[y][x] = c.index
	s.usedCrops = append(s.usedCrops, Crop{index: c.index, y: y, x: x, t: s.time})
	s.value += c.d - c.s + 1
	s.score += c.d - c.s + 1
	s.zhash ^= zobristTable[y][x] ^ zobristTable2[c.index]
	s.usedCheck[c.index] = true
	if pIndex != -1 {
		s.plotInfo[pIndex].size++
		s.plotInfo[pIndex].harvestTime = c.d
	}
	return nil
}

func (s *State) harvestALL() {
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			if s.cropIndex[i][j] == 0 {
				continue
			}
			cIndex := s.cropIndex[i][j]
			if Crops[cIndex-1].d == s.time {
				// harvest
				tmpC := s.cropIndex[i][j]
				s.cropIndex[i][j] = 0
				s.zhash ^= zobristTable[i][j] ^ zobristTable2[cIndex]
				pIndex := s.plotIndex[i][j]
				if pIndex >= 0 {
					s.plotInfo[pIndex].size--
					if s.plotInfo[pIndex].size < 0 {
						log.Println(i, j, tmpC, s.plotIndex[i][j])
						log.Printf("%+v\n", s.plotInfo[pIndex])
						log.Panicf("size < 0, %v\n", pIndex)
					}
					// update
					if s.plotInfo[pIndex].size == 0 {
						s.plotInfo[pIndex].harvestTime = 200
					} else {
						// 一番収穫時期が遅い作物の収穫時期に更新
						// １ターンに複数の作物を収穫するとき、その順番を考慮していないので,全探索する必要がある
						// すでに同じターンで収穫を行ったとき、s.plotInfo[pIndex].harvestTimeは更新されている
						if s.plotInfo[pIndex].harvestTime < s.time {
							continue
						}
						s.plotInfo[pIndex].harvestTime = 200
						for _, pos := range s.plotPoints[pIndex] {
							cIndex := s.cropIndex[pos[0]][pos[1]]
							if cIndex <= 0 {
								continue
							}
							s.plotInfo[pIndex].harvestTime = min(s.plotInfo[pIndex].harvestTime, Crops[cIndex-1].d)
						}
					}
				}
			}
		}
	}
}

type BeamQueue []State

func (q *BeamQueue) Push(s State) {
	*q = append(*q, s)
}

func (q *BeamQueue) Pop() State {
	rtn := (*q)[0]
	*q = (*q)[1:]
	return rtn
}

// バリエーションをふやす
// TimeAddisionを増やす
var TimeAddision int = 0
var MaxLength int = 200

func SetFlags() {
	if FLAG_TIMEADDITION != nil {
		TimeAddision = *FLAG_TIMEADDITION
	}
	if FLAG_MAXLENGTH != nil {
		MaxLength = *FLAG_MAXLENGTH
	}
}

func nextState(src State) (rtn []State) {
	if src.time == T {
		log.Println("time up")
		return []State{src}
	}
	// plotの順番をランダムにする
	pi := randomizedPlotIndices(len(src.plotInfo))
	// plant
	// この時間に植えれる作物をリストアップ (一つ先のも増やすと選択肢が増える)
	// まずは、startFirstScheduleのように、一つのプロットに同じ収穫時期の作物を植える
	for w := 0; w < 5; w++ {
		st := cloneState(src)
		timeCrops := filterTimeCrops(st)
		sortCropsByTime(timeCrops)
		allocateCropsToPlots(&st, timeCrops, pi)
		fillBlankCells(&st)
		//var added bool = false
		//harvest
		st.harvestALL()
		st.time++
		rtn = append(rtn, st)
		shuffle(pi)
	}
	return
}

// 通路も含めて、おけるマスを探す
func fillBlankCells(st *State) {
	var planted bool
	for {
		planted = false
		selectableCrops := selectCrops(st)
		blankCells := findBlankCells(st)
		//　３方向が壁または作物で下がれているか
		for _, pos := range blankCells {
			d := isThreeSideBlocked(*st, pos[0], pos[1])
			if d == -1 {
				continue
			}
			for _, c := range selectableCrops {
				if st.usedCheck[c.index] {
					continue
				}
				if c.d <= d {
					err := st.plant(c, pos[0], pos[1])
					if err != nil {
						log.Fatal(err)
					}
					planted = true
					break
				}
			}
		}
		if !planted {
			break
		}
	}
}

func selectCrops(st *State) (rtn []Crop) {
	for _, c := range Crops {
		if !st.usedCheck[c.index] && c.s == st.time {
			rtn = append(rtn, *c)
		}
	}
	sort.Slice(rtn, func(i, j int) bool {
		return rtn[i].d < rtn[j].d
	})
	return
}

func findBlankCells(st *State) (rtn [][2]int) {
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			if i == i0 && j == 0 {
				continue
			}
			if st.cropIndex[i][j] == 0 && st.plotIndex[i][j] == -1 {
				rtn = append(rtn, [2]int{i, j})
			}
		}
	}
	return
}

// 三方向が壁または作物で下がれているとき、最小のDを返す
// そうでないとき、-1を返す
func isThreeSideBlocked(st State, y, x int) int {
	var cnt int
	var minD int = 10000
	for _, d := range directions {
		ny, nx := y+d[0], x+d[1]
		if ny < 0 || ny >= H || nx < 0 || nx >= W {
			cnt++
			continue
		}
		if checkWall([2]int{y, x}, [2]int{ny, nx}) {
			cnt++
			continue
		}
		if st.cropIndex[ny][nx] != 0 {
			cnt++
			minD = min(Crops[st.cropIndex[ny][nx]-1].d, minD)
		}
	}
	if cnt == 3 {
		return minD
	}
	return -1
}

func chokudaiSarch() {
	zobristHashInit()
	hasDuplicate := make(map[uint64]bool)
	beam := make([]PriorityQueue, T+1)
	for i := MaxRegionSize - 3; i < MaxRegionSize+3; i++ {
		initialState := newState(i)
		log.Println(i, initialState.maxRegionSize)
		heap.Push(&beam[0], &Item{value: initialState, priority: initialState.value})
	}
	for time.Since(startTime).Seconds() < 1.6 {
		for t := 0; t < T-1; t++ {
			if len(beam[t]) == 0 {
				continue
			}
			//current := cloneState(beam[t][0].value)
			current := heap.Pop(&beam[t]).(*Item).value
			nexts := nextState(current)
			for _, next := range nexts {
				if !hasDuplicate[next.zhash] {
					heap.Push(&beam[t+1], &Item{value: next, priority: next.value})
				} else {
					hasDuplicate[next.zhash] = true
				}
			}
		}
	}
	best := heap.Pop(&beam[T-1]).(*Item)
	outputAns(best.value)
	log.Printf("score=%f\n", float64(1000000*best.value.value)/float64(H*W*T))
	log.Printf("maxRegionSize=%d\n", best.value.maxRegionSize)
}

func randomizedPlotIndices(lenPlotInfo int) []int {
	pi := make([]int, lenPlotInfo)
	for i := 0; i < lenPlotInfo; i++ {
		pi[i] = i
	}
	shuffle(pi)
	return pi
}

func filterTimeCrops(st State) []Crop {
	timeCrops := make([]Crop, 0)
	for _, c := range Crops {
		if !st.usedCheck[c.index] && c.s >= st.time && c.s <= st.time+TimeAddision {
			timeCrops = append(timeCrops, *c)
		}
	}
	return timeCrops
}

func sortCropsByTime(crops []Crop) {
	sort.Slice(crops, func(i, j int) bool {
		if crops[i].s == crops[j].s {
			return crops[i].d > crops[j].d
		}
		return crops[i].s < crops[j].s
	})
}

func allocateCropsToPlots(st *State, timeCrops []Crop, pi []int) {
	for _, crop := range timeCrops {
		minLength := 200
		minIndex := -1
		for _, pIndex := range pi {
			if isPlotAvailableForCrop(*st, pIndex, crop) {
				plotTimeDiff := st.plotInfo[pIndex].harvestTime - crop.d
				if plotTimeDiff < minLength {
					minLength = plotTimeDiff
					minIndex = pIndex
				}
			}
		}
		if minIndex != -1 {
			currentIndex := st.plotInfo[minIndex].size
			y, x := st.plotPoints[minIndex][currentIndex][0], st.plotPoints[minIndex][currentIndex][1]
			err := st.plant(crop, y, x)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func isPlotAvailableForCrop(st State, pIndex int, crop Crop) bool {
	if st.plotInfo[pIndex].size == 0 && st.plotInfo[pIndex].capacity > 0 {
		return true
	}
	return st.plotInfo[pIndex].harvestTime >= crop.d && st.plotInfo[pIndex].size < st.plotInfo[pIndex].capacity
}

var beamWidth int = 20

func beamSearch(plotPoints [][][2]int, plotsGrid [20][20]int) {
	zobristHashInit()
	dis := bfs2(plotsGrid)
	//log.Println(string(print(dis)))
	// 新たなdistanceを参考にplotPointsを-1から順番にする
	for i := 0; i < len(plotPoints); i++ {
		sort.Slice(plotPoints[i], func(j, k int) bool {
			return dis[plotPoints[i][j][0]][plotPoints[i][j][1]] > dis[plotPoints[i][k][0]][plotPoints[i][k][1]]
		})
	}
	initialState := newState(MaxRegionSize)
	queue := BeamQueue{initialState}
	loop := 0
	for len(queue) > 0 {
		var nextQueue BeamQueue
		//log.Println(loop, len(queue), queue[0].value)
		for _, s := range queue {
			for _, nextS := range nextState(s) {
				nextQueue.Push(nextS)
			}
		}
		sort.Slice(nextQueue, func(i, j int) bool {
			return nextQueue[i].value > nextQueue[j].value
		})
		if len(nextQueue) > beamWidth {
			nextQueue = nextQueue[:beamWidth]
		}
		queue = nextQueue
		//log.Println(len(queue), queue[0].value)
		if queue[0].time == T {
			break
		}
		loop++
	}
	best := queue[0]
	outputAns(best)
	log.Printf("score=%f\n", float64(1000000*best.value)/float64(H*W*T))
}

func outputAns(best State) {
	fmt.Println(len(best.usedCrops))
	for _, c := range best.usedCrops {
		fmt.Println(c.index, c.y, c.x, c.t)
	}
}

type Crop struct {
	index int
	s, d  int
	y, x  int
	t     int
}

type Plot struct {
	index       int       // 1~
	capacity    int       // 1~
	cells       *[][2]int // そのプロットに含まれるセル　順番に並んでる必要がある
	craps       []*Crop   // そのプロットに植えられている作物
	size        int
	harvestTime int // 最後に入れた作物の収穫時期
}

func (p *Plot) plant(c *Crop, t int) {
	p.craps = append(p.craps, c)
	p.harvestTime = c.d
	c.y, c.x = (*p.cells)[p.size][0], (*p.cells)[p.size][1]
	c.t = t
	p.size++
}

func (p *Plot) harvest(t int) {
	for len(p.craps) > 0 && p.craps[p.size-1].d == t {
		p.craps = p.craps[:p.size-1]
		p.size--
		if p.size == 0 {
			p.harvestTime = 200
		} else {
			p.harvestTime = p.craps[p.size-1].d
		}
	}
}

func startFirstSchedule(plotsGrid [20][20]int, plotPoints [][][2]int) {
	// sは小さい順に、dは大きい順にソート
	sort.Slice(Crops, func(i, j int) bool {
		if Crops[i].s == Crops[j].s {
			return Crops[i].d > Crops[j].d
		}
		return Crops[i].s < Crops[j].s
	})
	dis := bfs2(plotsGrid)
	log.Println(string(print(dis)))
	log.Println(string(print(plotsGrid)))
	// 新たなdistanceを参考にplotPointsを-1から順番にする
	for i := 0; i < len(plotPoints); i++ {
		sort.Slice(plotPoints[i], func(j, k int) bool {
			return dis[plotPoints[i][j][0]][plotPoints[i][j][1]] > dis[plotPoints[i][k][0]][plotPoints[i][k][1]]
		})
	}
	plots := make([]Plot, len(plotPoints))
	for i := 0; i < len(plotPoints); i++ {
		//log.Println(i, len(plotPoints[i]), plotPoints[i])
		plots[i] = Plot{i, len(plotPoints[i]), &plotPoints[i], make([]*Crop, 0, len(plotPoints[i])), 0, 200}
	}
	//log.Println(plots[1].cells)
	//for i := 0; i < len(*plots[1].cells); i++ {
	//log.Println(dis[(*plots[1].cells)[i][0]][(*plots[1].cells)[i][1]])
	//}
	for t := 1; t <= 100; t++ {
		// plant
		for i := 0; i < len(Crops); i++ {
			if Crops[i].s != t {
				continue
			}
			min_length := 200
			min_index := -1
			for j := 0; j < len(plots); j++ {
				// 作物の収穫時期がプロットにある作物の収穫時期より早くて、近いプロットにいれる
				if plots[j].harvestTime < Crops[i].d || plots[j].size == plots[j].capacity {
					continue
				}
				if plots[j].size > 0 && min_length > plots[j].harvestTime-Crops[i].d {
					min_length = plots[j].harvestTime - Crops[i].d
					min_index = plots[j].index
				} else {
					// 何も入ってないプロットに入れる
					min_length = 200
					min_index = plots[j].index
				}
			}
			if min_index != -1 {
				//log.Println(min_length)
				plots[min_index].plant(Crops[i], t)
			}
		}
		// harvest
		for i := 0; i < len(plots); i++ {
			if plots[i].size == 0 {
				continue
			}
			if plots[i].harvestTime == t {
				plots[i].harvest(t)
			}
		}
	}
	//for i := 0; i < len(crops); i++ {
	//log.Printf("%+v\n", crops[i])
	//}
}

// それぞれのプロット内で順番を決める
func bfs2(poltsGrid [20][20]int) (distance [20][20]int) {
	var queue [][2]int
	// 初期化
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			if poltsGrid[i][j] != -1 {
				distance[i][j] = 99
			} else {
				queue = append(queue, [2]int{i, j})
			}
		}
	}
	//print(distance)
	//log.Println(queue)
	// -1は全て０にする
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, d := range directions {
			y, x := current[0]+d[0], current[1]+d[1]
			if y < 0 || y >= H || x < 0 || x >= W {
				continue
			}
			if checkWall(current, [2]int{y, x}) {
				continue
			}
			// -1から他の全て or 同じプロット間しか移動できない
			if poltsGrid[current[0]][current[1]] != -1 && poltsGrid[y][x] != poltsGrid[current[0]][current[1]] {
				continue
			}
			if distance[y][x] > distance[current[0]][current[1]]+1 {
				distance[y][x] = distance[current[0]][current[1]] + 1
				queue = append(queue, [2]int{y, x})
			}
		}
	}
	return
}

// プロットに収穫時期が同じ作物をつくる
func schedule(plotsGrid [20][20]int, plotPoints [][][2]int) {
	sort.Slice(Crops, func(i, j int) bool {
		if Crops[i].d == Crops[j].d {
			return Crops[i].s < Crops[j].s
		}
		return Crops[i].d < Crops[j].d
	})
	harvestTime := make([]int, len(plotPoints))
	// t=1 ~ 100
	for t := 1; t <= 100; t++ {
		// log.Println("t=", t)
		// 作物の植え付け
		// 一つのプラントには同じ収穫時期の作物をうえる
		for i := 0; i < len(plotPoints); i++ {
			ht := -1
			// 収穫期がある＝すでに作物が植えられている
			if harvestTime[i] != 0 {
				continue
			}
			size := len(plotPoints[i])
			if size == 0 {
				continue
			}
			cnt := 0
			// log.Println("plot", i, "size", size)
			for index, c := range Crops {
				if c.s >= t && c.d > t && c.t == 0 && (ht == -1 || ht == c.d) {
					ht = c.d
					Crops[index].y, Crops[index].x = plotPoints[i][cnt][0], plotPoints[i][cnt][1]
					Crops[index].t = t
					// log.Println(crops[index])
					cnt++
					if size == cnt {
						break
					}
				}
			}
			// 収穫時期を登録
			if ht != -1 {
				harvestTime[i] = ht
			}
			// log.Printf("plot %d, size %d, cnt %d, ht %d\n", i, size, cnt, ht)
		}
		// harvest
		for i := 0; i < len(plotPoints); i++ {
			if harvestTime[i] == t {
				harvestTime[i] = 0
			}
		}
	}
}

var directions [][2]int = [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

func bfs() (distance [20][20]int) {
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			distance[i][j] = math.MaxInt
		}
	}
	distance[i0][0] = 0
	var queue [][2]int
	queue = append(queue, [2]int{i0, 0})
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, d := range directions {
			x, y := current[0]+d[0], current[1]+d[1]
			if x < 0 || x >= H || y < 0 || y >= W {
				continue
			}
			if (d[0] == 1 && h[x-1][y]) ||
				(d[0] == -1 && h[x][y]) ||
				(d[1] == 1 && v[x][y-1]) ||
				(d[1] == -1 && v[x][y]) {
				continue
			}
			if distance[x][y] > distance[current[0]][current[1]]+1 {
				distance[x][y] = distance[current[0]][current[1]] + 1
				queue = append(queue, [2]int{x, y})
			}
		}
	}
	return
}

func createRegions(dist [20][20]int, maxRegionSize int) (splited [20][20]int, areaPoints [][][2]int) {
	var num int = 1
	for {
		p, d := findFartest(dist, splited)
		if p[0] == -1 && p[1] == -1 {
			break
		}
		newRegion := expandRegeion(num, p, dist, &splited, maxRegionSize)
		if d == math.MaxInt { // 出入り口から到達不可能な場所
			addCellsToRegion(newRegion, &splited)
			updateDist([][2]int{}, &dist, splited)
			continue
		}
		newPath := createPath(newRegion, dist, &splited)
		updateDist(newPath, &dist, splited)
		num++
	}
	//print(splited)
	areaPoints = make([][][2]int, num+1)
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			num := splited[i][j]
			if num > 0 {
				areaPoints[num] = append(areaPoints[num], [2]int{i, j})
			}
		}
	}
	cnt := 0
	for i := 0; i <= num; i++ {
		//log.Println(i, len(areaPoints[i]))
		cnt += len(areaPoints[i])
	}
	//log.Println("total", cnt, float64(cnt)/float64(H*W))
	return
}

func findFartest(distance [20][20]int, regions [20][20]int) (p [2]int, maxDist int) {
	p = [2]int{-1, -1}
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			if distance[i][j] > maxDist && regions[i][j] == 0 {
				maxDist = distance[i][j]
				p[0], p[1] = i, j
			}
		}
	}
	//log.Println(p, maxDist)
	return
}

//var maxRegionSize int = 8

func expandRegeion(num int, p [2]int, dist [20][20]int, splited *[20][20]int, maxRegionSize int) (newRegion [][2]int) {
	//	log.Println("expand", num)
	splited[p[0]][p[1]] = num
	size := 1
	queue := [][2]int{{p[0], p[1]}}
	newRegion = append(newRegion, [2]int{p[0], p[1]})
	for len(queue) > 0 && size < maxRegionSize {
		cell := queue[0]
		queue = queue[1:]
		for _, d := range directions {
			ny, nx := cell[0]+d[0], cell[1]+d[1]
			if nx < 0 || nx >= W || ny < 0 || ny >= H {
				continue
			}
			if splited[ny][nx] != 0 {
				continue
			}
			if !checkWall(cell, [2]int{ny, nx}) {
				queue = append(queue, [2]int{ny, nx})
				newRegion = append(newRegion, [2]int{ny, nx})
				splited[ny][nx] = num
				size++
			}
			if size >= maxRegionSize {
				break
			}
		}
	}
	return
}

// createPath make path between newRegion and (i0, 0)
func createPath(newRegion [][2]int, dist [20][20]int, splited *[20][20]int) (path [][2]int) {
	//log.Println("createPath")
	// 1. newRegionの中で、(i0, 0)に一番近い点を探す
	var p [2]int = [2]int{-1, -1}
	minDist := math.MaxInt
	for _, q := range newRegion {
		if dist[q[0]][q[1]] < minDist {
			minDist = dist[q[0]][q[1]]
			p = q
		}
	}
	// 2.a (i0, 0)からpまでの最短経路を探す
	// 2.b dist[0][0]の距離が０

	for dist[p[0]][p[1]] > 0 {
		for _, d := range directions {
			y, x := p[0]+d[0], p[1]+d[1]
			if y < 0 || y >= H || x < 0 || x >= W || splited[y][x] > 0 {
				continue
			}
			if !checkWall(p, [2]int{y, x}) && dist[p[0]][p[1]]-dist[y][x] == 1 {
				splited[y][x] = -1
				path = append(path, [2]int{y, x})
				p = [2]int{y, x}
				break
			}
		}
	}
	//log.Println("end createPath")
	return
}

func updateDist(path [][2]int, dist *[20][20]int, splited [20][20]int) {
	var queue [][2]int
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			if splited[i][j] == -1 {
				(*dist)[i][j] = 0
				queue = append(queue, [2]int{i, j})
			} else if splited[i][j] == 0 {
				(*dist)[i][j] = math.MaxInt
			} else {
				(*dist)[i][j] = -1
			}
		}
	}
	for len(queue) > 0 {
		currnt := queue[0]
		queue = queue[1:]
		for _, d := range directions {
			y, x := currnt[0]+d[0], currnt[1]+d[1]
			if y < 0 || y >= H || x < 0 || x >= W || splited[y][x] != 0 {
				continue
			}
			if !checkWall(currnt, [2]int{y, x}) && (*dist)[y][x] > (*dist)[currnt[0]][currnt[1]]+1 {
				(*dist)[y][x] = (*dist)[currnt[0]][currnt[1]] + 1
				queue = append(queue, [2]int{y, x})
			}
		}
	}
}

// 孤立したエリアを近接エリアに統合する
func addCellsToRegion(cells [][2]int, regions *[20][20]int) {
	//log.Println(cells)
	nearestRegion := -1
	currentRegion := regions[cells[0][0]][cells[0][1]]
	// find nearest region
	for _, p := range cells {
		for _, d := range directions {
			y, x := p[0]+d[0], p[1]+d[1]
			if y < 0 || y >= H || x < 0 || x >= W {
				continue
			}
			if checkWall(p, [2]int{y, x}) {
				continue
			}
			if regions[y][x] > 0 && regions[y][x] != currentRegion {
				nearestRegion = regions[y][x]
				// ここで上書き
				for _, c := range cells {
					regions[c[0]][c[1]] = nearestRegion
				}
				return
			}
		}
	}
	log.Println("not found nearest region")
}

func checkWall(a [2]int, b [2]int) bool {
	switch {
	case a[0] == b[0] && a[1]-b[1] == -1:
		return v[a[0]][a[1]]
	case a[0] == b[0] && a[1]-b[1] == 1:
		return v[b[0]][b[1]]
	case a[0]-b[0] == -1 && a[1] == b[1]:
		return h[a[0]][a[1]]
	case a[0]-b[0] == 1 && a[1] == b[1]:
		return h[b[0]][b[1]]
	default:
		log.Fatalf("invalid wall check: %v, %v", a, b)
		return false // unreachable　code but need for linter
	}
}

func output() {
	usedCrops := make([]int, 0, len(Crops))
	for _, c := range Crops {
		if c.t != 0 {
			usedCrops = append(usedCrops, c.index)
		}
	}
	fmt.Println(len(usedCrops))
	for _, c := range Crops {
		if c.t != 0 {
			fmt.Println(c.index, c.y, c.x, c.t)
		}
	}
}

func copyCrops() (rtn []Crop) {
	for _, c := range Crops {
		rtn = append(rtn, Crop{c.index, c.s, c.d, c.y, c.x, c.t})
	}
	return
}

func resetCrops() {
	for i := 0; i < len(Crops); i++ {
		Crops[i].y = 0
		Crops[i].x = 0
		Crops[i].t = 0
	}
}

func score() {
	score := 0
	for _, c := range Crops {
		if c.t != 0 {
			score += c.d - c.s + 1
		}
	}
	log.Printf("score=%f\n", float64(1000000*score)/float64(H*W*T))
}

func analysis(grid [20][20]int) {
	valedCellCnt := 0
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			if grid[i][j] > 0 {
				valedCellCnt++
			}
		}
	}
	wallCnt := 0
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			if h[i][j] {
				wallCnt++
			}
			if v[i][j] {
				wallCnt++
			}
		}
	}
	log.Printf("K=%d, unvaled=%d wall=%d\n", K, H*W-valedCellCnt, wallCnt)
}

func print(grid [20][20]int) string {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	for i := 0; i <= 2*H; i++ {
		for j := 0; j <= 2*W; j++ {
			switch {
			case i%2 == 0 && j%2 == 0:
				buffer.WriteString("+")
			case i == 2*i0+1 && j == 0:
				buffer.WriteString(" ")
			case i == 0 || i == 2*H:
				buffer.WriteString("--")
			case j == 0 || j == 2*W:
				buffer.WriteString("|")
			case i%2 == 0:
				if h[i/2-1][(j-1)/2] {
					buffer.WriteString("--")
				} else {
					buffer.WriteString("  ")
				}
			case j%2 == 0:
				if v[(i-1)/2][j/2-1] {
					buffer.WriteString("|")
				} else {
					buffer.WriteString(" ")
				}
			default:
				// ここに数値を入れる
				buffer.WriteString(fmt.Sprintf("%2d", grid[(i-1)/2][(j-1)/2]))
			}
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// zobrist hashings
var zobristTable [20][20]uint64
var zobristTable2 [20 * 20 * 100]uint64

// Priority Queue
// An Item is something we manage in a priority queue.
type Item struct {
	value    State // The value of the item; arbitrary.
	priority int   // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
//func (pq *PriorityQueue) update(item *Item, value State, priority int) {
//item.value = value
//item.priority = priority
//heap.Fix(pq, item.index)
//}

func shuffle(data []int) {
	rand.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })
}
