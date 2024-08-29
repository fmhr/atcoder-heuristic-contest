package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"slices" // "golang.org/x/exp/slices"
	"sort"
	"strings"
	"time"
)

// A配列のPaからl個をB配列のPbに代入する
func signaleOpe(length, Pa, Pb int, A, B []int) {
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

type Path []int

// Floyd-Warshall Algorithm
// in.roadsから各都市間の最短経路をもとめる
func allPairsShortest(in Input) ([V][V]int, [V][V]int) {
	var dist [V][V]int
	var pred [V][V]int
	for i := 0; i < V; i++ {
		for j := 0; j < V; j++ {
			dist[i][j] = math.MaxInt32
			pred[i][j] = -1
			if i == j {
				dist[i][j] = 0
			}
		}
	}
	for _, road := range in.roads[:in.M] {
		u := road[0]
		v := road[1]
		dist[u][v] = 1
		dist[v][u] = 1
		pred[u][v] = u
		pred[v][u] = v
	}
	for k := 0; k < V; k++ {
		for u := 0; u < V; u++ {
			if dist[k][u] >= math.MaxInt32 {
				continue
			}
			for v := 0; v < V; v++ {
				if dist[k][v] >= math.MaxInt32 {
					continue
				}
				if newDist := dist[u][k] + dist[k][v]; newDist < dist[u][v] {
					dist[u][v] = newDist
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

// ルートにそって、配列Aをつくる
// ただし全ての都市を配列Aにいれなくてはいけない
func initialA(in Input, pred, dist [V][V]int) (A []int) {
	A = make([]int, 0, in.La)
	visitedCnt := 0
	var visited [V]int
	for i := 0; i < len(in.plan)-1 && in.La-len(A) > V-visitedCnt; i++ {
		u, v := in.plan[i], in.plan[i+1]
		root := constructShortestPath(u, v, pred, dist)
		for j := 1; j < len(root); j++ {
			if visited[root[j]] == 0 {
				visited[root[j]]++
				visitedCnt++
			}
			A = append(A, root[j])
			if in.La-len(A) == V-visitedCnt {
				break
			}
		}
	}
	// のこりのノードを順に
	//	for i := 0; i < V; i++ {
	//if visited[i] == 0 {
	//A = append(A, i)
	//visited[i] = 0
	//}
	//}
	log.Println(A)
	//まだ追加していない都市を隣接する都市の横に追加する
	unVisited := make([]int, 0)
	for i := 0; i < V; i++ {
		if visited[i] == 0 {
			unVisited = append(unVisited, i)
		}
	}
	for len(unVisited) > 0 {
	again:
		for i, u := range unVisited {
			for v := 0; v < V; v++ {
				if dist[u][v] == 1 {
					index := slices.Index(A, v)
					if index >= 0 {
						A = slices.Insert(A, index, u)
						unVisited = slices.Delete(unVisited, i, i+1)
						goto again
					}
				}
			}
		}
	}
	log.Println(A)
	return A
}

// u->vの最短経路列挙
// []intはuとvを含む
func findALLShortestPaths(dist [V][V]int, u, v int) (fullpath [][]int) {
	var queue [][]int
	queue = append(queue, []int{u})
	for len(queue) > 0 {
		current_path := queue[0]
		queue = queue[1:]
		current := current_path[len(current_path)-1]
		// 終了条件
		if current == v {
			fullpath = append(fullpath, current_path)
			continue
		}
		for k := 0; k < V; k++ {
			if dist[current][k] > 1 {
				continue
			}
			//　すでに経路に含まれている場合、パスする
			if slices.Contains(current_path, k) {
				continue
			}
			if dist[current][k]+dist[k][v] == dist[current][v] {
				newPath := make([]int, len(current_path))
				copy(newPath, current_path)
				newPath = append(newPath, k)
				queue = append(queue, newPath)
			}
		}
	}
	return fullpath
}

// 配列操作の操作幅
func getLength(in Input) (length []int) {
	for i := 1; i < in.Lb; i++ {
		length = append(length, i)
	}
	return length
}

// sliceIndexs return index value = v
func sliceIndexs(a []int, v int) (indexs []int) {
	for i := 0; i < len(a); i++ {
		if a[i] == v {
			indexs = append(indexs, i)
		}
	}
	return indexs
}

// func movement(root []int, A, B []int) (length, Pa, Pb int) {
// indexs := sliceIndexs(A, root[0])
// log.Println(indexs)
// return
// }
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var startTime time.Time

func main() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		log.SetOutput(io.Discard)
	}
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	startTime = time.Now()
	var output strings.Builder
	in := readInput()
	log.Printf("La=%v Lb=%v\n", in.La, in.Lb)
	//log.Println(in)
	dist, pred := allPairsShortest(in)
	_ = pred
	A := initialA(in, pred, dist)
	B := make([]int, in.Lb)
	for i := range B {
		B[i] = -1
	}
	//log.Println(len(A))
	output.WriteString(fmt.Sprintln(strings.Trim(fmt.Sprint(A), "[]")))
	signalOperations := 0
	for i := 0; i < V-1+1; i++ { // in.planは０を先頭に追加してサイズが601
		u, v := in.plan[i], in.plan[i+1]
		//log.Println(in.plan[i], "->", in.plan[i+1], "cost=", dist[u][v])
		//root := constructShortestPath(u, v, pred, dist) // １つのルート
		// root決め 配列Bの都市と重複するものを優先する
		roots := findALLShortestPaths(dist, u, v) // u, v間の全てのルート
		// 配列Bに
		cntStep := 0
		bestRoot := roots[0]
		//log.Println("配列B:", B)
		//log.Println("includeCnt=", cntStep)
		for j := 0; j < len(roots); j++ {
			cnt := 0
			for k := 0; k < len(roots[j]); k++ {
				if slices.Contains(B, roots[j][k]) {
					cnt++
				} else {
					break
				}
			}
			if cnt > cntStep {
				cntStep = cnt
				bestRoot = roots[j]
			}
		}
		root := bestRoot
		// root が小さい時、次の移動を先読みして配列操作の精度を上げる
		// 時間の割にスコアが伸びないのでコメントアウト
		rootSize := len(root)
		//if i+2 < V && rootSize < 4 {
		//nextRoot := constructShortestPath(u, v, pred, dist)
		//root = append(root, nextRoot[1:minInt(4, len(nextRoot))]...)
		//}
		//log.Println(root)
		for j := 1; j < rootSize; j++ {
			if slices.Contains(B, root[j]) {
				output.WriteString(fmt.Sprintln("m", root[j]))
			} else {
				// 配列操作 s l Pa Pb
				//log.Println("配列操作", root[j:])
				indexs := sliceIndexs(A, root[j]) // 配列Aのなかの候補(これを含まなければいけない) Paの候補
				lengths := getLength(in)          // 操作する幅の候補 lの候補
				actions := make([][4]int, 0)
				//log.Println("next->", root[j])
				rootNext := root[j:]
				for lindex := range lengths {
					l := lengths[lindex]
					for ii := range indexs {
						i := indexs[ii]
						for Pb := 0; Pb < in.Lb; Pb++ {
							// Pa は配列Aのスタート
							for Pa := i - l + 1; Pa <= i; Pa++ {
								if Pa < 0 || Pa+l > len(A) || Pb < 0 || Pb+l > len(B) {
									continue
								}
								var act [4]int
								act[0], act[1], act[2], act[3] = l, Pa, Pb, 0
								pb := make([]int, len(B))
								copy(pb, B)
								copy(pb[Pb:Pb+l], A[Pa:Pa+l])
								for j := 0; j < minInt(len(rootNext), in.Lb); j++ {
									if slices.Contains(pb, rootNext[j]) {
										act[3]++
									}
								}
								//log.Println(rootNext, pb, act)
								actions = append(actions, act)
							}
						}
					}
				}

				sort.Slice(actions, func(i, j int) bool { return actions[i][3] > actions[j][3] })
				//index := slices.Index(A, root[j])
				//length := len(B)
				//length = minInt(length, len(A)-index)
				length := actions[0][0]
				Pa := actions[0][1]
				Pb := actions[0][2]
				output.WriteString(fmt.Sprintln("s", length, Pa, Pb))
				output.WriteString(fmt.Sprintln("m", root[j]))
				signaleOpe(length, Pa, Pb, A, B)
				signalOperations++
			}
		}
	}
	fmt.Print(output.String())
	var sumLong int
	for i := 0; i < V; i++ {
		u, v := in.plan[i], in.plan[i+1]
		//log.Printf("%+v\n", dist[u][v])
		//ps := findALLShortestPaths(dist, u, v)
		//for j := range ps {
		//log.Println(j, ps[j])
		//}
		//log.Println(u, v, "距離", dist[u][v], "経路", len(ps))
		sumLong += dist[u][v]
	}
	log.Println("総距離", sumLong, "信号操作", signalOperations)
	//log.Println(A)
	//log.Println(len(A))
	log.Printf("Length=%v\n", sumLong)
	log.Printf("C=%v\n", signalOperations)
	elpseTime := time.Since(startTime)
	log.Printf("time=%v\n", elpseTime.Milliseconds())
}

// utils
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
