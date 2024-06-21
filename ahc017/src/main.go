package main

import (
	"fmt"
	"log"
)

var N, M, D, K int
var u [3001]int
var v [3001]int
var w [3001]int

func edges(e int) (int, int) {
	return u[e], v[e]
}

func main() {
	log.SetFlags(log.Lshortfile)
	// N頂点数 M辺数 D日数 K１日に工事可能な辺数
	fmt.Scan(&N, &M, &D, &K)
	var construction [3001]bool
	for i := 1; i <= M; i++ {
		fmt.Scan(&u[i], &v[i], &w[i])
	}
	var x [1001]int
	var y [1001]int
	for i := 1; i <= N; i++ {
		fmt.Scan(&x[i], &y[i])
	}
	// 出力　辺を工事する日付
	day := 1
	k := 0
	for i := 1; i <= M; i++ {
		fmt.Printf("%d ", day)
		k++
		if k == K {
			k = 0
			day++
		}
	}
	fmt.Println("")
	log.Printf("頂点数=%d, 辺数=%d, 日数=%d, 工事可能数=%d\n", N, M, D, K)
	dist0 := floydWarshall(u, v, w, construction)
	log.Println(dist0[1])
}

// アルゴリズムクイックリファレンスP181
// 使えない辺の距離を10^9とする
var UNREACH int = 1000000000

func floydWarshall(u, v, w [3001]int, construction [3001]bool) (dist [3001][3001]int) {
	for i := 0; i < 3001; i++ {
		for j := 0; j < 3001; j++ {
			dist[i][j] = UNREACH
		}
		dist[i][i] = 0
	}
	for i := 1; i <= M; i++ {
		if construction[i] {
			continue
		}
		dist[u[i]][v[i]] = w[i]
		dist[v[i]][u[i]] = w[i]
		if u[i] == 1 {
			log.Println(dist[u[i]][v[i]])
		}
	}
	for k := 1; k <= M; k++ {
		for i := 1; i <= M; i++ {
			if dist[k][i] == UNREACH {
				continue
			}
			for j := 1; j <= M; j++ {
				if dist[i][k]+dist[k][j] < dist[i][j] {
					dist[i][j] = dist[i][k] + dist[k][j]
				}
			}
		}
	}
	return dist
}

func dailyScore(dist0 [3001][3001]int, ngEdge []int) (score int) {
	for i := range ngEdge {
		u, v := edges(ngEdge[i])
		dist[u][v] = UNREACH
		dist[v][u] = UNREACH
	}
	return score
}
