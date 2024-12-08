package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"
)

type Input struct {
	N, T, sgm int
	w, h      []int
}

func (in Input) Clone() Input {
	t := in
	t.w = make([]int, in.N)
	t.h = make([]int, in.N)
	copy(t.w, in.w)
	copy(t.h, in.h)
	return t
}

func input() Input {
	var n, t, sgm int
	fmt.Scan(&n, &t, &sgm)
	w := make([]int, n)
	h := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&w[i], &h[i])
		//log.Println(i, w[i], h[i])
	}
	log.Printf("n=%d, t=%d, sgm=%d\n", n, t, sgm)
	return Input{n, t, sgm, w, h}
}

type CmdWithScore struct {
	cmd   Cmd
	score int
}

var beamWidth = 10

func BeamSearch(in Input, queryCnt *int) State {
	startTimeBS := time.Now()
	states := make([]State, 0)
	states = append(states, NewState(in))
	subStates := make([]State, 0)
	for t := 0; t < in.N; t++ {
		for w := 0; w < min(len(states), beamWidth); w++ {
			cmds := cmdGenerate(t)
			for _, cmd := range cmds {
				now := states[w].Clone()
				now.do(in, cmd, t)
				now.cmds = append(now.cmds, cmd)
				subStates = append(subStates, now)
			}
		}
		// ビームサーチ用の評価score_tでソート
		sort.Slice(subStates, func(i, j int) bool {
			return subStates[i].score_t < subStates[j].score_t
		})
		if t < in.N-1 {
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
	log.Printf("beam_score=%d\n", states[0].score)
	var w, h int
	log.Println("queryCnt", *queryCnt)
	for i := 0; i < len(states) && *queryCnt < in.T; i++ {
		fmt.Println(len(states[i].cmds))
		for _, cmd := range states[i].cmds {
			fmt.Println(cmd)
		}
		fmt.Scan(&w, &h)
		*queryCnt++
		log.Printf("estScore:%d, result:%d, deff:%d\n", states[i].score, w+h, states[i].score-w-h)
	}
	timeBS := time.Since(startTimeBS)
	log.Printf("bs_time=%.2f\n", timeBS.Seconds())
	return states[0]
}

func solver(in Input, queryCnt *int) {
	var measured_w, measured_h int
	// beam search
	_ = BeamSearch(in, queryCnt)
	log.Printf("queryCnt:%d /in.T:%d\n", queryCnt, in.T)
	log.Printf("rest querySize:%d\n", in.T-*queryCnt)
	for *queryCnt < in.T {
		fmt.Println(0)
		fmt.Scan(&measured_w, &measured_h)
		*queryCnt++
	}
}

type Cmd struct {
	p int  // 長方形の番号
	r bool // 1:90度回転
	d byte // U：下から上に配置 L:右から左に配置
	b int  // 基準となる長方形の番号
}

func (c Cmd) String() string {
	r := 0
	if c.r {
		r = 1
	}
	return fmt.Sprintf("%d %d %s %d", c.p, r, string(c.d), c.b)
}

// n: 追加する長方形
func cmdGenerate(n int) []Cmd {
	cmds := make([]Cmd, 0)
	for r := 0; r < 2; r++ {
		for d := 0; d < 2; d++ {
			for b := -1; b < n; b++ {
				cmds = append(cmds, Cmd{p: n, r: r == 1, d: "UL"[d], b: b})
			}
		}
	}
	return cmds
}

type Pos struct {
	x1, x2, y1, y2 int
	r              bool
	t              int
}

func (p *Pos) reset() {
	p.x1 = -1
	p.x2 = -1
	p.y1 = -1
	p.y2 = -1
	p.r = false
	p.t = -1
}

type State struct {
	turn           int
	pos            [100]Pos
	W, H           int
	W2, H2         int // 更新前 undo用
	score_t, score int // score_t = score + x2 + y2 評価用
	cmds           []Cmd
}

func (s State) Clone() State {
	t := s
	t.cmds = make([]Cmd, len(s.cmds))
	copy(t.cmds, s.cmds)
	return t
}

func NewState(in Input) State {
	s := State{}
	s.turn = 0
	for i := 0; i < 100; i++ {
		s.pos[i].reset()
	}
	s.W = 0
	s.H = 0
	s.W2 = 0
	s.H2 = 0
	s.score_t = 0
	s.score = 0
	s.cmds = make([]Cmd, 0, in.N)
	return s
}

func (s *State) undo(c Cmd) {
	s.pos[c.p].reset()
	s.W = s.W2
	s.H = s.H2
	s.score = s.W + s.H
}

func (s *State) do(in Input, c Cmd, t int) {
	// cmdのチェック
	if s.pos[c.p].t >= 0 {
		log.Println("c.p:", c.p, s.pos[c.p].t)
		log.Println("c:", c, s.pos[c.p].t)
		panic("already used")
	} else if c.b >= 0 && s.pos[c.b].t < 0 {
		log.Println(c.String())
		log.Printf("b=%d, t=%d\n", c.b, s.pos[c.b].t)
		panic("not used")
	}
	w, h := in.w[c.p], in.h[c.p]
	if c.r {
		w, h = h, w // 90度回転
	}

	var x1, x2, y1, y2 int
	var penalty int
	if c.d == 'U' {
		x1 = 0 // 基準になるx座標
		if c.b >= 0 {
			x1 = s.pos[c.b].x2
		}
		x2 = x1 + w
		y1 = 0
		for _, q := range s.pos {
			if q.t >= 0 {
				if max(x1, q.x1) < min(x2, q.x2) {
					y1 = max(y1, q.y2)
				} else {
					// 横をスレスレに通り抜けたとき
					c := max(x1, q.x1) - min(x2, q.x2)
					penalty += in.sgm / (c + 1)
				}
			}
		}
		y2 = y1 + h
		s.pos[c.p] = Pos{x1, x2, y1, y2, c.r, t}
	} else {
		y1 = 0 // 基準になるy座標
		if c.b >= 0 {
			y1 = s.pos[c.b].y2
		}
		y2 = y1 + h
		x1 = 0
		for _, q := range s.pos {
			if q.t >= 0 {
				if max(y1, q.y1) < min(y2, q.y2) {
					x1 = max(x1, q.x2)
				} else {
					// 縦をスレスレに通り抜けたとき
					c := max(y1, q.y1) - min(y2, q.y2)
					penalty += in.sgm / (c + 1)
				}
			}
		}
		x2 = x1 + w
		s.pos[c.p] = Pos{x1, x2, y1, y2, c.r, t}
	}
	s.W2 = s.W
	s.H2 = s.H

	s.W = max(s.W, s.pos[c.p].x2)
	s.H = max(s.H, s.pos[c.p].y2)
	s.score = s.W + s.H
	s.score_t = s.score + (min(x1, y1))/20 + penalty
}

func (s *State) query(in Input, cmd []Cmd) {
	for t, c := range cmd {
		s.do(in, c, t)
	}
}

func checkEstimate(in Input, est [][2]float64, stds [][2]float64) {
	input := make([][2]int, in.N)
	for i := 0; i < in.N; i++ {
		input[i][0] = in.w[i]
		input[i][1] = in.h[i]
	}
	trueSize := make([][2]int, in.N)
	sumErr1 := 0
	sumErr2 := 0
	for i := 0; i < in.N; i++ {
		for wh := 0; wh < 2; wh++ {
			fmt.Scan(&trueSize[i][wh])
			t := trueSize[i][wh]
			in := input[i][wh]
			est := int(est[i][wh])
			std := int(stds[i][wh])
			log.Printf("%d, true:%v, input:%v(%d), est:%v(%d) std:%v\n", i, t, in, in-t, est, est-t, std)
			sumErr1 += absInt(in - t)
			sumErr2 += absInt(est - t)
		}
	}
	log.Printf(" avgErr1=%d avgErr2=%d\n", sumErr1/in.N, sumErr2/in.N)
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

// estimaterはin.T-1回までqueryを使って推定する
func estimater(in Input, queryCnt *int) ([][2]float64, [][2]float64) {
	queryT := min(in.N, in.T-1) // 推定に使いクエリ回数
	*queryCnt = queryT
	estimateV := make([][2]float64, in.N)
	for i := 0; i < in.N; i++ {
		estimateV[i][0] = float64(in.w[i])
		estimateV[i][1] = float64(in.h[i])
	}
	puts := make([][]byte, 0)
	rolls := make([][]int, 0)
	var results [][2]float64
	for t := 0; t < queryT; t++ {
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
			}
		}
		var w, h float64
		fmt.Scan(&w, &h)
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
				estise[i].alpha[wh] = float64(in.w[i])
			} else {
				estise[i].alpha[wh] = float64(in.h[i])
			}
			estise[i].beta[wh] = 1.0
			estise[i].sigma2[wh] = float64(in.sgm * in.sgm)
		}
	}
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

	maxStep := 50000
	var samplese [100][2][]float64
	timNow := time.Now()
	var step int
	for step = 0; step < maxStep; step++ {
		elapsed := time.Since(timNow)
		if elapsed > 1000*time.Millisecond {
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
				betaNew := estise[x].beta[wh] + 0.5*float64(estise[x].mesuredCnt[wh])*(math.Pow(mean-mu, 2))
				estise[x].sigma2[wh] = 1 / sampleGamma(alphaNew, 1.0/betaNew)

				//new := rand.NormFloat64()*sigma2 + mean
				estimateV[x][wh] = mu
				samplese[x][wh] = append(samplese[x][wh], estimateV[x][wh])
			}
		}
	}
	burnIn := step / 3
	//log.Println(sigma)
	stds := make([][2]float64, in.N)
	for i := 0; i < in.N; i++ {
		for wh := 0; wh < 2; wh++ {
			mean := mean(samplese[i][wh][burnIn:])
			estimateV[i][wh] = mean
			std := std(samplese[i][wh][burnIn:])
			stds[i][wh] = std
			//log.Println(i, wh, int(mean), int(std))
		}
	}
	log.Printf("estTime=%.2f\n", time.Since(timNow).Seconds())
	return estimateV, stds
}

var ATCODER int

func main() {
	if os.Getenv("ATCODER") == "1" {
		ATCODER = 1
		log.SetOutput(io.Discard)
	}
	log.SetFlags(log.Lshortfile)
	startTIme := time.Now()
	in := input()
	insub := in.Clone()
	queryCnt := 0
	est, stds := estimater(in, &queryCnt)
	//	fmt.Println("0")
	//var w, h int
	//fmt.Scan(&w, &h)
	//log.Println("result", w, h)
	for i := 0; i < in.N; i++ {
		in.w[i] = int(est[i][0])
		in.h[i] = int(est[i][1])
	}
	solver(in, &queryCnt)
	elap := time.Since(startTIme)
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

func absInt(a int) int {
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

func sampleGamma(shape, scale float64) float64 {
	if shape <= 0 || scale <= 0 {
		log.Println(shape, scale)
		panic("shape and scale must be positive")
	}
	if shape < 1 {
		return sampleGamma(1.0+shape, scale) * math.Pow(rand.Float64(), 1.0/shape)
	}

	// Marsaglia and Tsang method
	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)
	for {
		x := rand.NormFloat64()
		v := 1.0 + c*x
		if v > 0 {
			v = v * v * v // v = v^3
			u := rand.Float64()
			if u < 1.0-0.0331*math.Pow(x, 4) {
				return d * v * scale
			}
			if math.Log(u) < 0.5*x*x+d*(1.0-v+math.Log(v)) {
				return d * v * scale
			}
		}
	}
}
