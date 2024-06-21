package main

import (
	"container/heap"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

var N, W, K, C int

// N = 200
// W 水源  1 ~ 4
// K 家 1 ~ 10
// C 体力 {1, 2, 4, 8, 16, 32, 61, 128}

// Sij 岩盤の硬さ10~5000は与えられない

type Point struct {
	y, x int
}

func (p Point) Cmp(q Point) (re int) {
	if p.y == q.y {
		if p.x == q.x {
			re = 0
		} else if p.x > q.x {
			re = 1
		} else {
			re = -1
		}
	} else if p.y > q.x {
		re = 1
	} else {
		re = -1
	}
	return
}

func distance(a, b Point) int {
	return abs(b.y-a.y) + abs(b.x-a.x)
}

// 水源と家は岩盤の硬さと反比例した確率で選ばれる
type House struct {
	Point
	water bool // if true, bedrock is breaked.
}

var WaterSources [4]Point
var houses [10]House

// waterway is cell with water
var waterway []Point

// mapping  --------------------------------------
type Cell struct {
	crushed  bool // 0:未開拓 1:破壊済み
	water    bool // 水源であっても岩盤があると水路にならない
	typ      int
	power    int // 採掘した累計力
	index    int
	estimate int
	digCount int
}

type Mapping [200][200]Cell

func (m Mapping) getPower(p Point) int {
	return m[p.y][p.x].power
}

func (m Mapping) wasBroken(p Point) bool {
	return m[p.y][p.x].crushed
}

func (m Mapping) guessCost(a, b Point) (cost int) {
	apower := mapping.getPower(a)
	if !mapping[a.y][a.x].crushed {
		apower = 50000
	}
	bpower := mapping.getPower(b)
	if !mapping[b.y][b.x].crushed {
		bpower = 50000
	}
	d := distance(a, b)
	cost = (apower + bpower) / 2 * d
	if d > DIS {
		cost += cost * d
	}
	if cost < 0 {
		log.Fatal(cost)
	}
	return
}

func (m *Mapping) openWater(p Point) {
	m[p.y][p.x].water = true
}

var totalCost int
var serverCnt int

func server(p Point, power int) (re int) {
	totalCost = totalCost + power + C
	serverCnt++
	fmt.Println(p.y, p.x, power)
	re = receiver()
	update(p.y, p.x, power, re)
	return re
}

func update(y, x, power, re int) {
	mapping[y][x].power += power
	mapping[y][x].digCount++
	// 岩盤破壊
	if re == 1 || re == 2 {
		// 次のタスクにすすめる
		if len(stackTask) > 0 {
			stackTask = stackTask[1:]
		}
		mapping[y][x].crushed = true
		// 水判定
		var water bool
		// 水源
		if mapping[y][x].water {
			water = true
		} else {
			cells := adjacentCells(Point{y, x})
			for _, c := range cells {
				water = mapping[c.y][c.x].water || water
			}
		}
		if water {
			mapping[y][x].water = true
			// 家があるとき
			if mapping[y][x].typ == HOUSE {
				houses[mapping[y][x].index].water = true
				//log.Println("水到達")
			}
			waterway = append(waterway, Point{y, x})
		}
	}
}

var mapping Mapping

func main() {
	log.SetFlags(log.Lshortfile)
	startTime := time.Now()
	initializes()
	InitialInput()
	paramaterInit()
	reactiveSolver()
	timeDistance := time.Since(startTime)
	log.Printf("N=%d W=%d K=%d C=%d\n", N, W, K, C)
	log.Printf("totalCost=%d,%d,%d\n", totalCost, serverCnt, serverCnt*C)
	log.Printf("time=%d\n", timeDistance.Milliseconds())
	//var crushedCellCnt int
	//var unCrushedCellCnt int
	//var totalDigCellCount int
	//var totalDigCount int
	//var totalDigCountCrushed int
	//var totalDigPower int
	//var CPower int
	//var totalPower int

	//for i := 0; i < 200; i++ {
	//for j := 0; j < 200; j++ {
	//c := mapping[i][j]
	//if c.crushed {
	//crushedCellCnt++
	//totalDigCountCrushed += c.digCount
	//totalPower += c.power
	//}
	//if c.power > 0 && !c.crushed {
	//unCrushedCellCnt++
	//}
	//if c.digCount > 0 {
	//totalDigCellCount++
	//totalDigCount += c.digCount
	//totalDigPower += c.power
	//}
	//}
	//}
	//CPower = totalDigCount * C
	//log.Println("crushedCellCnt=", crushedCellCnt, "unCrushedCellCnt=", unCrushedCellCnt)
	//log.Println("totalDigCellCount=", totalDigCellCount, "totalDigPower=", totalDigPower, "Cpower=", CPower)
	//log.Println("無駄ぼり回数=", totalDigCellCount-crushedCellCnt)
	//log.Println("壊すまでにかかった平均掘削回数=", float64(totalDigCountCrushed)/float64(crushedCellCnt))
	//log.Println("壊すまでにかかった平均パワー", float64(totalPower)/float64(crushedCellCnt))
	//log.Println("改善できそうなパワー")
}

func initializes() {
	for i := 0; i < 200; i++ {
		for j := 0; j < 200; j++ {
			mapping[i][j].index = -1
		}
	}
}

func InitialInput() {
	fmt.Scan(&N, &W, &K, &C)
	var a, b, c, d int
	for i := 0; i < W; i++ {
		fmt.Scan(&a, &b)
		WaterSources[i].y = a
		WaterSources[i].x = b
		mapping[a][b].typ = WATERSOURCE
		mapping[a][b].index = i
	}
	for i := 0; i < K; i++ {
		fmt.Scan(&c, &d)
		houses[i].y = c
		houses[i].x = d
		mapping[c][d].typ = HOUSE
		mapping[c][d].index = i
	}
	// initial setup
	for i := 0; i < W; i++ {
		waterway = append(waterway, WaterSources[i])
		mapping.openWater(WaterSources[i])
	}
}

func reactiveSolver() {
	var turn int
	trialDig()
	for {
		turn++
		next, y, x, P := turnSolver()
		if next == 0 {
			log.Println("next==0")
			break
		}
		re := server(Point{y, x}, P)
		//log.Println(re)
		// 0 : 岩盤が破壊できなかった
		// 1 : 岩盤が破壊できた　かつ　全ての家に水が流れていない
		// 2 : 石板が破壊できた  かつ　全ての家に水が流れた
		// -1 : 不正な入力
		if re == -1 || re == 2 {
			break
		}
	}
}

func receiver() (re int) {
	fmt.Scan(&re)
	return
}

var stackTask []Point

// next:次の場所の有無 1:あり 2：なし
func turnSolver() (next, y, x, P int) {
	if len(stackTask) > 0 {
		next, place := nextPlace()
		if next == 1 {
			power := calculatePower(place)
			return next, place.y, place.x, power
		}
	}
	if len(stackTask) == 0 {
		log.Fatal("stackTask is zero")
	}
	return
}

// すでに岩盤が破壊されている場所をとばす
// 次の場所が 1：ある 0:ない
func nextPlace() (int, Point) {
	for len(stackTask) > 0 {
		if mapping[stackTask[0].y][stackTask[0].x].crushed {
			update(stackTask[0].y, stackTask[0].x, 0, 1)
		} else {
			if len(stackTask) == 0 {
				log.Fatal("stackTask is nothing. TODO")
			}
			return 1, stackTask[0]
		}
	}
	return 0, Point{-1, -1}
}

var PP1 map[int]int = map[int]int{1: 12, 2: 16, 4: 16, 8: 16, 16: 16, 32: 64, 64: 64, 128: 64}
var PP1a map[int]int = map[int]int{1: 12, 2: 16, 4: 16, 8: 16, 16: 24, 32: 64, 64: 64, 128: 128}
var PP1b map[int]int = map[int]int{1: 12, 2: 16, 4: 14, 8: 16, 16: 24, 32: 45, 64: 50, 128: 50}
var PP2 map[int]int = map[int]int{1: 64, 2: 32, 4: 24, 8: 16, 16: 6, 32: 10, 64: 3, 128: 2}
var PP3 map[int]int = map[int]int{1: 8, 2: 8, 4: 8, 8: 8, 16: 8, 32: 8, 64: 8, 128: 8}
var P1a, P1b, P2 int

func paramaterInit() {
	P1a = PP1a[C]
	P1b = PP1b[C]
	P2 = PP2[C]
}

func calculatePower(place Point) (power int) {
	target := mapping[place.y][place.x]
	if target.crushed {
		log.Fatal("try dig broken bedrock.", place)
	}
	if target.power == 0 {
		if target.estimate == 0 {
			ac := aroundCell(place)
			if len(ac) > 0 {
				for _, acp := range ac {
					if mapping[acp.y][acp.x].crushed {
						power = mapping.getPower(acp) * 8 / 10
					}
				}
			}
		} else {
			return target.estimate
		}
	}
	if target.power == 0 && power == 0 {
		power = P1a
	} else if target.power > 0 {
		power = max(P1b, target.power/P2)
	}
	power = min(5000-target.power, power)
	return
}

func aroundCell(p Point) (next []Point) {
	for d := 0; d < 4; d++ {
		var np Point
		np.y = p.y + dy[d]
		np.x = p.x + dx[d]
		if np.y >= 0 && np.y < 200 && np.x >= 0 && np.x < 200 {
			next = append(next, np)
		}
	}
	return
}

// routeSearch is a to b, rにはaとbを含む
func routeSearch(a, b Point) (r []Point) {
	t := tan(a, b)
	vy := v1(a.y, b.y)
	vx := v1(a.x, b.x)
	r = append(r, a)
	for !(a.x == b.x && a.y == b.y) {
		if distance(a, b) == 1 {
			a = b
		} else {
			if t >= tan(a, b) && a.x != b.x {
				a.x += vx
			} else {
				a.y += vy
			}
		}
		r = append(r, a)
		if a.x > 200 || a.y > 200 {
			log.Fatal("over cell")
		}
	}
	return
}

func adjacentCells(p Point) (r []Point) {
	for i := 0; i < 4; i++ {
		n := Point{p.y + dy[i], p.x + dx[i]}
		if n.y >= 0 && n.y < 200 && n.x >= 0 && n.x < 200 {
			r = append(r, n)
		}
	}
	return
}

var DIS = 22

// 格子状にDIS間隔に中継点をおく
func trialPoints() (ps []Point) {
	for i := 0; ; i++ {
		y := (DIS/3 + i*DIS) / 2
		x0 := DIS/3 + (i%2)*(DIS/2)
		if y >= 200 {
			break
		}
		for j := 0; ; j++ {
			x := j*DIS + x0
			if x >= 200 {
				break
			}
			if mapping[y][x].typ == 0 {
				ps = append(ps, Point{y, x})
			}
		}
	}
	return
}

// 3点の重心に中継点を置く
func centroidOfTriangle(wh []Point) (re []Point) {
	for i := 0; i < W+K; i++ {
		for j := i + 1; j < W+K; j++ {
			for k := j + 1; k < W+K; k++ {
				y1, x1 := wh[i].y, wh[i].x
				y2, x2 := wh[j].y, wh[j].x
				y3, x3 := wh[k].y, wh[k].x
				y := (y1 + y2 + y3) / 3
				x := (x1 + x2 + x3) / 3
				var dist int = 100000000
				for _, a := range wh {
					dist = min(dist, distance(Point{y, x}, a))
				}
				if len(re) > 0 {
					for _, r := range re {
						dist = min(dist, distance(Point{y, x}, r))
					}
				}
				if (dist > 14 && dist < 30) || len(re) == 0 {
					re = append(re, Point{y, x})
				}
			}
		}
	}
	return
}

// TODO
// 家から近い水源を２つあげる（これで最低三角）
// どの家からも届かない水源は除外する
// 家と水源から近い補助点４つを候補にする
// 候補集合で凸包をつくる（家と水源より少しおおきな凸包)
// 凸包内の補助点のみ掘る
// 補助点のN %が有効になるまで堀続ける

var digLimitInside int = 200
var digLimitOutside int = 50

func trialDig() {
	// 水源、家の採掘　*使わない水源もコストを計算するために岩盤を壊す必要がある
	wh := make([]Point, 0)
	for i := 0; i < W; i++ {
		wh = append(wh, WaterSources[i])
	}
	for i := 0; i < K; i++ {
		wh = append(wh, houses[i].Point)
	}
	// check convex hull
	ch := convexHull(wh)
	// ----------------------------
	tp := trialPoints()
	//tp := centroidOfTriangle(wh)
	log.Println(len(tp))
	for i := 0; i < len(tp); i++ {
		//log.Println(tp[i], ch.inside(tp[i]))
		wh = append(wh, tp[i])
	}
	for i := 0; i < len(wh); i++ {
		var re int
		for re == 0 {
			if i < W {
			} else if i < W+K {
			} else {
				var limit int
				if ch.inside(wh[i]) {
					limit = digLimitInside
				} else {
					limit = digLimitOutside
				}
				if mapping.getPower(wh[i]) > limit {
					break
				}
			}
			power := calculatePower(wh[i])
			re = server(wh[i], power)
		}
		//if i > W && i < W+K {
		//digLimitInside = max(digLimitInside, mapping.getPower(wh[i])*5/10)
		//}
	}
	// stainer treee
	steinerTree(wh)
	// ----------------------------------
	return
	// グラフ化
	nodeMap := map[Point]int{}
	nodeList := make([]Node, 0)
	for i := 0; i < len(wh); i++ {
		t := mapping[wh[i].y][wh[i].x].typ
		if t == 0 {
			t = RELAYPOINT
		}
		nodeList = append(nodeList, Node{Point: wh[i], typ: t, index: i})
		nodeMap[wh[i]] = i
	}
	edges := make([]Edge, 0)
	for i := 0; i < len(nodeList); i++ {
		for j := i + 1; j < len(nodeList); j++ {
			s := nodeList[i]
			t := nodeList[j]
			//if distance(s.Point, t.Point) <= DIS*19/10 { // 端の水源、家はDIS以下に中継点がない可能性がある
			//log.Println(s, t, distance(s.Point, t.Point))
			cost := mapping.guessCost(s.Point, t.Point)
			edges = append(edges, Edge{i, j, cost})
			//}
		}
	}
	g := newGraph(len(nodeList))
	for _, e := range edges {
		g[e.from] = append(g[e.from], Edge{to: e.to, cost: e.cost})
		g[e.to] = append(g[e.to], Edge{to: e.from, cost: e.cost})
	}
	dists := make([][]int, W+K)
	prevs := make([][]int, W+K)
	for i := 0; i < W+K; i++ {
		dists[i], prevs[i] = g.dijkstar(len(nodeList), i)
	}
	// -----------------------------------------------------------
	// 水源と家の最小全域木
	kg := NewKruskalGraph(W+K, (W+K)*(W+K))
	kg.V = W + K
	kg.E = kg.V * kg.V
	for i := 0; i < kg.V; i++ {
		for j := 0; j < kg.V; j++ {
			e := &edge{}
			e.u = i
			e.v = j
			e.cost = dists[i][j]
			if i < W && j < W {
				e.cost = 0
			}
			kg.edges = append(kg.edges, *e)
		}
	}
	for i := 0; i < W; i++ {
		for j := i + 1; j < W; j++ {
			kg.uf.Unit(i, j)
		}
	}
	_ = kg.kruskal()
	var checked [200][200]bool
	for _, e := range kg.used {
		dists[e.u], prevs[e.u] = g.dijkstar(len(nodeList), e.u)
		r1 := restorRoute(e.u, e.v, prevs[e.u])
		for i := 1; i < len(r1); i++ {
			if checked[r1[i]][r1[i-1]] {
				continue
			}
			r2a, r2b := nodeList[r1[i]].Point, nodeList[r1[i-1]].Point
			g[r1[i]] = append(g[r1[i]], Edge{to: r1[i-1], cost: 1})
			g[r1[i-1]] = append(g[r1[i-1]], Edge{to: r1[i], cost: 1})
			// 中継点と中継点の間
			r2 := routeSearch(r2a, r2b)
			checked[r1[i]][r1[i-1]] = true
			checked[r1[i-1]][r1[i]] = true
			// ここの区間で推定を行う
			//estimateValue(r2)
			stackTask = append(stackTask, r2...)
		}
	}
	//log.Println(stackTask)
}

// restorRoute
func restorRoute(start, goal int, prev []int) (route []int) {
	route = append(route, goal)
	var now = goal
	for now != start {
		now = prev[now]
		route = append(route, now)
	}
	return
}

func estimateValue(route []Point) {
	s := mapping.getPower(route[0])
	t := mapping.getPower(route[len(route)-1])
	if s == 0 || t == 0 || !mapping.wasBroken(route[0]) || !mapping.wasBroken(route[len(route)-1]) {
		return
	}
	v := ((t - s) + len(route) - 1) / len(route)
	var k int
	k = s
	for i := 1; i < len(route)-1; i++ {
		k += v
		mapping[route[i].y][route[i].x].estimate = k
		//log.Println(i, mapping[route[i].y][route[i].x])
	}
}

// steiner tree solver -------------------------------------------------

func steinerTree(nodes []Point) {
	// wfで中間点を含む全経路距離をとる
	// 中継点のない最小全域木をみる
	// 中継点のaddとdeleteの焼きなまし
	// 経路復元
	n := len(nodes)
	g := make([][]int, n)
	for i := 0; i < n; i++ {
		g[i] = make([]int, n)
	}
	oldg := make([][]int, n)
	for i := 0; i < n; i++ {
		oldg[i] = make([]int, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				g[i][j] = 0
				oldg[i][j] = 0
			} else {
				g[i][j] = mapping.guessCost(nodes[i], nodes[j])
				tmp := g[i][j]
				oldg[i][j] = tmp
			}
		}
	}
	for k := 0; k < n; k++ {
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				g[i][j] = min(g[i][j], g[i][k]+g[k][j])
			}
		}
	}
	//for i := 0; i < n; i++ {
	//for j := i + 1; j < n; j++ {
	//log.Println(i, j, g[i][j])
	//}
	//}
	use := make([]bool, n)
	for i := 0; i < W+K; i++ {
		use[i] = true
	}
	size := W + K
	v, edges := smt(use, size, g)
	log.Println(v)
	log.Println(edges)
	for loop := 0; loop < 1000; loop++ {
		// swap
		if rand.Intn(10) < 3 {
			index1 := rand.Intn(n-W-K) + W + K
			index2 := rand.Intn(n-W-K) + W + K
			for i := 0; i < 10 && use[index1] == use[index2]; i++ {
				index2 = rand.Intn(n-W-K) + W + K
			}
			if use[index1] != use[index2] {
				use[index1], use[index2] = use[index2], use[index1]
				new, nedges := smt(use, size, g)
				if v > new {
					v = new
					edges = nedges
				}
			}
			continue
		}
		// ------
		var index = 0
		index = rand.Intn(n-W-K) + W + K
		for loop > 900 && !use[index] {
			index = rand.Intn(n-W-K) + W + K
			if rand.Intn(10) < 2 {
				break
			}
		}
		size += reverse(use, index)
		new, nedges := smt(use, size, g)
		// reset
		//var alpha int
		//if loop < 3000 {
		//alpha = new * rand.Intn(10) / 100
		//}
		if v > new {
			v = new
			edges = nedges
		} else {
			size += reverse(use, index)
		}
		//log.Println(v, new, size)
		//log.Println(use)
	}
	log.Println(edges)
	for i := 0; i < len(edges); i++ {
		u, v := edges[i][0], edges[i][1]
		path := restorePath(u, v, n, oldg, g)
		//log.Println(u, v, path)
		for j := 1; j < len(path); j++ {
			digpoints := routeSearch(nodes[path[j-1]], nodes[path[j]])
			//log.Println(digpoints)
			stackTask = append(stackTask, digpoints...)
		}
	}
}

func reverse(use []bool, index int) int {
	if use[index] {
		use[index] = false
		return -1
	} else {
		use[index] = true
		return 1
	}
}

func smt(use []bool, size int, g [][]int) (int, [][2]int) {
	nodes := make([]int, 0, size)
	for i := 0; i < len(use); i++ {
		if use[i] {
			nodes = append(nodes, i)
		}
	}
	n := size
	kg := NewKruskalGraph(n, n*n)
	kg.V = n
	kg.E = n * n
	for i := 0; i < kg.V; i++ {
		for j := 0; j < kg.V; j++ {
			e := &edge{}
			e.u, e.v, e.cost = i, j, g[nodes[i]][nodes[j]]
			if nodes[i] < W && nodes[j] < W {
				kg.uf.Unit(i, j)
			}
			kg.edges = append(kg.edges, *e)
		}
	}
	v := kg.kruskal()
	edge := make([][2]int, 0)
	for _, e := range kg.used {
		edge = append(edge, [2]int{nodes[e.u], nodes[e.v]})
	}
	return v, edge
}

func restorePath(start, end, n int, g1, g2 [][]int) (p []int) {
	//log.Println(start, end)
	p = append(p, start)
	curr := start
	for curr != end {
		stamp := make([]bool, n)
		for i := 0; i < n; i++ {
			if stamp[i] {
				continue
			}
			if i != curr && g1[curr][i]+g2[i][end] == g2[curr][end] {
				curr = i
				p = append(p, i)
				stamp[i] = true
				break
			}
		}
		//log.Println(curr, end)
	}
	return
}

// --------------------------------------------------------------------
// dijkstar
type Node struct {
	Point
	typ   int // 0:None 1:waterserver 2:home 3:relayPoint -1:unavailable
	index int
}

const (
	WATERSOURCE = 1
	HOUSE       = 2
	RELAYPOINT  = 3
	UNAVAILABLE = -1
)

type Edge struct {
	from int
	to   int
	cost int
}
type Graph [][]Edge

func newGraph(n int) Graph {
	g := make(Graph, n)
	for i := range g {
		g[i] = make([]Edge, 0)
	}
	return g
}

const inf int = 2 << 29

func (g *Graph) dijkstar(nodeSize, start int) (dist, prev []int) {
	dist = make([]int, nodeSize)
	prev = make([]int, nodeSize)
	for i := 0; i < nodeSize; i++ {
		dist[i] = inf
	}
	dist[start] = 0
	que := make(PriorityQueue, 0)
	heap.Push(&que, &Item{node: start, cost: 0})
	for que.Len() > 0 {
		p := heap.Pop(&que).(*Item)
		v := p.node
		if dist[v] < p.cost {
			continue
		}
		for _, e := range (*g)[v] {
			if dist[e.to] > dist[v]+e.cost {
				dist[e.to] = dist[v] + e.cost
				prev[e.to] = v
				heap.Push(&que, &Item{cost: dist[e.to], node: e.to})
			}
		}
	}
	return
}

type Item struct {
	node  int // The value of the item; arbitrary.
	cost  int // The priority of the item in the queue.
	index int // The index of the item in the heap.
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].cost < pq[j].cost
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// ------------------------------------
// kruskal

type edge struct {
	u, v, cost int
}

type kruskalGraph struct {
	V, E  int
	edges []edge
	used  []edge
	uf    *unionFind
}

type ByCost []edge

func (a ByCost) Len() int           { return len(a) }
func (a ByCost) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCost) Less(i, j int) bool { return a[i].cost < a[j].cost }

func NewKruskalGraph(v, e int) (k kruskalGraph) {
	k.V = v
	k.E = e
	k.uf = makeUnionFind(k.V)
	return
}

func (k *kruskalGraph) kruskal() int {
	sort.Sort(ByCost(k.edges))
	//log.Println(k.edges)
	cost := 0
	for i := 0; i < k.E; i++ {
		e := k.edges[i]
		if !k.uf.Same(e.u, e.v) {
			k.uf.Unit(e.u, e.v)
			cost += e.cost
			k.used = append(k.used, e)
		}
	}
	return cost
}

type unionFind struct {
	par  []int
	size []int
	link [][]int
}

func makeUnionFind(count int) *unionFind {
	par := make([]int, count)
	size := make([]int, count)
	link := make([][]int, count)
	for i := range link {
		link[i] = make([]int, 1)
	}
	for i := 0; i < count; i++ {
		par[i] = i
		size[i] = 1
		link[i][0] = i
	}
	return &unionFind{par: par, size: size, link: link}
}

func (uf *unionFind) Find(a int) int {
	for uf.par[a] != a {
		uf.par[a] = uf.par[uf.par[a]]
		a = uf.par[a]
	}
	return a
}

func (uf *unionFind) Unit(a, b int) {
	a = uf.Find(a)
	b = uf.Find(b)
	if a != b {
		if uf.size[a] > uf.size[b] {
			a, b = b, a
		}
		uf.par[a] = b
		uf.size[b] += uf.size[a]
		uf.link[b] = append(uf.link[b], uf.link[a]...)
	}
}

func (uf *unionFind) Same(a, b int) bool {
	return uf.Find(a) == uf.Find(b)
}

// -----------------------------------------
type Polygon []Point

func convexHull(pp Polygon) Polygon {
	p := make([]Point, len(pp))
	copy(p, pp)
	n := len(p)
	if n < 3 {
		return p
	}
	sort.Slice(p, func(i, j int) bool {
		if p[i].x == p[j].x {
			return p[i].y < p[j].y
		}
		return p[i].x < p[j].x
	})
	var upper Polygon
	upper = append(upper, p[0])
	upper = append(upper, p[1])
	for i := 2; i < n; i++ {
		upper = append(upper, p[i])
		for len(upper) >= 3 && calCp(upper[len(upper)-3], upper[len(upper)-2], upper[len(upper)-1]) < 0 {
			upper = upper[:len(upper)-2]
			upper = append(upper, p[i])
		}
	}
	var lower Polygon
	lower = append(lower, p[n-1])
	lower = append(lower, p[n-2])
	for i := n - 3; i >= 0; i-- {
		lower = append(lower, p[i])
		for len(lower) >= 3 && calCp(lower[len(lower)-3], lower[len(lower)-2], lower[len(lower)-1]) < 0 {
			lower = lower[:len(lower)-2]
			lower = append(lower, p[i])
		}
	}
	for i, l := range lower {
		if i != 0 && i != len(lower)-1 {
			upper = append(upper, l)
		}
	}
	return upper
}

func calCp(l0, l1, p Point) int {
	cp := (l1.x-l0.x)*(p.y-l0.y) - (l1.y-l0.y)*(p.x-l0.x)
	if cp == 0 {
		return 0
	} else if cp > 0 {
		return 1
	} else {
		return -1
	}
}

// https://tjkendev.github.io/procon-library/python/geometry/point_inside_polygon.html
func (p Polygon) inside(p0 Point) bool {
	var cnt int
	L := len(p)
	y, x := p0.y, p0.x
	for i := 1; i < L; i++ {
		y0, x0 := p[i-1].y, p[i-1].x
		y1, x1 := p[i].y, p[i].x
		x0 -= x
		y0 -= y
		x1 -= x
		y1 -= y
		cv := x0*x1 + y0*y1
		sv := x0*y1 - x1*y0
		if sv == 0 && cv == 0 {
			return true
		}
		if !(y0 < y1) {
			x0, x1 = x1, x0
			y0, y1 = y1, y0
		}
		if y0 <= 0 && 0 < y1 && x0*(y1-y0) > y0*(x1-x0) {
			cnt++
		}
	}
	return (cnt%2 == 1)
}

// -----------------------------------------
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func tan(a, b Point) float64 {
	return float64(abs(b.y-a.y)) / float64(abs(b.x-a.x))
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}

func v1(a, b int) int {
	if b-a >= 0 {
		return 1
	} else {
		return -1
	}
}

var dy [4]int = [4]int{1, 0, -1, 0}
var dx [4]int = [4]int{0, 1, 0, -1}

// checked
