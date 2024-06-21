package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"
)

// ./a.out -cpuprofile cpu.prof < tools/in/0000.txt > out.txt
// ./a.out -cpuprofile cpu.prof -memprofile mem.prof < tools/in/0000.txt > out.txt
// go tool pprof -http=localhost:8888 main cpu.prof
// go tool pprof -http=localhost:8888 main mem.prof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

// StartCPUProfile は、CPUプロファイルを開始する
func StartCPUProfile() func() {
	if *cpuprofile == "" {
		return func() {}
	}
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		log.Fatal("could not start CPU profile: ", err)
	}

	return func() {
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			log.Fatal("could not stop CPU profile: ", err)
		}
	}
}

// writeMemProfile は、メモリプロファイルを書き込む
func writeMemProfile() {
	if *memprofile == "" {
		return
	}
	f, err := os.Create(*memprofile)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close()
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

}

func init() {
	flag.Parse()
	if _, atcoder := os.LookupEnv("ATCODER"); atcoder {
		log.SetOutput(io.Discard)
		return
	}

	log.SetFlags(log.Lshortfile)
	//runtime.GOMAXPROCS(1) // 並列処理を抑制
	//debug.SetGCPercent(2000) // GCを抑制 2000% に設定
	rand.Seed(1) // 乱数のシードを固定することで、デバッグ時に再現性を持たせる
}

func main() {
	stopCPUProfile := StartCPUProfile()
	defer stopCPUProfile()
	defer writeMemProfile()
	// --- start
	startTime := time.Now()
	input := read()
	log.Printf("D=%d N=%d", input.D, input.N)
	//solver2(input)
	checkAllDayZeroChance(input)
	//searchReacts(input)
	elapseTime := time.Since(startTime)
	log.Printf("time=%f", float64(elapseTime)/float64(time.Second))
	log.Println("start", startTime.UnixNano(), "since", time.Since(startTime).Nanoseconds(), time.Since(startTime).Milliseconds())
	// --- end
}

type Event struct {
	width int
	hight int
	y, x  int
}

func (e Event) y2() int {
	return intMin(1000, e.y+e.hight)
}

func (e Event) x2() int {
	return e.x + e.width
}

func (e Event) size() int {
	return e.width * e.hight
}

type Input struct {
	W, D, N int
	request [50][50]int
}

func read() (input Input) {
	fmt.Scan(&input.W, &input.D, &input.N)
	for i := 0; i < input.D; i++ {
		for j := 0; j < input.N; j++ {
			// 逆順に入れてることに注意
			fmt.Scan(&input.request[i][j])
		}
	}
	return
}

func calcE(in Input) (e, eMax float64) {
	sumP := 0.0
	for i := 0; i < in.D; i++ {
		sumEvents := 0
		for j := 0; j < in.N; j++ {
			sumEvents += in.request[i][j]
		}
		//log.Println(i, sumEvents, float64(sumEvents)/float64(input.W*input.W))
		sumP += float64(sumEvents) / float64(in.W*in.W)
		eMax = math.Max(eMax, float64(sumEvents)/float64(in.W*in.W))
	}
	e = sumP / float64(in.D)
	log.Printf("e=%f", e)
	log.Printf("eMax=%f", eMax)
	log.Printf("preAvgWall=%f", predictAvgWall(in.N, e))
	return e, eMax
}

func solver2(in Input) {
	var events [50][50]Event
	e, eMax := calcE(in)
	log.Printf("e=%f", e)
	log.Printf("eMax=%f", eMax)
	predictAvgWall := predictAvgWall(in.N, e)
	log.Printf("preAvgWall=%f", predictAvgWall)
	meanWidth := 1000.0 / predictAvgWall
	variance := meanWidth * 1
	widthList := makeWidthListNormal(meanWidth, variance)
	bestWidthList := make([]int, len(widthList))
	bestInput := in
	copy(bestWidthList, widthList)
	var bestScore int = math.MaxInt64
	var bestEvents [50][50]Event
	startTime := time.Now()
	loopAllDay := 0
	for {
		widthList = makeWidthListNormal(meanWidth, variance)
		preWidthList := make([]int, 0)
		for day := 0; day < in.D; day++ {
			loop := 0
			dayBestCost := math.MaxInt64
			for {
				// イベントを割り振る
				//evt, cost := searchBuckets(in, widthList, day)
				evt, cost := matchEventsOneDay(in, day, widthList)
				hights := make([]int, len(widthList)) // 各widthの高さ
				bins := make([][]int, len(widthList)) // 各widthのイベント
				clear(hights)
				clear(bins)
				for i := 0; i < in.N; i++ {
					hight := (in.request[day][i] + widthList[evt[i]] - 1) / widthList[evt[i]]
					hights[evt[i]] += hight
					bins[evt[i]] = append(bins[evt[i]], i)
				}
				// eventsの更新 ここでサイズと位置が決まる
				for i := 0; i < len(widthList); i++ {
					hight2 := 0 // widthの高さ
					overHight := hights[i] - 1000
					for j := 0; j < len(bins[i]); j++ {
						event := bins[i][j]
						hight := (in.request[day][event] + widthList[i] - 1) / widthList[i]
						if hight < 1 {
							panic("hight < 1")
						}
						if overHight > 0 {
							// 1000を超える場所にあるイベントを縮小
							if overHight >= hight {
								// 超過分がイベントの高さより大きい場合は、イベントの高さを1にする
								overHight -= (hight - 1)
								hight = 1
							} else {
								// 超過分がイベントの高さより小さい場合は、イベントの高さを超過分だけ減らす
								hight -= overHight
								if hight < 1 {
									panic("hight < 1")
								}
								overHight = 0
							}
						} else if overHight < 0 {
							// 逆に1000を超えない場所にあるイベントを拡大して、1000にあわせる
							hight += (-overHight)
							overHight = 0
						}
						events[day][event].width = widthList[i]
						events[day][event].hight = hight
						events[day][event].x = Sum(widthList[:i])
						events[day][event].y = hight2
						hight2 += hight
					}
				}
				if cost == 0 || loop > 2000 {
					//log.Println(widthList)
					//log.Println(hights)
					//log.Println(events[day])
					preWidthList = make([]int, len(widthList))
					copy(preWidthList, widthList)
					//log.Println("day", day, "cost", cost)
					break
				}
				if dayBestCost > cost {
					dayBestCost = cost
					bestWidthList = make([]int, len(widthList))
					copy(bestWidthList, widthList)
					bestInput = in
				}
				// widthListを変更
				// 前日のwidthListをコピーしてから、ランダムに変更 パーテーションのコストを抑える
				if day == 0 || len(widthList) == 1 || loop > 10 {
					widthList = makeWidthListNormal(meanWidth, variance)
				} else {
					if loop%len(widthList) == 0 && day > 0 {
						widthList = make([]int, len(preWidthList))
						copy(widthList, preWidthList)
					}
					if len(widthList) == 1 {
						widthList = makeWidthListNormal(meanWidth, variance)
					} else {
						a := rand.Intn(len(widthList) - 1)
						asize := widthList[a]
						bsize := widthList[a+1]
						newSizeA := rand.Intn(asize+bsize-1) + 1
						widthList[a] = newSizeA
						widthList[a+1] = asize + bsize - newSizeA
						if widthList[a+1] == 0 {
							widthList[a+1], widthList[len(widthList)-1] = widthList[len(widthList)-1], widthList[a+1]
							widthList = widthList[:len(widthList)-1]
						}
					}
				}
				loop++
			}
		}
		s := score(in, events)
		if bestScore > s {
			bestScore = s
			bestWidthList = make([]int, len(widthList))
			copy(bestWidthList, widthList)
			bestInput = in
			bestEvents = events
		}
		if bestScore == 0 {
			break
		}
		loopAllDay++
		if time.Since(startTime) > 2*time.Second {
			break
		}
	}
	output(bestInput, bestEvents)
	score := score(in, bestEvents)
	log.Printf("loopAllDay=%d", loopAllDay)
	log.Printf("score=%d", score)
}

func solver(in Input) {
	var events [50][50]Event
	sumP := 0.0
	eMax := 0.0
	for i := 0; i < in.D; i++ {
		sumEvents := 0
		for j := 0; j < in.N; j++ {
			sumEvents += in.request[i][j]
		}
		//log.Println(i, sumEvents, float64(sumEvents)/float64(input.W*input.W))
		sumP += float64(sumEvents) / float64(in.W*in.W)
		eMax = math.Max(eMax, float64(sumEvents)/float64(in.W*in.W))
	}
	e := sumP / float64(in.D)
	log.Printf("e=%f", e)
	log.Printf("eMax=%f", eMax)
	log.Printf("preAvgWall=%f", predictAvgWall(in.N, e))
	numWall := 0
	var widthList []int
	preLenWidthLiest := predictAvgWall(in.N, e)
	meanWidth := 1000.0 / preLenWidthLiest
	variance := meanWidth * 1
	_ = variance
	for day := 0; day < in.D; day++ {
		widthList, events[day] = daySolverGreedy(&in, day, widthList, meanWidth, variance)
		numWall += len(widthList)
	}
	log.Println(arraySize)
	output(in, events)
	score := score(in, events)
	log.Printf("avgWall=%f", float64(numWall)/float64(in.D))
	log.Printf("score=%d", score)
}

// daySolverGreedy は、各日に対して、貪欲法で解く
// 各イベントと、矩形のサイズを比較して、ロスが少ないものから配置していく
func daySolverGreedy(in *Input, day int, widthList []int, meanWidth, variance float64) ([]int, [50]Event) {
	var events [50]Event
	if day == 0 {
		widthList = makeWidthListNormal(meanWidth, variance)
	}
	resetCnt := 0
	log.Println("day", day, "------------------------")
	//widthList = makeWidthListRand(preLenWidthLiest)
	// widthListに沿って、矩形を考える
	// totalHidht を使って、1000以下になるようにする
	totalHidht := make([]int, len(widthList))
	eindex := make([]int, in.N)
	for i := 0; i < in.N; i++ {
		eindex[i] = i
	}
	loop := 0
	var diffList [20][3]int // [widthIndex, diff, high]
	//diffList := make([][3]int, len(widthList))
	for {
		var allSolved bool = true
		for _, e := range eindex {
			clear(diffList[:])
			//diffList = make([][3]int, len(widthList))
			for i, w := range widthList {
				h := (in.request[day][e] + w - 1) / w
				if h >= 1000 {
					diffList[i] = [3]int{i, 999, 0}
				}
				diffList[i] = [3]int{i, w*h - in.request[day][e], h}
			}
			// Lossが小さい順にソート
			sort.Slice(diffList[:len(widthList)], func(i, j int) bool {
				return diffList[i][1] < diffList[j][1]
			})
			var solved bool
			for i := 0; i < len(diffList); i++ {
				h := (in.request[day][e] + widthList[diffList[i][0]] - 1) / widthList[diffList[i][0]]
				if totalHidht[diffList[i][0]]+h <= 1000 {
					y := totalHidht[diffList[i][0]]
					totalHidht[diffList[i][0]] += h
					solved = true
					// Inputの更新
					events[e].width = widthList[diffList[i][0]]
					events[e].hight = h
					sumWidth := Sum(widthList[:diffList[i][0]])
					events[e].x = sumWidth
					events[e].y = y
					break
				}
			}
			if !solved {
				allSolved = false
				break
			}
		}
		if allSolved {
			break
		}
		widthList = makeWidthListNormal(meanWidth, variance)
		// 順番をランダム
		//rand.Shuffle(len(eindex), func(i, j int) { eindex[i], eindex[j] = eindex[j], eindex[i] })
		// リセット
		totalHidht = make([]int, len(widthList))
		resetCnt++
		loop++
	}
	log.Println(widthList)
	log.Println(totalHidht)
	//log.Println("totalSize", totalSize, "diff", totalSize-in.W*in.W, "resetCnt", resetCnt)
	return widthList, events
}

const maxWidthNum = 20 // 縦に分割するときの最大数

type subState struct {
	evetns [50]int // どこに配置するか
	hights [20]int
	cost   int // このSAの中で使う評価値 小さいほど良い
}

// calcValue は、矩形の配置が正しいかを評価する
func (s subState) calcValue(in Input, widthList []int, day int) (v int) {
	w := len(widthList)
	var hights [20]int
	for i := 0; i < in.N; i++ {
		hight := (in.request[day][i] + widthList[s.evetns[i]] - 1) / widthList[s.evetns[i]]
		hights[s.evetns[i]] += hight
	}
	for i := 0; i < w; i++ {
		if hights[i] > 1000 {
			v += ((hights[i] - 1000) * widthList[i])
		}
	}
	if v < 0 {
		panic("v < 0")
	}
	return
}

func (s *subState) move(in Input, widthList []int, day, e, w int) {
	s.hights[s.evetns[e]] -= (in.request[day][e] + widthList[s.evetns[e]] - 1) / widthList[s.evetns[e]]
	s.hights[w] += (in.request[day][e] + widthList[w] - 1) / widthList[w]
	s.cost = s.calcValue(in, widthList, day)
}

func newSubState(in Input, widthList []int) (s subState) {
	for i := 0; i < in.N; i++ {
		s.evetns[i] = rand.Intn(len(widthList))
		hight := (in.request[0][i] + widthList[s.evetns[i]] - 1) / widthList[s.evetns[i]]
		s.hights[s.evetns[i]] += hight
	}
	return
}

// 縦に分割したときに、それぞれのイベントをどこに配置するかを決める
// 縦のサイズが1000以下になれば良い
func searchBuckets(in Input, widthList []int, day int) ([50]int, int) {
	if len(widthList) == 0 {
		return [50]int{}, 0
	}
	//startTemp := 30000.0
	//endTemp := 10.0
	state := newSubState(in, widthList)
	state.cost = state.calcValue(in, widthList, day)
	maxStep := 100
	minCost := state.cost
	bestState := state
	for i := 0; i < maxStep; i++ {
		temp := (1 - (float64(i) / float64(maxStep))) * 0.0
		if i%2 == 0 {
			// 配置を変更
			e := rand.Intn(in.N)
			w := uint16(rand.Intn(len(widthList)))
			if state.evetns[e] == int(w) {
				continue
			}
			old := state.evetns[e]
			state.move(in, widthList, day, e, int(w))

			//temp := startTemp + (endTemp-startTemp)*float64(i)/float64(maxStep)
			if minCost > state.cost || temp > rand.Float64() {
				minCost = state.cost
				bestState = state
			} else {
				state.move(in, widthList, day, e, old)
			}
		} else {
			// 配置のSWAP
			a := rand.Intn(in.N)
			b := rand.Intn(in.N)
			if a == b {
				continue
			}
			oldaPos := state.evetns[a]
			oldbPos := state.evetns[b]
			state.move(in, widthList, day, a, oldbPos)
			//temp := startTemp + (endTemp-startTemp)*float64(i)/float64(maxStep)
			if minCost > state.cost || temp > rand.Float64() {
				minCost = state.cost
				bestState = state
			} else {
				state.move(in, widthList, day, a, oldaPos)
				state.move(in, widthList, day, b, oldbPos)
			}
		}
		if minCost < 0 {
			panic("minCost < 0")
		}

		if minCost == 0 {
			break
		}
		if i%10 == 0 {
			state = bestState
		}
	}
	//log.Println(bestScore, bestScore, widthList)

	return bestState.evetns, int(minCost)
}

// matchEventsOneDay は、一日分のイベントをマッチングさせる
func matchEventsOneDay(in Input, day int, widths []int) ([50]uint8, int) {
	if len(widths) == 0 {
		return [50]uint8{}, 0
	}

	var match [50]uint8
	var heights [20]int
	var indexList [50]int
	for i := 0; i < in.N; i++ {
		indexList[i] = in.N - 1 - i
	}

	var lossList [50][40][3]int32 // [event][wideSize][widthIndex, loss, high]
	for i := 0; i < in.N; i++ {
		for j := 0; j < len(widths); j++ {
			height := (in.request[day][i] + widths[j] - 1) / widths[j]
			if height >= 1000 {
				// 1000に収まらない
				lossList[i][j][0] = int32(j)
				lossList[i][j][1] = 9999
				lossList[i][j][2] = 0
			} else {
				lossList[i][j][0] = int32(j)
				lossList[i][j][1] = int32(widths[j]*height - in.request[day][i])
				lossList[i][j][2] = int32(height)
			}
		}
		sort.Slice(lossList[i][:len(widths)], func(a, b int) bool {
			return lossList[i][a][1] < lossList[i][b][1]
		})
	}

	bestCost := math.MaxInt64
	var bestSet [50]uint8
	loop := 0
	for loop < 5 {
		var allMatch bool = true
		//for num, i := range indexList {
		for k := 0; k < in.N; k++ {
			i := indexList[k]
			var matchBool bool
			for j := 0; j < len(widths); j++ {
				v := lossList[i][j]
				if heights[v[0]] < 1000 {
					heights[v[0]] += int(v[2])
					match[i] = uint8(v[0])
					matchBool = true
					break
				}
			}
			if !matchBool {
				allMatch = false
				break
			}
		}
		if allMatch {
			sumOverHightW := 0
			for i := 0; i < len(widths); i++ {
				sumOverHightW += intMax(0, heights[i]-1000) * widths[i]
			}
			if sumOverHightW == 0 {
				return match, 0
			}
			if bestCost > sumOverHightW {
				bestCost = sumOverHightW
				bestSet = match
			}
		}
		rand.Shuffle(in.N, func(i, j int) { indexList[i], indexList[j] = indexList[j], indexList[i] })
		for i := 0; i < len(widths); i++ {
			heights[i] = 0
		}
		loop++
	}
	return bestSet, bestCost
}

// makeWidthListRand は、widthListをランダムに生成
func makeWidthListRand(widthNum float64) (widthList []int) {
	num, frac := math.Modf(widthNum)
	//log.Println(widthNum, num, frac)
	if rand.Float64() < frac {
		num++
	}
	//log.Println(widthNum, num, frac)
	points := make([]int, int(num))
	for i := 0; i < int(num); i++ {
		points[i] = rand.Intn(999) + 1
	}
	sort.Ints(points)
	//log.Println(points)
	pre := 0
	for _, p := range points {
		if p-pre < 1 || p == 1000 {
			continue
		}
		widthList = append(widthList, p-pre)
		pre = p
	}
	widthList = append(widthList, 1000-pre)
	sort.Slice(widthList, func(i, j int) bool {
		return widthList[i] > widthList[j]
	})
	//log.Println(len(widthList), widthList)
	return
}

// makeWidthListNormal は、widthListを正規分布に従って生成
func makeWidthListNormal(mean, variance float64) (widthList []int) {
	sum := 0
	for sum < 1000 {
		width := int(NormDistSampleMin1(mean, variance))
		if width < 1 {
			continue
		}
		widthList = append(widthList, intMin(1000-sum, width))
		sum += width
	}
	sort.Slice(widthList, func(i, j int) bool {
		return widthList[i] > widthList[j]
	})
	return
}

// addWalls adds walls for the rectangle.
func addWalls(y1, x1, y2, x2 int, vs, hs *BitArray) {
	// vertical
	for y := y1; y < y2 && y < 1000; y++ {
		if x1 > 0 {
			vs.Set(y, x1-1)
		}
		if x2 < 1000 {
			vs.Set(y, x2-1)
		}
	}
	// horizontal
	for x := x1; x < x2 && x < 1000; x++ {
		if y1 > 0 {
			hs.Set(y1-1, x)
		}
		if y2 < 1000 {
			hs.Set(y2-1, x)
		}
	}
}

// predictAvgWall は、Nとeを受け取り、壁の個数を予測
func predictAvgWall(N int, e float64) float64 {
	coef_e := -32.39385659
	coef_N := 0.28385756
	intercept := 31.42203874072525
	//avgWallを計算
	avgWall := coef_e*e + coef_N*float64(N) + intercept
	return avgWall
}

func score(in Input, events [50][50]Event) int {
	cost := 1
	var vs, hs, vs2, hs2 BitArray
	for day := 0; day < in.D; day++ {
		for j := 0; j < in.N; j++ {
			// 面積が足りない場合のペナルティコスト
			if in.request[day][j] > events[day][j].size() {
				cost += (in.request[day][j] - events[day][j].size()) * 100
			}
			y1 := events[day][j].y
			x1 := events[day][j].x
			y2 := events[day][j].y2()
			x2 := events[day][j].x2()
			addWalls(y1, x1, y2, x2, &vs, &hs)
		}
		if day == 0 {
			//初日はコスト:0
			//cost += vs.PopCount()
			//cost += hs.PopCount()
		} else {
			tmpCost := 0
			tmpCost += vs.XorPopCount(vs2)
			tmpCost += hs.XorPopCount(hs2)
			//log.Println("day", day, "cost", tmpCost)
			cost += tmpCost
		}
		//log.Println(vs.PopCount(), hs.PopCount())
		vs2 = vs
		hs2 = hs
		vs.Reset()
		hs.Reset()
	}
	return cost
}

func output(in Input, events [50][50]Event) {
	for i := 0; i < in.D; i++ {
		for j := 0; j < in.N; j++ {
			fmt.Print(events[i][j].y, events[i][j].x, events[i][j].y2(), events[i][j].x2(), " ")
		}
		fmt.Println("")
	}
}

// ΣMax(Aij)が1000*1000を超えない時、すべての日において通用する区切り型を探す
// Cost=1になる
func checkAllDayZeroChance(in Input) {
	var maxEvents [50]int
	for i := 0; i < in.D; i++ {
		//log.Println(in.request[i])
		for j := 0; j < in.N; j++ {
			maxEvents[j] = intMax(maxEvents[j], in.request[i][j])
		}
	}
	log.Println(Sum(maxEvents[:]), maxEvents)
	if Sum(maxEvents[:]) <= 1000*1000 {
		// すべての日において通用する区切り型が存在する
		log.Println("all day zero chance")
	} else {
		log.Println("not all day zero chance")
	}
}

// --- util ---
func Sum(list []int) (sum int) {
	for _, v := range list {
		sum += v
	}
	return
}

func Max(list []int) (max int) {
	for _, v := range list {
		if v > max {
			max = v
		}
	}
	return
}

func Min(list []int) (min int) {
	min = math.MaxInt64
	for _, v := range list {
		if v < min {
			min = v
		}
	}
	return
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// bitArrayを管理するためのセット
const widthBits = 1000
const totalBits = 1000 * 1000
const uint64Size = 64
const arraySize = totalBits / uint64Size

type BitArray [arraySize]uint64

func (b *BitArray) Set(y, x int) {
	index := y*widthBits + x
	b[index/uint64Size] |= 1 << (index % uint64Size)
}

func (b *BitArray) Unset(y, x int) {
	index := y*widthBits + x
	b[index/uint64Size] &= ^(1 << (index % uint64Size))
}

func (b *BitArray) Get(y, x int) bool {
	index := y*widthBits + x
	return b[index/uint64Size]&(1<<(index%uint64Size)) != 0
}

func (b BitArray) PopCount() (count int) {
	for i := 0; i < arraySize; i++ {
		count += bits.OnesCount64(b[i])
	}
	return count
}

func (b BitArray) XorPopCount(a BitArray) (count int) {
	for i := 0; i < arraySize; i++ {
		count += bits.OnesCount64(b[i] ^ a[i])
	}
	return count
}

func (b *BitArray) Reset() {
	for i := 0; i < arraySize; i++ {
		b[i] = 0
	}
}

// NormDistSampleMin1 は、正規分布に従う乱数を生成して、1以上の値を返す
// mean 平均 variance 分散
func NormDistSampleMin1(mean, variance float64) float64 {
	stdDev := math.Sqrt(variance)
	return math.Max(1, generateNormalDistributionSample(mean, stdDev))
}

// mean 平均 stdDev 標準偏差
// stdDev = math.Sqrt(variance)
func generateNormalDistributionSample(mean, stdDev float64) float64 {
	return rand.NormFloat64()*stdDev + mean
}
