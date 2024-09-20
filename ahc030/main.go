package main

import (
	"container/list"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"slices"
	"sort"
	"time"
)

// TODO upsolve中。動かない。
func main() {
	startTime := time.Now()
	input := readInput()
	// 時間調整のために難易度（＝必要ターン数）を推定する。可能な配置数が多く、epsが大きいほど難易度が高い
	numPossibilities := 1
	for i := 0; i < input.m; i++ {
		numPossibilities *= (input.n - input.maxI[i]) * (input.n - input.maxJ[i])
	}
	// capacity=エントロピーの平均？
	capacity := 1.0*input.eps*math.Log2(input.eps) + (1.0-input.eps)*math.Log2((1.0-input.eps))
	difficulty := math.Log2(float64(numPossibilities)) / capacity
	// ２つのポリオミノの位置を入れ替える操作を行うために、入れ替えた際にどれだけ１をずらせば良いかを予め計算しておく
	// swap[p][q] := pとq+Δができるだけ一致するようなΔ
	swap := make([][][]int, input.m, input.m)
	for i := 0; i < input.m; i++ {
		swap[i] = make([][]int, input.m, input.m)
	}
	for p := 0; p < input.m; p++ {
		bs := make([]bool, input.n2, input.n2)
		for _, ij := range input.ts[p] {
			bs[ij] = true
		}
		for q := 0; q < input.m; q++ {
			if p == q {
				continue
			}
			list := make([][3]int, 0, input.n2)
			for di := -input.maxI[q]; di < input.n-input.maxI[p]; di++ {
				for dj := -input.maxJ[q]; dj < input.n-input.maxJ[p]; dj++ {
					count := 0
					for _, ij := range input.ts[q] {
						i := ij/uint(input.n) + uint(di)
						j := ij%uint(input.n) + uint(dj)
						if i < uint(input.n) && j < uint(input.n) && bs[i*uint(input.n)+j] {
							count++
						}
					}
					list = append(list, [3]int{count, di, dj})
				}
			}
			sort.Slice(list, func(i, j int) bool {
				if list[i][0] == list[j][0] {
					if list[i][1] == list[j][1] {
						return list[i][2] < list[j][2]
					}
					return list[i][1] < list[j][1]
				}
				return list[i][0] < list[j][0]
			})
			list = list[:4]
			ijList := make([]int, 4, 4)
			for i := 0; i < 4; i++ {
				ijList[i] = list[i][1]*input.n + list[i][2]
			}
			swap[p][q] = ijList
		}
	}

	sim := NewSim(input)
	state := NewState(input)
	// 尤度が高い配置候補を覚えておき、更新してく
	pool := make([]Entry, 0, 0)
	ITER := int(math.Min(math.Round(30000.0*math.Pow(2, 160.0/difficulty)), 5000000))
	for t := 0; ; t++ {
		fmt.Printf("# time = %.3f", time.Since(startTime).Seconds())
		if sim.rem == 0 {
			log.Println("!log giveup 1", input.n, input.m)
			break
		}
		if time.Since(startTime) > 2900*time.Millisecond {
			sim.giveup()
			break
		}
		// 各配置の対数尤度を計算
		for _, e := range pool {
			if len(e.count) == 0 && len(e.ps) != 0 {
				e.count = input.get_count(e.ps)
			}
			e.prob = sim.ln_prob(state, e.count, e.ps)
		}
		sort.Slice(pool, func(i, j int) bool {
			return pool[i].prob < pool[j].prob
		})
		set := make(map[uint64]float64)
		for _, e := range pool {
			set[e.hash] = e.prob
		}
		if t == 0 {
			// 1ターン目はすべての配置が等確率なのでランダムに候補を生成する
			for i := 0; i < ITER; i++ {
				for p := 0; p < input.m; p++ {
					i := rand.Intn(input.n - input.maxI[p])
					j := rand.Intn(input.n - input.maxJ[p])
					state.move_to(uint(p), uint(i*input.n+j))
				}
				if _, ok := set[state.hash]; !ok {
					set[state.hash] = 0.0
					pool = append(pool, Entry{
						hash:  state.hash,
						prob:  0.0,
						ps:    append([]uint(nil), state.ps...),
						count: append([]uint8(nil), state.count...),
					})
				}
			}
		} else {
			// １番尤度が高い配置からスタートし、焼きなまし法を実行することで配置行をを沢山生成する
			crt := pool[0].prob
			for p := 0; p < input.m; p++ {
				state.move_to(uint(p), pool[0].ps[p])
			}
			fmt.Printf("# %.4f -> ", crt)
			max := crt
			T0 := 2.0
			T1 := 1.0
			// TLEにならないように、残り時間が少なくなったら反復回数を減らす
			ITER = int(float64(ITER) * (3.0 - math.Min(1.0, time.Since(startTime).Seconds())))
			for t := 0; t < ITER; t++ {
				temp := T0 + (T1-T0)*float64(t)/float64(ITER)
				coin := rand.Float64() * 10.0
				if coin < 3.0 {
					// ポリオミノをランダムに選び、上下左右に１マス移動
					p := rand.Intn(input.m)
					dij := DIJ[rand.Intn(4)]
					i2 := state.ps[p]/uint(input.n) + uint(dij[0])
					j2 := state.ps[p]%uint(input.n) + uint(dij[1])
					if i2 < uint(input.n-input.maxI[p]) && j2 < uint(input.n-input.maxJ[p]) {
						bk := state.ps[p]
						state.move_to(uint(p), i2*uint(input.n)+j2)
						next, ok := set[state.hash]
						if !ok {
							next = sim.ln_prob_state(&state)
							set[state.hash] = next
							if next-max >= 10.0 {
								pool = append(pool, Entry{
									hash:  state.hash,
									prob:  next,
									ps:    append([]uint(nil), state.ps...),
									count: append([]uint8(nil), state.count...),
								})
							}
						}
						if crt <= next || rand.Float64() < math.Exp((next-crt)/temp) {
							crt = next
						} else {
							state.move_to(uint(p), bk)
						}
					}
				} else if coin < 4.0 {
					// ポリオミノをランダムに選び、ランダムな位置に移動
					p := rand.Intn(input.m)
					i2 := rand.Intn(input.n - input.maxI[p])
					j2 := rand.Intn(input.n - input.maxJ[p])
					bk := slices.Clone(state.ps) // * スライスのコピーはこの方法が正しい。他の箇所も同様
					state.move_to(uint(p), uint(i2*input.n+j2))
					next, ok := set[state.hash]
					if !ok {
						next = sim.ln_prob_state(&state)
						set[state.hash] = next
						if next-max >= -10.0 {
							pool = append(pool, Entry{
								hash:  state.hash,
								prob:  next,
								ps:    append([]uint(nil), state.ps...),
								count: append([]uint8(nil), state.count...),
							})
						}
					}
					if crt <= next || rand.Float64() < math.Exp((next-crt)/temp) {
						crt = next
					} else {
						state.ps = bk
					}
				} else {
					// ポリオミノを２つランダムに選び、互いの位置を交換
					// TODO
				}
			}
		}
	}

	log.Println("input", input)
	log.Println("time", time.Since(startTime))
}

// 焼きなましの差分計算用のデータ
type State struct {
	ts [][]uint
	// 各ポリオミノの配置
	ps []uint
	// 各マスの埋蔵量(失敗した占い結果と同じ領域火を判断するのに使用しており、失敗するまでは不要なので計算をのぞいている)
	count []uint8
	// これまでの各占いに対する合計埋蔵量
	query_count []uint8
	// [p][ij][q] := p番のポリオミノを(i, j)に配置した時の、q番の占いに対する合計埋蔵量の増加量
	pij_to_queru_count [][][]uint8
	hashes             [][]uint64
	hash               uint64
}

func NewState(input Input) (state State) {
	src := rand.NewSource(8904281)
	rng := rand.New(src)
	hashes := make([][]uint64, input.m)
	for i := 0; i < input.m; i++ {
		hashes[i] = make([]uint64, input.n2)
	}
	for p := 0; p < input.m; p++ {
		// 同じ図形は同じハッシュ値を使用することで、同じ形の図形の位置を入れ替えた状態を同一視する。
		// 同じ図形が同じ位置にある状態が正解の場合に、hashが衝突して解が見つからなくなるが、稀なので
		// そのようなレアケースは捨てて速度を優先
		if p > 0 && equlaUints(input.ts[p-1], input.ts[p]) {
			hashes[p] = hashes[p-1]
		} else {
			for ij := 0; ij < input.n2; ij++ {
				hashes[p][ij] = rng.Uint64()
			}
		}
	}
	var hash uint64
	for p := 0; p < input.m; p++ {
		hash ^= hashes[p][0]
	}
	sub_pij_to_queru_count := make([][][]uint8, input.m)
	for p := 0; p < input.m; p++ {
		sub_pij_to_queru_count[p] = make([][]uint8, input.n2)
	}
	state = State{
		ts:                 input.ts,
		ps:                 make([]uint, input.m),
		count:              make([]uint8, 0),
		query_count:        make([]uint8, 0),
		pij_to_queru_count: sub_pij_to_queru_count,
		hashes:             hashes,
		hash:               hash,
	}
	return state
}

// p番目のポリオミノを(i, j)に移動する
func (self *State) move_to(p uint, tij uint) {
	self.hash ^= self.hashes[p][self.ps[p]] ^ self.hashes[p][tij]
	for i := range self.pij_to_queru_count[p][self.ps[p]] {
		sub := self.pij_to_queru_count[p][self.ps[p]][i]
		add := self.pij_to_queru_count[p][tij][i]
		self.query_count[i] += add - sub
	}
	if len(self.count) > 0 {
		for _, ij := range self.ts[p] {
			self.count[ij+self.ps[p]] -= 1
			self.count[ij+tij] += 1
		}
	}
	self.ps[p] = tij
}

// 占ったマス集合qsを追加
func (self *State) add_query(input Input, qs []uint) {
	in_query := make([]bool, input.n2)
	for _, ij := range qs {
		in_query[ij] = true
	}
	for p := 0; p < input.m; p++ {
		for di := 0; di < input.n-input.maxI[p]; di++ {
			for dj := 0; dj < input.n-input.maxJ[p]; dj++ {
				dij := uint(di*input.n + dj)
				c := 0
				for _, ij := range input.ts[p] {
					if in_query[ij+dij] {
						c++
					}
				}
				self.pij_to_queru_count[p][dij] = append(self.pij_to_queru_count[p][dij], uint8(c))
			}
		}
	}
	count := input.get_count(qs)
	var c uint8
	for _, ij := range qs {
		c += count[ij]
	}
	self.query_count = append(self.query_count, c)
}

type Input struct {
	n, m  int
	eps   float64
	total int
	maxI  []int
	maxJ  []int
	ts    [][]uint
	n2    int
}

func (self *Input) get_count(ps []uint) []uint8 {
	count := make([]uint8, self.n2)
	for p := 0; p < len(ps); p++ {
		pij := ps[p]
		for _, ij := range self.ts[p] {
			count[ij+pij] += 1
		}
	}
	return count
}

func readInput() (input Input) {
	var n, m int
	var eps float64

	fmt.Scan(&n, &m, &eps)
	ts2 := make([][][2]int, m, m)
	var total int
	for i := 0; i < m; i++ {
		var d int
		fmt.Scan(&d)
		ts2[i] = make([][2]int, d, d)
		for j := 0; j < d; j++ {
			fmt.Scan(&ts2[i][j][0], &ts2[i][j][1])
		}
		total += d
	}
	// ポリオ味ののサイズでソート
	sort.Slice(ts2, func(i, j int) bool {
		return len(ts2[i]) < len(ts2[j])
	})
	// ポリオミノの大きさを保存する
	maxI := make([]int, m, m)
	maxJ := make([]int, m, m)
	for i := 0; i < m; i++ {
		maxI[i] = maxInt(maxI[i], ts2[i][0][0])
	}
	for i := 0; i < m; i++ {
		maxJ[i] = maxInt(maxJ[i], ts2[i][0][1])
	}
	ts := make([][]uint, m, m)
	for i := 0; i < m; i++ {
		ts[i] = make([]uint, len(ts2[i]), len(ts2[i]))
		for j := 0; j < len(ts2[i]); j++ {
			ts[i][j] = uint(ts2[i][j][0]*n + ts2[i][j][1])
		}
	}
	input = Input{n, m, eps, total, maxI, maxJ, ts, n * n}
	return input
}

// 配置候補の情報
type Entry struct {
	hash  uint64
	prob  float64
	ps    []uint
	count []uint8
}

const EPS float64 = 1e-6

// 各クエリ処理を行うためのデータ
type Sim struct {
	n     int
	n2    int
	m     int
	total int
	eps   float64
	query [][]int
	// (i, j, resp)
	mined [][2]int
	// probs[k][S][r] := kマスに対してクエリして真の値がSであるときに（lb＋r)が返ってくる確率、そのln)
	// ある程度外れると確率はほぼ０になるので、高速化のため、必要な範囲だけ保持している
	probs_lb [][]int
	probs    [][][][2]float64
	// ln_probs_query[q][S] := q番目のクエリに対して真の根がSであるときに実際の返答が返ってくる確率のln
	ln_probs_query [][]float64
	// 失敗した推測履歴
	failed [][]int
	// 残りクエリ数
	rem int
}

func NewSim(input Input) (sim Sim) {
	n := input.n
	n2 := input.n2
	m := input.m
	total := input.total
	eps := input.eps
	query := make([][]int, 0, 0)
	mined := make([][2]int, 0, 0)
	probs_lb := make([][]int, input.n2+1, input.n2+1)
	probs_ub := 0.0
	for i := 0; i <= input.n2; i++ {
		probs_lb[i] = make([]int, input.total+1, input.total+1)
	}
	probs := make([][][][2]float64, input.n2+1, input.n2+1)
	for i := 0; i <= input.n2; i++ {
		probs[i] = make([][][2]float64, input.total+1, input.total+1)
	}
	for k := 1; k < input.n2; k++ {
		for S := 0; S < input.total; S++ {
			mu := float64(k-S)*input.eps + float64(S)*(1.0-input.eps)
			sigma := math.Sqrt(float64(k) * input.eps * (1.0 - input.eps))
			for r := int(math.Round(mu)); r >= 0; r-- {
				var prob float64
				if r == 0 {
					prob = probability_in_range(mu, sigma, -100.0, float64(r)+0.5)
				} else {
					prob = probability_in_range(mu, sigma, float64(r)-0.5, float64(r)+0.5)
				}
				if prob < EPS {
					break
				}
				probs[k][S] = append(probs[k][S], [2]float64{prob, math.Log(prob)})
				probs_ub = math.Max(probs_ub, float64(r)+1)
			}
		}
	}
	log.Println("n2=", input.n2)
	log.Println("total=", input.total)
	log.Println("ub=", probs_ub)
	ln_probs_query := make([][]float64, 0, 0)
	failed := make([][]int, 0, 0)
	rem := total
	sim = Sim{n, n2, m, total, eps, query, mined, probs_lb, probs, ln_probs_query, failed, rem}
	return sim
}

func (self *Sim) Ans(T []int) bool {
	if self.rem == 0 {
		log.Println("!log giveup 1", self.n, self.m)
		os.Exit(1)
	}
	self.rem -= 1
	fmt.Printf("a %d", len(T))
	for _, t := range T {
		fmt.Printf(" %d %d ", t/self.n, t%self.n)
		fmt.Println()
	}
	var ret int
	fmt.Scan(&ret)
	if ret == 1 {
		return true
	}
	self.failed = append(self.failed, T)
	return false
}

func (self *Sim) Query(ps []int) int {
	if self.rem == 0 {
		log.Println("!log giveup 2", self.n, self.m)
		os.Exit(1)
	}
	self.rem -= 1
	fmt.Printf("q %d", len(ps))
	for _, ij := range ps {
		fmt.Printf(" %d %d", ij/self.n, ij%self.n)
	}
	fmt.Println()
	var ret int
	fmt.Scan(&ret)
	self.query = append(self.query, ps)
	probs := make([]float64, self.total+1, self.total+1)
	// すべてのSについて、queryの結果を反映させる
	for S := 0; S <= self.total; S++ {
		k := float64(len(ps))
		mu := (k-float64(S))*self.eps + float64(S)*(1.0-self.eps)
		sigma := math.Sqrt(k * self.eps * (1.0 - self.eps))
		var prob float64
		if ret == 0 {
			prob = probability_in_range(mu, sigma, -100.0, float64(ret)+0.5)
		} else {
			prob = probability_in_range(mu, sigma, float64(ret)-0.5, float64(ret)+0.5)
		}
		probs[S] = math.Log(prob)
	}
	for i := 0; i < len(probs)-1; i++ {
		if !math.IsInf(probs[i], 0) && math.IsInf(probs[i+1], 0) {
			probs[i+1] = probs[i] - 10.0
		}
	}
	for i := len(probs) - 1; i > 0; i-- {
		if !math.IsInf(probs[i], 0) && math.IsInf(probs[i-1], 0) {
			probs[i-1] = probs[i] - 10.0
		}
	}
	self.ln_probs_query = append(self.ln_probs_query, probs)
	return ret
}

func (self *Sim) Mine(i, j int) int {
	if self.rem == 0 {
		log.Println("!log giveup 3", self.n, self.m)
		os.Exit(1)
	}
	self.rem -= 1
	fmt.Printf("q 1 %d %d\n", i, j)
	var ret int
	fmt.Scan(&ret)
	self.mined = append(self.mined, [2]int{i*self.n + j, ret})
	return ret
}

// 埋蔵量がvsで各ポリオミノのいちがpsのときの対数尤度を計算
func (self *Sim) ln_prob(state State, vs []uint8, ps []uint) float64 {
	for _, qs := range self.failed {
		for _, ij := range qs {
			c_flag := false
			if vs[ij] == 0 {
				c_flag = true
			}
			if c_flag {
				break
			}
		}
		size := 0
		for ij := 0; ij < self.n2; ij++ {
			if vs[ij] > 0 {
				size += 1
			}
		}
		if size == len(qs) {
			return -1e20
		}
	}
	prob := 0.0
	for q := 0; q < len(self.query); q++ {
		count := 0
		for p, ij := range ps {
			count += int(state.pij_to_queru_count[p][ij][q])
		}
		prob += self.ln_probs_query[q][count]
	}
	return prob
}

// 状態stateのときの対数尤度を計算
func (self *Sim) ln_prob_state(state *State) float64 {
	for _, qs := range self.failed {
		continue_flag := false
		for _, ij := range qs {
			if state.count[ij] == 0 {
				continue_flag = true
				break
			}
		}
		if continue_flag {
			continue
		}
		size := 0
		for ij := 0; ij < self.n2; ij++ {
			if state.count[ij] > 0 {
				size += 1
			}
		}
		if size == len(qs) {
			return -1e20
		}
	}
	sum := 0.0
	for i := range self.ln_probs_query {
		sum += self.ln_probs_query[i][state.query_count[i]]
	}
	return sum
}

// 実行切れしたときはBFSで掘る。滅多に実行されない上に相対スコアへの影響が小さいので手抜き
func (self *Sim) giveup() {
	log.Println("!log giveup 4", self.n, self.m)
	que := list.New()
	que.PushBack([2]int{self.n / 2, self.n / 2})
	list := make([]int, 0, 0)
	rem := self.total
	used := make([][]bool, self.n)
	for i := 0; i < self.n; i++ {
		used[i] = make([]bool, self.n)
	}
	for que.Len() > 0 {
		ij := que.Front().Value.([2]int)
		i := ij[0]
		j := ij[1]
		if !used[i][j] {
			used[i][j] = true
			continue
		}
		ret := self.Mine(i, j)
		if ret > 0 {
			list = append(list, i*self.n+j)
			rem -= ret
			if rem == 0 {
				break
			}
		}
		for _, dij := range DIJ {
			i2 := i + dij[0]
			j2 := j + dij[1]
			if i2 < self.n && j2 < self.n {
				if ret == 0 {
					que.PushBack([2]int{i2, j2})
				} else {
					que.PushFront([2]int{i2, j2})
				}
			}
		}
		self.Ans(list)
	}
}

// util
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func normal_cdf(x, mean, std_dev float64) float64 {
	return 0.5 * (1.0 + math.Erf((x-mean)/(std_dev*math.Sqrt2)))
}

// 正規分布の確率密度関数 aからbまでの確率を求める
func probability_in_range(mean, std_dev, a, b float64) float64 {
	if mean < a {
		return probability_in_range(mean, std_dev, 2.0*mean-b, 2.0*mean-a)
	}
	p_a := normal_cdf(a, mean, std_dev)
	p_b := normal_cdf(b, mean, std_dev)
	return p_b - p_a
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

var DIJ = [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

func equlaUints(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
