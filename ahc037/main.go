package main

import (
	"fmt"
	"log"
	"math"
	"sort"
)

const (
	N = 1000
)

type Input struct {
	sodas [1000]soda
	L     int
}

type soda struct {
	x, y     int
	parent   int
	children []int
	cost     int // parentからのコスト
	created  bool
}

func searchMini(u []soda, n int) (p int) {
	c := u[n]
	miniCost := (c.x + c.y) * 2
	for i := 0; i < len(u); i++ {
		if i == n {
			continue
		}
		if c.x >= u[i].x && c.y >= u[i].y {
			cost := c.x - u[i].x + c.y - u[i].y
			if cost < miniCost {
				miniCost = cost
				p = i
			}
		}
	}
	return
}

type ans struct {
	out  [][4]int
	cost int
}

func (a ans) Score(L int) int {
	return int(math.Round(1000000 * (float64((N * L)) / float64(1+a.cost))))
}

func readInput() (in Input) {
	_N := 0
	fmt.Scan(&_N)
	for i := 0; i < N; i++ {
		fmt.Scan(&in.sodas[i].x, &in.sodas[i].y)
		in.L = maxInt(in.L, in.sodas[i].x)
		in.L = maxInt(in.L, in.sodas[i].y)
	}
	return in
}

// x = seet, y = carbon とすると、
// x'>=x y'>=y なので、小さいものからつくっていく

func solve(in Input) {
	sort.Slice(in.sodas[:], func(i, j int) bool {
		return in.sodas[i].x < in.sodas[j].x
	})

	used := map[[2]int]bool{}
	S := make([]soda, 0, N+1)
	S = append(S, soda{x: 0, y: 0, created: true})
	for i := 0; i < N; i++ {
		S = append(S, in.sodas[i])
		used[[2]int{in.sodas[i].x, in.sodas[i].y}] = true
	}

	// 中間地点になる100追加する
	//for i := 0; i < 200; i++ {
	//x, y := rand.Intn(1000000000), rand.Intn(1000000000)
	//S = append(S, soda{x: x, y: y})
	//}
	lenSize := len(S)
again:

	for i := 1; i < len(S); i++ {
		p := searchMini(S, i)
		S[i].parent = p
		S[p].children = append(S[p].children, i)
	}

	// chiledrenが２以上のものをさがす
	for i := 0; i < len(S); i++ {
		if len(S[i].children) >= 2 {
			//log.Println(i, S[i].x, S[i].y, S[i].children)
			// 2つの子供の中間地点を作る
			//　短い法を探す
			var p, a, b Point
			p.x, p.y = S[i].x, S[i].y
			a.x, a.y = S[S[i].children[0]].x, S[S[i].children[0]].y
			b.x, b.y = S[S[i].children[1]].x, S[S[i].children[1]].y
			// yの大きい法をaとする
			if a.y < b.y {
				a, b = b, a
			}
			// p-aがp-bよりも短い場合
			if distance(p, a) < distance(p, b) {
				// aからp-bに垂線を引いた点を求める
				y := int(math.Floor(findIntersection(p, b, a.x)))
				x := a.x
				//log.Println(p, x >= p.x, y >= p.y)
				//log.Println(a, a.x >= x, a.y >= y)
				//log.Println(b, b.x >= x, b.y >= y)
				//log.Println(x, y)
				if x == p.x && y == p.y {
					//log.Println("Pass")
					continue
				}
				if _, ok := used[[2]int{x, y}]; ok {
					continue
				}
				S = append(S, soda{x: x, y: y})
				used[[2]int{x, y}] = true
			} else if distance(p, a) > distance(p, b) {
				x := int(math.Floor(findIntersectionY(p, a, b.y)))
				y := b.y
				//log.Println(p, x >= p.x, y >= p.y)
				//log.Println(a, a.x >= x, a.y >= y)
				//log.Println(b, b.x >= x, b.y >= y)
				//log.Println(x, y)
				if x == p.x && y == p.y {
					//log.Println("Pass")
					continue
				}
				if _, ok := used[[2]int{x, y}]; ok {
					continue
				}
				S = append(S, soda{x: x, y: y})
				used[[2]int{x, y}] = true
			}
		}
	}
	if lenSize != len(S) {
		for i := 0; i < len(S); i++ {
			S[i].parent = 0
			S[i].children = make([]int, 0)
		}
		log.Println("again", len(S))
		lenSize = len(S)
		goto again
	}

	for i := 1; i < len(S); i++ {
		p := searchMini(S, i)
		S[i].parent = p
		//S[p].children = append(S[p].children, i)
	}

	var a ans
	var createSoda func(i int)
	createSoda = func(i int) {
		if S[i].created {
			return
		}
		p := S[S[i].parent]
		if !p.created {
			createSoda(S[i].parent)
			//fmt.Println(p.x, p.y, S[i].x, S[i].y)
		}
		a.out = append(a.out, [4]int{p.x, p.y, S[i].x, S[i].y})
		a.cost += S[i].x - p.x + S[i].y - p.y
		S[i].created = true
	}
	// 1000個作る 10001個目以降は中継地点
	for i := 1; i < N+1; i++ {
		createSoda(i)
	}
	log.Println(len(a.out), a.cost, a.Score(in.L))
	fmt.Println(len(a.out))
	for i := 0; i < len(a.out); i++ {
		fmt.Println(a.out[i][0], a.out[i][1], a.out[i][2], a.out[i][3])
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	in := readInput()
	solve(in)
}

// utils
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type Point struct {
	x, y int
}

// 2点間の距離
func distance(p1, p2 Point) float64 {
	return math.Sqrt(float64((p1.x-p2.x)*(p1.x-p2.x) + (p1.y-p2.y)*(p1.y-p2.y)))
}

// 線分pb上に点aからx軸に垂線を引いた時の交点
func findIntersection(p, b Point, ax int) float64 {
	// 線分PBが垂直な場合（x座標が同じ場合）の処理
	if p.x == b.x {
		if p.x == ax {
			// 線分PBとx=axが重なる場合、任意の点（ここではPを選択）を返す
			return float64(p.y)
		}
		// 交点が存在しない場合
		return math.NaN()
	}

	// 線分PBの傾きと切片を計算
	m := float64(b.y-p.y) / float64(b.x-p.x)
	c := float64(p.y) - float64(m)*float64(p.x)

	// 交点のy座標を計算
	y := m*float64(ax) + c
	return y
}

func findIntersectionY(p, b Point, ay int) float64 {
	// 線分PBが水平な場合（y座標が同じ場合）の処理
	if p.y == b.y {
		if p.y == ay {
			// 線分PBとy=ayが重なる場合、任意の点（ここではPを選択）を返す
			return float64(p.x)
		}
		// 交点が存在しない場合
		return math.NaN()
	}

	// 線分PBの傾きと切片を計算
	m := float64(b.x-p.x) / float64(b.y-p.y)
	c := float64(p.x) - m*float64(p.y)

	// 交点のx座標を計算
	x := m*float64(ay) + c

	// 交点が線分PB上にあるかチェック
	minY := math.Min(float64(p.y), float64(b.y))
	maxY := math.Max(float64(p.y), float64(b.y))
	if float64(ay) < minY || float64(ay) > maxY {
		return math.NaN()
	}

	return x
}
