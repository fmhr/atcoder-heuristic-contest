package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

var startTime time.Time
var frand = rand.New(rand.NewSource(1333))

func main() {
	startTime = time.Now()
	in := readInput()
	log.Printf("M=%d L=%d W=%d\n", in.M, in.L, in.W)
	solver(in)
	log.Printf("elapsed=%.2f\n", float64(time.Since(startTime).Microseconds())/1000)
}

type Point struct {
	Y, X int
}

// 大小関係がわかればいいので、√を取らない
func dist2(a, b Point) int {
	return (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y)
}

// 小数点以下切り捨て
func dist(a, b Point) int {
	return int(math.Floor(math.Sqrt(float64(dist2(a, b)))))
}

type City struct {
	Point
	ID int
}

// answerの出力用
type AnsGroup struct {
	Citys []int
	Edges [][2]int
	Cost  int
}

func (a AnsGroup) calcScore(cities []City) int {
	// エッジの長さの合計
	score := 0
	for _, edge := range a.Edges {
		score += dist(cities[edge[0]].Point, cities[edge[1]].Point)
	}
	return score
}

// Output()は、回答形式に合わせてStringに変換する
func (a AnsGroup) Output() (str string) {
	for i, city := range a.Citys {
		if i < len(a.Citys)-1 {
			str += fmt.Sprintf("%d ", city)
		} else {
			str += fmt.Sprintf("%d\n", city)
		}
	}
	for _, edge := range a.Edges {
		str += fmt.Sprintf("%d %d\n", edge[0], edge[1])
	}
	return str
}

func solver(in Input) {
	sortedGroup := make([]int, in.M)
	for i := 0; i < in.M; i++ {
		sortedGroup[i] = in.G[i]
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sortedGroup)))
	mapping := makeMapping(in.G[:in.M], sortedGroup)

	// 都市の初期値は範囲の中心とする
	var cities [N]City
	for i := 0; i < N; i++ {
		cities[i].ID = i
		cities[i].Y = (in.lxrxlyry[i*4+2] + in.lxrxlyry[i*4+3]) / 2
		cities[i].X = (in.lxrxlyry[i*4+0] + in.lxrxlyry[i*4+1]) / 2
	}
	center := Point{Y: 10000 / 2, X: 10000 / 2}
	var used [N]bool
	sortByCenterCities := make([]City, N)
	copy(sortByCenterCities, cities[:])
	sort.Slice(sortByCenterCities[:], func(i, j int) bool {
		return dist2(center, sortByCenterCities[i].Point) > dist2(center, sortByCenterCities[j].Point)
	})
	tmp := make([]City, N)
	copy(tmp, cities[:])
	ansGrops := make([]AnsGroup, in.M)
	for i := 0; i < in.M; i++ {
		// グループのrootを決める
		groupRoot := -1
		for _, city := range sortByCenterCities {
			if !used[city.ID] {
				groupRoot = city.ID
				used[city.ID] = true
				ansGrops[i].Citys = append(ansGrops[i].Citys, city.ID)
				break
			}
		}
		// rootからの距離が近い順にソートする
		sort.Slice(tmp[:], func(i, j int) bool {
			return dist2(cities[groupRoot].Point, tmp[i].Point) < dist2(cities[groupRoot].Point, tmp[j].Point)
		})
		// グループに都市を追加する
		// Edgesは、グループのrootと都市を結ぶエッジ
		for _, city := range tmp {
			if len(ansGrops[i].Citys) >= sortedGroup[i] {
				break
			}
			if !used[city.ID] {
				ansGrops[i].Citys = append(ansGrops[i].Citys, city.ID)
				used[city.ID] = true
			}
		}
		//log.Println("i:", i, "groupRoot:", groupRoot, "requre:", sortedGroup[i], "cities:", len(ansGrops[i].Citys))
		tmpCity := make([]City, len(ansGrops[i].Citys))
		for j, city := range ansGrops[i].Citys {
			tmpCity[j] = cities[city]
		}
		ansGrops[i].Edges = createMST(tmpCity)
		if len(ansGrops[i].Citys) > 2 {
			// query test
			q := make([]int, 0, in.L)
			for j := 0; j < in.L && j < len(ansGrops[i].Citys); j++ {
				q = append(q, ansGrops[i].Citys[j])
			}
			edge := query(q)
			log.Println(in.L, len(ansGrops[i].Citys), q, edge)
			if in.L >= len(ansGrops[i].Citys) {
				ansGrops[i].Edges = edge
			}
		}
	}
	// クエリの終了
	fmt.Println("!")
	for i := 0; i < in.M; i++ {
		fmt.Print(ansGrops[mapping[i]].Output())
	}
	return

	// kd-treeを作成する
	//kdt := NewKDTree(cities[:])
	//printTree(kdt.Root, 0)

	// 全都市間の距離を計算する
	// Groupの都市数が多い順に、都市間が短いエッジを結ぶ
	// Kruskal法に近いけど、エッジはGrpupと新しい都市を結ぶ時のみつなぐ
	//	var allDistance [N][N]int
	//allEdge := make([]Edge, 0, N*(N-1)/2)
	//for i := 0; i < N; i++ {
	//allDistance[i][i] = 0
	//for j := i + 1; j < N; j++ {
	//allDistance[i][j] = (cities[i].X-cities[j].X)*(cities[i].X-cities[j].X) + (cities[i].Y-cities[j].Y)*(cities[i].Y-cities[j].Y)
	//allDistance[j][i] = allDistance[i][j]
	//allEdge = append(allEdge, Edge{From: i, To: j, Weight: allDistance[i][j]})
	//}
	//}
	//// allEdgeを重みが小さい順にソートする
	//sort.Slice(allEdge, func(i, j int) bool {
	//return allEdge[i].Weight < allEdge[j].Weight
	//})

	//uf := NewUnionFind(N) // 全体の森
	////var used [N]bool

	//ansGroup := make([]AnsGroup, in.M)
	//for i := 0; i < in.M; i++ {
	//// sortedGroup[i]の都市をグループにする
	//size := sortedGroup[i]
	//var root int = -1
	//if size == 1 {
	//// size=1の時は、rootのみを探す
	//for j := 0; j < N; j++ {
	//if uf.size[j] == 1 && !used[j] {
	//root = j
	//used[j] = true
	//break
	//}
	//}
	//if root == -1 {
	//log.Println("Error: root not found")
	//}
	//ansGroup[i].Citys = append(ansGroup[i].Citys, cities[root].ID)
	//} else {
	//for j := 0; j < size; j++ {
	//if j == 0 {
	//// sortedGroup[i]の最初の都市は、グループの親(root)にする
	//for _, edge := range allEdge {
	//if uf.size[edge.From] == 1 && uf.size[edge.To] == 1 && !used[edge.From] && !used[edge.To] {
	//// 両方ともグループに属していない
	//root = edge.From
	//ansGroup[i].Citys = append(ansGroup[i].Citys, cities[edge.From].ID)
	//used[edge.From] = true
	//break
	//}
	//}
	//if root == -1 {
	//log.Println("Error: root not found")
	//}
	//} else {
	//// rootと繋がる都市を探す
	//for _, edge := range allEdge {
	//if uf.Find(edge.From) == root && uf.Find(edge.To) != root && uf.size[edge.To] == 1 {
	//// rootに繋がる都市を見つけた
	//ansGroup[i].Citys = append(ansGroup[i].Citys, cities[edge.To].ID)
	//ansGroup[i].Edges = append(ansGroup[i].Edges, [2]int{cities[edge.From].ID, cities[edge.To].ID})
	//uf.Union(edge.From, edge.To)
	//used[edge.To] = true
	//ansGroup[i].Cost += edge.Weight
	//break
	//}
	//if uf.Find(edge.To) == root && uf.Find(edge.From) != root && uf.size[edge.From] == 1 {
	//// rootに繋がる都市を見つけた
	//ansGroup[i].Citys = append(ansGroup[i].Citys, cities[edge.From].ID)
	//ansGroup[i].Edges = append(ansGroup[i].Edges, [2]int{cities[edge.To].ID, cities[edge.From].ID})
	//uf.Union(edge.To, edge.From)
	//used[edge.From] = true
	//ansGroup[i].Cost += edge.Weight
	//break
	//}
	//}
	//}
	//}
	//}
	//log.Println("request Size=", size, "root=", root, "size=", uf.size[root], "cost=", ansGroup[i].Cost, "cities=", ansGroup[i].Citys)
	//}
	//// 推定座標でスコアを計算
	//var score int
	//for i := 0; i < in.M; i++ {
	//score += ansGroup[i].calcScore(cities[:])
	//}
	//log.Printf("score=%d\n", score)

}

const (
	N = 800 // 都市の個数
	Q = 400 // クエリの個数
)

type Input struct {
	M        int        // 都市のグループの数 1<= M <= 400
	L        int        // クエリの都市の最大数 1<= L <= 15
	W        int        //　二次元座標の最大値 500 <= W <= 2500
	G        [400]int   // 各グループの都市の数 1<= G[i] <= N(800) i= 0..M-1
	lxrxlyry [N * 4]int // 各都市の座標 0 <= lxrxlyry[i] <= W
	// lxrxlyry[i] = (lx, rx, ly, ry) i=0..N-1
}

// 固定入力はとばす
func readInput() (in Input) {
	var n, q int
	fmt.Scan(&n, &in.M, &q, &in.L, &in.W)
	for i := 0; i < in.M; i++ {
		fmt.Scan(&in.G[i])
	}
	for i := 0; i < N*4; i++ {
		fmt.Scan(&in.lxrxlyry[i])
	}
	return in
}

type UnionFind struct {
	parent []int // 親ノードのインデックス
	size   []int // 各グループのサイズ rootのindexにアクセスする root以外は０にする
}

func NewUnionFind(n int) *UnionFind {
	uf := &UnionFind{
		parent: make([]int, n),
		size:   make([]int, n),
	}
	for i := 0; i < n; i++ {
		uf.parent[i] = i
		uf.size[i] = 1
	}
	return uf
}

// Findは、xの親ノードを返す
func (uf *UnionFind) Find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.Find(uf.parent[x])
	}
	return uf.parent[x]
}
func (uf *UnionFind) Union(x, y int) {
	rootX := uf.Find(x)
	rootY := uf.Find(y)
	if rootX != rootY {
		uf.parent[rootY] = rootX
	}
	uf.size[rootX] += uf.size[rootY]
	uf.size[rootY] = 0 // rootYはrootXに統合されたので、サイズは0にする
}
func (uf *UnionFind) Same(x, y int) bool {
	return uf.Find(x) == uf.Find(y)
}

type Edge struct {
	From, To int
	Weight   int
}
type Edges []Edge

func (e Edges) Len() int {
	return len(e)
}
func (e Edges) Less(i, j int) bool {
	return e[i].Weight < e[j].Weight
}
func (e Edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func Kruskal(n int, edges Edges) (int, []Edge) {
	uf := NewUnionFind(n)
	sort.Sort(edges)
	var mst []Edge
	mstWeight := 0
	for _, edge := range edges {
		if !uf.Same(edge.From, edge.To) {
			uf.Union(edge.From, edge.To)
			mst = append(mst, edge)
			mstWeight += edge.Weight
		}
	}
	return mstWeight, mst
}

// cityからkruskal用のedgeを作成して、最小全域木を求める
// kruskal用にcityを0からインデックスを振り直す
// edgesには、cityのIDに変換して返す

func createMST(cities []City) [][2]int {
	newIndex := make([]int, len(cities))
	for i := 0; i < len(cities); i++ {
		newIndex[i] = cities[i].ID
	}
	edges := make(Edges, 0)
	for i := 0; i < len(cities); i++ {
		for j := i + 1; j < len(cities); j++ {
			weight := (cities[i].X-cities[j].X)*(cities[i].X-cities[j].X) + (cities[i].Y-cities[j].Y)*(cities[i].Y-cities[j].Y)
			edges = append(edges, Edge{From: i, To: j, Weight: weight})
		}
	}
	cost, mst := Kruskal(len(cities), edges)
	_ = cost
	//log.Printf("cost=%d\n", cost)
	newEdge := make([][2]int, len(mst))
	for i := 0; i < len(mst); i++ {
		newEdge[i][0] = newIndex[mst[i].From]
		newEdge[i][1] = newIndex[mst[i].To]
	}
	return newEdge
}

func makeMapping(a, b []int) []int {
	if len(a) != len(b) {
		log.Fatal("makeMapping: length mismatch")
	}
	mapping := make([]int, len(a))
	used := make([]bool, len(a))
	for i, v := range a {
		for j, w := range b {
			if v == w && !used[j] {
				mapping[i] = j
				used[j] = true
				break
			}
		}
	}
	return mapping
}

// kd-tree
type Node struct {
	Point
	Index int
	Left  *Node
	Right *Node
}

type KDTree struct {
	Root *Node
}

func NewKDTree(cities []City) *KDTree {
	points := make([]Point, len(cities))
	for i, city := range cities {
		points[i] = Point{Y: city.Y, X: city.X}
	}
	nodes := make([]int, len(cities))
	for i := range nodes {
		nodes[i] = i
	}
	return &KDTree{Root: buildTree(points, nodes, 0)}
}

func buildTree(cities []Point, indices []int, depth int) *Node {
	if len(indices) == 0 {
		return nil
	}

	axis := depth % 2

	sort.Slice(indices, func(i, j int) bool {
		if axis == 0 {
			return cities[indices[i]].X < cities[indices[j]].X
		}
		return cities[indices[i]].Y < cities[indices[j]].Y
	})

	mid := len(indices) / 2

	return &Node{
		Point: cities[indices[mid]],
		Index: indices[mid],
		Left:  buildTree(cities, indices[:mid], depth+1),
		Right: buildTree(cities, indices[mid+1:], depth+1),
	}
}

func printTree(node *Node, depth int) {
	if node == nil {
		return
	}
	fmt.Printf("%s(%d, %d)\n", string(' '+depth*2), node.X, node.Y)
	printTree(node.Left, depth+1)
	printTree(node.Right, depth+1)
}

// query
func query(cities []int) (edges [][2]int) {
	str := fmt.Sprintf("? %d", len(cities))
	for _, city := range cities {
		str += fmt.Sprintf(" %d", city)
	}
	fmt.Println(str)
	for i := 0; i < len(cities)-1; i++ {
		var a, b int
		fmt.Scan(&a, &b)
		edges = append(edges, [2]int{a, b})
	}
	return edges
}
