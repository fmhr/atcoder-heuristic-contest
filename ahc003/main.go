package main

import (
	"bufio"
	"container/heap"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
)

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

var sc = bufio.NewScanner(os.Stdin)
var buff []byte

func nextInt() int {
	sc.Scan()
	i, err := strconv.Atoi(sc.Text())
	if err != nil {
		panic(err)
	}
	return i
}

func nextFloat64() float64 {
	sc.Scan()
	f, err := strconv.ParseFloat(sc.Text(), 64)
	if err != nil {
		panic(err)
	}
	return f
}

func init() {
	sc.Split(bufio.ScanWords)
	sc.Buffer(buff, bufio.MaxScanTokenSize*1024)
	log.SetFlags(log.Lshortfile)
}

// https://golang.org/pkg/runtime/pprof/
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

var local = flag.Bool("local", false, "if local")

func main() {
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

	// ... rest of the program ...
	if *local {
		localTester()
	} else {
		solver()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

type Point struct {
	i, j int
}

func (p *Point) move(m byte) {
	p.i = p.i + di[direction[m]]
	p.j = p.j + dj[direction[m]]
}

type Ask struct {
	s Point
	t Point
	a float64
	e float64
}

var direction = map[byte]int{'D': 0, 'R': 1, 'U': 2, 'L': 3}
var di = [4]int{1, 0, -1, 0}
var dj = [4]int{0, 1, 0, -1}
var dd = [4]byte{'D', 'R', 'U', 'L'}

var h [30][30]int
var v [30][30]int

func localTester() {
	// input testcase
	for i := 0; i < 30; i++ {
		for j := 0; j < 29; j++ {
			h[i][j] = nextInt()
		}
	}
	for i := 0; i < 29; i++ {
		for j := 0; j < 30; j++ {
			v[i][j] = nextInt()
		}
	}
	asks := make([]Ask, 1000)
	for i := 0; i < 1000; i++ {
		asks[i].s.i = nextInt()
		asks[i].s.j = nextInt()
		asks[i].t.i = nextInt()
		asks[i].t.j = nextInt()
		asks[i].a = nextFloat64()
		asks[i].e = nextFloat64()
	}
	var score float64
	for k := 0; k < 1000; k++ {
		path := query(asks[k].s.i, asks[k].s.j, asks[k].t.i, asks[k].t.j)
		fmt.Println(path)
		b := compute_path_length(asks[k].s, asks[k].t, path)
		score = score*0.998 + (asks[k].a)/float64(b)
	}
	score = math.Round(2312311.0 * score)
	log.Printf("%f\n", score)
}

func compute_path_length(start, goal Point, route string) (dest int) {
	now := start
	for i := 0; i < len(route); i++ {
		var next Point
		d := direction[route[i]]
		next.i = now.i + di[d]
		next.j = now.j + dj[d]
		switch d {
		case 0: // D
			dest += v[now.i][now.j]
		case 1: // R
			dest += h[now.i][now.j]
		case 2: // U
			dest += v[next.i][next.j]
		case 3: // L
			dest += h[next.i][next.j]
		}
		now = next
	}
	return
}

// 直線的に動く暫定
func query(si, sj, ti, tj int) (route string) {
	if si-ti < 0 {
		route += strings.Repeat("D", ti-si)
	} else {
		route += strings.Repeat("U", si-ti)
	}
	if sj-tj < 0 {
		route += strings.Repeat("R", tj-sj)
	} else {
		route += strings.Repeat("L", sj-tj)
	}
	return route
}

func randomSolver(q *QueryRecord, pr *PathRecord) []byte {
	si := q.start.i
	sj := q.start.j
	ti := q.stop.i
	tj := q.stop.j
	var now Point
	now.i = si
	now.j = sj
	rb := make([]byte, absInt(si-ti)+absInt(sj-tj))
	cnt := 0
	for !(now.i == ti && now.j == tj) {
		r := ""
		if now.i < ti {
			r += "D"
		} else if now.i > ti {
			r += "U"
		}
		if now.j < tj {
			r += "R"
		} else if now.j > tj {
			r += "L"
		}
		if len(r) == 0 {
			panic("Errorrrrr")
		}
		for i := 0; i < len(r); i++ {
			pr.AddAppeared(now, r[i])
		}
		rb[cnt] = r[rand.Intn(len(r))]
		pr.AddAppeared(now, rb[cnt])
		now.move(rb[cnt])
		cnt++
	}
	return rb
}

func sampleUCB(p Path) float64 {
	//v := 0.0
	//log.Println(math.Sqrt(math.Log(float64(p.numOfAppeared)) / float64(2*p.numOfSelected)))
	//v = float64(p.SampleAverage)
	//v = float64(p.SampleAverage) - math.Sqrt(math.Log(float64(p.numOfAppeared))/float64(2*p.numOfSelected))
	return float64(p.SampleAverage)
}

func greedySolver(q *QueryRecord, pr *PathRecord) (int, []byte) {
	cost := 0
	si := q.start.i
	sj := q.start.j
	ti := q.stop.i
	tj := q.stop.j
	var now Point
	now.i = si
	now.j = sj
	rb := make([]byte, absInt(si-ti)+absInt(sj-tj))
	cnt := 0
	for !(now.i == ti && now.j == tj) {
		r := ""
		if now.i < ti {
			r += "D"
		} else if now.i > ti {
			r += "U"
		}
		if now.j < tj {
			r += "R"
		} else if now.j > tj {
			r += "L"
		}
		if len(r) == 0 {
			panic("Errorrrrr")
		}
		for i := 0; i < len(r); i++ {
			pr.AddAppeared(now, r[i])
		}
		//
		ps := make([]Path, len(r))
		nouse := -1
		for i := 0; i < len(r); i++ {
			y, x := getIj(now, r[i])
			ps[i] = pr.getPath(y, x, r[i])
			ps[i].index = i
			if ps[i].numOfSelected == 0 {
				nouse = i
			}
		}
		if nouse != -1 {
			rb[cnt] = r[nouse]
			cost += 1
		} else {
			sort.Slice(ps, func(i, j int) bool {
				return sampleUCB(ps[i]) < sampleUCB(ps[j])
			})
			rb[cnt] = r[ps[0].index]
			cost += ps[0].SampleAverage
		}
		//rb[cnt] = r[rand.Intn(len(r))]
		pr.AddSelected(now, rb[cnt])
		now.move(rb[cnt])
		cnt++
	}
	return cost, rb
}

// worchal floyd ----------------------------------------------
const inf int = 2 << 29

type Graph struct {
	cost [900][900]int
	size int
}

var next [900][900]int
var g Graph

func toindex(i, j int) int {
	return i*30 + j
}
func fromindex(k int) (int, int) {
	return k / 30, k % 30
}
func buildGraph(pr PathRecord) {
	for i := 0; i < 900; i++ {
		for j := 0; j < 900; j++ {
			if i == j {
				g.cost[i][j] = 0
			}
			g.cost[i][j] = inf
		}
	}
	// X軸方向
	for i := 0; i < 30; i++ {
		for j := 0; j < 29; j++ {
			a := toindex(i, j)
			b := toindex(i, j+1)
			if pr.h[i][j].numOfSelected > 0 {
				g.cost[a][b] = pr.h[i][j].SampleAverage
				g.cost[b][a] = pr.h[i][j].SampleAverage
			} else {
				g.cost[a][b] = 100000000
				g.cost[b][a] = 100000000
			}
		}
	}
	// Y軸方向
	for i := 0; i < 29; i++ {
		for j := 0; j < 30; j++ {
			a := toindex(i, j)
			b := toindex(i+1, j)
			if pr.v[i][j].numOfSelected > 0 {
				g.cost[a][b] = pr.v[i][j].SampleAverage
				g.cost[b][a] = pr.v[i][j].SampleAverage
			} else {
				g.cost[a][b] = 100000000
				g.cost[b][a] = 100000000
			}
		}
	}
}

func warchalFloyd() {
	for i := 0; i < 900; i++ {
		for j := 0; j < 900; j++ {
			next[i][j] = j
		}
	}
	for k := 0; k < 900; k++ {
		for i := 0; i < 900; i++ {
			for j := 0; j < 900; j++ {
				if g.cost[i][j] > g.cost[i][k]+g.cost[k][j] {
					g.cost[i][j] = g.cost[i][k] + g.cost[k][j]
					next[i][j] = next[i][k]
				}
			}
		}
	}
}

func routeRestor(start, stop int) []int {
	route := make([]int, 0)
	for cur := start; cur != stop; cur = next[cur][stop] {
		route = append(route, cur)
	}
	route = append(route, stop)
	return route
}

func toMoves(route []int) (move string) {
	for i := 0; i < len(route)-1; i++ {
		switch route[i+1] - route[i] {
		case 1:
			move += "R"
		case 30:
			move += "D"
		case -1:
			move += "L"
		case -30:
			move += "U"
		}
	}
	return
}

/// ------------------------------------------------------
type QueryRecord struct {
	start  Point
	stop   Point
	move   []byte
	result int
}

type Path struct {
	numOfAppeared int
	numOfSelected int
	SampleAverage int
	index         int
}

type PathRecord struct {
	h    [30][30]Path // y,i方向
	v    [30][30]Path // x,j方向
	time int
}

type tmpPath struct {
	i, j int
	move byte
}

func (pr PathRecord) getCntSelected() int {
	cnt := 0
	for i := 0; i < 30; i++ {
		for j := 0; j < 30; j++ {
			if pr.h[i][j].numOfSelected > 0 {
				cnt++
			}
		}
	}
	for i := 0; i < 30; i++ {
		for j := 0; j < 30; j++ {
			if pr.v[i][j].numOfSelected > 0 {
				cnt++
			}
		}
	}
	return cnt
}

func (pr PathRecord) getPath(i, j int, move byte) Path {
	if move == 'D' || move == 'U' {
		return pr.v[i][j]
	} else {
		return pr.h[i][j]
	}
}

func (pr *PathRecord) setDistance(p tmpPath, d int) {
	if p.move == 'D' || p.move == 'U' {
		pr.v[p.i][p.j].SampleAverage = d
	} else {
		pr.h[p.i][p.j].SampleAverage = d
	}
}

func getIj(now Point, move byte) (int, int) {
	var i, j int
	if move == 'D' || move == 'R' {
		i = now.i
		j = now.j
	} else if move == 'U' || move == 'L' {
		i = now.i + di[direction[move]]
		j = now.j + dj[direction[move]]
	}
	return i, j
}

func (pr *PathRecord) AddAppeared(now Point, move byte) {
	i, j := getIj(now, move)
	if move == 'D' || move == 'U' {
		pr.v[i][j].numOfAppeared++
	} else if move == 'R' || move == 'L' {
		pr.h[i][j].numOfAppeared++
	}
}

func (pr *PathRecord) AddSelected(now Point, move byte) {
	i, j := getIj(now, move)
	if move == 'D' || move == 'U' {
		pr.v[i][j].numOfSelected++
	} else if move == 'R' || move == 'L' {
		pr.h[i][j].numOfSelected++
	}
}

func (pr *PathRecord) AddAverage(now Point, move byte, dis int) {
	i, j := getIj(now, move)
	if move == 'D' || move == 'U' {
		if pr.v[i][j].numOfSelected == 1 {
			pr.v[i][j].SampleAverage = dis
		} else {
			pr.v[i][j].SampleAverage = pr.v[i][j].SampleAverage*(pr.v[i][j].numOfSelected-1) + dis
			pr.v[i][j].SampleAverage = pr.v[i][j].SampleAverage / pr.v[i][j].numOfSelected
		}
	} else if move == 'R' || move == 'L' {
		if pr.h[i][j].numOfSelected == 1 {
			pr.h[i][j].SampleAverage = dis
		} else {
			pr.h[i][j].SampleAverage = pr.h[i][j].SampleAverage*(pr.h[i][j].numOfSelected-1) + dis
			pr.h[i][j].SampleAverage = pr.h[i][j].SampleAverage / pr.h[i][j].numOfSelected
		}
	}
}

func (pr *PathRecord) ReflectResult(q QueryRecord) {
	now := q.start
	average := q.result / len(q.move)
	for i := 0; i < len(q.move); i++ {
		pr.AddAverage(now, q.move[i], average)
		now.move(q.move[i])
	}
}

// dijkstra go
type Edge struct {
	to, cost int
}
type Dijkstra struct {
	edges [30][]Edge
	prev  [30]int
}

func (g *Dijkstra) buildGridEdge(pr PathRecord) {
	for i := 0; i < 30; i++ {
		for j := 0; j < 30; j++ {
			for d := 0; d < 4; d++ {
				ni := i + di[d]
				nj := j + dj[d]
				if ni >= 0 && ni < 30 && nj >= 0 && nj < 30 {
					path := pr.getPath(i, j, dd[d])
					cost := path.SampleAverage
					g.edges[i*30+j] = append(g.edges[i*30+j], Edge{to: ni*30 + nj, cost: cost})
				}
			}
		}
	}
	for i := 0; i < 30; i++ {
		g.prev[i] = -1
	}
}

// TODO テスト
// n:numNode s:source
func (g *Dijkstra) do(s int) {
	d := make([]int, 30)
	for i := 0; i < 30; i++ {
		d[i] = inf
	}
	d[s] = 0
	pq := make(PriorityQueue, 0)
	heap.Push(&pq, &Item{node: s, cost: 0})
	for pq.Len() > 0 {
		p := heap.Pop(&pq).(*Item)
		v := p.node
		if d[v] < p.cost {
			continue
		}
		for _, e := range g.edges[v] {
			if d[e.to] > d[v]+e.cost {
				d[e.to] = d[v] + e.cost
				heap.Push(&pq, &Item{cost: d[e.to], node: e.to})
				g.prev[e.to] = v
			}
		}
	}

}
func (g *Dijkstra) getPath(s, t int) (p []int) {
	g.do(s)
	for t != -1 {
		p = append(p, g.prev[t])
		t = g.prev[t]
	}
	for i := 0; i < len(p)/2; i++ {
		p[i], p[len(p)-i-1] = p[len(p)-i-1], p[i]
	}
	return
}

func solver() {
	var pr PathRecord
	var i int
	for i = 0; i < 1000; i++ {
		var cost int
		var q QueryRecord
		q.start.i = nextInt()
		q.start.j = nextInt()
		q.stop.i = nextInt()
		q.stop.j = nextInt()
		cost, q.move = greedySolver(&q, &pr)
		fmt.Println(string(q.move))
		q.result = nextInt()
		pr.ReflectResult(q)
		_ = cost
		if i%8 == 0 {
			cnt := float64(pr.getCntSelected())
			if cnt/1750 > 0.96 {
				i++
				break
			}
		}
	}
	log.Printf("turn=%d\n", i)
	if i < 1000 {
		buildGraph(pr)
		warchalFloyd()
		for ; i < 1000; i++ {
			var q QueryRecord
			q.start.i = nextInt()
			q.start.j = nextInt()
			q.stop.i = nextInt()
			q.stop.j = nextInt()
			s := toindex(q.start.i, q.start.j)
			t := toindex(q.stop.i, q.stop.j)
			path := routeRestor(s, t)
			fmt.Println(toMoves(path))
			q.move = []byte(toMoves(path))
			q.result = nextInt()
		}
	}
}

// An Item is something we manage in a priority queue.
type Item struct {
	node int // The value of the item; arbitrary.
	cost int // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
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

func (pq *PriorityQueue) update(item *Item, value int, priority int) {
	item.node = value
	item.cost = priority
	heap.Fix(pq, item.index)
}
