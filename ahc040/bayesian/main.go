package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
)

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

func main() {
	log.SetFlags(log.Lshortfile)
	// 測定値 (観測サンプル)
	y1Samples := []float64{30187.14, 30301.36, 30476.78, 30358.49, 30121.85, 30320.11, 30515.69, 30248.62,
		30383.36, 30262.47, 30420.13, 30179.74, 30487.25, 30284.19, 30396.08, 30366.85,
		30449.31, 30265.94, 30450.55, 30359.47}
	y2Samples := []float64{69923.95, 70268.57, 70105.64, 70320.42, 70114.06, 70234.67, 70049.89, 70321.96,
		70063.88, 70212.84, 70328.73, 70256.94, 70123.35, 70044.01, 70284.82, 70352.49,
		70213.57, 70102.98, 70314.24, 70178.34}
	y3Samples := []float64{79908.65, 80035.97, 80122.95, 79954.76, 79832.68, 80101.43, 80025.88, 80047.69,
		79963.38, 80074.61, 79863.21, 80031.37, 79987.03, 80126.74, 80043.79, 80025.49,
		80056.18, 79950.12, 80112.64, 79973.45}

	// 測定器の標準偏差
	sigma := 5000.0

	// ギブスサンプリングの設定
	numIterations := 2000
	burnIn := 100
	a, b, c := 0.0, 0.0, 0.0
	aSamples, bSamples, cSamples := []float64{}, []float64{}, []float64{}

	rand.Seed(42) // 乱数シード

	// ギブスサンプリング
	for t := 0; t < numIterations; t++ {
		// 1. aの更新
		bMean := mean(y1Samples) - b
		cMean := mean(y3Samples) - b - c
		aMean := (bMean + cMean) / 2.0
		a = rand.NormFloat64()*sigma/math.Sqrt(2.0) + aMean

		// 2. bの更新
		aMean = mean(y1Samples) - a
		cMean = mean(y2Samples) - c
		bMean = (aMean + cMean) / 2.0
		b = rand.NormFloat64()*sigma/math.Sqrt(2.0) + bMean

		// 3. cの更新
		bMean = mean(y2Samples) - b
		aMean = mean(y3Samples) - a - b
		cMean = (bMean + aMean) / 2.0
		c = rand.NormFloat64()*sigma/math.Sqrt(2.0) + cMean

		// 結果を保存
		aSamples = append(aSamples, a)
		bSamples = append(bSamples, b)
		cSamples = append(cSamples, c)
		log.Println(t, a, b, c)

		// 100ステップごとに出力
		if t > 5 && (t+1)%1000 == 0 {
			aMeanCurrent := mean(aSamples[burnIn:])
			bMeanCurrent := mean(bSamples[burnIn:])
			cMeanCurrent := mean(cSamples[burnIn:])

			aStdCurrent := std(aSamples[burnIn:])
			bStdCurrent := std(bSamples[burnIn:])
			cStdCurrent := std(cSamples[burnIn:])

			fmt.Printf("ステップ %d:\n", t+1)
			fmt.Printf("  a の推定値: 平均 = %.2f, 標準偏差 = %.2f\n", aMeanCurrent, aStdCurrent)
			fmt.Printf("  b の推定値: 平均 = %.2f, 標準偏差 = %.2f\n", bMeanCurrent, bStdCurrent)
			fmt.Printf("  c の推定値: 平均 = %.2f, 標準偏差 = %.2f\n", cMeanCurrent, cStdCurrent)
			fmt.Println(strings.Repeat("-", 40))
		}
	}
}
