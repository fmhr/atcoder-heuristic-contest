package main

import (
	"flag"
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

var (
	pram1 *float64 // 伸ばすときの比率
	pram2 *float64 // 縮めるときの比率
	pram3 *float64 // エッジが一致したときの分散更新係数
	pram4 *float64 // 分散更新係数
)

func init() {
	for i := 0; i < 800; i++ {
		hashTable[i] = rand.Uint64()
	}
	pram1 = flag.Float64("pram1", 1.194, "伸ばすときの比率")
	pram2 = flag.Float64("pram2", 0.724, "縮めるときの比率")
	pram3 = flag.Float64("pram3", 0.976, "エッジが一致したときの分散更新係数")
	pram4 = flag.Float64("pram4", 0.993, "分散更新係数")

	flag.Parse()

	log.Println("pram1=", *pram1)
	log.Println("pram2=", *pram2)
	log.Println("pram3=", *pram3)
	log.Println("pram4=", *pram4)
}

func main() {
	startTime = time.Now()
	in := parseInput()
	log.Printf("M=%d L=%d W=%d\n", in.M, in.L, in.W)
	// クエリで一度に使える頂点の数は,L L*Q/Nが各頂点で使える数
	// Q = 400 N = 800 なので L/2
	// ただし１度のクエリで、複数の辺で使われる頂点はその分多く更新される
	//log.Printf("NUM=%f\n", float64(in.L*in.Q)/float64(800))
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

func distanceSquared(a, b [2]float64) float64 {
	return (a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1])
}

// answerの出力用
type AnsGroup struct {
	Cities []int
	Edges  [][2]int
	Cost   int
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
	ID int
	//Center      [2]float64 // x, y
	Mean        [2]float64 // x, y
	Variance    [2]float64
	Ract        [4]float64
	SumVariance float64 // 分散の合計
	UpdateCount int     // 更新回数
}

// 分散の合計（算術平均）を更新する
func (cs *CityState) UpdateSumVariance() {
	cs.SumVariance = (cs.Variance[0] + cs.Variance[1]) / 2
}

type MSTResult struct {
	Vertex []int
	Edges  [][2]int
	V2     BitSet
	Hash   uint64
}

func NewMSTResult(v []int, e [][2]int) MSTResult {
	var r MSTResult
	r.Vertex = make([]int, len(v))
	copy(r.Vertex, v)
	r.Edges = make([][2]int, len(e))
	copy(r.Edges, e)
	r.V2 = *NewBitSet(800)
	for _, v := range v {
		r.V2.Set(v)
	}
	r.Hash = makeHash(v)
	return r
}

// クエリをつかって、都市の座標を推定する
// usableQは使えるクエリの数
func estimatePhase(in Input, usableQ int, cityStates []CityState) {
	// 既に中心都市として使用したIDを記録する配列
	usedAsCenters := make([]bool, in.N)

	// クエリを使って位置推定を更新していく処理
	for usableQ > 0 {
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

		hash := makeHash(queryCities)
		_, ok := queryResultMap[hash]
		if ok {
			// 重複したクエリはスキップ
			//log.Println("Duplicate query detected, skipping.", queryCities, hit.Vertex)
			continue
		}

		// 推定座標でMSTを作成
		estimateResult := createMST(queryCities, cityStates)

		// クエリを実行
		queryResult := sendQuery(queryCities)
		usableQ--
		//log.Println("Query result:", queryResult)

		// クエリ結果をMSTに変換
		mst := NewMSTResult(queryCities, queryResult)
		dup, ok := queryResultMap[mst.Hash]
		if !ok {
			queryResultMap[mst.Hash] = mst
		} else {
			log.Println("Duplicate query result detected", dup.Vertex)
			panic("Duplicate query result detected")
		}

		// クエリ結果を使って位置推定を更新
		updateCityPositions(cityStates, queryCities, queryResult, estimateResult)
		//if os.Getenv("ATCODER") != "1" {
		//rmse := calcRMSE(cityStates, in.trueXY)
		//log.Printf("%d RMSE: %f", usableQ, rmse)
		//}
	}

	// クエリの履歴を使って、都市の位置を更新
	for i := 0; i < 50; i++ {
		for _, mst := range queryResultMap {
			estiEdges := createMST(mst.Vertex, cityStates)
			//log.Println("estiEdges:", estiEdges)
			updateCityPositions(cityStates, mst.Vertex, mst.Edges, estiEdges)
		}
		if os.Getenv("ATCODER") != "1" {
			rmse := calcRMSE(cityStates, in.trueXY)
			log.Printf("update %d RMSE: %f", i, rmse)
		}
	}

	// 最終的な推定精度を評価
	if os.Getenv("ATCODER") != "1" {
		rmse := calcRMSE(cityStates, in.trueXY)
		log.Printf("Estimate Phase Final RMSE: %f", rmse)
	}
	//for _, mst := range queryResultMap {
	//log.Printf("Vertex: %v Edges: %d", mst.Vertex, len(mst.Edges))
	//}
}

// クエリ結果からベイズ更新を行う関数
func updateCityPositions(cityStates []CityState, queriedCities []int, edges [][2]int, estimateResult [][2]int) {
	// エッジ比較
	common, onlyEstimated, onlyActual := compareEdges(estimateResult, edges)

	for _, edge := range common {
		city1 := cityStates[edge[0]]
		city2 := cityStates[edge[1]]

		// 分散を更新する
		city1.Variance[0] *= (*pram3)
		city1.Variance[1] *= (*pram3)
		city2.Variance[0] *= (*pram3)
		city2.Variance[1] *= (*pram3)

		// 更新カウントを増やす
		city1.UpdateCount++
		city2.UpdateCount++
		// 分散の合計を更新
		city1.UpdateSumVariance()
		city2.UpdateSumVariance()
	}

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
		targetDist := currentDist * (*pram1)

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
		targetDist := currentDist * (*pram2)

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

	// 更新カウントを増やす
	city.UpdateCount++

	// NaNチェック（更新後）
	if math.IsNaN(city.Mean[0]) || math.IsNaN(city.Mean[1]) {
		log.Printf("Warning: NaN detected in city %d after update", city.ID)
		// 更新前の値に戻す
		city.Mean[0] -= moveX
		city.Mean[1] -= moveY
		city.UpdateCount-- // 更新をロールバック
		return
	}

	// 矩形の範囲内に収める
	city.Mean[0] = math.Max(city.Ract[0], math.Min(city.Ract[1], city.Mean[0]))
	city.Mean[1] = math.Max(city.Ract[2], math.Min(city.Ract[3], city.Mean[1]))

	// 分散を更新（0.95をかける）
	minVariance := 0.01 // 最小分散（あまりに小さくしないために）
	city.Variance[0] = math.Max(minVariance, city.Variance[0]*(*pram4))
	city.Variance[1] = math.Max(minVariance, city.Variance[1]*(*pram4))

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
	time1 := time.Now()
	cityStates := make([]CityState, in.N)
	for i := 0; i < in.N; i++ {
		cityStates[i].ID = i

		// 矩形の座標を取得
		rx := float64(in.lxrxlyry[i*4+1]) // 右
		lx := float64(in.lxrxlyry[i*4+0]) // 左
		ry := float64(in.lxrxlyry[i*4+3]) // 上
		ly := float64(in.lxrxlyry[i*4+2]) // 下

		// 矩形の中心
		//cityStates[i].Center[0] = (rx + lx) / 2
		//cityStates[i].Center[1] = (ry + ly) / 2

		// 初期平均値は矩形の中心
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

	// buildPhaseで使うクエリの数
	reserveQuery := 0
	for i := 0; i < in.M; i++ {
		if in.G[i] > 2 && in.L >= in.G[i] {
			reserveQuery++
		}
	}

	estimatePhase(in, in.Q-reserveQuery, cityStates)
	time2 := time.Now()
	log.Println("estimatePhaseTime:", time2.Sub(time1).Seconds())

	sortedGroup := make([]int, in.M)
	for i := 0; i < in.M; i++ {
		sortedGroup[i] = in.G[i]
	}

	bestAns := make([]AnsGroup, in.M)
	bestMapping := make([]int, in.M)
	bestScore := 100000000.0
	ansGroups := make([]AnsGroup, in.M)
	mapping := make([]int, in.M)
	loop := 20
	if in.M == 1 {
		loop = 1
	}
	for k := 0; k < loop; k++ {
		frand.Shuffle(len(sortedGroup), func(i, j int) {
			sortedGroup[i], sortedGroup[j] = sortedGroup[j], sortedGroup[i]
		})
		mapping = makeMapping(in.G[:in.M], sortedGroup)

		centerf := [2]float64{5000.0, 5000.0}
		var used [N]bool
		citiesSortedByCenter := make([]CityState, N)
		copy(citiesSortedByCenter, cityStates[:])
		sort.Slice(citiesSortedByCenter[:], func(i, j int) bool {
			return distanceSquared(citiesSortedByCenter[i].Mean, centerf) > distanceSquared(citiesSortedByCenter[j].Mean, centerf)
		})
		tmp := make([]CityState, N)
		copy(tmp, cityStates[:])
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
				//return distSquared(oldCities[groupRoot].Point, tmp[i].Point) < distSquared(oldCities[groupRoot].Point, tmp[j].Point)
				return distanceSquared(cityStates[groupRoot].Mean, tmp[i].Mean) < distanceSquared(cityStates[groupRoot].Mean, tmp[j].Mean)
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
			ansGroups[i].Edges = createMST(ansGroups[i].Cities, cityStates)
		}
		// 推定座標でcostの計算
		allCost := 0.0
		for i := 0; i < in.M; i++ {
			for j := 0; j < len(ansGroups[i].Edges); j++ {
				allCost += math.Sqrt(distanceSquared(cityStates[ansGroups[i].Edges[j][0]].Mean, cityStates[ansGroups[i].Edges[j][1]].Mean))
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
			log.Println("bestScore=", bestScore)
		}
	}
	// queryを使ったedgeの最適化
	log.Printf("localScore=%d\n", int(bestScore))
	for i := 0; i < in.M; i++ {
		if len(bestAns[i].Cities) > 2 && in.L >= len(bestAns[i].Cities) {
			bestAns[i].Edges = sendQuery(bestAns[i].Cities)
		}
	}

	time3 := time.Now()
	log.Println("buildPhasetime:", time3.Sub(time2).Seconds())

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
		// 初期のRMSEは W/4.24　程度
		log.Printf("RMSE=%.2f W/4.24:%.2f\n", calcRMSE(cityStates, in.trueXY), float64(in.W)/4.24)
	}
	//log.Printf("queryCount=%d\n", queryCount)
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

func createMST(cities []int, cityStates []CityState) [][2]int {

	// 頂点を0~len(cities)-1に振り直してMSTを作成
	edges := make(Edges, 0)
	for i := 0; i < len(cities); i++ {
		for j := i + 1; j < len(cities); j++ {
			city1 := cities[i]
			city2 := cities[j]
			weight := distanceSquared(cityStates[city1].Mean, cityStates[city2].Mean)
			edges = append(edges, Edge{From: i, To: j, Weight: weight})
		}
	}
	cost, mst := runKruskal(len(cities), edges)
	_ = cost
	//log.Printf("cost=%d\n", cost)
	newEdge := make([][2]int, len(mst))
	for i := 0; i < len(mst); i++ {
		from := cities[mst[i].From]
		to := cities[mst[i].To]
		newEdge[i][0] = intMin(from, to)
		newEdge[i][1] = intMax(from, to)
	}
	// ソートして返す
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

// query
var queryCount int
var queryResultMap = make(map[uint64]MSTResult)

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

type BitSet struct {
	bits []uint64
	size int
}

func NewBitSet(size int) *BitSet {
	// uint64 は 64bit なので必要な数を確保
	numWords := (size + 63) / 64
	return &BitSet{
		bits: make([]uint64, numWords),
		size: size,
	}
}

var hashTable [800]uint64

func makeHash(cities []int) uint64 {
	hash := uint64(0)
	for i := 0; i < len(cities); i++ {
		hash ^= uint64(cities[i]) ^ hashTable[cities[i]]
	}
	return hash
}

// ビットを立てる
func (bs *BitSet) Set(i int) {
	if i > bs.size {
		panic("index out of range")
	}
	if i >= 0 && i < bs.size {
		word, bit := i/64, uint(i%64)
		bs.bits[word] |= 1 << bit
	}
}

// ビットを落とす
func (bs *BitSet) Clear(i int) {
	if i >= 0 && i < bs.size {
		word, bit := i/64, uint(i%64)
		bs.bits[word] &^= 1 << bit
	}
}

// ビットをトグルする（反転）
func (bs *BitSet) Toggle(i int) {
	if i >= 0 && i < bs.size {
		word, bit := i/64, uint(i%64)
		bs.bits[word] ^= 1 << bit
	}
}

// ビットを取得する
func (bs *BitSet) Get(i int) bool {
	if i >= 0 && i < bs.size {
		word, bit := i/64, uint(i%64)
		return bs.bits[word]&(1<<bit) != 0
	}
	return false
}

// サイズ取得
func (bs *BitSet) Len() int {
	return bs.size
}

func calcRMSE(cityStates []CityState, trueXY [800][2]float64) float64 {
	sumSqErr := 0.0
	for i := 0; i < len(cityStates); i++ {
		estX := cityStates[i].Mean[0]
		estY := cityStates[i].Mean[1]
		realX := trueXY[i][0]
		realY := trueXY[i][1]

		dx := estX - realX
		sumSqErr += dx * dx

		dy := estY - realY
		sumSqErr += dy * dy
	}
	// 平均二乗誤差（Mean Squared Error）
	mse := sumSqErr / float64(2*len(cityStates)) // 都市数×座標2次元分で割る
	// 平均二乗誤差の平方根（Root Mean Squared Error）
	rmse := math.Sqrt(mse)
	return rmse
}
