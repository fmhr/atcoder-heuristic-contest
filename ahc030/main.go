package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"time"
)

func extentConfirmWithHeap(mapping *[20][20]squareInfo) {
	// 確定したマスに隣接したマスの中から、最も平均の高いマスを選び、占いを行う

}

// exteneComfirm : 確定したマスに隣接したマスを占い、確定したマスを広げる
func extentComfirm(mapping *[20][20]squareInfo) {
	sumConfirmed := 0
	var visited [20][20]bool
	stack := make([][2]int, 0)
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			if mapping[i][j].variance == 0 && mapping[i][j].mean >= 1 {
				stack = append(stack, [2]int{i, j})
				sumConfirmed += int(mapping[i][j].mean)
			}
		}
	}
	for len(stack) > 0 {
		now := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for d := 0; d < 4; d++ {
			ny, nx := now[0]+dy[d], now[1]+dx[d]
			if ny < 0 || ny >= islandSize || nx < 0 || nx >= islandSize {
				continue
			}
			if mapping[ny][nx].variance == 0 {
				continue
			}
			if visited[ny][nx] {
				continue
			}
			// ここで占いを行う
			a, _ := divine([][2]int{{ny, nx}})
			mapping[ny][nx].update(float64(a), 1*errP*(1-errP), 1)
			if a > 0 {
				stack = append(stack, [2]int{ny, nx})
				sumConfirmed += a
			}
		}
	}
}

type samplingInfo struct {
	SamMean  int
	Variance float64
	Points   [][2]int
	Size     int
}

func newSamplingInfo(sum, num int, points [][2]int) samplingInfo {
	variance := float64(num) * errP * (1 - errP)
	return samplingInfo{SamMean: sum, Variance: variance, Points: points, Size: num}
}

// mapppingの可視化
func visualizeMapping(mapping [20][20]squareInfo) {
	log.Println("mean")
	for i := 0; i < islandSize; i++ {
		var str string
		for j := 0; j < islandSize; j++ {
			str += fmt.Sprintf("%.2f ", mapping[i][j].mean)
		}
		log.Println(str)
	}
	log.Println("variance")
	for i := 0; i < islandSize; i++ {
		var str string
		for j := 0; j < islandSize; j++ {
			str += fmt.Sprintf("%.2f ", mapping[i][j].variance)
		}
		log.Println(str)
	}
	log.Println("confirmed")
	for i := 0; i < islandSize; i++ {
		var str string
		for j := 0; j < islandSize; j++ {
			if mapping[i][j].variance == 0 {
				str += fmt.Sprintf("%d ", int(mapping[i][j].mean))
			} else {
				str += "x "
			}
		}
		log.Println(str)
	}
}

func divideGrid(si, sj, ei, ej int) (regionPoints [4][][2]int) {
	log.Println("divideGrid", si, sj, ei, ej)
	size := ei - si
	half := size/2 + (size % 2)
	// 4つの領域に分ける
	// 1 3
	// 2 4
	// 奇数にも対応（1以外の領域が大きくなる）
	regionStart := [4][2]int{{si, sj}, {si + half, sj}, {si, sj + half}, {si + half, sj + half}}
	regionEnd := [4][2]int{{si + half, sj + half}, {ei, sj + half}, {si + half, ej}, {ei, ej}}

	for k := 0; k < 4; k++ {
		startI, startJ := regionStart[k][0], regionStart[k][1]
		endI, endJ := regionEnd[k][0], regionEnd[k][1]
		for i := startI; i < endI; i++ {
			for j := startJ; j < endJ; j++ {
				regionPoints[k] = append(regionPoints[k], [2]int{i, j})
			}
		}
	}
	return regionPoints
}

type squareInfo struct {
	mean     float64
	variance float64
	times    int
}

// numはサンプリングした時の数
func (s *squareInfo) update(newMean, newVariance float64, sampleSize int) {
	if sampleSize == 1 {
		// サンプリング(占い)を１マスで行った場合は正確な値が出る
		s.mean = newMean
		s.variance = 0
		s.times = 1
		return
	} else if s.variance == 0 {
		// 既に正確な値が出ている場合は何もしない
		//log.Println("already accurate")
		//s.times++
		return
	}
	s.mean = (s.mean*float64(s.times) + newMean) / float64(s.times+1)
	s.variance = (s.variance*float64(s.times) + newVariance) / float64(s.times+1)
	s.times++
}

// 始めの情報だけで初期値を設定する
func initializeMapping() (mapping [20][20]squareInfo) {
	sumOils := 0
	for i := 0; i < numberOfOil; i++ {
		sumOils += oilFieldSize[i]
	}
	sumMean := float64(sumOils) / float64(islandSize*islandSize)
	sumVariance := float64(islandSize*islandSize) * errP * (1 - errP)
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			mapping[i][j] = squareInfo{mean: sumMean, variance: sumVariance, times: 1}
		}
	}
	return
}

// reSampling : 確定したマスでサンプリングの結果を更新して、新しいマッピングを作成する
func reSampling(mapping [20][20]squareInfo, samplingLogs []samplingInfo) (newMapping [20][20]squareInfo) {
	newMapping = initializeMapping()
	// サンプリングログを確認して、確定マスでサンプリング記録を更新する
	for i := 0; i < len(samplingLogs); i++ {
		for j := 0; j < len(samplingLogs[i].Points); j++ {
			p := samplingLogs[i].Points[j]
			// 確定マスが存在するので、Logを更新する
			if mapping[p[0]][p[1]].variance == 0.0 {
				samplingLogs[i].SamMean -= int(mapping[p[0]][p[1]].mean)
				samplingLogs[i].Variance -= 1 * errP * (1 - errP)
				newMapping[p[0]][p[1]] = mapping[p[0]][p[1]] // 確定マスの情報をコピー
				// Sizeは精度に関係するので更新しない
				// Pointsの確定マスはスキップされるのでここでは削除しない
			}
		}
	}
	// 新しいサンプリングで新しいマッピングを作成する
	for i := 0; i < len(samplingLogs); i++ {
		if samplingLogs[i].Size == 0 {
			continue
		}
		for j := 0; j < len(samplingLogs[i].Points); j++ {
			p := samplingLogs[i].Points[j]
			v := float64(samplingLogs[i].Size) * errP * (1 - errP)
			if v == 0 {
				log.Println(p, samplingLogs[i].SamMean, samplingLogs[i].Size, v)
			}
			newMapping[p[0]][p[1]].update(float64(samplingLogs[i].SamMean)/float64(samplingLogs[i].Size), v, samplingLogs[i].Size)
		}
	}
	//visualizeMapping(newMapping)
	return
}

func solver_mini_case() {
	samplingLogs := make([]samplingInfo, 0)
	mapping := initializeMapping()
	// 地図上で油田の占める割合が小さいケースでの実装
	// 油田がありそうなところを探して、その周囲を探索する
	// 全体を４分割して、最初の１マスを決める
	subgrids := divideGrid(0, 0, islandSize, islandSize)
	for {
		maxDivine := 0
		maxDivineIndex := 0
		for k := 0; k < 4; k++ {
			a, cost := divine(subgrids[k])
			log.Printf("divine:%d size:%d num:%d cost: %.3f\n", k, len(subgrids[k]), a, cost)
			for i := 0; i < len(subgrids[k]); i++ {
				variance := float64(len(subgrids[k])) * errP * (1 - errP)
				mapping[subgrids[k][i][0]][subgrids[k][i][1]].update(float64(a)/float64(len(subgrids[k])), variance, len(subgrids[k]))
				samplingLogs = append(samplingLogs, newSamplingInfo(a, len(subgrids[k]), subgrids[k]))
			}
			if a > maxDivine {
				maxDivine = a
				maxDivineIndex = k
			}
		}
		next := subgrids[maxDivineIndex]
		if len(next) <= 1 {
			if len(next) == 0 {
				break
			}
			if mapping[next[0][0]][next[0][1]].variance == 0.0 {
				break
			}
		}
		subgrids = divideGrid(next[0][0], next[0][1], next[len(next)-1][0]+1, next[len(next)-1][1]+1)
		for k := 0; k < 4; k++ {
			log.Println("next", k, "size", len(subgrids[k]), subgrids[k])
		}
	}
	visualizeMapping(mapping)
	// 確定した油田のマスを広げていく
	extentComfirm(&mapping)
	log.Println("extentComfirm")

	visualizeMapping(mapping)
	log.Println("totalCost", totalCost)
	//mapping = reSampling(mapping, samplingLogs)
	log.Println("reSampling")
	visualizeMapping(mapping)
	// 油田が全て見つかるまで繰り返す
	var numOil int
	for numOil != sumOils {
		// 油田の出る確率が大きいものを探す
		maxMean := -10000.0
		maxPoints := make([][2]int, 0)
		for i := 0; i < islandSize; i++ {
			for j := 0; j < islandSize; j++ {
				if mapping[i][j].variance != 0 {
					if mapping[i][j].mean > maxMean {
						maxMean = mapping[i][j].mean
						maxPoints = [][2]int{{i, j}}
					} else if mapping[i][j].mean == maxMean {
						maxPoints = append(maxPoints, [2]int{i, j})
					}
				}
			}
		}
		log.Println("maxMean", maxMean, maxPoints)
		// ４分割して、最初の１マスを決める
		for {
			var subgrids [4][][2]int
			subgrids[0] = maxPoints[:len(maxPoints)/4]
			subgrids[1] = maxPoints[len(maxPoints)/4 : len(maxPoints)/2]
			subgrids[2] = maxPoints[len(maxPoints)/2 : len(maxPoints)*3/4]
			subgrids[3] = maxPoints[len(maxPoints)*3/4:]
			maxIndex := 0
			maxMean := -10000.0
			for i := 0; i < 4; i++ {
				sumMean, cost := divine(subgrids[i])
				sampleSize := len(subgrids[i])
				if sampleSize == 0 {
					continue
				}
				newMean := float64(sumMean) / float64(sampleSize)
				_ = cost
				for j := 0; j < len(subgrids[i]); j++ {
					newVariance := float64(sampleSize) * errP * (1 - errP)
					mapping[subgrids[i][j][0]][subgrids[i][j][1]].update(newMean, newVariance, sampleSize)
				}
				if newMean > maxMean {
					maxMean = newMean
					maxIndex = i
				}
			}
			maxPoints = subgrids[maxIndex]
			log.Println(len(maxPoints), maxPoints)
			if len(maxPoints) <= 1 {
				break
			}
			log.Println(subgrids[maxIndex])
		}
		//visualizeMapping(mapping)
		extentComfirm(&mapping)
		log.Println("extentComfirm")
		visualizeMapping(mapping)
		numOil = 0
		for i := 0; i < islandSize; i++ {
			for j := 0; j < islandSize; j++ {
				if mapping[i][j].variance == 0 {
					numOil += int(mapping[i][j].mean)
				}
			}
		}
		log.Println("sumOils:", sumOils, "numOil", numOil)
	}
	log.Println("sumOils:", sumOils, "numOil", numOil)
	// 結果を出力する
	ans := make([][20][2]int, 0, numOil)
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			if mapping[i][j].variance == 0 && mapping[i][j].mean > 0 {
				ans = append(ans, [20][2]int{{i, j}})
			}
		}
	}
	fmt.Printf("a %d ", len(ans))
	for i := 0; i < len(ans); i++ {
		fmt.Printf("%d %d ", ans[i][0][0], ans[i][0][1])
	}
	fmt.Println("")
	var rtn int
	fmt.Scan(&rtn)
	log.Println(rtn)
}

var ansF [][20][2]int
var cntCheckFilled int

// 探索済みで1以上の観測地点にその値が入っているか確認する
func checkFilled(confirmed [20][20]int, field [20][20]int) bool {
	cntCheckFilled++
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			if confirmed[i][j] > 0 && confirmed[i][j]-1 != field[i][j] {
				return false
			}
		}
	}
	return true
}

// 深さ優先探索で条件に合うものを全探索する
func searchOilField(confirmed [20][20]int, field [20][20]int, oilNum int, ans [20][2]int) {
	if oilNum == numberOfOil {
		if checkFilled(confirmed, field) {
			//log.Println("find!", ans[:numberOfOil])
			//r := output(ans)
			//if r == 1 {
			//os.Exit(0)
			//}
			//log.Println(r)
			ansF = append(ansF, ans)
		}
		return
	}
	// 全マスを探索する
	var x, y int
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			var ng bool
			var k int
			// TODO 油田がひとつもconfirmedとヒットしないときは、探索をスキップする
			for ; k < oilFieldSize[oilNum]; k++ {
				// ここに油田があると仮定して, confirmedと矛盾がないか調べる
				y = i + oilField[oilNum][k][0]
				x = j + oilField[oilNum][k][1]
				if y < 0 || y >= islandSize {
					ng = true
					break
				}
				if x < 0 || x >= islandSize {
					ng = true
					break
				}
				// confirmedが0の場所は未探索なので、矛盾しない
				if confirmed[y][x] == 0 {
					continue
				}
				// confirmedが1以上の場所は探索済みなので、矛盾がないか調べる
				// +1したときconfirmedより大きくなる場合は矛盾する
				if confirmed[y][x]-1 > field[y][x] {
					field[y][x]++
				} else {
					ng = true
					break
				}
			}
			if !ng {
				// ここに油田があると仮定して、次の油田を探索する
				ans[oilNum][0] = i
				ans[oilNum][1] = j
				searchOilField(confirmed, field, oilNum+1, ans)
				// ここでbreakすると、他のありえる場所を探索しない
			}
			// 油田が矛盾するor他の可能性を探すためにのUndo
			for k--; k >= 0; k-- {
				field[i+oilField[oilNum][k][0]][j+oilField[oilNum][k][1]]--
			}
		}
	}
}

// 油田の存在するすべてのマスを列挙する
func output(ans [20][2]int) (rtn int) {
	var fiels [20][20]bool
	for i := 0; i < numberOfOil; i++ {
		for j := 0; j < oilFieldSize[i]; j++ {
			fiels[ans[i][0]+oilField[i][j][0]][ans[i][1]+oilField[i][j][1]] = true
		}
	}
	cnt := 0
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			if fiels[i][j] {
				cnt++
			}
		}
	}
	fmt.Print("a ", cnt, " ")
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			if fiels[i][j] {
				fmt.Print(i, " ", j, " ")
			}
		}
	}
	fmt.Print("\n")
	fmt.Scan(&rtn)
	if rtn == 0 {
		totalCost++
	}
	return rtn
}

var totalCost float64

// 1マスの油田の存在を確認する
func singleDivine(x, y int) (v int, c float64) {
	fmt.Println("q 1", x, y)
	fmt.Scan(&v)
	c = 1
	totalCost++
	return
}

// 複数マスの油田の存在を占う
func divine(set [][2]int) (sumMean int, cost float64) {
	if len(set) == 0 {
		log.Println("divine empty set")
		return 0, 0
	}
	if len(set) == 1 {
		return singleDivine(set[0][0], set[0][1])
	}
	fmt.Print("q ", len(set), " ")
	for i := 0; i < len(set); i++ {
		fmt.Print(set[i][0], " ", set[i][1], " ")
	}
	fmt.Println()
	fmt.Scan(&sumMean)
	k := len(set)
	cost = 1 / math.Log2(float64(k))
	totalCost += 1 / math.Log2(float64(k))
	return sumMean, cost
}

// すべてのマスの油田の存在を確認する
func solver() {
	var confirmed [20][20]int
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			//if i%3 == 0 && j%3 == 0 {
			//continue
			//}
			r, _ := singleDivine(i, j)
			confirmed[i][j] = r + 1
		}
	}
	// ここでconfirmedの状態を確認する
	cnt := 0
	for i := 0; i < islandSize; i++ {
		for j := 0; j < islandSize; j++ {
			if confirmed[i][j] > 1 {
				cnt++
			}
		}
	}
	log.Printf("confirmeK=%f", float64(cnt)/float64(islandSize*islandSize))
	log.Println("solver")
	for i := 0; i < islandSize; i++ {
		log.Println(confirmed[i][:islandSize])
	}
	var f [20][20]int
	var ans [20][2]int
	searchOilField(confirmed, f, 0, ans)
	log.Println(ansF[0][:numberOfOil])
	log.Printf("ansSize=%d", len(ansF))
	for i := 0; i < len(ansF); i++ {
		r := output(ansF[i])
		// 1なら終了
		if r == 1 {
			break
		}
	}
	log.Printf("totalCost=%f", totalCost)
	log.Printf("CheckFill=%d", cntCheckFilled)
}

var islandSize, numberOfOil int
var errP float64
var oilField [20][][2]int // 最大２０個の油田と、それぞれの油田の座標
var oilFieldSize [20]int
var sumOils int

func read(r *os.File) {
	fmt.Fscan(r, &islandSize, &numberOfOil, &errP)
	ds := make([]int, numberOfOil)
	for i := 0; i < numberOfOil; i++ {
		fmt.Fscan(r, &oilFieldSize[i])
		ds[i] = oilFieldSize[i]
		oilField[i] = make([][2]int, oilFieldSize[i])
		for j := 0; j < oilFieldSize[i]; j++ {
			fmt.Fscan(r, &oilField[i][j][0], &oilField[i][j][1])
		}
		// visualize
		var vis [20][20]int
		for j := 0; j < oilFieldSize[i]; j++ {
			vis[oilField[i][j][0]][oilField[i][j][1]] = 1
		}
		//log.Println("visualize", i, "th oilField:")
		//for j := 0; j < islandSize; j++ {
		//log.Println(vis[j][:islandSize])
		//}
	}
	maxD := intsMax(0, ds...)
	minD := intsMin(ds[0], ds...)
	abgD := intsMean(ds...)
	sumD := intsSum(ds...)
	sumOils = sumD
	log.Printf("size=%d, oils=%d, errorP=%f d_max=%d d_min=%d d_avg=%f\n", islandSize, numberOfOil, errP, maxD, minD, abgD)
	log.Printf("sumD=%d K=%v\n", sumD, float64(sumD)/float64(islandSize*islandSize))
}

// ./bin/main -cpuprofile cpu.out < tools/in/0000.txt && go tool pprof -http=localhost:8888 a.out cpu.out
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	///////////////////////////////
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
	////////////////////////////////
	log.SetFlags(log.Lshortfile)
	s := time.Now()
	read(os.Stdin)
	//solver()
	solver_mini_case()
	elp := time.Since(s)
	log.Printf("time=%v s", elp.Seconds())
}

// 以下、ユーティリティ
func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func intsMax(a int, b ...int) int {
	for i := 0; i < len(b); i++ {
		a = intMax(a, b[i])
	}
	return a
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intsMin(a int, b ...int) int {
	for i := 0; i < len(b); i++ {
		a = intMin(a, b[i])
	}
	return a
}

func intsMean(a ...int) float64 {
	var sum int
	for i := 0; i < len(a); i++ {
		sum += a[i]
	}
	return float64(sum) / float64(len(a))
}

func intsSum(a ...int) int {
	var sum int
	for i := 0; i < len(a); i++ {
		sum += a[i]
	}
	return sum
}

var dy = []int{0, 1, 0, -1}
var dx = []int{1, 0, -1, 0}
