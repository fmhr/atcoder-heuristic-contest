package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"time"

	"gonum.org/v1/gonum/mat"
)

var (
	ATCODER = "0"
)

func init() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		ATCODER = "1"
	}
}

// 矩形の精度を推定で高める
func updateInput(input Input) []float64 {
	// whの平均と分散を計算
	means := make([]float64, input.N*2)
	vars := make([]float64, input.N*2)
	for i := 0; i < input.N; i++ {
		means[i*2] = float64(input.wh[i][0])
		means[i*2+1] = float64(input.wh[i][1])
	}
	means, vars = estimate(means, float64(input.sigma))
	_, _ = means, vars
	for i := 0; i < input.N*2; i++ {
		log.Printf("%.0f ±%.0f\n", means[i], math.Sqrt(vars[i]))
	}
	// 十時型に並べてクエリすることで和の計測を行う
	// 現在の事後分布における計測結果の分散が大きくなるようにクエリを行う
	NumWidth := 10
	rng := rand.New(rand.NewPCG(0, 0))
	stime := getTime()
	iter := 0
	for input.turn+NumWidth < input.T {
		best := make([][][]Cmd, 0)
		max := 0.0
		for len(best[0][0]) == 0 ||
			getTime()-stime < 0.5*float64(input.turn+1)/(float64(input.T)-float64(NumWidth)) {
			iter++
			cmds := make([]Cmd, 0)
			cmds = append(cmds, Cmd{p: 0, r: false, d: false, b: ^uint(0)})
			row := make([]int, 0)
			col := make([]int, 0)
			rb := ^1
			cb := ^1
			sim := NewSim()
			// この先は, Estimator->Simulatorの順で実装する必要がある
		}
	}
	return means
}

// 初期入力を生成方法からベイズ推定で補正
// 10000~50000を10刻みで、それを真の値としたときの観測値の事後確率を計算
// 40001パターンの事後分布から周辺化された平均と分散を返す
// (4001*入力数)の計算量
func estimate(ys []float64, sigma float64) ([]float64, []float64) {
	// Lの範囲を設定
	lMin := 10000
	lMax := 50000
	lStep := 10

	// 事前分布の対数
	// lMinからlMaxまでlStep刻みで一様分布とみなして対数を取る
	logPLPrior := -math.Log(float64((lMax-lMin)/lStep + 1))
	//log.Println("logPLPrior", logPLPrior)

	// 各観測値に対して、Lに依存しない値を事前に計算
	betas := make([]float64, len(ys))
	cdfBetas := make([]float64, len(ys))
	phiBetas := make([]float64, len(ys))
	//log.Println("i, y, beta, cdf, phi")
	for i, y := range ys {
		// 分布の上限 100000 を基準に、観測値 y との差を標準偏差 sigma でスケーリング
		// /sigama にすることで、スケール後の標準偏差が1になる
		// 0~100000の中に入る確率
		beta := (100000.0 - y) / sigma
		log.Println(y, beta)
		betas[i] = beta
		cdfBetas[i] = NormalCDF(beta)
		phiBetas[i] = NormalPDF(beta)
		//log.Printf("%d %.0f %f %f %f\n", i, y, beta, cdfBetas[i], phiBetas[i])
	}
	// Lの事後対数確率
	LlogPosterior := make([][2]float64, 0, (lMax-lMin)/lStep+1)
	for l := lMin; l <= lMax; l += lStep {
		// x の事前分布の対数
		nl := float64(100000.0 - l + 1)
		logPxPrior := -math.Log(nl)

		// 各観測値に対する対数尤度の計算
		logLikelihoodSum := 0.0 // 対数尤度の和
		valid := true
		for i, yi := range ys {
			cdfBeta := cdfBetas[i]
			// l~100000の中に入る確率  標準化
			alpha := (float64(l) - yi) / sigma
			cdfAlpha := NormalCDF(alpha)  // 値がl以下になる確率
			cdfDiff := cdfBeta - cdfAlpha // 値がl~100000の中に入る確率

			if cdfDiff <= 0.0 {
				valid = false
				break
			}

			// 対数尤度の計算
			logLikeLihoodI := logPxPrior + math.Log(cdfDiff)
			logLikelihoodSum += logLikeLihoodI
		}
		if !valid {
			continue
		}

		// 事後分布の対数を計算 (事前分布の対数 + 対数尤度の和)
		logPosterior := logPLPrior + logLikelihoodSum
		// Lと対数事後確率を保存
		LlogPosterior = append(LlogPosterior, [2]float64{float64(l), logPosterior})
	}

	// 分母の対数和を計算
	logPs := make([]float64, len(LlogPosterior))
	for i, lp := range LlogPosterior {
		logPs[i] = lp[1]
	}
	logDenominator := LogSumExp(logPs)
	//log.Println("logDenominator", logDenominator)

	// 各x_i の事後確率と分散
	xmeans := make([]float64, 0, len(ys))
	xvars := make([]float64, 0, len(ys))

	// 各x_i の条件付き平均と分散
	numLen := len(LlogPosterior)
	xCondMeans := make([][]float64, len(ys))
	xCondVars := make([][]float64, len(ys))
	for i := 0; i < len(ys); i++ {
		xCondMeans[i] = make([]float64, 0, numLen)
		xCondVars[i] = make([]float64, 0, numLen)
	}

	// 各 L に対して事後確率を計算し、各　x_i の条件付き平均と分散を計算
	lPosterior := make([]float64, numLen)
	for i, lp := range LlogPosterior {
		lValue, LogPosterior := lp[0], lp[1]
		// 事後確率を計算
		logPL := LogPosterior - logDenominator
		pl := math.Exp(logPL) // 対数を外す
		lPosterior[i] = pl

		// 各 x_i の条件付き平均と分散を計算
		for i := range ys {
			beta := betas[i]
			phiBeta := phiBetas[i]
			cdfBeta := cdfBetas[i]

			alpha := (lValue - ys[i]) / sigma
			cdfAlpha := NormalCDF(alpha)
			phiAlpha := NormalPDF(alpha)

			cdfDiff := cdfBeta - cdfAlpha
			if cdfDiff <= 0.0 {
				log.Println("cdfDiff <= 0.0")
				continue
			}

			phiDiff := phiAlpha - phiBeta

			// 条件付き平均と分散
			mean := ys[i] + sigma*(phiDiff/(cdfBeta-cdfAlpha))
			xCondMeans[i] = append(xCondMeans[i], mean)

			variance := sigma * sigma * (1.0 + (alpha*phiAlpha-beta*phiBeta)/cdfDiff - (phiDiff/cdfDiff)*(phiDiff/cdfDiff))
			xCondVars[i] = append(xCondVars[i], variance)
		}
	}

	// 各 x_i の周辺化された平均を計算
	for i := range ys {
		meanSum := 0.0
		for j, l := range lPosterior {
			meanSum += l * xCondMeans[i][j]
		}
		xmeans = append(xmeans, meanSum)
	}

	// 各 x_i の周辺化された分散を計算
	for i := range ys {
		mean := xmeans[i]
		varSum := 0.0
		for j, l := range lPosterior {
			meanDiff := xCondMeans[i][j] - mean
			varSum += l * (xCondVars[i][j] + meanDiff*meanDiff)
		}
		xvars = append(xvars, varSum)
	}
	return xmeans, xvars
}

// v0 １つの観測値に対する事後確率を計算
// return 平均, 標準偏差
func estimateV0(yMin int, y int, std float64) (float64, float64) {
	// Lの範囲を設定
	lMin := max(yMin-int(std*2), 10000)
	lMax := 100000
	lStep := 100
	// 事前分布を(lMax-lMin)/lStep+1個作る
	// 事前分布の平均のリスト(10000~100000)
	llist := make([]float64, 0, (lMax-lMin)/lStep+1)
	for l := lMin; l <= lMax; l += lStep {
		llist = append(llist, float64(l))
	}

	// それぞれの尤度の計算 P(y|L)
	linkhoods := make([]float64, len(llist))
	for i, l := range llist {
		linkhoods[i] = normalPDF(float64(y), l, std)
		//log.Println(y, l, linkhoods[i])
	}
	// 事前分布 (一様分布) P(L)
	prior := 1.0 / float64(len(llist))

	// 事後分布の計算 P(L|y)
	posterior := make([]float64, len(llist))
	sumPosterior := 0.0
	for i := 0; i < len(llist); i++ {
		posterior[i] = linkhoods[i] * prior // 尤度 * 事前分布
		sumPosterior += posterior[i]
	}
	for i := 0; i < len(llist); i++ {
		// 正規化 (事後確率の和が1になるように総和で割る)
		posterior[i] /= sumPosterior
	}
	// 平均の計算(事後分布)
	mean := 0.0
	for i := 0; i < len(llist); i++ {
		//log.Println(llist[i], posterior[i], llist[i]*posterior[i])
		mean += llist[i] * posterior[i] // L * P(L|y)
	}
	// 分散の計算(事後分布)
	variance := 0.0
	for i := 0; i < len(llist); i++ {
		variance += (llist[i] - mean) * (llist[i] - mean) * posterior[i] // (L - mean)^2 * P(L|y)
	}
	//log.Printf("mean: %.0f, std: %.0f\n", mean, math.Sqrt(variance))
	return mean, math.Sqrt(variance)
}

// 対数をつけて計算することで、精度を高める
func estimateV0WithLog(y int, std float64) (float64, float64) {
	lMin := 10000
	lMax := 100000
	lStep := 10
	llist := make([]float64, 0, (lMax-lMin)/lStep+1)
	for l := lMin; l <= lMax; l += lStep {
		llist = append(llist, float64(l))
	}
	// 対数尤度 P(y|L)= -log(σ√{2*π}) - {(y - L)^2}/{2*σ^2}
	logLikelihoods := make([]float64, len(llist))
	for i, l := range llist {
		logLikelihoods[i] = normalPDFLog(float64(y), l, std)
	}
	// 対数事前分布 P(L) = -log(n)
	// 1/n は対数で -log(n) になる
	logPrior := -math.Log(float64(len(llist)))
	// 対数事後分布 P(L|y) ∝ P(y|L) + P(L)
	logPosteriors := make([]float64, len(llist))
	for i := 0; i < len(llist); i++ {
		logPosteriors[i] = logLikelihoods[i] + logPrior
	}
	// 正規化定数
	LogSumPosterior := LogSumExp(logPosteriors)
	// 正規化された事後分布 対数をとる
	posteriors := make([]float64, len(llist))
	for i := 0; i < len(llist); i++ {
		posteriors[i] = math.Exp(logPosteriors[i] - LogSumPosterior)
	}
	// 平均
	mean := 0.0
	for i := 0; i < len(llist); i++ {
		mean += llist[i] * posteriors[i]
	}
	// 分散
	variance := 0.0
	for i := 0; i < len(llist); i++ {
		variance += (llist[i] - mean) * (llist[i] - mean) * posteriors[i]
	}
	log.Printf("mean: %.0f, std: %.0f\n", mean, math.Sqrt(variance))
	return mean, math.Sqrt(variance)
}

type Cmd struct {
	p uint // piece index
	r bool // rotation
	d bool // direction
	b uint // base piece index
}

type Input struct {
	turn      int
	N, T      int
	sigma     int
	wh        [][2]int // w, h
	whv       [][2]Val
	estimator *Estimator
}

func readInput() Input {
	var N, T int
	var sigma int
	fmt.Scan(&N, &T, &sigma)
	log.Printf("N: %d, T: %d, sigma: %d\n", N, T, sigma)
	wh := make([][2]int, N)
	for i := 0; i < N; i++ {
		fmt.Scan(&wh[i][0], &wh[i][1])
	}
	return NewInput(N, T, sigma, wh)
}

func NewInput(N, T, sigma int, wh [][2]int) Input {
	whv := make([][2]Val, N)
	means := make([]float64, 2*N)
	variances := make([]float64, 2*N)
	sigma2 := float64(sigma * sigma)

	for i, pair := range wh {
		w, h := float64(pair[0]), float64(pair[1])
		whv[i] = [2]Val{
			{mean: w, sigma: sigma2}, {mean: h, sigma: sigma2}} // ?
		means[i*2] = w
		variances[i*2] = sigma2
		means[i*2+1] = h
		variances[i*2+1] = sigma2
	}
	return Input{
		N: N, T: T, sigma: sigma, wh: wh, whv: whv,
		estimator: NewEstimator(means, variances, float64(sigma)),
	}
}

const TL = 2.9

func main() {
	getTime()
	input := readInput()
	means := updateInput(input)
	_ = means
	log.Println("time", getTime())

	var estMean []float64
	var estStd []float64
	yMin := 1000000
	yMax := 0
	for i := 0; i < input.N*2; i++ {
		yMin = min(yMin, input.wh[i/2][i%2])
		yMax = max(yMax, input.wh[i/2][i%2])
	}
	for i := 0; i < input.N; i++ {
		for j := 0; j < 2; j++ {
			m, sd := estimateV0WithLog(input.wh[i][j], float64(input.sigma))
			estMean = append(estMean, m)
			estStd = append(estStd, sd)
		}
	}
	log.Println("time", getTime())
	var w, h int
	for i := 0; i < input.T; i++ {
		fmt.Print("0\n")
		fmt.Scan(&w, &h)
	}
	if ATCODER != "1" {
		trueWH := make([]int, input.N*2)
		for i := 0; i < input.N*2; i++ {
			fmt.Scan(&trueWH[i])
		}
		checkEstimate(input, estMean, trueWH)
		// wataさんのを再現
		log.Println("wataさんのを再現")
		checkEstimate(input, means, trueWH)
	}
}

func checkEstimate(in Input, estMean []float64, trueWH []int) {
	sum1 := 0.0
	sum2 := 0.0
	for i := 0; i < in.N*2; i++ {
		//s := "good"
		//if math.Abs(estMean[i]-float64(trueWH[i])) > math.Abs(float64(in.wh[i/2][i%2])-float64(trueWH[i])) {
		//s = "bad"
		//}
		//log.Printf("true: %d, input:%6d(%6d), est: %.0f(%5d) %s\n", trueWH[i], in.wh[i/2][i%2], in.wh[i/2][i%2]-trueWH[i], estMean[i], int(estMean[i])-trueWH[i], s)
		sum1 += (estMean[i] - float64(trueWH[i])) * (estMean[i] - float64(trueWH[i]))
		sum2 += (float64(in.wh[i/2][i%2]) - float64(trueWH[i])) * (float64(in.wh[i/2][i%2]) - float64(trueWH[i]))
	}
	log.Println("RMSE est  ", math.Sqrt(sum1/float64(in.N*2)))
	log.Println("RMSE input", math.Sqrt(sum2/float64(in.N*2)))
}

var STIME time.Time

func getTime() float64 {
	if STIME == (time.Time{}) {
		STIME = time.Now()
	}
	if ATCODER != "1" {
		// うちのPCは、ジャッジサーバより２倍ぐらい早い（願望）
		return time.Since(STIME).Seconds() * 2.0
	}
	return time.Since(STIME).Seconds()
}

// NormalPDF は標準正規分布の確率密度関数（PDF）を計算します
// 数式: f(x) = (1 / √(2π)) * e^(-(x^2)/2)
// x の周辺に観測値が出現する「密度」。数値そのものは確率ではない。
// x は平均値付近で大きな値を取り、それ以外の場所では小さな値を取る。
// xには正規化された値を入れるひつようがある
func NormalPDF(x float64) float64 {
	return math.Exp(-x*x/2.0) / math.Sqrt(2.0*math.Pi)
}

// NormalCDF は標準正規分布の累積分布関数（CDF）を計算します
// 数式: Φ(x) = (1/2) * [1 + erf(x/√2)]
// erf(x) は誤差関数です
// 値が x 以下になる確率（累積したもの）。0～1の間の値を取る。
func NormalCDF(x float64) float64 {
	return 0.5 * (1.0 + math.Erf(x/math.Sqrt(2.0)))
}

// 一番原始的？な正規分布の確率密度関数
//
//	(1 / (σ√{2π})) * e^(-(x - μ)^2 / 2σ^2)
func normalPDF(x, mu, sigma float64) float64 {
	//log.Println("normalPDF", x, mu, sigma)
	//log.Println("1 / (σ√{2π})", 1/(sigma*math.Sqrt(2*math.Pi)))
	//log.Printf("%f, %f, %f\n", (x-mu)*(x-mu), 2*sigma*sigma, -((x-mu)*(x-mu))/(2*sigma*sigma))
	return (1 / (math.Sqrt(2 * math.Pi * sigma * sigma))) *
		math.Exp(-((x-mu)*(x-mu))/(2*sigma*sigma))
}

// logP(y|L) = -log(sigma√{2*π}) - {(y - L)^2}/{2*σ^2}
func normalPDFLog(x, mu, sigma float64) float64 {
	return -math.Log(sigma*math.Sqrt(2*math.Pi)) - ((x-mu)*(x-mu))/(2*sigma*sigma)
}

// 対数を使う
func normalPDF2(x, mu, sigma float64) float64 {
	logPDF := -math.Log(sigma*math.Sqrt(2*math.Pi)) - ((x-mu)*(x-mu))/(2*sigma*sigma)
	return math.Exp(logPDF)
}

// LogSumExp は対数の和を計算します
// overflow を防ぐ
// logP(y) = log(Σ exp(logP(y|L) + logP(L))) 一般式
// logP(y) = M + log(Σ exp(logP(y|L) + logP(L) - M)) に変形 M = max(logP(L|y))
func LogSumExp(logs []float64) float64 {
	if len(logs) == 0 {
		return math.Inf(-1)
	}
	max := logs[0]
	for _, logv := range logs {
		if logv > max {
			max = logv
		}
	}
	var sum float64
	for _, logv := range logs {
		sum += math.Exp(logv - max)
	}
	return max + math.Log(sum)
}

// T型に配置して、観測値からそれぞれの平均と分散を推定する
type Val struct {
	mean  float64
	sigma float64
	a     []int // 一緒に観測れたもののindex
}

func NewVal() *Val {
	return &Val{
		mean:  0.0,
		sigma: 0.0,
		a:     make([]int, 0),
	}
}

func (v Val) clone() *Val {
	return &Val{
		mean:  v.mean,
		sigma: v.sigma,
		a:     append([]int{}, v.a...),
	}
}

func (v *Val) Add(i int, estimator *Estimator) {
	mean := estimator.mean.At(i, 0)
	variance := estimator.convariance.At(i, i)
	v.mean += mean
	// new Variance
	newVar := v.sigma*v.sigma + variance
	// add covariance terms
	// 一緒に観測されたもの(j)とiの共分散を足す
	for _, j := range v.a {
		newVar += 2 * estimator.convariance.At(i, j) // なぜ２倍？
	}
	v.sigma = math.Sqrt(newVar)
	v.a = append(v.a, i)
}

// UpperBound は上限を計算します
func (v *Val) ub(d float64) float64 {
	return v.mean + v.sigma*d
}

// LowerBound は下限を計算します
func (v *Val) lb(d float64) float64 {
	return v.mean - v.sigma*d
}

// Estimator implements Kalman Filter
type Estimator struct {
	mean        *mat.VecDense // 1次元 平均
	convariance *mat.Dense    // 2次元 共分散行列
	sigma2      float64
	n           int // サイズ
}

func NewEstimator(means []float64, vars []float64, sigma float64) *Estimator {
	n := len(means)
	sigma2 := sigma * sigma
	// Create mean vector
	mean := mat.NewVecDense(n, means)
	// Create diagonal covariance matris
	covariance := NewDiagonalDense(n, vars) // 対角行列(n*n)

	return &Estimator{
		mean:        mean,
		convariance: covariance,
		sigma2:      sigma2,
		n:           n,
	}
}

func (e *Estimator) Update(S []int, t float64) {
	// Create observation vector h 観察行列
	h := mat.NewVecDense(e.n, nil)
	for _, idx := range S {
		h.SetVec(idx, 1.0)
	}
	// Calculate S_i = h^T * Σ * h + σ^2
	var siMatrix mat.Dense
	siMatrix.Mul(h.T(), e.convariance)               // h^T * Σ 観察行列の逆と共分散行列 (Mul:行列の積)
	si := mat.Dot(siMatrix.RowView(0), h) + e.sigma2 // (h^T * Σ) * h + σ^2 (Dot:ベクトルの内積)
	// Calculate K = Σ * h / S_i カルマンゲイン
	k := mat.NewVecDense(e.n, nil)
	k.MulVec(e.convariance, h) // Σ * h 共分散行列と観察行列の積
	k.ScaleVec(1.0/si, k)      // Σ * h / S_i カルマンゲイン
	// Calculate residual r = t - h^T * μ 観測値と観察行列の積
	residual := t - mat.Dot(h, e.mean)
	// Update mean μ_new = μ_old + K * r 平均の更新
	e.mean.AddScaledVec(e.mean, residual, k)
	// Update Covariance Σ_new = Σ_old - K * h^T * ∑_old 共分散行列の更新
	var khT, update mat.Dense
	khT.Mul(k, h.T())
	update.Mul(&khT, e.convariance)
	e.convariance.Sub(e.convariance, &update)
}

// NewDiagonalDense は対角行列をmat.Denseで生成します
// mat.NewDiagDense は体格行列をvectorで生成するので使えない
func NewDiagonalDense(n int, v []float64) *mat.Dense {
	dense := mat.NewDense(n, n, nil)
	for i := 0; i < n; i++ {
		dense.Set(i, i, v[i])
	}
	return dense
}

// Simulater
type Sim struct {
	pos  [][4]*Val // (x1, x2, y1, y2)
	done []bool
	W    *Val
	H    *Val
}

// 初期化
func NewSim(n int) *Sim {
	pos := make([][4]*Val, n)
	for i := range pos {
		pos[i] = [4]*Val{NewVal(), NewVal(), NewVal(), NewVal()}
	}
	return &Sim{
		pos:  pos,
		done: make([]bool, n),
		W:    NewVal(),
		H:    NewVal(),
	}
}

// Put attemps to place a rectangle according to the command
func (s *Sim) Put(input *Input, c Cmd, strict bool) bool {
	d := 1.5 // 1.5倍の幅で判定
	w := int(c.p * 2)
	h := int(c.p*2 + 1)

	// 回転
	if c.r {
		w, h = h, w
	}

	touch := false

	if !c.d { // place to the right
		var x1 *Val
		if c.b == ^uint(0) {
			x1 = NewVal()
		} else {
			x1 = s.pos[c.b][1]
		}
		x2 := x1.clone()
		x2.Add(w, input.estimator)
		var y1 *Val
		// すでに設置済みのものとの衝突判定
		for q := range c.p {
			if !s.done[q] || q == c.b {
				continue
			}
			if s.pos[q][1].ub(d) > x1.lb(d) && x2.ub(d) > s.pos[q][0].lb(d) {
				if s.pos[q][3].mean > y1.mean {
					if strict {
						touch = s.pos[q][1].lb(d) < x1.ub(d) || x2.lb(d) < s.pos[q][0].ub(d)
					}
					y1 = s.pos[q][3].clone()
				}
			}
			if touch {
				return false
			}
			y2 := y1.clone()
			y2.Add(h, input.estimator)
			s.pos[c.p] = [4]*Val{x1, x2, y1, y2}
			// wip
		}
	}
}
