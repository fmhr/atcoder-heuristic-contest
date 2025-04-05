package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
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
	in := parseInput()
	log.Printf("M=%d L=%d W=%d\n", in.M, in.L, in.W)
	solve(in)
	log.Printf("elapsed=%.2f\n", float64(time.Since(startTime).Microseconds())/1000)
}

type Point struct {
	Y, X float64
}

// 大小関係がわかればいいので、√を取らない
func distSquared(a, b Point) float64 {
	return (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y)
}

// 小数点以下切り捨て
func distance(a, b Point) int {
	return int(math.Floor(math.Sqrt(float64(distSquared(a, b)))))
}

type City struct {
	ID int
	Point
}

// answerの出力用
type AnsGroup struct {
	Cities []int
	Edges  [][2]int
	Cost   int
}

func (a AnsGroup) calcScore(cities []City) int {
	// エッジの長さの合計
	score := 0
	for _, edge := range a.Edges {
		score += distance(cities[edge[0]].Point, cities[edge[1]].Point)
	}
	return score
}

// Output()は、回答形式に合わせてStringに変換する
func (a AnsGroup) Output() (str string) {
	for i, city := range a.Cities {
		if i < len(a.Cities)-1 {
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

type CityState struct {
	ID          int
	Mean        [2]float64 // x, y
	Variance    [2]float64
	Ract        [4]float64
	SumVariance float64 // 分散の合計
}

func (cs CityState) toCity() City {
	return City{ID: cs.ID, Point: Point{Y: cs.Mean[1], X: cs.Mean[0]}}
}

// 分散の合計（算術平均）を更新する
func (cs *CityState) UpdateSumVariance() {
	cs.SumVariance = (cs.Variance[0] + cs.Variance[1]) / 2
}

// クエリをつかって、都市の座標を推定する
// usableQは使えるクエリの数
func estimatePhase(in Input, usableQ int) {
	cityStates := make([]CityState, in.N)
	for i := 0; i < in.N; i++ {
		cityStates[i].ID = i

		// 矩形の座標を取得
		rx := float64(in.lxrxlyry[i*4+1]) // 右
		lx := float64(in.lxrxlyry[i*4+0]) // 左
		ry := float64(in.lxrxlyry[i*4+3]) // 上
		ly := float64(in.lxrxlyry[i*4+2]) // 下

		// 平均値は矩形の中心
		cityStates[i].Mean[0] = (rx + lx) / 2
		cityStates[i].Mean[1] = (ry + ly) / 2

		// 分散は一様分布の理論値 (b-a)²/12 を使用
		cityStates[i].Variance[0] = math.Pow(rx-lx, 2) / 12.0
		cityStates[i].Variance[1] = math.Pow(ry-ly, 2) / 12.0

		// 最小分散を設定（数値安定性のため）
		minVariance := 1.0
		cityStates[i].Variance[0] = math.Max(minVariance, cityStates[i].Variance[0])
		cityStates[i].Variance[1] = math.Max(minVariance, cityStates[i].Variance[1])

		// 矩形情報も保存
		cityStates[i].Ract[0] = lx
		cityStates[i].Ract[1] = rx
		cityStates[i].Ract[2] = ly
		cityStates[i].Ract[3] = ry

		// 分散の合計（算術平均）を計算
		cityStates[i].UpdateSumVariance()
	}

	// 既に中心都市として使用したIDを記録する配列
	usedAsCenters := make([]bool, in.N)

	// クエリを使って位置推定を更新していく処理
	for q := 0; q < usableQ; q++ {
		// SumVarianceが大きい都市をソートしたスライスを作成
		sortedCities := make([]CityState, 0, in.N)

		// まだ中心として使用していない都市だけをソート対象にする
		for i := 0; i < in.N; i++ {
			if !usedAsCenters[i] {
				sortedCities = append(sortedCities, cityStates[i])
			}
		}

		// 分散の大きい順（不確かさが大きい順）にソート
		sort.Slice(sortedCities, func(i, j int) bool {
			return sortedCities[i].SumVariance > sortedCities[j].SumVariance
		})

		// 最も分散の大きい都市を中心として選択
		centerCity := sortedCities[0]
		//log.Printf("Query %d: center city ID=%d, SumVariance=%.2f\n", q, centerCity.ID, centerCity.SumVariance)

		// この都市を中心として使用済みとマーク
		usedAsCenters[centerCity.ID] = true

		// centerCityからの距離が近い順に他の都市をソート
		cityDistances := make([]struct {
			CityState
			Distance float64
		}, in.N)

		for i, city := range cityStates {
			dx := centerCity.Mean[0] - city.Mean[0]
			dy := centerCity.Mean[1] - city.Mean[1]
			distance := math.Sqrt(dx*dx + dy*dy)
			cityDistances[i] = struct {
				CityState
				Distance float64
			}{city, distance}
		}

		// 距離が近い順にソート
		sort.Slice(cityDistances, func(i, j int) bool {
			return cityDistances[i].Distance < cityDistances[j].Distance
		})

		// クエリに使用する都市リストを作成（最大 in.L 個）
		querySize := intMin(in.L, in.N)
		queryCities := make([]int, querySize)
		queryCities[0] = centerCity.ID // 最初は中心都市

		// 残りの都市を距離が近い順に追加
		cityIndex := 1
		for i := 0; i < len(cityDistances) && cityIndex < querySize; i++ {
			cityID := cityDistances[i].ID
			if cityID != centerCity.ID { // 中心都市は既に追加済み
				queryCities[cityIndex] = cityID
				cityIndex++
			}
		}
		//log.Println("Query cities:", queryCities)

		tmpCities := make([]City, len(queryCities))
		for i, cityID := range queryCities {
			tmpCities[i] = cityStates[cityID].toCity()
		}
		estimateResult := createMST(tmpCities)
		//log.Println("estimateResult:", estimateResult)

		// クエリを実行
		queryResult := sendQuery(queryCities)
		//log.Println("Query result:", queryResult)

		// エッジの比較分析（ソート済みリストを利用した効率的な方法）
		common, onlyEstimated, onlyActual := compareEdges(estimateResult, queryResult)
		_ = common
		_ = onlyEstimated
		_ = onlyActual
		//log.Printf("Common edges: %v (%d)", common, len(common))
		//log.Printf("Only estimated: %v (%d)", onlyEstimated, len(onlyEstimated))
		//log.Printf("Only actual: %v (%d)", onlyActual, len(onlyActual))

		// 一致率の計算
		//matchRate := float64(len(common)) / float64(len(estimateResult)) * 100.0
		//log.Printf("Match rate: %.2f%%", matchRate)

		// クエリ結果を使って位置推定を更新
		updateCityPositions(cityStates, queryCities, queryResult)
	}

	// 最終的な推定精度を評価
	if os.Getenv("ATCODER") != "1" {
		sumSqErr := 0.0
		for i := 0; i < in.N; i++ {
			estX := cityStates[i].Mean[0]
			estY := cityStates[i].Mean[1]
			realX := in.trueXY[i][0]
			realY := in.trueXY[i][1]

			// X座標の二乗誤差
			dx := estX - realX
			sumSqErr += dx * dx

			// Y座標の二乗誤差
			dy := estY - realY
			sumSqErr += dy * dy
			if math.IsNaN(sumSqErr) {
				log.Printf("%+v\n", cityStates[i])
				panic("NaN detected")
			}
		}
		log.Println("Sum of squared errors:", sumSqErr)
		// 平均二乗誤差（Mean Squared Error）
		mse := sumSqErr / float64(2*in.N) // 都市数×座標2次元分で割る

		// 平均二乗誤差の平方根（Root Mean Squared Error）
		rmse := math.Sqrt(mse)

		log.Printf("Estimate Phase Final RMSE: %f", rmse)

		// 各クエリ後のRMSE推移を表示するための変数を追加してもよい
	}
}

// クエリ結果からベイズ更新を行う関数
func updateCityPositions(cityStates []CityState, queriedCities []int, edges [][2]int) {
	// 現在の推定位置に基づくシティのマップを作成
	tmpCities := make([]City, len(queriedCities))
	for i, cityID := range queriedCities {
		tmpCities[i] = cityStates[cityID].toCity()
	}
	estimateResult := createMST(tmpCities)

	// エッジ比較
	common, onlyEstimated, onlyActual := compareEdges(estimateResult, edges)
	_ = common

	// 推定結果にのみ存在するエッジに対して調整を行う (10%伸ばす)
	for _, edge := range onlyEstimated {
		// エッジの端点となる都市の状態を取得
		city1 := cityStates[edge[0]]
		city2 := cityStates[edge[1]]

		// エッジの現在の長さを計算
		p1 := Point{X: city1.Mean[0], Y: city1.Mean[1]}
		p2 := Point{X: city2.Mean[0], Y: city2.Mean[1]}
		currentDistSq := distSquared(p1, p2)
		currentDist := math.Sqrt(currentDistSq)

		// 調整後の長さ（10%伸ばす）
		targetDist := currentDist * 1.10

		// エッジの方向ベクトル
		dx := p2.X - p1.X
		dy := p2.Y - p1.Y

		// 単位ベクトル
		factor := 1.0
		if currentDist > 0 {
			factor = 1.0 / currentDist
		}
		ux := dx * factor
		uy := dy * factor

		// エッジを伸ばす量（片側5%ずつ）
		stretchAmount := (targetDist - currentDist) / 2.0

		// 都市1を逆方向に動かす（分散に応じて重み付け）
		totalVariance1 := city1.Variance[0] + city1.Variance[1]
		totalVariance2 := city2.Variance[0] + city2.Variance[1]
		totalVariance := totalVariance1 + totalVariance2

		// 分散が大きい方が多く動くよう重み付け
		weight1 := totalVariance1 / totalVariance
		weight2 := totalVariance2 / totalVariance

		// 都市1の移動
		moveX1 := -ux * stretchAmount * weight1
		moveY1 := -uy * stretchAmount * weight1

		// 都市2の移動
		moveX2 := ux * stretchAmount * weight2
		moveY2 := uy * stretchAmount * weight2

		// 都市1の位置と分散の更新
		updateCityPosition(&cityStates[edge[0]], moveX1, moveY1)

		// 都市2の位置と分散の更新
		updateCityPosition(&cityStates[edge[1]], moveX2, moveY2)
	}

	// 実際のエッジに合わせて、存在しない推定エッジは縮める処理
	for _, edge := range onlyActual {
		// エッジの端点となる都市の状態を取得
		city1 := cityStates[edge[0]]
		city2 := cityStates[edge[1]]

		// エッジの現在の長さを計算
		p1 := Point{X: city1.Mean[0], Y: city1.Mean[1]}
		p2 := Point{X: city2.Mean[0], Y: city2.Mean[1]}
		currentDistSq := distSquared(p1, p2)
		currentDist := math.Sqrt(currentDistSq)

		// 調整後の長さ（10%縮める）
		targetDist := currentDist * 0.90

		// エッジの方向ベクトル
		dx := p2.X - p1.X
		dy := p2.Y - p1.Y

		// 単位ベクトル
		factor := 1.0
		if currentDist > 0 {
			factor = 1.0 / currentDist
		}
		ux := dx * factor
		uy := dy * factor

		// エッジを縮める量（片側5%ずつ）
		shrinkAmount := (currentDist - targetDist) / 2.0

		// 分散に応じて重み付け
		totalVariance1 := city1.Variance[0] + city1.Variance[1]
		totalVariance2 := city2.Variance[0] + city2.Variance[1]
		totalVariance := totalVariance1 + totalVariance2

		// 分散が大きい方が多く動くよう重み付け
		weight1 := totalVariance1 / totalVariance
		weight2 := totalVariance2 / totalVariance

		// 都市1の移動（都市2の方に近づける）
		moveX1 := ux * shrinkAmount * weight1
		moveY1 := uy * shrinkAmount * weight1

		// 都市2の移動（都市1の方に近づける）
		moveX2 := -ux * shrinkAmount * weight2
		moveY2 := -uy * shrinkAmount * weight2

		// 都市1の位置と分散の更新
		updateCityPosition(&cityStates[edge[0]], moveX1, moveY1)

		// 都市2の位置と分散の更新
		updateCityPosition(&cityStates[edge[1]], moveX2, moveY2)
	}
}

// 都市の位置を移動し、分散も更新する関数
func updateCityPosition(city *CityState, moveX, moveY float64) {
	// NaNチェック（入力）
	if math.IsNaN(moveX) || math.IsNaN(moveY) {
		log.Printf("Warning: NaN input detected for city %d: moveX=%.6f, moveY=%.6f",
			city.ID, moveX, moveY)
		return
	}

	// 位置の更新
	city.Mean[0] += moveX
	city.Mean[1] += moveY

	// NaNチェック（更新後）
	if math.IsNaN(city.Mean[0]) || math.IsNaN(city.Mean[1]) {
		log.Printf("Warning: NaN detected in city %d after update", city.ID)
		// 更新前の値に戻す
		city.Mean[0] -= moveX
		city.Mean[1] -= moveY
		return
	}

	// 矩形の範囲内に収める
	city.Mean[0] = math.Max(city.Ract[0], math.Min(city.Ract[1], city.Mean[0]))
	city.Mean[1] = math.Max(city.Ract[2], math.Min(city.Ract[3], city.Mean[1]))

	// 分散を更新（0.99をかける）
	minVariance := 0.1 // 最小分散（あまりに小さくしないために）
	city.Variance[0] = math.Max(minVariance, city.Variance[0]*0.99)
	city.Variance[1] = math.Max(minVariance, city.Variance[1]*0.99)

	// 分散の合計を更新
	city.UpdateSumVariance()
}

// クエリ結果と推定結果のエッジ比較関数（ソート済み配列前提の効率的実装）
func compareEdges(estimateResult, queryResult [][2]int) (common, onlyEstimated, onlyActual [][2]int) {
	i, j := 0, 0

	// 両方のエッジリストをソート順に比較
	for i < len(estimateResult) && j < len(queryResult) {
		// エッジの比較
		compResult := compareEdge(estimateResult[i], queryResult[j])

		if compResult == 0 {
			// 両方に存在
			common = append(common, estimateResult[i])
			i++
			j++
		} else if compResult < 0 {
			// 推定結果にのみ存在
			onlyEstimated = append(onlyEstimated, estimateResult[i])
			i++
		} else {
			// クエリ結果にのみ存在
			onlyActual = append(onlyActual, queryResult[j])
			j++
		}
	}

	// 残りの推定結果エッジを処理
	for i < len(estimateResult) {
		onlyEstimated = append(onlyEstimated, estimateResult[i])
		i++
	}

	// 残りのクエリ結果エッジを処理
	for j < len(queryResult) {
		onlyActual = append(onlyActual, queryResult[j])
		j++
	}

	return
}

// エッジの比較関数（ソート順に基づく）
func compareEdge(a, b [2]int) int {
	if a[0] != b[0] {
		return a[0] - b[0]
	}
	return a[1] - b[1]
}

func solve(in Input) {
	estimatePhase(in, in.Q-100)

	sortedGroup := make([]int, in.M)
	for i := 0; i < in.M; i++ {
		sortedGroup[i] = in.G[i]
	}

	bestAns := make([]AnsGroup, in.M)
	bestMapping := make([]int, in.M)
	bestScore := 100000000
	ansGroups := make([]AnsGroup, in.M)
	mapping := make([]int, in.M)
	var cities [N]City
	loop := 20
	if in.M == 1 {
		loop = 1
	}
	for k := 0; k < loop; k++ {
		//sort.Sort(sort.Reverse(sort.IntSlice(sortedGroup)))
		frand.Shuffle(len(sortedGroup), func(i, j int) {
			sortedGroup[i], sortedGroup[j] = sortedGroup[j], sortedGroup[i]
		})
		mapping = makeMapping(in.G[:in.M], sortedGroup)
		// 都市の初期値は範囲の中心とする
		for i := 0; i < N; i++ {
			cities[i].ID = i
			cities[i].Y = float64((in.lxrxlyry[i*4+2] + in.lxrxlyry[i*4+3])) / 2
			cities[i].X = float64((in.lxrxlyry[i*4+0] + in.lxrxlyry[i*4+1])) / 2
		}
		center := Point{Y: 10000 / 2, X: 10000 / 2}
		var used [N]bool
		citiesSortedByCenter := make([]City, N)
		copy(citiesSortedByCenter, cities[:])
		sort.Slice(citiesSortedByCenter[:], func(i, j int) bool {
			return distSquared(center, citiesSortedByCenter[i].Point) > distSquared(center, citiesSortedByCenter[j].Point)
		})
		tmp := make([]City, N)
		copy(tmp, cities[:])
		ansGroups = make([]AnsGroup, in.M)
		for i := 0; i < in.M; i++ {
			// グループのrootを決める
			groupRoot := -1
			for _, city := range citiesSortedByCenter {
				if !used[city.ID] {
					groupRoot = city.ID
					used[city.ID] = true
					ansGroups[i].Cities = append(ansGroups[i].Cities, city.ID)
					break
				}
			}
			// rootからの距離が近い順にソートする
			sort.Slice(tmp[:], func(i, j int) bool {
				return distSquared(cities[groupRoot].Point, tmp[i].Point) < distSquared(cities[groupRoot].Point, tmp[j].Point)
			})
			// グループに都市を追加する
			// Edgesは、グループのrootと都市を結ぶエッジ
			for _, city := range tmp {
				if len(ansGroups[i].Cities) >= sortedGroup[i] {
					break
				}
				if !used[city.ID] {
					ansGroups[i].Cities = append(ansGroups[i].Cities, city.ID)
					used[city.ID] = true
				}
			}
			//log.Println("i:", i, "groupRoot:", groupRoot, "requre:", sortedGroup[i], "cities:", len(ansGrops[i].Citys))
			tmpCity := make([]City, len(ansGroups[i].Cities))
			for j, city := range ansGroups[i].Cities {
				tmpCity[j] = cities[city]
			}
			ansGroups[i].Edges = createMST(tmpCity)
		}
		// 推定座標でcostの計算
		allCost := 0
		for i := 0; i < in.M; i++ {
			for j := 0; j < len(ansGroups[i].Edges); j++ {
				allCost += distance(cities[ansGroups[i].Edges[j][0]].Point, cities[ansGroups[i].Edges[j][1]].Point)
			}
		}
		//log.Printf("estCost=%d\n", allCost)
		if allCost < bestScore {
			bestScore = allCost
			bestAns = make([]AnsGroup, in.M)
			for i := 0; i < in.M; i++ {
				bestAns[i].Cities = make([]int, len(ansGroups[i].Cities))
				bestAns[i].Edges = make([][2]int, len(ansGroups[i].Edges))
				bestAns[i].Cost = ansGroups[i].Cost
				copy(bestAns[i].Cities, ansGroups[i].Cities)
				copy(bestAns[i].Edges, ansGroups[i].Edges)
			}
			copy(bestMapping, mapping[:])
		}
	}
	// queryを使ったedgeの最適化
	log.Println("bestScore=", bestScore)
	for i := 0; i < in.M; i++ {
		if len(bestAns[i].Cities) > 2 && in.L >= len(bestAns[i].Cities) {
			bestAns[i].Edges = sendQuery(bestAns[i].Cities)
		}
	}

	//log.Println("mapping=", mapping)
	//log.Println("ansGrops=", ansGrops)
	// クエリの終了
	fmt.Println("!")
	for i := 0; i < in.M; i++ {
		fmt.Print(bestAns[bestMapping[i]].Output())
	}
	// kd-treeを作成する
	//kdt := NewKDTree(cities[:])
	//printTree(kdt.Root, 0)

	// 全都市間の距離を計算する
	// Groupの都市数が多い順に、都市間が短いエッジを結ぶ
	// Kruskal法に近いけど、エッジはGrpupと新しい都市を結ぶ時のみつなぐ
	//	var allDistance [N][N]int
	//allEdge := make([]Edge, 0, N*(N-1)/2)

	// -------------------------------------------------------------------
	if os.Getenv("ATCODER") != "1" {
		sumSqErr := 0.0
		for i := 0; i < N; i++ {
			estX := float64(cities[i].X)
			estY := float64(cities[i].Y)
			realX := in.trueXY[i][0]
			realY := in.trueXY[i][1]
			sumSqErr += (estX - realX) * (estX - realX)
			sumSqErr += (estY - realY) * (estY - realY)
		}
		mse := sumSqErr / float64(N) // 平均二乗誤差
		rmse := math.Sqrt(mse)       // 平均二乗誤差の平方根
		// 初期のRMSEは W/4.24　程度
		log.Printf("RMSE=%.2f\n", rmse)
	}
	log.Printf("queryCount=%d\n", queryCount)
}

const (
	N = 800 // 都市の個数
	Q = 400 // クエリの個数
)

type Input struct {
	N        int
	M        int // 都市のグループの数 1<= M <= 400
	Q        int
	L        int           // クエリの都市の最大数 1<= L <= 15
	W        int           //　二次元座標の最大値 500 <= W <= 2500
	G        [400]int      // 各グループの都市の数 1<= G[i] <= N(800) i= 0..M-1
	lxrxlyry [N * 4]int    // 各都市の座標 0 <= lxrxlyry[i] <= W
	trueXY   [N][2]float64 // 実際の座標
}

// 固定入力はとばす
func parseInput() (in Input) {
	fmt.Scan(&in.N, &in.M, &in.Q, &in.L, &in.W)
	for i := 0; i < in.M; i++ {
		fmt.Scan(&in.G[i])
	}
	for i := 0; i < N*4; i++ {
		fmt.Scan(&in.lxrxlyry[i])
	}
	if os.Getenv("ATCODER") != "1" {
		for i := 0; i < N; i++ {
			fmt.Scan(&in.trueXY[i][0], &in.trueXY[i][1])
		}
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
	Weight   float64
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
func runKruskal(n int, edges Edges) (float64, []Edge) {
	uf := NewUnionFind(n)
	sort.Sort(edges)
	var mst []Edge
	mstWeight := 0.0
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
			weight := distSquared(cities[i].Point, cities[j].Point)
			edges = append(edges, Edge{From: i, To: j, Weight: weight})
		}
	}
	cost, mst := runKruskal(len(cities), edges)
	_ = cost
	//log.Printf("cost=%d\n", cost)
	newEdge := make([][2]int, len(mst))
	for i := 0; i < len(mst); i++ {
		from := newIndex[mst[i].From]
		to := newIndex[mst[i].To]

		// 小さい方のIDを先に配置する（クエリと同じ順序に）
		if from > to {
			from, to = to, from
		}

		newEdge[i][0] = from
		newEdge[i][1] = to
	}
	sort.Slice(newEdge, func(i, j int) bool {
		if newEdge[i][0] == newEdge[j][0] {
			return newEdge[i][1] < newEdge[j][1]
		}
		return newEdge[i][0] < newEdge[j][0]
	})
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
	fmt.Printf("%s(%f, %f)\n", fmt.Sprint(' '+depth*2), node.X, node.Y)
	printTree(node.Left, depth+1)
	printTree(node.Right, depth+1)
}

// query
var queryCount int

func sendQuery(cities []int) (edges [][2]int) {
	if queryCount >= Q {
		panic("query count over")
	}
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
	queryCount++
	return edges
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
