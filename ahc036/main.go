package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"slices" // "golang.org/x/exp/slices"
	"strings"
)

func main() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		log.SetOutput(io.Discard)
	}
	var output strings.Builder
	in := readInput()
	dist, pred := allPairsShortest(in)
	A := make([]int, in.La)
	B := make([]int, in.Lb)
	for i := 0; i < V; i++ {
		A[i] = i
	}
	for i := range B {
		B[i] = -1
	}
	output.WriteString(fmt.Sprintln(strings.Trim(fmt.Sprint(A), "[]")))
	cntMove := 0
	cntSingleOpe := 0
	for i := 0; i < V-1+1; i++ { // in.planは０を先頭に追加してサイズが601
		u, v := in.plan[i], in.plan[i+1]
		log.Println(in.plan[i], "->", in.plan[i+1], "cost=", dist[u][v])
		root := constructShortestPath(u, v, pred, dist)
		log.Println(root)
		for j := 1; j < len(root); j++ {
			if slices.Contains(B, root[j]) {
				output.WriteString(fmt.Sprintln("m", root[j]))
			} else {
				index := slices.Index(A, root[j])
				//size := len(B)
				size := 1
				output.WriteString(fmt.Sprintln("s", size, index, 0))
				output.WriteString(fmt.Sprintln("m", root[j]))
				singleOpe(1, index, 0, A, B)
				cntSingleOpe++
			}
			cntMove++
		}
	}
	fmt.Print(output.String())
	log.Println(cntMove, cntSingleOpe)
	log.Println("2")
}

// A配列のPaからl個をB配列のPbに代入する
func singleOpe(length, Pa, Pb int, A, B []int) {
	if len(A)-Pa-length < 0 {
		log.Fatal(len(A), Pa, length)
	}
	if len(B)-Pb-length < 0 {
		log.Fatal(len(B), Pb, length)
	}
	for i := 0; i < length; i++ {
		B[Pb+i] = A[Pa+i]
	}
}

const (
	_N          int = 600
	_T          int = 600
	MaxRoadSize int = 600 * 3 // MaxRoadSize
	V           int = 600
)

type Input struct {
	N     int             //都市の数 N=600
	M     int             //道路の本数 N-1<=M<=3*N-6
	T     int             //訪問する都市の数 T=600
	La    int             // 配列Aの長さ N<=La<=2*N
	Lb    int             // 配列Bの長さ 4<=Lb<=24
	roads [600 * 3][2]int // u,vの都市間をつなぐ道路
	plan  [601]int        // 初期位置の0を先頭に追加する
}

func readInput() (in Input) {
	fmt.Scan(&in.N, &in.M, &in.T, &in.La, &in.Lb)
	for i := 0; i < in.M; i++ {
		fmt.Scan(&in.roads[i][0], &in.roads[i][1])
	}
	in.plan[0] = 0
	for i := 0; i < _T; i++ {
		fmt.Scan(&in.plan[i+1])
	}
	return in
}

// Floyd-Warshall Algorithm
// in.roadsから各都市間の最短経路をもとめる
func allPairsShortest(in Input) ([V][V]int, [V][V]int) {
	inf := math.MaxInt / 4
	var dist [V][V]int
	var pred [V][V]int
	for i := 0; i < V; i++ {
		for j := 0; j < V; j++ {
			dist[i][j] = inf
			pred[i][j] = -1
		}
		dist[i][i] = 0
	}
	for i := 0; i < MaxRoadSize; i++ {
		u := in.roads[i][0]
		v := in.roads[i][1]
		dist[u][v] = 1
		pred[u][v] = u
		dist[v][u] = 1
		pred[v][u] = v

	}
	for k := 0; k < V; k++ {
		for u := 0; u < V; u++ {
			for v := 0; v < V; v++ {
				newLength := dist[u][k] + dist[k][v]
				if newLength < dist[u][v] {
					dist[u][v] = newLength
					pred[u][v] = pred[k][v]
				}
			}
		}
	}
	return dist, pred
}

func constructShortestPath(s, t int, pred [V][V]int, dist [V][V]int) []int {
	reversePath := make([]int, 0, dist[s][t]+2)
	current := t

	// 逆順に経路を辿る
	for current != s {
		if t == -1 {
			panic("No Path found")
		}
		reversePath = append(reversePath, current)
		current = pred[s][current]
	}
	reversePath = append(reversePath, s)
	// 辿る順に並び替える
	for i, j := 0, len(reversePath)-1; i < j; i, j = i+1, j-1 {
		reversePath[i], reversePath[j] = reversePath[j], reversePath[i]
	}
	return reversePath
}
