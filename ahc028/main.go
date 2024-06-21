package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
)

var startTime time.Time
var timeLimit time.Duration = time.Duration(1.850 * float64(time.Second))

// ./a.out -cpuprofile cpu.out < tools/in/0000.txt && go tool pprof -http=localhost:8888 a.out cpu.out
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

//var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	///////////////////////////////
	flag.Parse()
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
	solver()
	log.Printf("time=%f\n", time.Since(startTime).Seconds())
}

type Point struct {
	y, x int
}

func (p Point) String() string {
	return fmt.Sprintf("%d %d", p.y, p.x)
}

const N int = 15
const M int = 200

var startPoint Point
var keyboard [N][N]byte
var Words [M]string
var points [26][]Point

func read(r io.Reader) {
	var n, m int
	fmt.Fscan(r, &n, &m)
	fmt.Fscan(r, &startPoint.y, &startPoint.x)
	for i := 0; i < N; i++ {
		var keys string
		fmt.Fscan(r, &keys)
		for j := 0; j < N; j++ {
			keyboard[i][j] = keys[j]
		}
	}
	for i := 0; i < M; i++ {
		fmt.Fscan(r, &Words[i])
	}
	for i := 0; i < 26; i++ {
		for j := 0; j < N; j++ {
			for k := 0; k < N; k++ {
				if keyboard[j][k] == byte('A'+i) {
					points[i] = append(points[i], Point{j, k})
				}
			}
		}
	}
}

func solver() {
	read(os.Stdin)
	// wordsの順番を変更して、最終的な文字列を最小にする
	// 文字の重複によって、文字列は縮む
	// shortest Superstring problem
	// 前後の文字列と繋がっていない文字は順番を変更できるので、キーボードの位置を考慮して、順番を変える
	//log.Printf("len=%d\n", len(result))
	//str := greedyOrder(result, points, startPoint)
	bestRoot := []Point{}
	bestScore := 0
	var loop int
	var maxLoop int = 1
	timeLimitForLoop := time.Duration(int(timeLimit)/maxLoop) * time.Millisecond
	for i := 0; i < maxLoop; i++ {
		ws := shortestSuperstring(Words[:], 2)
		ws = shortestMerge(ws)
		ws = shortestSuperstring(ws[:], 1)
		//log.Println(result)
		log.Println(timeLimit, time.Since(startTime))
		log.Println(timeLimit - time.Since(startTime))
		t := minTime(time.Duration(timeLimitForLoop), timeLimit-time.Since(startTime))
		str := beamSearchOrder(ws, startPoint, t)
		rtn2, _ := dpRoot(str, startPoint, true)
		if score(rtn2) > bestScore {
			bestScore = score(rtn2)
			bestRoot = rtn2
		}
		rand.Shuffle(len(Words), func(i, j int) {
			Words[i], Words[j] = Words[j], Words[i]
		})
		elp := time.Since(startTime).Seconds()
		aveerageTime := elp / float64(i+1)
		if elp+aveerageTime+0.1 > timeLimit.Seconds() {
			break
		}
		loop++
		log.Println(bestScore, len(ws))
	}
	for i := 0; i < len(bestRoot); i++ {
		fmt.Println(bestRoot[i])
	}
	log.Printf("score=%d\n", bestScore)
}

var X int = 1

// shortestSuperstring : 単語の文字の重複を考慮して、結合する
// 4 ~ mini 文字の重複までを条件にする // 初期状態で重複なしの5文字の単語集合なので
func shortestSuperstring(words []string, mini int) (w []string) {
	w = make([]string, len(words))
	copy(w, words)
	for {
		var restart bool
		for k := 4; k >= mini; k-- {
			for i := 0; i < len(w); i++ {
				for j := 0; j < len(w); j++ {
					if i == j {
						continue
					}
					if w[i][len(w[i])-k:] == w[j][:k] {
						_, costi := dpRootCache(w[i], false)
						_, costj := dpRootCache(w[j], false)
						newWord := w[i] + w[j][k:]
						_, cost := dpRootCache(newWord, false)
						if costi+costj+X >= cost && rand.Float64() < 1.0 {
							//log.Println(words[i], words[j], costi, costj, cost)
							w[i] = newWord
							w[j] = w[len(w)-1]
							w = w[:len(w)-1]
							restart = true
							break
						}
					}
				}
				if restart {
					break
				}
			}
			if restart {
				break
			}
		}
		if !restart {
			break
		}
	}
	return w
}

// shortestMerge : 単語の最短経路に絞って、結合する
// 一文字の重複（前後の文字が同じ）のみを考慮する // 2文字以上の重複は結合しておく
func shortestMerge(words []string) []string {
	roots := make([][]Point, len(words))
	// まずは、単語の最短経路を求める
	for i := 0; i < len(words); i++ {
		roots[i], _ = dpRoot(words[i], Point{-1, -1}, true)
	}
	loop := true
	for loop {
		loop = false
		for i := 0; i < len(words); i++ {
			for j := 0; j < len(words); j++ {
				if i == j {
					continue
				}
				if roots[i][len(roots[i])-1] == roots[j][0] {
					//log.Println(roots[i][len(roots[i])-1], roots[j][0])
					// 結合する
					newWord := words[i] + words[j][1:]
					words[i] = newWord
					words[j] = words[len(words)-1]
					words = words[:len(words)-1]
					roots[i] = append(roots[i], roots[j][1:]...)
					roots[j] = roots[len(roots)-1]
					roots = roots[:len(roots)-1]
					loop = true
					break
				}
			}
			if loop {
				break
			}
		}
	}
	return words
}

type Node struct {
	used        [200]bool
	str         string
	cost        int
	baseCostSum int
	lastPoint   Point
	trueScore   int
}

// すべての単語が使われている前提
func (n *Node) calcScore() int {
	if n.trueScore != 0 {
		return n.trueScore
	}
	root, _ := dpRoot(n.str, startPoint, true)
	n.trueScore = score(root)
	return n.trueScore
}

func goalCheck(n *Node, m int) bool {
	for i := 0; i < m; i++ {
		if !n.used[i] {
			return false
		}
	}
	return true
}

func baseCostSum(n Node) (sum int) {
	for i := 0; i < len(Words); i++ {
		if !n.used[i] {
			_, cst := dpRootCache(Words[i], false)
			sum += cst
		}
	}
	return sum
}

func generateNodes(n Node, words []string) (nodes []Node) {
	nodes = make([]Node, 0, len(words))
	var str strings.Builder
	for i := 0; i < len(words); i++ {
		if n.used[i] {
			continue
		}
		root, cst := dpRoot(words[i], n.lastPoint, true)
		_, baseCst := dpRootCache(words[i], false)
		cst -= baseCst
		str.Reset()
		str.Grow(len(n.str) + len(words[i]))
		str.WriteString(n.str)
		if len(n.str) > 1 && n.str[len(n.str)-1] == words[i][0] {
			str.WriteString(words[i][1:])
		} else {
			str.WriteString(words[i])
		}
		node := Node{n.used, str.String(), n.cost + cst, n.baseCostSum - baseCst, root[len(root)-1], 0}
		node.used[i] = true
		nodes = append(nodes, node)
	}
	return nodes
}

func beamSearchOrder(words []string, start Point, localTimeLimit time.Duration) string {
	log.Println(localTimeLimit)
	localStartTime := time.Now()
	beamWidth := 20
	minBeamWidth := 1
	maxBeamWidth := 300
	maxGeneration := len(words)
	var lastTime time.Time = time.Now()
	initialNode := Node{[200]bool{}, "", 0, 0, start, 0}
	initialNode.baseCostSum = baseCostSum(initialNode)
	nodes := make([]Node, 0, 200)
	nodes = append(nodes, initialNode)
	nodesSub := make([]Node, 0, 200)
	loop := len(words)
	withCnt := 0
	for generation := 0; generation < loop; generation++ {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].cost < nodes[j].cost
		})
		for i := 0; i < min(beamWidth, len(nodes)); i++ {
			n := generateNodes(nodes[i], words)
			nodesSub = append(nodesSub, n...)
		}
		nodes = make([]Node, len(nodesSub))
		copy(nodes, nodesSub)
		if goalCheck(&nodes[0], len(words)) {
			break
		}
		nodesSub = nodesSub[:0]
		//beamWidthを変更する
		numLeaves := len(nodes)
		now := time.Now()
		if timeLimit > 0.0 && numLeaves > beamWidth && generation <= maxGeneration {
			// この世代での経過時間
			elapsed := now.Sub(lastTime).Seconds()
			// 1beamWidthあたりの経過時間
			timePerWidth := elapsed / float64(beamWidth)
			// 残り時間
			timeLeft := localTimeLimit.Seconds() - now.Sub(localStartTime).Seconds()
			// 残りの世代あたりの使用可能時間
			timePerGenerationLeft := timeLeft / float64(maxGeneration-generation-1)
			// 残り時間からbeamWidthを計算
			beamWidth = int(math.Round(timePerGenerationLeft/timePerWidth) * 0.9)
			beamWidth = max(minBeamWidth, min(beamWidth, maxBeamWidth))
			//log.Println(generation, "elapsed", elapsed, "timePerWidth", timePerWidth, "TimeLeft", TimeLeft, "timePerGenerationLeft", timePerGenerationLeft, "beamWidth", beamWidth)
			log.Println(generation, "beamWidth", beamWidth)
		}
		lastTime = time.Now()
		withCnt += beamWidth
	}
	//log.Println(time.Since(startTime).Seconds())
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].cost < nodes[j].cost
	})
	nodes = nodes[:min(len(nodes), 200)]
	// スコア計算が結構時間がかかるので、注意
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].calcScore() > nodes[j].calcScore()
	})
	//log.Printf("t2=%f\n", t2)
	//log.Println(len(nodes))
	//log.Println(time.Since(startTime).Seconds())
	log.Printf("withCnt=%d\n", withCnt)
	return nodes[0].str
}

const start_temp = 5.0
const end_temp = 0.0
const maxTimeSeconsds = 1.9

func SARoot(word string) (bestSolution []int) {
	iterations := 0
	wordNum := make([]int, len(word))
	for i := 0; i < len(word); i++ {
		wordNum[i] = int(word[i] - 'A')
	}
	currentSolution := make([]int, len(word))
	for i := 0; i < len(word); i++ {
		currentSolution[i] = rand.Intn(len(points[wordNum[i]]))
	}
	bestSolution = make([]int, len(word))
	copy(bestSolution, currentSolution)
	newSolution := make([]int, len(word))
	best := float64(rootLength(word, currentSolution, points))
	for {
		currentTime := time.Since(startTime).Seconds()
		if currentTime > maxTimeSeconsds {
			break
		}
		copy(newSolution, currentSolution)
		w := rand.Intn(len(word))
		n := rand.Intn(len(points[wordNum[w]]))
		newSolution[w] = n
		currentEnergy := float64(rootLength(word, currentSolution, points))
		newEnergy := float64(rootLength(word, newSolution, points))

		// 受理確率を計算
		temp := start_temp + (end_temp-start_temp)*currentTime/maxTimeSeconsds
		acceptanceProbability := math.Exp((currentEnergy - newEnergy) / temp)
		if newEnergy <= currentEnergy || rand.Float64() < acceptanceProbability {
			currentEnergy = newEnergy
			copy(currentSolution, newSolution)
		}

		if currentEnergy < best {
			best = currentEnergy
			copy(bestSolution, newSolution)
		}
		//log.Println(best, newEnergy, currentEnergy, temp, acceptanceProbability)
		iterations++
	}
	log.Printf("iterations=%d\n", iterations)
	return bestSolution
}

func dpRoot(word string, startP Point, needRoot bool) ([]Point, int) {
	dp := make([][N][N]int, len(word))
	root := make([][N][N]Point, len(word))
	//var dp [M * 5][N][N]int
	//var root [M * 5][N][N]Point
	for i := 0; i < len(word); i++ {
		for j := 0; j < N; j++ {
			for k := 0; k < N; k++ {
				dp[i][j][k] = math.MaxInt32
			}
		}
	}
	if startP.y != -1 {
		for i := 0; i < len(points[word[0]-'A']); i++ {
			p := points[word[0]-'A'][i]
			dp[0][p.y][p.x] = distance(startP, p)
		}
	} else {
		for i := 0; i < len(points[word[0]-'A']); i++ {
			dp[0][points[word[0]-'A'][i].y][points[word[0]-'A'][i].x] = 0
		}
	}
	for l := 1; l < len(word); l++ {
		a := points[word[l-1]-'A']
		b := points[word[l]-'A']
		for i := 0; i < len(a); i++ {
			for j := 0; j < len(b); j++ {
				cost := dp[l-1][a[i].y][a[i].x] + distance(a[i], b[j])
				if cost < dp[l][b[j].y][b[j].x] {
					dp[l][b[j].y][b[j].x] = cost
					if needRoot {
						root[l][b[j].y][b[j].x] = a[i]
					}
				}
			}
		}
	}
	minCostIndex := 0
	minCost := dp[len(word)-1][points[word[len(word)-1]-'A'][0].y][points[word[len(word)-1]-'A'][0].x]
	for i := 1; i < len(points[word[len(word)-1]-'A']); i++ {
		if dp[len(word)-1][points[word[len(word)-1]-'A'][i].y][points[word[len(word)-1]-'A'][i].x] < minCost {
			minCostIndex = i
			minCost = dp[len(word)-1][points[word[len(word)-1]-'A'][i].y][points[word[len(word)-1]-'A'][i].x]
		}
	}
	if !needRoot {
		return nil, minCost
	}
	rootPoint := make([]Point, len(word))
	rootPoint[len(word)-1] = points[word[len(word)-1]-'A'][minCostIndex]
	for i := len(word) - 2; i >= 0; i-- {
		rootPoint[i] = root[i+1][rootPoint[i+1].y][rootPoint[i+1].x]
	}
	return rootPoint, minCost
}

var dpRootCacheMap map[string]int

func dpRootCache(word string, needRoot bool) ([]Point, int) {
	if dpRootCacheMap == nil {
		dpRootCacheMap = make(map[string]int)
	}
	if cache, ok := dpRootCacheMap[word]; ok && !needRoot {
		return nil, cache
	} else {
		root, cost := dpRoot(word, Point{-1, -1}, needRoot)
		dpRootCacheMap[word] = cost
		return root, cost
	}
}

func score(ans []Point) (score int) {
	score = 10000
	cost := distance(startPoint, ans[0]) + 1
	for i := 0; i < len(ans)-1; i++ {
		cost += distance(ans[i], ans[i+1]) + 1
	}
	//log.Printf("score=%d cost=%d\n", score-cost, cost)
	return score - cost
}

func rootLength(word string, root []int, points [26][]Point) (length int) {
	for i := 0; i < len(word)-1; i++ {
		length += distance(points[word[i]-'A'][root[i]], points[word[i+1]-'A'][root[i+1]])
	}
	return length
}

func distance(p1, p2 Point) int {
	return abs(p1.y-p2.y) + abs(p1.x-p2.x)
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func minTime(t1, t2 time.Duration) time.Duration {
	if t1 < t2 {
		return t1
	}
	return t2
}
