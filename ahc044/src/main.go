package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var ATCODER bool     // AtCoder環境かどうか
var frand *rand.Rand // シード固定乱数生成器 rand.Seed()は無効 go1.24~

func init() {
	log.SetFlags(log.Lshortfile)
	frand = rand.New(rand.NewSource(0))
}

var startTime time.Time

func main() {
	if os.Getenv("ATCODER") == "1" {
		ATCODER = true
		log.Println("on AtCoder")
		log.SetOutput(io.Discard)
	}
	var memStats runtime.MemStats

	defer runtime.ReadMemStats(&memStats)
	defer log.Printf("HeapAlloc: %d bytes\n", memStats.HeapAlloc)
	startTime = time.Now()
	log.SetFlags(log.Lshortfile)
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	in := readInput(reader)
	solver(in, writer)
	log.Printf("time=%v\n", time.Since(startTime).Milliseconds())
}

const (
	N = 100
	L = 500000
)

func solver(in *Input, writer io.Writer) {
	var a, b [100]int

	var bestA, bestB [100]int
	for i := 0; i < N; i++ {
		a[i] = frand.Intn(N)
		b[i] = frand.Intn(N)
	}
	bestScore := Score(in.T, a, b)
	bestA = a
	bestB = b

	iter := 0
	T0 := 100.0         // 初期温度
	Tf := 1.0           // 終了温度
	timeLimit := 1990.0 // 実行時間制限(ms)
	downScale := 100
	elapsed := float64(time.Since(startTime).Milliseconds())
	for {
		iter++
		elapsed = float64(time.Since(startTime).Milliseconds())
		if elapsed > 1800 {
			downScale = 10
		}
		if elapsed > timeLimit {
			break
		}

		// 温度を時間経過で下げる
		T := T0 + (Tf-T0)*(elapsed/timeLimit)

		// 現在の解をコピー
		currentA := a
		currentB := b

		// 近傍の解を生成
		if frand.Intn(2) == 0 {
			a[frand.Intn(N)] = frand.Intn(N)
		} else {
			b[frand.Intn(N)] = frand.Intn(N)
		}

		// スコアを計算
		score := ScoreMini(in.T, a, b, downScale)

		// スコアの変化量を計算
		delta := float64(score - bestScore)

		// 確率的に遷移を決定
		if delta > 0.0 {
			// スコアが改善する場合は必ず遷移
			bestScore = score
			bestA = a
			bestB = b
			log.Println("update best score:", bestScore, "iter:", iter)
		} else {
			// スコアが悪化する場合は、確率 exp(delta/T) で遷移
			prob := math.Exp(delta / T)
			if frand.Float64() < prob {
				// 遷移する場合
				bestScore = score
				bestA = a
				bestB = b
				log.Println("move to worse score:", score, "iter:", iter, "prob:", prob)
			} else {
				// 遷移しない場合、元の解に戻す
				a = currentA
				b = currentB
			}
		}
	}

	log.Printf("iters=%d\n", iter)
	for i := 0; i < N; i++ {
		_, _ = fmt.Fprintln(writer, bestA[i], bestB[i])
	}
}

func Score(targets, a, b [100]int) int {
	actualCounts := make([]int, N)

	currentEmployee := 0
	actualCounts[currentEmployee]++

	// Simulate L-1 more weeks
	for week := 1; week < L; week++ {
		// Determine next employee based on current employee's cleaning count
		if actualCounts[currentEmployee]%2 == 1 { // Odd count
			currentEmployee = a[currentEmployee]
		} else { // Even count
			currentEmployee = b[currentEmployee]
		}

		// Update cleaning count for the new employee
		actualCounts[currentEmployee]++
	}

	// Calculate error: sum of absolute differences between actual and target counts
	totalError := 0
	for i := 0; i < N; i++ {
		error := int(math.Abs(float64(actualCounts[i] - targets[i])))
		totalError += error
	}

	// Final score is 10^6 - error (guaranteed to be non-negative)
	score := 1000000 - totalError
	if score < 0 {
		score = 0 // Just in case, ensure score is non-negative
	}

	return score
}

func ScoreMini(targets, a, b [100]int, downScale int) int {
	actualCounts := make([]int, N)

	currentEmployee := 0
	actualCounts[currentEmployee]++

	// Simulate L-1 more weeks
	for week := 1; week < L/downScale; week++ {
		// Determine next employee based on current employee's cleaning count
		if actualCounts[currentEmployee]%2 == 1 { // Odd count
			currentEmployee = a[currentEmployee]
		} else { // Even count
			currentEmployee = b[currentEmployee]
		}

		// Update cleaning count for the new employee
		actualCounts[currentEmployee]++
	}

	// Calculate error: sum of absolute differences between actual and target counts
	totalError := 0
	for i := 0; i < N; i++ {
		error := int(math.Abs(float64(actualCounts[i]*downScale - targets[i])))
		totalError += error
	}

	// Final score is 10^6 - error (guaranteed to be non-negative)
	score := 1000000 - totalError
	if score < 0 {
		score = 0 // Just in case, ensure score is non-negative
	}

	return score
}

type Input struct {
	_N int
	_L int
	T  [100]int
}

func readInput(reader *bufio.Reader) *Input {
	in := &Input{}
	_, _ = fmt.Fscan(reader, &in._N, &in._L)
	for i := 0; i < in._N; i++ {
		_, _ = fmt.Fscan(reader, &in.T[i])
	}
	return in
}
