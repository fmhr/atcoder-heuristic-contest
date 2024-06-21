package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

var startTime time.Time

func main() {
	log.SetFlags(log.Lshortfile)
	rand.Seed(2)
	log.Println("V2")
	startTime = time.Now()
	readParameters()
	solver()
	timeDistance := time.Since(startTime)
	log.Println(timeDistance)
}

var dy = []int{0, 1, 0, -1, 0, 1, 1, -1, -1, 2, 0, -2, 0}
var dx = []int{0, 0, 1, 0, -1, 1, -1, -1, 1, 0, 2, 0, -2}

var L int // size of the grie
var N int // numver of exit cells
var S int // standard deviation
var exitCells [][2]int
var temperatures [100][100]int

func readParameters() {
	fmt.Scan(&L, &N, &S)
	log.Printf("L=%d N=%d S=%d\n", L, N, S)
	exitCells = make([][2]int, N)
	for i := 0; i < N; i++ {
		fmt.Scan(&exitCells[i][0], &exitCells[i][1])
	}
}

const DataCheckNum = 9

func solver() {
	placement()
	// wormholeData has the five mean of the temperature of each exit cell
	//	  3
	//  4 0 2
	//    1
	wormholeData := mesureAll()
	ans := predict(wormholeData)
	// answer
	fmt.Println("-1 -1 -1")
	for i := 0; i < N; i++ {
		fmt.Println(ans[i])
	}
	// after answer
	cost := CalcPlacementCost()
	log.Println("placementCost=", cost)
}

// ------------------ placement ------------------
// placement sets the air conditioning unit
// randomly
func placement() {
	// set
	for i := 0; i < L; i++ {
		for j := 0; j < L; j++ {
			temperatures[i][j] = rand.Intn(1000)
		}
	}
	// output
	for i := 0; i < L; i++ {
		for j := 0; j < L; j++ {
			fmt.Print(temperatures[i][j], " ")
		}
		fmt.Println()
	}
}

func CalcPlacementCost() (cost int) {
	for i := 0; i < L; i++ {
		for j := 0; j < L; j++ {
			a := temperatures[i][j] - temperatures[(i+1)%L][j]
			b := temperatures[i][j] - temperatures[i][(j+1)%L]
			cost += ((a * a) + (b * b))
		}
	}
	return cost
}

// ------------------ mesurment ------------------
// mesurment mesures the temperature
func mesurment(i, y, x int) (m mesurmentResult) {
	m.i = i
	m.y = y
	m.x = x
	fmt.Println(i, y, x)
	fmt.Scan(&m.mean)
	return
}

type mesurmentResult struct {
	i, y, x, mean int
}

func mesureN(i, y, x, n int) mesurmentResult {
	sumResult := 0
	for j := 0; j < n; j++ {
		r := mesurment(i, y, x)
		sumResult += r.mean
	}
	return mesurmentResult{i, y, x, sumResult / n}
}

// mesureI mesures the wormhole
func mesureI(i int) []int {
	// mesurment max 10000 times
	n := 10000 / DataCheckNum / N
	var fingurePrint []int
	for d := 0; d < DataCheckNum; d++ {
		//fingurePrint[d] = mesureN(i, dy[d], dx[d], n).mean
		fingurePrint = append(fingurePrint, mesureN(i, dy[d], dx[d], n).mean)
	}
	return fingurePrint
}

func mesureAll() [][]int {
	wormholeData := make([][]int, 0)
	for i := 0; i < N; i++ {
		//wormholeData[i] = mesureI(i)
		wormholeData = append(wormholeData, mesureI(i))
	}
	return wormholeData
}

// ------------------ predict ------------------

func predict(wormholeData [][]int) []int {
	return simulatedAnnealing(wormholeData)
}

func simulatedAnnealing(wormholeData [][]int) []int {
	state := make([]int, N)
	for i := 0; i < N; i++ {
		state[i] = i
	}
	// colunさんのコードを参考にしたログ
	nextThreshold := 5
	var up, same, down, force, iterations int
	//var startTemp float64
	//var endTemp float64
	var timeLimit time.Duration = 2900 * time.Millisecond
	startValue := calcEvaluationValue(state, wormholeData)
	var loopCount int

	log.Println("          up      same      down  force iterations")
	var startTemp float64 = 1000
	var endTemp float64 = 1
	var temp float64 = startTemp
	_ = temp
	for {
		timeDistance := time.Since(startTime)
		if timeDistance > timeLimit {
			break
		}
		//
		temp = startTemp + (endTemp-startTemp)*float64(timeDistance)/float64(timeLimit)
		// 2SWAP
		a := rand.Intn(N)
		b := rand.Intn(N)
		for a == b {
			b = rand.Intn(N)
		}
		state[a], state[b] = state[b], state[a]

		newValue := calcEvaluationValue(state, wormholeData)
		prob := math.Exp((float64(newValue) - float64(startValue)) / temp)
		//log.Println("prob", prob, newValue, startValue)
		if newValue > startValue {
			up++
		} else if newValue == startValue {
			same++
		} else {
			down++
		}
		if prob > rand.Float64() {
			if newValue < startValue {
				force++
			}
			startValue = newValue
		} else {
			state[a], state[b] = state[b], state[a]
		}
		iterations++
		loopCount++
		// ------------------ log ------------------
		for i := nextThreshold; i <= 100; i += 10 {
			if timeDistance > timeLimit*time.Duration(nextThreshold)/100 {
				log.Printf("SA(%2d%%): %6d  %6d  %6d  %6d %d\n", nextThreshold, up, same, down, force, iterations)
				up, same, down, force, iterations = 0, 0, 0, 0, 0
				nextThreshold += 10
			}
		}
		// ------------------ log ------------------
	}
	return state
}

// calcEvaluationValue calculates the evaluation value
// Sum of the difference between the measured temperature and the true temperature.
func calcEvaluationValue(s []int, data [][]int) (rtn int) {
	for i := 0; i < N; i++ {
		for k := 0; k < DataCheckNum; k++ {
			y := (exitCells[s[i]][0] + dy[k] + L) % L
			x := (exitCells[s[i]][1] + dx[k] + L) % L
			rtn += absInt(temperatures[y][x] - data[i][k])
		}
	}
	return -rtn
}

// TODO speeder calcEvaluationValue
// swapするところだけ再計算する

// ------------------ util ------------------
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
