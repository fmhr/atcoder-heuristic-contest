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
	"time"
)

type Input struct {
	N, T int
	sgm  int32
	w, h []int32
}

func (in Input) Clone() Input {
	t := in
	t.w = make([]int32, in.N)
	t.h = make([]int32, in.N)
	copy(t.w, in.w)
	copy(t.h, in.h)
	return t
}

func input() Input {
	var n, t int
	var sgm int32
	fmt.Scan(&n, &t, &sgm)
	w := make([]int32, n)
	h := make([]int32, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&w[i], &h[i])
		//log.Println(i, w[i], h[i])
	}
	log.Printf("n=%d, t=%d, sgm=%d\n", n, t, sgm)
	return Input{n, t, sgm, w, h}
}

type CmdWithScore struct {
	cmd   Cmd
	score int32
}

type CmdNode struct {
	cmd    Cmd
	parent *CmdNode
}

func (c *CmdNode) cmds() []Cmd {
	cmds := make([]Cmd, 0)
	for p := c; p != nil; p = p.parent {
		cmds = append(cmds, p.cmd)
	}
	// reverse
	for i, j := 0, len(cmds)-1; i < j; i, j = i+1, j-1 {
		cmds[i], cmds[j] = cmds[j], cmds[i]
	}
	return cmds
}

type CMDTree struct {
	root *CmdNode
}

func (c *CMDTree) addTo(p *CmdNode) {
	c.root = p
}

var beamWidth = 200

func BeamSearch(in Input, queryCnt *int) State {
	states := make([]State, 0)
	cmdTree := CMDTree{}
	posTree := PosTree{}
	states = append(states, State{W: 0, H: 0, score: 0, score_t: 0})
	states[0].cmdp = cmdTree.root
	states[0].posP = posTree.root
	subStates := make([]State, 0)
	for t := 0; t < in.N; t++ {
		for w := 0; w < min(len(states), beamWidth); w++ {
			cmds := cmdGenerate(int8(t))
			posts := states[w].posP.posList() // これまでの配置
			now := states[w].Clone()
			for _, cmd := range cmds {
				penalty, pos := now.do(in, cmd, t, posts)
				if penalty == 0 {
					new := now.Clone()
					new.cmdp = &CmdNode{cmd, states[w].cmdp}
					new.posP = &PosNode{t, pos, states[w].posP}
					subStates = append(subStates, new)
				}
				now.undo()
			}
		}
		// ビームサーチ用の評価score_tでソート
		sort.Slice(subStates, func(i, j int) bool {
			return subStates[i].score_t < subStates[j].score_t
		})

		if t < in.N-4 {
			states = subStates[:min(len(subStates), beamWidth)]
		} else {
			states = subStates
		}
		subStates = make([]State, 0)
	}
	// スコアによるソート
	sort.Slice(subStates, func(i, j int) bool {
		return subStates[i].score < subStates[j].score
	})
	var w, h int32
	bestScore := int32(1e9)
	bestTime := 0
	for i := 0; i < len(states) && *queryCnt < in.T; i++ {
		cmds := states[i].cmdp.cmds()
		fmt.Println(len(cmds))
		for _, cmd := range cmds {
			fmt.Println(cmd)
		}
		fmt.Scan(&w, &h)
		*queryCnt++
		//log.Printf("estScore:%d, result:%d, deff:%d\n", states[i].score, w+h, w+h-states[i].score)
		if bestScore > w+h {
			bestScore = w + h
			bestTime = i
		}
	}
	log.Println("bestScore:", bestScore, "bestQueryNum:", bestTime)
	return states[0]
}

func solver(in Input, queryCnt *int) {
	var measured_w, measured_h int
	// beam search
	_ = BeamSearch(in, queryCnt)
	for *queryCnt < in.T {
		fmt.Println(0)
		fmt.Scan(&measured_w, &measured_h)
		*queryCnt++
	}
}

type Cmd struct {
	p int8 // 長方形の番号
	r int8 // 1:90度回転
	d byte // U：下から上に配置 L:右から左に配置
	b int8 // 基準となる長方形の番号
}

func (c Cmd) String() string {
	r := 0
	if c.r == 1 {
		r = 1
	}
	return fmt.Sprintf("%d %d %s %d", c.p, r, string(c.d), c.b)
}

// n: 追加する長方形
func cmdGenerate(n int8) []Cmd {
	cmds := make([]Cmd, 0)
	var r, d, b int8
	for r = 0; r < 2; r++ {
		for d = 0; d < 2; d++ {
			for b = -1; b < int8(n); b++ {
				cmds = append(cmds, Cmd{p: int8(n), r: r, d: "UL"[d], b: b})
			}
		}
	}
	return cmds
}

type Pos struct {
	x1, x2, y1, y2 int32
	t              int32
	r              int8
}

func (p *Pos) reset() {
	p.x1 = -1
	p.x2 = -1
	p.y1 = -1
	p.y2 = -1
	p.r = 0
	p.t = -1
}

type PosNode struct {
	index  int
	pos    Pos
	parent *PosNode
}

func (p *PosNode) posList() []Pos {
	posList := make([]Pos, 0)
	for q := p; q != nil; q = q.parent {
		posList = append(posList, q.pos)
	}
	for i, j := 0, len(posList)-1; i < j; i, j = i+1, j-1 {
		posList[i], posList[j] = posList[j], posList[i]
	}
	return posList
}

type PosTree struct {
	root *PosNode
}

type State struct {
	turn           int32
	W, H           int32
	W2, H2         int32 // 更新前 undo用
	score_t, score int32 // score_t = score + x2 + y2 評価用
	score2         int32 // undo用
	cmdp           *CmdNode
	posP           *PosNode
}

func (s State) Clone() State {
	t := s
	return t
}

func (s *State) undo() {
	s.W = s.W2
	s.H = s.H2
	s.score = s.W + s.H
}

func (s *State) do(in Input, c Cmd, t int, posts []Pos) (penalty float64, pos Pos) {
	w, h := in.w[c.p], in.h[c.p]
	if c.r == 1 {
		w, h = h, w // 90度回転
	}

	var x1, x2, y1, y2 int32
	collision := 0
	if c.d == 'U' {
		x1 = 0 // 基準になるx座標
		if c.b >= 0 {
			x1 = posts[c.b].x2
		}
		x2 = x1 + w
		y1 = 0
		for _, q := range posts {
			if q.t >= 0 {
				if max32(x1, q.x1) < mini32(x2, q.x2) {
					y1 = max32(y1, q.y2)
					//if collision == 0 {
					//// 重なった部分が小さすぎる場合、ペナルティを追加する
					//var diff int32
					//if x1 > q.x1 {
					//diff = max32(x2, q.x2) - mini32(x1, q.x1)
					//} else {
					//diff = max32(x2, q.x2) - mini32(x1, q.x1)
					//}
					//if diff > 0 && diff < int32(in.sgm)*2 {
					//penalty += float64(in.sgm*2) / float64(diff)
					//return penalty
					//}
					//}
					collision++
				} else if collision == 0 {
					// ギリギリすり抜けたときのペナルティ
					var diff int32
					if x1 < q.x1 {
						diff = q.x1 - x2
					} else {
						diff = x1 - q.x2
					}
					if diff > 0 && diff < int32(in.sgm)*2 {
						penalty += float64(in.sgm*2) / float64(diff)
					}
				}
			}
		}
		y2 = y1 + h
		pos = Pos{x1: x1, x2: x2, y1: y1, y2: y2, r: c.r, t: int32(t)}
	} else {
		y1 = 0 // 基準になるy座標
		if c.b >= 0 {
			y1 = posts[c.b].y2
		}
		y2 = y1 + h
		x1 = 0
		for _, q := range posts {
			if q.t >= 0 {
				if max32(y1, q.y1) < mini32(y2, q.y2) {
					x1 = max32(x1, q.x2)
					//if collision == 0 {
					//var diff int32
					//if y1 > q.y1 {
					//diff = max32(y2, q.y2) - mini32(y1, q.y1)
					//} else {
					//diff = max32(y2, q.y2) - mini32(y1, q.y1)
					//}
					//if diff > 0 && diff < int32(in.sgm)*2 {
					//penalty += float64(in.sgm*2) / float64(diff)
					//}
					//}
					collision++
				} else if collision == 0 {
					var diff int32
					if y1 < q.y1 {
						diff = q.y1 - y2
					} else {
						diff = y1 - q.y2
					}
					if diff > 0 && diff < int32(in.sgm)*2 {
						penalty += float64(in.sgm*2) / float64(diff)
					}
				}
			}
		}
		x2 = x1 + w
		pos = Pos{x1: x1, x2: x2, y1: y1, y2: y2, r: c.r, t: int32(t)}
	}
	s.W2 = s.W
	s.H2 = s.H
	s.W = max32(s.W, pos.x2)
	s.H = max32(s.H, pos.y2)
	s.score = s.W + s.H
	s.score_t = s.score + (x1+y1)/20
	return penalty, pos
}

func checkEstimate(in Input, est [][2]float64, stds [][2]float64) {
	input := make([][2]int32, in.N)
	for i := 0; i < in.N; i++ {
		input[i][0] = in.w[i]
		input[i][1] = in.h[i]
	}
	trueSize := make([][2]int32, in.N)
	var sumErr1, sumErr2 int
	for i := 0; i < in.N; i++ {
		for wh := 0; wh < 2; wh++ {
			fmt.Scan(&trueSize[i][wh])
			t := trueSize[i][wh]
			in := input[i][wh]
			est := int32(est[i][wh])
			std := stds[i][wh]
			log.Printf("%d, true:%v, input:%v(%d), est:%v(%d) std:%v\n", i, t, in, in-t, est, est-t, std)
			sumErr1 += int(in-t) * int(in-t)
			sumErr2 += int(est-t) * int(est-t)
		}
	}
	log.Println("出力後に実際のスコアと比較")
	log.Printf("avgErr1=%d avgErr2=%d\n", int(math.Sqrt(float64(sumErr1))), int(math.Sqrt(float64(sumErr2))))
}

// w, hの両方を持つ
type EstimateValue struct {
	mesuredCnt [2]int     // 0:w, 1:h
	mesureSum  [2]float64 // 測定したときの結果の合計
	partyCnt   [2][]int   // 他の長方形が一緒に測定した回数 長方形の番号*2 (w, h)
	sigma2     [2]float64
	alpha      [2]float64
	beta       [2]float64
}

// 観測結果
type Observation struct {
	used   [200]bool
	result float64
}

// estimaterはin.T-1回までqueryを使って推定する
func estimater(in Input, queryCnt *int) ([][2]float64, [][2]float64) {
	queryT := in.T // 推定に使いクエリ回数
	*queryCnt = queryT
	estimateV := make([][2]float64, in.N)
	for i := 0; i < in.N; i++ {
		estimateV[i][0] = float64(0)
		estimateV[i][1] = float64(0)
	}
	puts := make([][]byte, 0)
	rolls := make([][]int, 0)
	var results [][2]float64
	var observations []Observation

	for t := 0; t < queryT; t++ {
		var ulist [200]bool // 下から上に配置、結果はw
		var llist [200]bool // 右から左に配置、結果はh
		// なんこの長方形を使うか
		m := in.N
		ns := selectRandom(in.N, m)
		put := make([]byte, in.N)
		roll := make([]int, in.N)
		for i := 0; i < in.N; i++ {
			put[i] = '.'
		}
		for _, i := range ns {
			// それぞれの長方形をw, hのどちらかに配置する
			put[i] = "UL"[rand.Intn(2)]
			roll[i] = rand.Intn(2)
		}
		fmt.Println(len(ns))
		for i := 0; i < in.N; i++ {
			if put[i] != '.' {
				fmt.Printf("%d %d %s %d\n", i, roll[i], string(put[i]), -1)
				//log.Printf("%d %d %s %d\n", i, 0, string(put[i]), -1)
				if put[i] == 'U' {
					ulist[i*2+(1+roll[i])%2] = true
				} else if put[i] == 'L' {
					llist[i*2+roll[i]] = true
				}
			}
		}
		//log.Println("ulist", ulist[0], ulist[1])
		//log.Println("llist", llist[0], llist[1])
		var w, h float64
		fmt.Scan(&w, &h)
		obw := Observation{used: ulist, result: w}
		obh := Observation{used: llist, result: h}
		observations = append(observations, obw)
		observations = append(observations, obh)
		results = append(results, [2]float64{w, h})
		puts = append(puts, put)
		rolls = append(rolls, roll)
	}

	// 推定
	// 長方形のw,hを順番に推定する
	// 上の測定回数をまず数える
	estise := make([]EstimateValue, in.N)
	for i := 0; i < in.N; i++ {
		for wh := 0; wh < 2; wh++ {
			estise[i].partyCnt[wh] = make([]int, in.N*2)
			if wh == 0 {
				estise[i].alpha[wh] = 3
			} else {
				estise[i].alpha[wh] = 3
			}
			estise[i].beta[wh] = (3 - 1) * float64(in.sgm*in.sgm)
			estise[i].sigma2[wh] = float64(in.sgm * in.sgm)
		}
	}
	log.Println(estise[0].sigma2[0])
	for k := 0; k < len(puts); k++ {
		put := puts[k]
		roll := rolls[k]
		// k回目の測定結果
		first := true // 一番最初の長方形は、両方測定される
		for i := 0; i < len(put); i++ {
			// 0だけは常に両方測定される
			if put[i] == '.' {
				continue
			}
			if put[i] == 'L' || first {
				wh := 0 + roll[i]
				estise[i].mesuredCnt[wh]++
				estise[i].mesureSum[wh] += results[k][0]
				party := slicesIndex(put, 'L') // 一緒に測定された長方形の番号
				for _, j := range party {
					// j番目の長方形がi番目の長方形と一緒に測定された回数
					// 0:w, 1:h
					ab := j*2 + 0 + roll[j]
					estise[i].partyCnt[wh][ab]++
				}
			}
			if put[i] == 'U' || first {
				wh := 1 - roll[i]
				estise[i].mesuredCnt[wh]++
				estise[i].mesureSum[wh] += results[k][1]
				party := slicesIndex(put, 'U')
				for _, j := range party {
					ab := j*2 + 1 - roll[j]
					estise[i].partyCnt[wh][ab]++
				}
			}
			first = false
		}
	}
	for i := 0; i < in.N; i++ {
		estise[i].partyCnt[0][i*2] = 0 // 自分自身は加算しない 更新するので
		estise[i].partyCnt[1][i*2+1] = 0
		estise[i].mesureSum[0] += float64(in.w[i])
		estise[i].mesureSum[1] += float64(in.h[i])
		estise[i].mesuredCnt[0] += 1
		estise[i].mesuredCnt[1] += 1
	}
	// 例 x番目のwを推定する
	//	for x := 0; x < in.N; x++ {
	//for wh := 0; wh < 2; wh++ {
	//log.Printf("estise %d %.2f\n", estise[x].mesuredCnt[wh], estise[x].mesureSum[wh])
	//log.Println("partyCnt", estise[x].partyCnt[wh])
	//}
	//}

	maxStep := 5000000
	var samplese [100][2][]float64
	timNow := time.Now()
	var step int
	burnIn := step / 3
	for step = 0; step < maxStep; step++ {
		elapsed := time.Since(timNow)
		if elapsed > 2000*time.Millisecond {
			break
		}
		for x := 0; x < in.N; x++ {
			for wh := 0; wh < 2; wh++ {
				//log.Printf("estise %d %.2f\n", estise[x].mesuredCnt[wh], estise[x].mesureSum[wh])
				//log.Println("partyCnt", estise[x].partyCnt[wh], x, wh)
				value := estise[x].mesureSum[wh] // 0番目のwの測定結果の合計
				// valueから他の長方形のwを引く
				for i := 0; i < in.N*2; i++ {
					// i番目とx番目が一緒に測定された回数 * 推定値
					a := i / 2
					b := i % 2
					value -= float64(estise[x].partyCnt[wh][i]) * estimateV[a][b]
				}
				// muの更新
				mean := value / float64(estise[x].mesuredCnt[wh]) // 0番目のwの参加回数を引いて平均を取る
				mu := rand.NormFloat64()*math.Sqrt(estise[x].sigma2[wh]) + mean
				// sigma^2の更新
				alphaNew := estise[x].alpha[wh] + float64(estise[x].mesuredCnt[wh])/2
				betaNew := estise[x].beta[wh]
				for i := 0; i < len(observations); i++ {
					if observations[i].used[x*2+wh] {
						mu_t := observations[i].result
						for j := 0; j < in.N*2; j++ {
							if i != j && observations[i].used[j] {
								mu_t -= estimateV[j/2][j%2]
							}
						}
						betaNew += math.Pow(mu_t-mu, 2) / 2
					}
				}
				estise[x].sigma2[wh] = sampleGamma(alphaNew, 1/betaNew)
				estise[x].alpha[wh] = alphaNew
				estise[x].beta[wh] = betaNew
				estise[x].sigma2[wh] = float64(in.sgm*in.sgm) / 2

				//new := rand.NormFloat64()*sigma2 + mean
				estimateV[x][wh] = mu
				samplese[x][wh] = append(samplese[x][wh], estimateV[x][wh])
			}
		}
	}
	log.Println("STEP", step)
	stds := make([][2]float64, in.N)
	stdSum := 0
	for i := 0; i < in.N; i++ {
		for wh := 0; wh < 2; wh++ {
			mean := mean(samplese[i][wh][burnIn:])
			estimateV[i][wh] = mean
			std := std(samplese[i][wh][burnIn:])
			stds[i][wh] = std
			stdSum += int(std)
		}
	}
	for i := 0; i < in.N; i++ {
		for j := 0; j < 2; j++ {
			//log.Printf("%d %d sigma2 %.2f\n", i, j, estise[i].sigma2[j])
		}
	}
	return estimateV, stds
}

var ATCODER int
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		ATCODER = 1
		log.SetOutput(io.Discard)
	}
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

	startTIme := time.Now()
	in := input()
	insub := in.Clone()
	queryCnt := 0
	est, stds := estimater(in, &queryCnt)
	for i := 0; i < in.N; i++ {
		in.w[i] = int32(est[i][0])
		in.h[i] = int32(est[i][1])
	}
	estTime := time.Since(startTIme).Seconds()
	log.Printf("estTime=%.3f estQueryUse:%d\n", estTime, queryCnt)
	solver(in, &queryCnt)
	elap := time.Since(startTIme)
	log.Printf("bsTime=%.3f s\n", elap.Seconds()-estTime)
	log.Printf("time=%.2f ms\n", elap.Seconds())
	if ATCODER != 1 {
		checkEstimate(insub, est, stds)
	}
}

// util
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mini32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func abs32(a int32) int32 {
	if a < 0 {
		return -a
	}
	return a
}

// UまたはLが出てるindex または最初にでてくるUL
func slicesIndex(slices []byte, v byte) (indexes []int) {
	for i, s := range slices {
		if s == v {
			indexes = append(indexes, i)
		} else if s != '.' && len(indexes) == 0 {
			indexes = append(indexes, i)
		}
	}
	return
}

// 平均値を計算
func mean(samples []float64) float64 {
	sum := 0.0
	for _, v := range samples {
		sum += v
	}
	return sum / float64(len(samples))
}

// 標準偏差を計算
func std(samples []float64) float64 {
	m := mean(samples)
	var sum float64
	for _, v := range samples {
		sum += (v - m) * (v - m)
	}
	return math.Sqrt(sum / float64(len(samples)))
}

// 0からnの範囲からm個の整数をランダムに選んで返す関数
func selectRandom(n, m int) []int {
	// 0からnの整数をスライスに格納
	nums := make([]int, n)
	for i := 0; i < n; i++ {
		nums[i] = i
	}

	// ランダムにシャッフル
	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})
	r := nums[:m]
	sort.Ints(r)
	// 最初のm個を返す
	return r
}

// メトロポリス法で受理するかどうかを判断する関数
func shouldAccept(xCurrent, xProposed, sigma float64) bool {
	if xProposed < 0 {
		return false
	}
	// 現在の値と提案された値を使って受理確率を計算
	currentTerm := math.Pow(xCurrent, 2) / (2 * math.Pow(sigma, 2))
	proposedTerm := math.Pow(xProposed, 2) / (2 * math.Pow(sigma, 2))
	alpha := math.Exp(-(proposedTerm - currentTerm))

	// 一様乱数を生成
	randomValue := rand.Float64() // 0から1の乱数

	// 乱数が受理確率以下なら受理
	return randomValue <= alpha
}

// ガンマ分布から乱数を生成する関数
func sampleGamma(alpha, beta float64) float64 {
	if alpha <= 0 || beta <= 0 {
		log.Println(alpha, beta)
		panic("shape and scale must be positive")
	}
	if alpha < 1 {
		return sampleGamma(1.0+alpha, beta) * math.Pow(rand.Float64(), 1.0/alpha)
	}

	// Marsaglia and Tsang method
	d := alpha - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)
	for {
		x := rand.NormFloat64()
		v := 1.0 + c*x
		if v > 0 {
			v = v * v * v // v = v^3
			u := rand.Float64()
			if u < 1.0-0.0331*math.Pow(x, 4) {
				return d * v * beta
			}
			if math.Log(u) < 0.5*x*x+d*(1.0-v+math.Log(v)) {
				return d * v * beta
			}
		}
	}
}

// 逆ガンマ分布から乱数を生成する関数
func sampleInverseGamma(alpha, beta float64) float64 {
	// ガンマ分布の逆数を取る
	gammaSample := sampleGamma(alpha, beta)
	return 1.0 / gammaSample
}
