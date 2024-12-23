package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

var (
	ATCODER = "0"
	LOCAL   = "0"
)

func init() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		ATCODER = "1"
	}
	if os.Getenv("LOCAL") == "1" {
		LOCAL = "1"
	}
}

// 矩形の精度を推定で高める
func updateInput(input Input) {
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
}

// 初期入力を生成方法からベイズ推定で補正
// 10000~50000を10刻みで、それを真の値としたときの観測値の事後確率を計算
// 40001パターンの事後分布から周辺化された平均と分散を返す
// (4001*入力数)の計算量
func estimate(ys []float64, sigma float64) ([]float64, []float64) {
	// Lの範囲を設定
	lMin := 10000
	lMax := 50000 // TODO: 100000
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
		// 0~100000の中に入る確率
		beta := (100000.0 - y) / sigma
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
func estimateV0(y int, std float64) (float64, float64) {
	// Lの範囲を設定
	lMin := 10000
	lMax := 100000
	lStep := 10
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
	log.Printf("mean: %.0f, std: %.0f\n", mean, math.Sqrt(variance))
	return mean, math.Sqrt(variance)
}

type Input struct {
	N, T  int
	sigma int
	wh    [][2]int // w, h
	wv    [][2]int // wh, sigma^2
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
	wv := make([][2]int, 2*N)
	for i := 0; i < 2*N; i++ {
		wv[i][0] = wh[i/2][i%2]
		wv[i][1] = sigma * sigma
	}
	return Input{N, T, sigma, wh, wv}
}

const TL = 2.9

func main() {
	getTime()
	input := readInput()
	updateInput(input)
	log.Println("get time: ", getTime())
	t := time.Now()
	for i := 0; i < input.N; i++ {
		for j := 0; j < 2; j++ {
			estimateV0(input.wh[i][j], float64(input.sigma))
		}
	}
	log.Println("estimateV0 all elaps time: ", time.Since(t))
	log.Println("get time: ", getTime())
	//estimateV0(input.wh[0][0], float64(input.sigma))
	elp := getTime()
	log.Println(elp)
	var w, h int
	for i := 0; i < input.T; i++ {
		fmt.Print("0\n")
		fmt.Scan(&w, &h)
	}
}

var STIME time.Time

func getTime() float64 {
	if STIME == (time.Time{}) {
		STIME = time.Now()
	}
	if LOCAL == "1" {
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
func normalPDF(x, mu, sigma float64) float64 {
	return (1 / (math.Sqrt(2 * math.Pi * sigma * sigma))) *
		math.Exp(-((x-mu)*(x-mu))/(2*sigma*sigma))
}

// overflow を防ぐための近似式
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
