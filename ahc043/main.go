package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

var ATCODER bool     // AtCoder環境かどうか
var frand *rand.Rand // シード固定乱数生成器 rand.Seed()は無効 go1.24~

func init() {
	log.SetFlags(log.Lshortfile)
	frand = rand.New(rand.NewSource(1))
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
	_ = in
	//ans := beamSearch(*in)
	ans := ChokudaiSearch(*in)
	_, _ = fmt.Fprintln(writer, ans)
	//log.Printf("in=%+v\n", in)
	log.Printf("time=%v\n", time.Since(startTime).Milliseconds())
}

type bsAction struct {
	path []Pos
	typ  []int8
}

func ChokudaiSearch(in Input) string {
	stations := ChooseStationPositionFast(in)
	log.Printf("stations=%d", len(stations))
	edges, _ := constructMSTRailway(in, stations)

	// すべての駅の場所と、それらをつなぐエッジを行動にする
	allAction := make([]bsAction, 0, len(stations)+len(edges))
	for _, s := range stations {
		allAction = append(allAction, bsAction{path: []Pos{s}, typ: []int8{STATION}})
	}
	for _, e := range edges {
		allAction = append(allAction, bsAction{path: e.Path, typ: e.Rail})
	}

	initialState := newBsState(&in, len(allAction))
	// chokudai search
	pq := make([]*PriorityQueue2, len(allAction)+1)
	for i := 0; i < len(allAction)+1; i++ {
		pq[i] = &PriorityQueue2{}
		heap.Init(pq[i])
	}
	heap.Push(pq[0], initialState)

	bestState := initialState.Clone()
	bestScore := initialState.state.score
	loopCnt := 0
	var elapsedTime time.Duration
	for {
		newCnt := 0
		for i := 0; i < len(allAction); i++ {
			if pq[i].Len() == 0 {
				continue
			}
			cur := heap.Pop(pq[i]).(*bsState)
			if i > 1 && cur.state.income == 0 {
				continue
			}
			// 残っているアクションを実行する
			for j := 0; j < len(cur.restActions); j++ {
				act := &allAction[cur.restActions[j]]
				ok, connect := cur.state.field.CanBuildRail(act.path, act.typ)
				if (ok && connect) || (ok && i == 0) {
					p := make([]Pos, 0, len(act.path))
					t := make([]int8, 0, len(act.typ))
					for k := 0; k < len(act.path); k++ {
						typ := act.typ[k]
						now := cur.state.field.cell[act.path[k].Index()]
						if typ == now || now == STATION {
							continue
						}
						p = append(p, act.path[k])
						t = append(t, typ)
					}
					if len(p) == 0 {
						cur.restActions = append(cur.restActions[:j], cur.restActions[j+1:]...)
						continue
					}
					cost := calBuildCost(t)
					if cost > cur.state.money+cur.state.income*(in.T-cur.state.turn) {
						continue
					}
					// 駅は最後に建てる
					if t[0] == STATION && len(t) > 1 {
						if t[len(t)-1] == STATION && len(t) > 2 {
							t[0], t[len(t)-2] = t[len(t)-2], t[0]
							p[0], p[len(p)-2] = p[len(p)-2], p[0]
						} else {
							t[0], t[len(t)-1] = t[len(t)-1], t[0]
							p[0], p[len(p)-1] = p[len(p)-1], p[0]
						}
					}
					newState := cur.Clone()
					for k := 0; k < len(p); k++ {
						for buildCost[t[k]] > newState.state.money {
							// 建築費用までお金を貯める
							newState.state.do(Action{Kind: DO_NOTHING}, in, k == len(p)-1)
						}
						err := newState.state.do(Action{Kind: t[k], Y: p[k].Y, X: p[k].X}, in, k == len(p)-1)
						if err != nil {
							panic(err)
						}
					}
					// delete restActions
					newState.restActions = append(newState.restActions[:j], newState.restActions[j+1:]...)
					heap.Push(pq[i+1], newState)
					newCnt++
					if newState.state.score > bestScore {
						bestState = newState.Clone()
						bestScore = newState.state.score
						//log.Println(i, j, "bestScore", bestScore)
					}
				}
				elapsedTime = time.Since(startTime)
			}
			if elapsedTime > 2900*time.Millisecond {
				break
			}
		}
		if elapsedTime > 2900*time.Millisecond {
			break
		}
		loopCnt++
	}
	log.Printf("loop=%d\n", loopCnt)
	log.Println("bestScore", bestScore)
	log.Println(bestState.state.field.ToString())

	sb := strings.Builder{}
	for _, act := range bestState.state.actions {
		if act.Kind == DO_NOTHING && act.Num > 0 {
			log.Println("DO_NOTHING:", act.Num)
			for i := 0; i < int(act.Num); i++ {
				sb.WriteString("-1\n")
			}
		}
		sb.WriteString(act.String())
	}
	for i := 0; i < in.T-bestState.state.turn; i++ {
		sb.WriteString("-1\n")
	}
	return sb.String()
}

// 駅の位置を受け取る
// 1.各駅ごとに４までの近い駅を探す
// 2.src->dstで駅間を繋ぐ経路を探す
// 駅間の経路が使われた回数をカウントする
// よく使われたものからfieldに追加していく
// 更新されたfieldでつくられた経路でグラフを作るo
// src->dstで経路を探す 駅から駅までの経路を１つのActionとする
func BuildGraph(in Input, stations []Pos) {
	// 中央の駅から始める
	sort.Slice(stations, func(i, j int) bool {
		return distance(Pos{Y: N / 2, X: N / 2}, stations[i]) < distance(Pos{Y: N / 2, X: N / 2}, stations[j])
	})
	for i, s := range stations {
		log.Printf("station[%d]=%+v dist=%d\n", i, s, distance(Pos{Y: N / 2, X: N / 2}, s))
	}
	// 1
	g := make([][]DijEdge, len(stations))
	for i := 0; i < len(stations); i++ {
		for j := 0; j < len(stations); j++ {
			if i == j {
				continue
			}
			dist := distance(stations[i], stations[j])
			g[i] = append(g[i], DijEdge{To: j, Cost: dist})
		}
		sort.Slice(g[i], func(k, j int) bool {
			return g[i][k].Cost < g[i][j].Cost
		})
		//g[i] = g[i][:4] // 4までに限定すると総距離は当然伸びる
		//if len(g[i]) > 10 {
		//g[i] = g[i][:10]
		//}
	}
	// 2
	// 前処理で駅間の経路を求める
	// station_path_min_max[i][j] = iからjまでの経路　値は駅のindex
	// j>i
	station_path_min_max := make([][][]int, len(stations))
	for i := 0; i < len(stations); i++ {
		station_path_min_max[i] = make([][]int, len(stations))
		for j := i + 1; j < len(stations); j++ {
			//log.Printf("%d->%d L1=%d\n", i, j, distance(stations[i], stations[j]))
			path, dist := dijkstra(g, i, j)
			//log.Printf("dist=%d path=%+v\n", dist, path)
			if len(path) != 0 {
				station_path_min_max[i][j] = path
			}
			if dist > distance(stations[i], stations[j]) {
				//log.Printf("L1:%d dist:%d\n", distance(stations[i], stations[j]), dist)
			}
		}
	}
	// src->dstの最も近い駅を探す
	// 各セルに対して、最も近い駅を探す
	// 重複は上書きしてしまう
	var stationGrid [2500]int // 駅のindex
	for i := 0; i < 2500; i++ {
		stationGrid[i] = -1
	}
	for i, s := range stations {
		stationGrid[index(s.Y, s.X)] = i
		for d := 0; d < 13; d++ {
			y, x := s.Y+ddy[d], s.X+ddx[d]
			if y >= 0 && y < N && x >= 0 && x < N {
				stationGrid[index(s.Y, s.X)] = i
			}
		}
	}
	//log.Println(gridToString(stationGrid))
	// すべてのsrc->dstを回して、経路が何回使われたかをカウントする
	countPath_min_max := map[[2]int]int{} // 駅間のカウント
	var sumDist int
	for i := 0; i < in.M; i++ {
		start := stationGrid[in.src[i].Y*50+in.src[i].X]
		end := stationGrid[in.dst[i].Y*50+in.dst[i].X]
		path := station_path_min_max[min(start, end)][max(start, end)]
		dist := 0
		// pathは駅を辿っている
		// pathの中の駅間はL1距離なので,distance()をつかって距離を計算する
		// len(path)は駅の数
		for j := 1; j < len(path); j++ {
			dist += distance(stations[path[j-1]], stations[path[j]])
			a, b := min(path[j-1], path[j]), max(path[j-1], path[j])
			countPath_min_max[[2]int{a, b}]++
		}
		log.Printf("%d L1:%d 駅間L1:%d dist:%d\n", i, distance(in.src[i], in.dst[i]), distance(stations[start], stations[end]), dist)
		sumDist += dist
	}
	log.Println("総距離", sumDist)
	log.Println("countPath_min_max", len(countPath_min_max))
	// 使われた回数が多いものからfieldに追加していく
	// tmpCoutPathをつかって、sortする
	tmpCoutPath := make([][3]int, 0, len(countPath_min_max))
	for k, v := range countPath_min_max {
		tmpCoutPath = append(tmpCoutPath, [3]int{k[0], k[1], v})
	}
	sort.Slice(tmpCoutPath, func(i, j int) bool {
		return tmpCoutPath[i][2] > tmpCoutPath[j][2]
	})
	for _, v := range tmpCoutPath {
		log.Printf("to:%d from:%d count:%d\n", v[0], v[1], v[2])
	}
	// 使われた回数が多いものからfieldに追加していく
	f := NewField(50)
	for s := range stations {
		err := f.Build(Action{Kind: STATION, Y: stations[s].Y, X: stations[s].X})
		if err != nil {
			panic("invalid station")
		}
	}
	for _, v := range tmpCoutPath {
		stationIndexs := station_path_min_max[v[0]][v[1]]
		if len(stationIndexs) == 0 {
			continue
		}
		to := stations[stationIndexs[0]]
		from := stations[stationIndexs[1]]
		path := f.FindNewPath(from, to)
		typ := f.SelectRails(path)
		if len(path) == 0 {
			continue
		}
		//log.Println(from, to, len(path), distance(from, to))
		//log.Println(float64(len(path)) / float64(distance(from, to)) * 100)
		ok, _ := f.CanBuildRail(path, typ)
		if !ok {
			log.Println("can't build rail")
			continue
		}
		for i := 0; i < len(path); i++ {
			if typ[i] == STATION || f.cell[path[i].Index()] == STATION {
				continue
			}
			if err := f.Build(Action{Kind: typ[i], Y: path[i].Y, X: path[i].X}); err != nil {
				log.Println("can't build rail")
			}
		}
	}
	log.Println(f.ToString())
}

// 問題固有の固定値
const (
	N            = 50
	COST_STATION = 5000
	COST_RAIL    = 100
)

// ４方向
const (
	// 0:UP, 1:RIGHT, 2:DOWN, 3:LEFT
	UP    = 0
	RIGHT = 1
	DOWN  = 2
	LEFT  = 3
)

const (
	// Action
	DO_NOTHING int8 = -1
)

// セルの種類
const (
	EMPTY           int8 = -1
	STATION         int8 = 0
	RAIL_HORIZONTAL int8 = 1
	RAIL_VERTICAL   int8 = 2
	RAIL_LEFT_DOWN  int8 = 3
	RAIL_LEFT_UP    int8 = 4
	RAIL_RIGHT_UP   int8 = 5
	RAIL_RIGHT_DOWN int8 = 6
	WALL            int8 = 7 // テストの障害物として使う
)

// UP, RIGHT, DOWN, LEFT
var dy = []int8{-1, 0, 1, 0}
var dx = []int8{0, 1, 0, -1}

// int16ToString は、レールタイプのintの種類を文字列に変換する
// EMPTY = DO_NOTHING = -1 に注意
func int16ToString(a int8) string {
	switch a {
	case EMPTY:
		return "EMPTY"
	case STATION:
		return "STATION"
	case RAIL_HORIZONTAL:
		return "RAIL_HORIZONTAL"
	case RAIL_VERTICAL:
		return "RAIL_VERTICAL"
	case RAIL_LEFT_DOWN:
		return "RAIL_LEFT_DOWN"
	case RAIL_LEFT_UP:
		return "RAIL_LEFT_UP"
	case RAIL_RIGHT_UP:
		return "RAIL_RIGHT_UP"
	case RAIL_RIGHT_DOWN:
		return "RAIL_RIGHT_DOWN"
	case WALL:
		return "OTHER"
	}
	return "UNKNOWN"
}

func isRail(kind int8) bool {
	return kind >= RAIL_HORIZONTAL && kind <= RAIL_RIGHT_DOWN
}

// railToString は、[]intのレールの種類を文字列に変換する
func railToString(rails []int8) string {
	var sb strings.Builder
	for _, rail := range rails {
		sb.WriteString(" ")
		sb.WriteString(railMap[rail])
	}
	return sb.String()
}

var railMap = map[int8]string{
	EMPTY:           ".",
	STATION:         "◎",
	RAIL_HORIZONTAL: "─",
	RAIL_VERTICAL:   "│",
	RAIL_LEFT_DOWN:  "┐",
	RAIL_LEFT_UP:    "┘",
	RAIL_RIGHT_UP:   "└",
	RAIL_RIGHT_DOWN: "┌",
	WALL:            "#",
}

var buildCost = map[int8]int{
	EMPTY:           0,            // EMPTY
	STATION:         COST_STATION, // STATION
	RAIL_HORIZONTAL: COST_RAIL,
	RAIL_VERTICAL:   COST_RAIL,
	RAIL_LEFT_DOWN:  COST_RAIL,
	RAIL_LEFT_UP:    COST_RAIL,
	RAIL_RIGHT_UP:   COST_RAIL,
	RAIL_RIGHT_DOWN: COST_RAIL,
	WALL:            0, // other
}

// calBuildCost は、[]actの建設コストを計算する
func calBuildCost(act []int8) (cost int) {
	for _, a := range act {
		if val, ok := buildCost[a]; ok {
			cost += val
		} else {
			log.Printf("calBuildCost: invalid kind:%d\n", a)
			panic("calBuildCost: invalid kind")
		}
	}
	return
}

type Action struct {
	Kind    int8
	Y, X    int8
	Num     int16
	comment *string
}

func (a Action) String() (str string) {
	if a.Kind == DO_NOTHING {
		str = fmt.Sprintf("%d", a.Kind)
	} else {
		str = fmt.Sprintf("%d %d %d", a.Kind, a.Y, a.X)
	}
	str = *a.comment + str + "\n"
	return
}

type Field struct {
	cell [N * N]int8
	//stations []Pos
	coverd BitSet //駅によってカバーされた位置
}

// n == 50
// 全マスをノードとして、UFをもつ
func NewField(n int) *Field {
	if n != 50 {
		panic("n need to be 50")
	}
	f := new(Field)
	for i := int8(0); i < int8(n); i++ {
		for j := int8(0); j < int8(n); j++ {
			f.cell[index(i, j)] = EMPTY
		}
	}
	//f.stations = make([]Pos, 0)
	return f
}

func (f *Field) Clone() *Field {
	if f == nil {
		return nil
	}
	newField := &Field{
		cell: [N * N]int8{},
		//stations: make([]Pos, len(f.stations)),
	}
	//copy(newField.stations, f.stations)
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			newField.cell[i*50+j] = f.cell[i*50+j]
		}
	}
	newField.coverd = f.coverd
	return newField
}

// typeToString は、posのセルの種類を返す 表示用のレールの記号
//func (f Field) typeToString(pos Pos) string {
//return railMap[f.cell[pos.Y][pos.X]]
//}

// ToString は、Fieldを文字列に変換する
func (f Field) ToString() string {
	str := "view cellString()\n"
	for i := 0; i < N; i++ {
		str += fmt.Sprintf("%2d ", i)
		for j := 0; j < N; j++ {
			str += railMap[f.cell[i*50+j]]
		}
		str += "\n"
	}
	return str
}

func (f *Field) Build(act Action) error {
	if act.Kind == DO_NOTHING {
		return nil
	}
	// 範囲外
	if act.Kind < 0 || act.Kind > 6 {
		//panic("invalid kind:" + fmt.Sprint(act.Kind))
		return fmt.Errorf("invalid kind:%s", fmt.Sprint(act.Kind))
	}
	// すでになにか立っていて、問題がある
	yx := index(act.Y, act.X)
	if f.cell[yx] != EMPTY {
		if !(act.Kind == STATION && f.cell[yx] >= 1 && f.cell[yx] <= 6) {
			// 駅は線路の上に建てることができる
			//log.Println(f.cellString())
			log.Printf("try to build: typ:%d Y:%d X:%d but already built %d\n", act.Kind, act.Y, act.X, f.cell[yx])
			return fmt.Errorf("already built")
		}
		if isRail(act.Kind) && f.cell[yx] == STATION {
			panic("駅の上に線路を建てることはできません")
		}
		if isRail(act.Kind) && isRail(f.cell[yx]) && act.Kind != f.cell[yx] {
			panic("線路の上に線路を建てることはできません")
		}
		if act.Kind == f.cell[yx] {
			panic("同じ線路を建てることはできません")
		}
	}
	// 建設
	f.cell[yx] = act.Kind

	// 駅がカバーしている範囲を更新
	if act.Kind == STATION {
		//f.stations = append(f.stations, Pos{Y: act.Y, X: act.X})
		for d := 0; d < 13; d++ {
			y, x := act.Y+ddy[d], act.X+ddx[d]
			if y >= 0 && y < N && x >= 0 && x < N {
				f.coverd.Set(index(y, x))
			}
		}
	}
	return nil
}

// IsNearStation 駅,路線をつかって、a,bがつながっているかを返す
// a, b　はHOME, WORKSPACE
func (f Field) IsNearStation(a, b Pos) int {
	if f.coverd.Get(index(a.Y, a.X)) && f.coverd.Get(index(b.Y, b.X)) {
		return 2
	} else if f.coverd.Get(index(a.Y, a.X)) || f.coverd.Get(index(b.Y, b.X)) {
		return 1
	}
	return 0
}

// 2点間の最短経路を返す (a から b へ)
// bはEMPTYまたはSTATION
// fieldの線路または駅を通って移動するための関数
func (f *Field) FindNewPath(a, b Pos) (path []Pos) {
	// a から b への最短経路を返す
	// field=EMPTY なら移動可能 それ以外は移動不可
	var dist [N * N]int
	for i := 0; i < N*N; i++ {
		dist[i] = 10000
	}
	dist[int(a.Y)*50+int(a.X)] = 0
	var que []Pos
	que = append(que, a)
	for len(que) > 0 {
		p := que[0]
		que = que[1:]
		for d := 0; d < 4; d++ {
			y, x := p.Y+dy[d], p.X+dx[d]
			yx := index(y, x)
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if f.cell[yx] != EMPTY && f.cell[yx] != STATION {
				continue
			}
			if dist[int(y)*50+int(x)] > dist[int(p.Y)*50+int(p.X)]+1 {
				dist[int(y)*50+int(x)] = dist[int(p.Y)*50+int(p.X)] + 1
				que = append(que, Pos{Y: y, X: x})
			}
		}
	}
	if dist[int(b.Y)*50+int(b.X)] == 10000 {
		return nil
	}
	// b から a への経路を復元
	path = append(path, b)
	for path[len(path)-1] != a {
		p := path[len(path)-1]
		for d := 0; d < 4; d++ {
			y, x := p.Y+dy[d], p.X+dx[d]
			yx := index(y, x)
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if y == a.Y && x == a.X {
				path = append(path, Pos{Y: y, X: x})
				break
			}
			if f.cell[yx] != EMPTY && f.cell[yx] != STATION {
				continue
			}
			if dist[int(y)*50+int(x)] == dist[int(p.Y)*50+int(p.X)]-1 {
				path = append(path, Pos{Y: y, X: x})
				break
			}
		}
	}
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
}

// 2点間の最短経路を返す (a から b へ)
// paht[0]とpath[len(path)-1]は駅
// 駅間の路線の線路の種類を返す
func (f *Field) SelectRails(path []Pos) (types []int8) {
	if len(path) == 0 {
		return nil
	}
	types = make([]int8, len(path))
	types[0] = STATION
	types[len(path)-1] = STATION
	for i := 1; i < len(path)-1; i++ {
		y0, x0 := path[i-1].Y, path[i-1].X
		y1, x1 := path[i].Y, path[i].X
		y2, x2 := path[i+1].Y, path[i+1].X
		if y0 == y1 && y1 == y2 {
			types[i] = RAIL_HORIZONTAL
		} else if x0 == x1 && x1 == x2 {
			types[i] = RAIL_VERTICAL
		} else if (y0 < y1 && x1 < x2) || (y2 < y1 && x1 < x0) {
			types[i] = RAIL_RIGHT_UP
		} else if (y0 < y1 && x1 > x2) || (y2 < y1 && x1 > x0) {
			types[i] = RAIL_LEFT_UP
		} else if (y0 > y1 && x1 < x2) || (y2 > y1 && x1 < x0) {
			types[i] = RAIL_RIGHT_DOWN
		} else if (y0 > y1 && x1 > x2) || (y2 > y1 && x1 > x0) {
			types[i] = RAIL_LEFT_DOWN
		} else {
			panic("invalid path")
		}
	}
	return
}

// railが繋がる向きを返す,dy,dxに対応
// 0:UP, 1:RIGHT, 2:DOWN, 3:LEFT
func railDirection(rail int8) []int {
	switch rail {
	case RAIL_HORIZONTAL:
		return []int{1, 3}
	case RAIL_VERTICAL:
		return []int{0, 2}
	case RAIL_LEFT_DOWN:
		return []int{3, 2}
	case RAIL_LEFT_UP:
		return []int{0, 3}
	case RAIL_RIGHT_UP:
		return []int{0, 1}
	case RAIL_RIGHT_DOWN:
		return []int{1, 2}
	case STATION, EMPTY:
		return []int{0, 1, 2, 3}
	}
	return nil
}

// checkConnec はレールの接続ルールを判定する
func checkConnec(railType int8, direction int, isStart bool) bool {
	if railType == STATION {
		return true
	}
	switch direction {
	case UP:
		if isStart {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_UP || railType == RAIL_RIGHT_UP
		} else {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_DOWN || railType == RAIL_RIGHT_DOWN
		}
	case RIGHT:
		if isStart {
			return railType == RAIL_HORIZONTAL || railType == RAIL_RIGHT_DOWN || railType == RAIL_RIGHT_UP
		} else {
			return railType == RAIL_HORIZONTAL || railType == RAIL_LEFT_DOWN || railType == RAIL_LEFT_UP
		}
	case DOWN:
		if isStart {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_DOWN || railType == RAIL_RIGHT_DOWN
		} else {
			return railType == RAIL_VERTICAL || railType == RAIL_LEFT_UP || railType == RAIL_RIGHT_UP
		}
	case LEFT:
		if isStart {
			return railType == RAIL_HORIZONTAL || railType == RAIL_LEFT_DOWN || railType == RAIL_LEFT_UP
		} else {
			return railType == RAIL_HORIZONTAL || railType == RAIL_RIGHT_DOWN || railType == RAIL_RIGHT_UP
		}
	}
	return false
}

// canMove は、aからbに移動可能かを返す
// dist[2500]を参照してpathを返すときに、セルの種類もチェックする必要がある
// aからbの向きに移動できる && bがaから受けいることができるか
func (f Field) canMove(a, b Pos) bool {
	if distance(a, b) != 1 {
		log.Println("distance", a, b, distance(a, b))
		return false
	}
	if f.cell[index(b.Y, b.X)] == WALL {
		return false
	}
	// directionはaからbに移動する向き
	direction := UP
	switch {
	case a.Y == b.Y && a.X < b.X:
		direction = RIGHT
	case a.Y == b.Y:
		direction = LEFT
	case a.Y < b.Y:
		direction = DOWN
	}
	// aからbに移動する向きが繋がっているか
	x, y := a.X+dx[direction], a.Y+dy[direction]
	if !(x == b.X && y == b.Y) {
		log.Fatal("canMove: invalid direction")
	}
	// aのレールが繋がっていなかったらfalse
	if isRail(f.cell[index(a.Y, a.X)]) {
		if !checkConnec(f.cell[index(a.Y, a.X)], direction, true) {
			return false
		}
	}
	// bのレールが繋がっていなかったらfalse
	if isRail(f.cell[index(b.Y, b.X)]) {
		if !checkConnec(f.cell[index(b.Y, b.X)], direction, false) {
			return false
		}
	}
	return true
}

// CanBuildRail は、pathをたどって、線路を敷くことができるかを返す
// pathは駅から駅までの経路
// 線路の上に駅は建てることができる
// 線路の上に種類の違う線路を建てることはできない
// 最初と最後に駅があるPathを受け取る
// 既存のなにかと連結するか
func (f Field) CanBuildRail(path []Pos, typ []int8) (bool, bool) {
	connect := false
	for i := 0; i < len(path); i++ {
		y, x := path[i].Y, path[i].X
		// 完成形と線路の形が違うのはNG
		if isRail(f.cell[index(y, x)]) && isRail(typ[i]) {
			if f.cell[index(y, x)] != typ[i] {
				return false, false
			}
		}
		// どちらかが駅の時
		if typ[i] == STATION {
			if isRail(f.cell[index(y, x)]) || f.cell[index(y, x)] == STATION {
				connect = true
			}
		}
		if f.cell[index(y, x)] != EMPTY {
			connect = true
		}
	}
	return true, connect
}

// 駅a, bが繋がることができるか、できないときnil,できるときはすべてのpath
// すでに建築済みの路線もpathに含まれる
func (f Field) canConnect(a, b Pos) ([]Pos, error) {
	// distの更新
	var dist [2500]int
	for i := 0; i < 2500; i++ {
		dist[i] = 10000
	}
	dist[int(a.Y)*50+int(a.X)] = 0
	q := []Pos{a}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		if p == b {
			break
		}

		direction := railDirection(f.cell[p.Index()])

		if len(direction) == 0 {
			log.Println(f.cell[p.Index()])
			panic("invalid rail")
		}
		for _, d := range direction {
			y, x := p.Y+dy[d], p.X+dx[d]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			if f.cell[index(y, x)] == EMPTY || f.cell[index(y, x)] == STATION {
				if dist[int(y)*50+int(x)] > dist[int(p.Y)*50+int(p.X)]+1 {
					dist[int(y)*50+int(x)] = dist[int(p.Y)*50+int(p.X)] + 1
					q = append(q, Pos{Y: y, X: x})
				}
			}
			if isRail(f.cell[index(y, x)]) && checkConnec(f.cell[index(y, x)], int(d), false) {
				if dist[int(y)*50+int(x)] > dist[int(p.Y)*50+int(p.X)]+1 {
					dist[int(y)*50+int(x)] = dist[int(p.Y)*50+int(p.X)] + 1
					q = append(q, Pos{Y: y, X: x})
				}
			}
		}
	}
	// 繋がっていない
	if dist[int(b.Y)*50+int(b.X)] == 10000 {
		return nil, fmt.Errorf("can't reach")
	}
	// b から a への経路を復元
	path := []Pos{b}
MAKEPATH:
	for {
		p := path[len(path)-1]
		direction := [4]int{0, 1, 2, 3}
		// このシャッフルで多様性を持たせる
		randShuffle(4, func(i, j int) { direction[i], direction[j] = direction[j], direction[i] })
		for _, d := range direction {
			y, x := p.Y+dy[d], p.X+dx[d]
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				// 場外
				continue
			}
			if y == a.Y && x == a.X {
				// 駅に到達
				path = append(path, Pos{Y: y, X: x})
				break MAKEPATH
			}
			if dist[int(y)*50+int(x)] == dist[int(p.Y)*50+int(p.X)]-1 {
				if f.canMove(p, Pos{Y: y, X: x}) {
					path = append(path, Pos{Y: y, X: x})
					break
				}
			}
		}
	}
	return path, nil
}

var ErrNotEnoughMoney = fmt.Errorf("not enough money")

type State struct {
	field           *Field
	money           int
	turn            int
	income          int
	score           int // 最終ターンでの予想スコア
	potentialIncome int
	actions         []Action
	connected       []bool // in.Mが接続済みかどうか
}

func (s *State) Clone() *State {
	newActions := make([]Action, len(s.actions))
	copy(newActions, s.actions)
	newConnected := make([]bool, len(s.connected))
	copy(newConnected, s.connected)
	newState := &State{
		field:           s.field.Clone(),
		money:           s.money,
		turn:            s.turn,
		income:          s.income,
		score:           s.score,
		potentialIncome: s.potentialIncome,
		actions:         newActions,
		connected:       newConnected,
	}
	return newState
}

func NewState(in *Input) *State {
	s := new(State)
	s.field = NewField(in.N)
	s.money = in.K
	s.connected = make([]bool, in.M)
	return s
}

func (s *State) do(act Action, in Input, last bool) error {
	if s.money < buildCost[act.Kind] {
		return ErrNotEnoughMoney
	}
	if act.Kind != DO_NOTHING {
		err := s.field.Build(act)
		if err != nil {
			log.Println("acttype:", act.Kind, "pos:", act.Y, act.X)
			log.Println("build error", err)
			return err
		}
		s.money -= buildCost[act.Kind]
		// 駅が建築された時に、隣接するsrc,dstが追加されて収入が増える
		if act.Kind == STATION && last {
			for i := 0; i < in.M; i++ {
				if !s.connected[i] {
					ab := s.field.IsNearStation(in.src[i], in.dst[i])
					if ab == 2 {
						s.income += in.income[i]
						s.connected[i] = true
					} else if ab == 1 {
						s.potentialIncome += in.income[i]
					}
				}
			}
		}
	}
	s.turn++
	s.money += s.income
	s.score = s.money + s.income*(in.T-s.turn)
	//if !ATCODER {
	if act.comment == nil {
		act.comment = new(string)
	}
	*act.comment = fmt.Sprintf("#turn=%d, \n#money=%d, \n#income=%d\n #Score=%d\n",
		s.turn, s.money, s.income, s.score)
	//}
	if len(s.actions) > 0 {
		if s.actions[len(s.actions)-1].Kind == DO_NOTHING && act.Kind == DO_NOTHING {
			s.actions[len(s.actions)-1].Num++
			return nil
		}
	}
	s.actions = append(s.actions, act)
	return nil
}

type Pos struct {
	Y, X int8
}

func (p Pos) Index() int {
	return int(p.Y)*50 + int(p.X)
}

func index(y, x int8) int {
	return int(y)*50 + int(x)
}

func (p Pos) add(a Pos) Pos {
	return Pos{Y: p.Y + a.Y, X: p.X + a.X}
}

func (p Pos) Clone() Pos {
	return Pos{Y: p.Y, X: p.X}
}

func distance(a, b Pos) int {
	return absint(int(a.X)-int(b.X)) + absint(int(a.Y)-int(b.Y))
}

type Pair [2]Pos

// uniquePair は、p1, p2の順番を統一してのPairを返す
func uniquePair(p1, p2 Pos) Pair {
	if p1.Y < p2.Y || p1.Y == p2.Y && p1.X < p2.X {
		return Pair{p1, p2}
	}
	return Pair{p2, p1}
}

// stationの周辺
var ddy = [13]int8{0, -1, 0, 1, 0, -1, 1, 1, -1, -2, 0, 2, 0}
var ddx = [13]int8{0, 0, 1, 0, -1, 1, 1, -1, -1, 0, 2, 0, -2}

// すべての駅を繋ぐ鉄道を敷設する
// MSTクラスカル法を使っているが、簡易距離と制約によって、無駄なエッジが作られることがある
// ここのエッジは短いが、まれに駅を挟む
func constructMSTRailway(in Input, stations []Pos) ([]mstEdge, *Field) {
	numStations := int(len(stations))
	stationIndexMap := make(map[Pos]int)
	for i, s := range stations {
		stationIndexMap[s] = int(i)
	}
	// 決めておいた駅を建設する
	field := NewField(in.N)
	for i := int(0); i < numStations; i++ {
		err := field.Build(Action{Kind: STATION, Y: stations[i].Y, X: stations[i].X})
		if err != nil {
			log.Println("fatal build station:", stations[i])
			log.Println("station build error", err)
			panic(err)
		}
	}

	// マンハッタン距離を使って,全駅間の暫定距離を求める
	edges := []mstEdge{}
	for i := int(0); i < numStations; i++ {
		for j := i + 1; j < numStations; j++ {
			dist := distance(stations[i], stations[j])
			edges = append(edges, mstEdge{From: i, To: j, L1: dist})
		}
	}
	//////////////////////////
	// MSTを求める
	// コストが小さい順にソート
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].L1 < edges[j].L1
	})

	// UnionFindで連結成分を管理
	uf := NewUnionFind()
	// Kruskal法で最小全域木を求める
	mstEdges := []mstEdge{}
	for _, edge := range edges {
		// すでに連結されている場合はスキップ
		if uf.same(int(edge.From), int(edge.To)) {
			continue
		}
		// 連結可能か確認
		path, _ := field.canConnect(stations[edge.From], stations[edge.To])
		// ここではフィールドを使っているので、path=nil,でも素通り
		if path != nil {
			types := field.SelectRails(path)
			ok, _ := field.CanBuildRail(path, types)
			if !ok {
				continue
			}
			// すでに建築済みまたは駅がある場合はスキップ
			for i := 1; i < len(path)-1; i++ {
				if field.cell[path[i].Index()] == STATION || field.cell[path[i].Index()] == types[i] {
					// すでに駅があり線路が必要ない時 または、すでに建築予定の線路がある時
					continue
				}
				err := field.Build(Action{Kind: types[i], Y: path[i].Y, X: path[i].X})
				if err != nil {
					panic(err)
				}
			}
			uf.unite(int(edge.From), int(edge.To))
			edge.Path = path
			edge.Rail = types
			mstEdges = append(mstEdges, edge)
		}
	}
	log.Println(field.ToString())

	///////////////////////////////////
	// 全てに駅間を繋ぐエッジを作る
	// 使わない場所はOTHERにして、線路と駅をコピーする
	field2 := NewField(in.N) // mstEdgesで使われている場所だけを使う
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			field2.cell[i*50+j] = WALL
		}
	}
	for _, edge := range mstEdges {
		for _, p := range edge.Path {
			if isRail(field.cell[p.Index()]) || field.cell[p.Index()] == STATION {
				field2.cell[p.Index()] = EMPTY // WALLの強制解除
				_ = field2.Build(Action{Kind: field.cell[p.Index()], Y: p.Y, X: p.X})
				// 駅->線路の順番で建築するときエラーを吐くが無視できる
			}
		}
	}
	///////////////////////////////////
	// MST木の上で、すべての家と職場を繋ぐ駅のエッジを作る
	// src,dstの対応する駅を探して、その間を繋ぐエッジを作る
	unique := make(map[Pair]bool)
	for i := 0; i < in.M; i++ {
		src, dst := in.src[i], in.dst[i]
		statinsSrc := make([]Pos, 0)
		statinsDst := make([]Pos, 0)
		for _, s := range stations {
			if distance(s, src) <= 2 {
				statinsSrc = append(statinsSrc, s)
			}
			if distance(s, dst) <= 2 {
				statinsDst = append(statinsDst, s)
			}

		}
		if len(statinsSrc) == 0 || len(statinsDst) == 0 {
			// すべての家と職場は駅から距離２以内にある
			panic("no station")
		}
		for _, s0 := range statinsSrc {
			for _, s1 := range statinsDst {
				if unique[uniquePair(s0, s1)] {
					continue
				}
				if s0 == s1 {
					continue
				}
				path, err := field2.canConnect(s0, s1)
				if err != nil {
					panic(err)
				}
				unique[uniquePair(s0, s1)] = true
				types := field2.SelectRails(path)
				dist := distance(s0, s1)
				edge := mstEdge{From: stationIndexMap[s0], To: stationIndexMap[s1], Path: path, Rail: types, L1: dist}
				mstEdges = append(mstEdges, edge)
				//log.Println("append", s0, s1, dist)
			}
		}
	}
	log.Println("edgeNum", len(mstEdges), "extraEdgeNum", len(unique))
	return mstEdges, field
}

// ConstructGreedyRailway は、貪欲法で鉄道を構築する
// 盤面を４分割にして、左上の駅は右下の駅と繋ぐ,他も同様
func ConstructGreedyRailway(in Input, stations []Pos) ([]mstEdge, *Field) {
	f := NewField(in.N)
	for _, s := range stations {
		err := f.Build(Action{Kind: STATION, Y: s.Y, X: s.X})
		if err != nil {
			panic(err)
		}
	}
	log.Println(f.ToString())

	directions := map[string][]int8{
		"LT": {1, 2}, // 右, 下
		"RT": {2, 3}, // 下, 左
		"LB": {0, 1}, // 上, 右
		"RB": {0, 3}, // 上, 左
	}

	var q []Pos
NEXTSTATION:
	for _, s := range stations {
		q = make([]Pos, 0)
		var region string
		if s.Y < N/2 && s.X < N/2 {
			region = "LT"
		} else if s.Y < N/2 && s.X >= N/2 {
			region = "RT"
		} else if s.Y >= N/2 && s.X < N/2 {
			region = "LB"
		} else {
			region = "RB"
		}

		var used [50][50]bool
		used[s.Y][s.X] = true
		q = append(q, s)

		for len(q) > 0 {
			p := q[0]
			q = q[1:]
			for _, d := range directions[region] {
				next := Pos{Y: p.Y + dy[d], X: p.X + dx[d]}
				if next.Y < 0 || next.Y >= N || next.X < 0 || next.X >= N {
					continue
				}
				if used[next.Y][next.X] {
					continue
				}
				if f.cell[next.Index()] == EMPTY {
					used[next.Y][next.X] = true
					q = append(q, next)
				} else if f.cell[next.Index()] == STATION {
					path, err := f.canConnect(s, next)
					if err != nil {
						panic(err)
					}
					types := f.SelectRails(path)
					for k := 1; k < len(path)-1; k++ {
						if f.cell[path[k].Index()] == STATION {
							continue
						}
						if isRail(types[k]) && f.cell[path[k].Index()] != types[k] {
							err := f.Build(Action{Kind: types[k], Y: path[k].Y, X: path[k].X})
							if err != nil {
								log.Println("k", k)
								log.Println(types)
								panic(err)
							}
						}
					}
					continue NEXTSTATION
				}
			}
		}
	}
	return nil, f
}

func ChooseStationPositionFast(in Input) (poss []Pos) {
	poss = make([]Pos, 0, intMax(in.M, 100))
	sumPoints := in.M * 2
	var grid [2500]int
	for i := 0; i < in.M; i++ {
		grid[index(in.src[i].Y, in.src[i].X)]++
		grid[index(in.dst[i].Y, in.dst[i].X)]++
	}
	var coverd [2500]bool
	coverdPoints := 0

	for coverdPoints < sumPoints {
		bestPos := Pos{Y: 0, X: 0}
		bestHit := 0
		for i := int8(0); i < 50; i++ {
			for j := int8(0); j < 50; j++ {
				hit := 0
				if coverd[index(i, j)] {
					continue
				}
				for k := 0; k < 13; k++ {
					y, x := i+ddy[k], j+ddx[k]
					if y < 0 || y >= 50 || x < 0 || x >= 50 {
						continue
					}
					if coverd[index(y, x)] {
						continue
					}
					hit += grid[index(y, x)]
				}
				if hit > bestHit {
					bestHit = hit
					bestPos = Pos{Y: i, X: j}
				}
			}
		}
		if bestHit == 0 {
			panic("no station position")
		}
		poss = append(poss, bestPos)
		coverdPoints += bestHit
		for k := 0; k < 13; k++ {
			y, x := int(bestPos.Y)+int(ddy[k]), int(bestPos.X)+int(ddx[k])
			if y < 0 || y >= 50 || x < 0 || x >= 50 {
				continue
			}
			coverd[y*50+x] = true
		}
	}
	sort.Slice(poss, func(i, j int) bool {
		if poss[i].Y == poss[j].Y {
			return poss[i].X < poss[j].X
		}
		return poss[i].Y < poss[j].Y
	})
	return poss
}

type bsState struct {
	state       State
	restActions []uint
}

func newBsState(in *Input, numActions int) *bsState {
	new := &bsState{
		state:       *NewState(in),
		restActions: make([]uint, numActions),
	}
	for i := 0; i < numActions; i++ {
		new.restActions[i] = uint(i)
	}
	return new
}

func (s *bsState) Clone() *bsState {
	clonedState := &bsState{
		state:       *s.state.Clone(),
		restActions: make([]uint, len(s.restActions)),
	}
	copy(clonedState.restActions, s.restActions)
	return clonedState
}

type Input struct {
	N      int   // 縦長 N=50
	M      int   // 人数 50<=M<=1600
	K      int   // 初期資金 11000<=K<=20000
	T      int   // ターン数 T=800
	src    []Pos // 人の初期位置
	dst    []Pos // 人の目的地
	income []int // 人の収入
}

func readInput(re *bufio.Reader) *Input {
	var in Input
	fmt.Fscan(re, &in.N, &in.M, &in.K, &in.T)
	src := make([]Pos, in.M)
	dst := make([]Pos, in.M)
	income := make([]int, in.M)
	for i := 0; i < in.M; i++ {
		fmt.Fscan(re, &src[i].Y, &src[i].X, &dst[i].Y, &dst[i].X)
		income[i] = int(distance(src[i], dst[i]))
	}
	log.Printf("readInput: N=%v, M=%v, K=%v, T=%v\n", in.N, in.M, in.K, in.T)
	in.src = src
	in.dst = dst
	in.income = income
	return &in
}

func absint(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type UnionFind struct {
	par [2500]int
}

func (uf *UnionFind) Clone() *UnionFind {
	if uf == nil {
		return nil
	}
	newUF := new(UnionFind)
	newUF.par = uf.par
	return newUF
}

// NewUnionFind は、UnionFindを初期化して返す
// 2500固定
func NewUnionFind() *UnionFind {
	uf := new(UnionFind)
	for i := int(0); i < 2500; i++ {
		uf.par[i] = i
	}
	return uf
}

func (uf *UnionFind) root(a int) int {
	if uf.par[a] == a {
		return a
	}
	uf.par[a] = uf.root(uf.par[a])
	return uf.par[a]
}

func (uf *UnionFind) same(a, b int) bool {
	return uf.root(int(a)) == uf.root(int(b))
}

func (uf *UnionFind) unite(a, b int) {
	a = uf.root(a)
	b = uf.root(b)
	if a == b {
		return
	}
	uf.par[a] = b
}

func gridToString(grid [2500]int) (str string) {
	str = "showGrid()\n"
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			if grid[i*50+j] == 10000 {
				str += "## "
			} else {
				str += fmt.Sprintf("%2d ", grid[i*50+j])
			}
		}
		str += "\n"
	}
	return str
}

// MST用
type mstEdge struct {
	From, To int
	L1       int // マンハッタン距離
	Path     []Pos
	Rail     []int8
}

type Graph struct {
	NumNodes int
	Edges    []mstEdge
}

// BitSet は 2500個の bool を uint64 で管理する構造体
type BitSet struct {
	bits [40]uint64 // 2500 / 64 = 40
}

// Set 指定したインデックスのビットを 1 (true) にする
func (b *BitSet) Set(index int) {
	if index < 0 || index >= 2500 {
		panic("index out of range")
	}
	b.bits[index/64] |= (1 << (index % 64))
}

// Clear 指定したインデックスのビットを 0 (false) にする
func (b *BitSet) Clear(index int) {
	if index < 0 || index >= 2500 {
		panic("index out of range")
	}
	b.bits[index/64] &^= (1 << (index % 64))
}

// Get 指定したインデックスのビットの状態を取得する
func (b *BitSet) Get(index int) bool {
	if index < 0 || index >= 2500 {
		panic("index out of range")
	}
	return (b.bits[index/64] & (1 << (index % 64))) != 0
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func randShuffle(n int, swap func(i int, j int)) {
	for i := 0; i < n; i++ {
		j := frand.Intn(n)
		swap(i, j)
	}
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// chokudai search
// 状態を表す構造体
// 優先度付きキューの実装
type PriorityQueue2 []*bsState

func (pq PriorityQueue2) Len() int { return len(pq) }

func (pq PriorityQueue2) Less(i, j int) bool {
	return pq[i].state.score > pq[j].state.score
}

func (pq PriorityQueue2) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue2) Push(x interface{}) {
	*pq = append(*pq, x.(*bsState))
}

func (pq *PriorityQueue2) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}

// //////////////////////////////////
// Dijkstra library
type DijEdge struct {
	To, Cost int
}

type Item struct {
	cost, nodeCount, node int
}

type PriorityQueue []Item

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].cost == pq[j].cost {
		return pq[i].nodeCount > pq[j].nodeCount
	}
	return pq[i].cost < pq[j].cost
}
func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(Item)
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// return path, cost
func dijkstra(graph [][]DijEdge, start, goal int) ([]int, int) {
	n := len(graph)
	dist := make([]int, n)
	nodeCount := make([]int, n)
	prev := make([]int, n)
	for i := range dist {
		dist[i] = 10000
		nodeCount[i] = 0
		prev[i] = -1
	}
	dist[start] = 0
	nodeCount[start] = 1
	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, Item{cost: 0, nodeCount: 1, node: start})
	for pq.Len() > 0 {
		current := heap.Pop(pq).(Item)
		cost, count, node := current.cost, current.nodeCount, current.node
		if node == goal {
			break
		}
		for _, edge := range graph[node] {
			nowCost := cost + edge.Cost
			nowCount := count + 1
			if nowCost < dist[edge.To] || nowCost == dist[edge.To] && nowCount > nodeCount[edge.To] {
				dist[edge.To] = nowCost
				nodeCount[edge.To] = nowCount
				prev[edge.To] = node
				heap.Push(pq, Item{cost: nowCost, nodeCount: nowCount, node: edge.To})
			}
		}
	}
	// 経路復元
	path := []int{}
	for node := goal; node != -1; node = prev[node] {
		path = append(path, node)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	if len(path) == 0 || path[0] != start {
		log.Println("cannot reach goal", start, "to", goal)
		return nil, 10000000
	}
	return path, dist[goal]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
