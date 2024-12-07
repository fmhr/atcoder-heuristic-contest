package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"
)

type Input struct {
	N, T, sgm int
	w, h      []int
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

var beamWidth = 20

func BeamSearch(in Input, queryCnt *int) State {
	states := make([]State, 0)
	states = append(states, NewState(in))
	subStates := make([]State, 0)
	for t := 0; t < in.N; t++ {
		if t > in.N-4 {
			beamWidth = 40
		}
		for w := 0; w < min(len(states), beamWidth); w++ {
			cmds := cmdGenerate(t)
			for _, cmd := range cmds {
				now := states[w].Clone()
				now.do(in, cmd, t, 0)
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
	for i := 0; i < len(states) && *queryCnt < in.T; i++ {
		fmt.Println(len(states[i].cmds))
		for _, cmd := range states[i].cmds {
			fmt.Println(cmd)
		}
		fmt.Scan(&w, &h)
		*queryCnt++
		log.Printf("estScore:%d, result:%d, deff:%d\n", states[i].score, w+h, states[i].score-w-h)
	}

	return states[0]
}

func solver(in Input) {
	queryCnt := 0
	var measured_w, measured_h int
	// beam search
	_ = BeamSearch(in, &queryCnt)
	log.Printf("queryCnt:%d /in.T:%d\n", queryCnt, in.T)
	log.Printf("rest querySize:%d\n", in.T-queryCnt)
	for queryCnt < in.T {
		fmt.Println(0)
		fmt.Scan(&measured_w, &measured_h)
		queryCnt++
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

func (s *State) do(in Input, c Cmd, t int, clearance int) {
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
	if c.d == 'U' {
		x1 = 0 // 基準になるx座標
		if c.b >= 0 {
			x1 = s.pos[c.b].x2
		}
		x2 = x1 + w
		y1 = 0
		for _, q := range s.pos {
			if q.t >= 0 && max(x1, q.x1) < min(x2, q.x2)-clearance {
				y1 = max(y1, q.y2)
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
			if q.t >= 0 && max(y1, q.y1) < min(y2, q.y2)-clearance {
				x1 = max(x1, q.x2)
			}
		}
		x2 = x1 + w
		s.pos[c.p] = Pos{x1, x2, y1, y2, c.r, t}
	}
	s.W2 = s.W
	s.H2 = s.H
	penalty := 0
	if s.W < s.pos[c.p].x2 || s.H < s.pos[c.p].y2 {
		penalty += 10000000
	}
	s.W = max(s.W, s.pos[c.p].x2)
	s.H = max(s.H, s.pos[c.p].y2)
	s.score = s.W + s.H
	s.score_t = s.score + (min(x1, y1))%100 + penalty
}

func (s *State) query(in Input, cmd []Cmd) {
	for t, c := range cmd {
		s.do(in, c, t, in.sgm)
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
			//std := int(stds[i][wh])
			//log.Printf("%d, true:%v, input:%v(%d), est:%v(%d) std:%v\n", i, t, in, in-t, est, est-t, std)
			sumErr1 += absInt(in - t)
			sumErr2 += absInt(est - t)
		}
	}
	log.Printf("in.Sigm=%d, avgErr1=%d avgErr2=%d\n", in.sgm, sumErr1/in.N, sumErr2/in.N)
}

type EstimateValue struct {
	mesuredCnt [2]int     // 0:w, 1:h
	mesureSum  [2]float64 // 測定したときの結果の合計
	partyCnt   [2][]int   // 他の長方形が一緒に測定した回数
}

// estimaterはin.T-1回までqueryを使って推定する
func estimater(in Input) ([][2]float64, [][2]float64) {
	estimateV := make([][2]float64, in.N)
	for i := 0; i < in.N; i++ {
		estimateV[i][0] = float64(in.w[i])
		estimateV[i][1] = float64(in.h[i])
	}
	puts := make([][]byte, 0)
	var results [][2]float64
	for t := 0; t < in.T; t++ {
		// なんこの長方形を使うか
		m := in.N
		ns := selectRandom(in.N, m)
		put := make([]byte, in.N)
		for i := 0; i < in.N; i++ {
			put[i] = '.'
		}
		for _, i := range ns {
			// それぞれの長方形をw, hのどちらかに配置する
			put[i] = "UL"[rand.Intn(2)]
		}
		fmt.Println(len(ns))
		for i := 0; i < in.N; i++ {
			if put[i] != '.' {
				fmt.Printf("%d %d %s %d\n", i, 0, string(put[i]), -1)
				//log.Printf("%d %d %s %d\n", i, 0, string(put[i]), -1)
			}
		}
		var w, h float64
		fmt.Scan(&w, &h)
		results = append(results, [2]float64{w, h})
		puts = append(puts, put)
	}

	// 推定
	// 長方形のw,hを順番に推定する
	// 上の測定回数をまず数える
	estise := make([]EstimateValue, in.N)
	for i := 0; i < in.N; i++ {
		estise[i].partyCnt[0] = make([]int, in.N)
		estise[i].partyCnt[1] = make([]int, in.N)
	}
	for k, p := range puts {
		// k回目の測定結果
		first := true // 一番最初の長方形は、両方測定される
		for i, d := range p {
			// 0だけは常に両方測定される
			if d == '.' {
				continue
			}
			if d == 'L' || first {
				estise[i].mesuredCnt[0]++
				estise[i].mesureSum[0] += results[k][0]
				party := slicesIndex(p, 'L') // 一緒に測定された長方形の番号
				for _, j := range party {
					estise[i].partyCnt[0][j]++
				}
			}
			if d == 'U' || first {
				estise[i].mesuredCnt[1]++
				estise[i].mesureSum[1] += results[k][1]
				party := slicesIndex(p, 'U')
				for _, j := range party {
					estise[i].partyCnt[1][j]++
				}
			}
			first = false
		}
	}
	for i := 0; i < in.N; i++ {
		estise[i].partyCnt[0][i] = 0 // 自分自身は加算しない 更新するので
		estise[i].partyCnt[1][i] = 0
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

	maxStep := 100000
	burnIn := maxStep / 3
	sigma := float64(in.sgm)
	var samplese [100][2][]float64
	for step := 0; step < maxStep; step++ {
		for x := 0; x < in.N; x++ {
			for wh := 0; wh < 2; wh++ {
				//log.Printf("estise %d %.2f\n", estise[x].mesuredCnt[wh], estise[x].mesureSum[wh])
				//log.Println("partyCnt", estise[x].partyCnt[wh], x, wh)
				value := estise[x].mesureSum[wh] // 0番目のwの測定結果の合計
				// valueから他の長方形のwを引く
				for i := 0; i < in.N; i++ {
					// i番目とx番目が一緒に測定された回数 * 推定値
					value -= float64(estise[x].partyCnt[wh][i]) * estimateV[i][wh]
				}
				mean := value / float64(estise[x].mesuredCnt[wh]) // 0番目のwの参加回数を引いて平均を取る
				sigma2 := sigma / float64(estise[x].mesuredCnt[wh])
				new := rand.NormFloat64()*sigma2/math.Sqrt(2.0) + mean
				estimateV[x][wh] = new
				samplese[x][wh] = append(samplese[x][wh], estimateV[x][wh])
			}
		}
	}
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
	return estimateV, stds
}

func main() {
	log.SetFlags(log.Lshortfile)
	startTIme := time.Now()
	in := input()
	//est, stds := estimater(in)
	//checkEstimate(in, est, stds)
	solver(in)
	elap := time.Since(startTIme)
	log.Printf("time_ms=%d ms\n", elap.Milliseconds())
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
