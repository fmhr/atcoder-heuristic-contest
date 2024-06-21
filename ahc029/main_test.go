package main

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func TestGenerateCard(t *testing.T) {
	for i := 0; i < 100; i++ {
		//L := rand.Intn(20)
		L := 0
		M := rand.Intn(8-2) + 2
		x0 := rand.Intn(20) + 1
		x1 := rand.Intn(10) + 1
		x2 := rand.Intn(10) + 1
		x3 := rand.Intn(5) + 1
		x4 := rand.Intn(3) + 1
		card := generateCard(L, M, x0, x1, x2, x3, x4)
		//fmt.Println(L, M, x0, x1, x2, x3, x4)
		fmt.Println(card)
	}
}

// sample invest card
// Investカードのコストを計算する
func TestInvestCard(t *testing.T) {
	sumPrice := 0
	var N int = 1000
	for i := 0; i < N; i++ {
		//L := rand.Intn(20)
		L := 0
		M := rand.Intn(8-2) + 2
		card := generateCard(L, M, 0, 0, 0, 0, 1)
		//fmt.Println(L, M, x0, x1, x2, x3, x4)
		fmt.Printf("%+v \n", card)
		sumPrice += int(card.Cost)
	}
	fmt.Println("sumPrice:", sumPrice, "average:", float64(sumPrice)/float64(N))
}

// sample WorkSingle card
func TestWorkSingleCard(t *testing.T) {
	sumPrice := 0
	var N int = 1000
	p := make([]float64, N)
	for i := 0; i < N; i++ {
		L := 0
		M := rand.Intn(8-2) + 2
		card := generateCard(L, M, 1, 0, 0, 0, 0)
		//fmt.Println(L, M, x0, x1, x2, x3, x4)
		fmt.Printf("%v  %f\n", card, float64(card.Workforce)/float64(card.Cost))
		sumPrice += int(card.Cost)
		p[i] = float64(card.Workforce) / float64(card.Cost)
	}
	fmt.Println("sumPrice:", sumPrice, "average:", float64(sumPrice)/float64(N))
	fmt.Println("average:", average(p), "median:", median(p), "max:", max(p), "min:", min(p), "stddev:", stddev(p))
}

// sample WorkAll card
func TestWorkAllCard(t *testing.T) {
	sumPrice := 0
	for i := 0; i < 100; i++ {
		L := 0
		M := rand.Intn(8-2) + 2
		card := generateCard(L, M, 0, 1, 0, 0, 0)
		fmt.Printf("%+v \n", card)
		sumPrice += int(card.Cost)
	}
	fmt.Println("sumPrice:", sumPrice, "average:", float64(sumPrice)/100)
}

// sample CancelSingle card
func TestCancelSingle(t *testing.T) {
	var N int = 1000
	prices := make([]float64, N)
	for i := 0; i < N; i++ {
		L := 0
		M := rand.Intn(8-2) + 2
		card := generateCard(L, M, 0, 0, 1, 0, 0)
		fmt.Printf("%+v \n", card)
		prices[i] = float64(card.Cost)
	}
	fmt.Println("average:", average(prices), "median:", median(prices), "max:", max(prices), "min:", min(prices), "stddev:", stddev(prices))
}

func TestGenerateProject(t *testing.T) {
	for i := 0; i < 10; i++ {
		L := 0
		p := generateProject(L)
		fmt.Println(p)
	}
}

func TestSampleProject(t *testing.T) {
	var N int = 1000
	pp := make([]float64, N)
	for i := 0; i < N; i++ {
		L := 0
		p := generateProject(L)
		//fmt.Println(p, float64(p.Value)/float64(p.Workload))
		fmt.Printf("workload:%3d, value:%3d, cospa:%f\n", p.Workload, p.Value, float64(p.Value)/float64(p.Workload))
		pp[i] = float64(p.Value) / float64(p.Workload)
	}
	fmt.Println("average:", average(pp), "median:", median(pp), "max:", max(pp), "min:", min(pp), "stddev:", stddev(pp))
}

func average(p []float64) float64 {
	sum := 0.0
	for _, v := range p {
		sum += v
	}
	return sum / float64(len(p))
}
func max(p []float64) float64 {
	max := 0.0
	for _, v := range p {
		if max < v {
			max = v
		}
	}
	return max
}

func min(p []float64) float64 {
	min := 100000.0
	for _, v := range p {
		if min > v {
			min = v
		}
	}
	return min
}
func stddev(p []float64) float64 {
	ave := average(p)
	sum := 0.0
	for _, v := range p {
		sum += (v - ave) * (v - ave)
	}
	return sum / float64(len(p))
}
func median(p []float64) float64 {
	sort.Float64s(p)
	return p[len(p)/2]
}
