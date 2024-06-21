package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

	"golang.org/x/exp/slices"
)

// (Score)sum=5146933076.00 avarage=5146933.08 log=22.361667

type Comparison int

const (
	GreaterThan Comparison = 1
	LessThan    Comparison = -1
	Equal       Comparison = 0
	Unknown     Comparison = 2
)

var comparisonOperators map[string]Comparison = map[string]Comparison{">": GreaterThan, "<": LessThan, "=": Equal}

func init() {
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			balanceCache[i][j] = Unknown
		}
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	rand.Seed(0)
	start := time.Now()
	solver()
	duration := time.Since(start)
	log.Printf("time=%f\n", duration.Seconds())
}

var N, D, Q int
var Qleft int

func solver() {
	// 数分割問題の貪欲法がつかえそう
	// https://scmopt.github.io/opt100/76npp.html

	// 入力
	_, err := fmt.Scan(&N, &D, &Q)
	if err != nil {
		log.Fatal(err)
	}

	Qleft = Q
	log.Printf("N=%d D=%d Q=%d", N, D, Q)

	// レンジアイテムをもとめる (アイテムの内一番大きいものと一番小さいものの差)
	//rengeItems := findMinimums(10)
	log.Println(Q, Qleft)
	//アイテムを全てソートでつなげる
	for i := 0; i < N-1; i++ {
		_, err := balance(i, i+1, &Qleft)
		if err != nil {
			log.Fatal(err)
		}
	}

	tmpItems := make([]int, N)
	for i := 0; i < N; i++ {
		tmpItems[i] = i
	}

	log.Println(Q, Qleft)
	time := int(float64(Q) * 0.1)
	log.Println(time)
	items := estimateItem()
	var sumWeight int
	for i := 0; i < N; i++ {
		sumWeight += items[i].weight
	}
	log.Println(sumWeight)
	//q := Q / 10
	loop := 0
	var a, b Bug
	for Qleft > 0 {
		loop++
		if loop%minInt(1000, 100) == 1 {
			sumWeight = 0
			items = estimateItem()
			for i := 0; i < N; i++ {
				sumWeight += items[i].weight
			}
			rand.Shuffle(len(items), func(i, j int) {
				items[i], items[j] = items[j], items[i]
			})
		}
		limit := rand.Intn(sumWeight / D)
		a.reset()
		b.reset()
		var i int
		for i < N-1 && (a.weight < limit || b.weight < limit) {
			if a.weight < limit {
				a.add(items[i])
				i++
			}
			if b.weight < limit {
				b.add(items[i])
				i++
			}
		}
		//log.Println(a, b, limit)
		_, err := balanceItems(a.items, b.items, &Qleft)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(Q, Qleft, loop)
	}
	tmpBugs := solverWeight(estimate())
	bags := make([][]int, D)
	for i := 0; i < D; i++ {
		bags[i] = tmpBugs[i].items
	}
	//	// かごに入れていく
	//for i := 0; i < N; i++ {
	//j := i % D
	//if i/D%2 == 0 {
	//j = D - 1 - j
	//}
	//bags[j] = append(bags[j], i)
	//}
	fmt.Println("#c ", outputBags(N, bags))
	fmt.Println(outputBags(N, bags))
	return
	var moveItem [2]int   // [試行回数, 受理回数]
	var swapItem [2]int   // [試行回数, 受理回数]
	var swapMinMax [2]int // [試行回数, 受理回数]
	for loop < 1000 {
		loop++
		func() {
			// moveItem
			// カゴをランダムに選ぶ
			// 大きい方から小さい方にアイテムを移動する
			// 移動後もカゴの大小関係がかわらないものを選ぶ
			// * 二つのカゴの大小関係が縮まる
			moveItem[0]++
			i := rand.Intn(D - 1)
			j := i + 1 + rand.Intn(D-i-1)
			// i < j
			cmp1, err := balanceItems(bags[i], bags[j], &Qleft)
			if err != nil {
				if errors.Is(err, ErrNoLeftNoCache) {
					return
				}
				log.Fatal(err)
			}
			if cmp1 == Equal {
				return
			} else if cmp1 == GreaterThan {
				bags[i], bags[j] = bags[j], bags[i]
				// bag[i] < bag[j]の関係にする
			}
			randItem := rand.Intn(len(bags[j]))
			item := bags[j][randItem]
			bags[j] = append(bags[j][:randItem], bags[j][randItem+1:]...) // delete
			bags[i] = append(bags[i], item)                               // insert
			cmp2, err := balanceItems(bags[i], bags[j], &Qleft)
			if err != nil {
				if errors.Is(err, ErrNoLeftNoCache) {
					// 測定不能
					// 直前の操作を戻す
					bags[i] = bags[i][:len(bags[i])-1] // delete
					bags[j] = append(bags[j], item)    // insert
					return
				}
				log.Fatal(err)
			}
			if cmp2 == GreaterThan {
				// 戻す
				bags[i] = bags[i][:len(bags[i])-1] // delete
				bags[j] = append(bags[j], item)    // insert
				return
			}
			fmt.Println("#c ", outputBags(N, bags))
			moveItem[1]++
			//balanceCache = warshallFloyd(balanceCache)
		}()
		if Q-Qleft < N/2 {
			continue
		}

		var leftZero int
		func() {
			swapItem[0]++
			// カゴをランダムに選ぶ
			// 大きな方の小さなアイテムと小さな方の大きなアイテムを交換する
			i := rand.Intn(D - 1)
			j := i + 1 + rand.Intn(D-i-1)
			// i < j
			cmp1, err := balanceItems(bags[i], bags[j], &leftZero)
			if err != nil {
				if errors.Is(err, ErrNoLeftNoCache) {
					return
				}
				log.Fatal(err)
			}
			if cmp1 == Equal {
				return
			} else if cmp1 == GreaterThan {
				// bag[i] < bag[j]の関係にする
				bags[i], bags[j] = bags[j], bags[i]
			}
			indexI := rand.Intn(len(bags[i]))
			//indexJ := rand.Intn(len(bags[j]))
			itemI := bags[i][indexI]
			//itemJ := bags[j][indexJ]
			//cmp2, err := balance(itemI, itemJ, &leftZero)
			//if err != nil {
			//if errors.Is(err, ErrNoLeftNoCache) {
			//return
			//}
			//log.Fatal(err)
			//}
			//if !(cmp2 == LessThan) {
			//return
			//}
			var indexJ int
			for indexJ < len(bags[j]) {
				itemJ := bags[j][indexJ]
				cmp2, err := balance(itemI, itemJ, &leftZero)
				if err != nil {
					if errors.Is(err, ErrNoLeftNoCache) {
						return
					}
					log.Fatal(err)
				}
				if cmp2 == GreaterThan {
					break
				}
				indexJ++
			}
			if indexJ == len(bags[j]) {
				return
			}

			// SWAP
			bags[i][indexI], bags[j][indexJ] = bags[j][indexJ], bags[i][indexI]
			cmp3, err := balanceItems(bags[i], bags[j], &leftZero)
			if err != nil {
				if errors.Is(err, ErrNoLeftNoCache) {
					// 戻す
					bags[i][indexI], bags[j][indexJ] = bags[j][indexJ], bags[i][indexI]
					return
				}
				log.Fatal(err)
			}
			if !(cmp3 == LessThan) {
				// 戻す
				bags[i][indexI], bags[j][indexJ] = bags[j][indexJ], bags[i][indexI]
				return
			}
			swapItem[1]++
			balanceCache = warshallFloyd(balanceCache)
		}()
		if D == 2 {
			continue
		}
		func() {
			// 最大のカゴから最小のカゴにアイテムを移動する
			// 最大のカゴのアイテムの大きいものから試す
			// 最大のカゴがbug[1]よりおおきく、最小のカゴがbug[D-2]より小さいとき受理
			// bag[0] < bag[1] < ... < bag[D-1]
			swapMinMax[0]++
			index1 := rand.Intn(len(bags[0]))
			index2 := rand.Intn(len(bags[D-1]))
			item1 := bags[0][index1]
			item2 := bags[D-1][index2]
			cmp1, err := balance(item1, item2, &Qleft)
			if err != nil {
				if errors.Is(err, ErrNoLeftNoCache) {
					return
				}
				log.Fatal(err)
			}
			// 選ばれたアイテムは item1 > item2でないといけない
			if !(cmp1 == GreaterThan) {
				return
			}
			// SWAP
			bags[0][index1], bags[D-1][index2] = bags[D-1][index2], bags[0][index1]
			cmp2, err := balanceItems(bags[0], bags[D-1-1], &leftZero)
			if err != nil {
				if errors.Is(err, ErrNoLeftNoCache) {
					// 戻す
					bags[0][index1], bags[D-1][index2] = bags[D-1][index2], bags[0][index1]
					return
				}
				log.Fatal(err)
			}
			if !(cmp2 == LessThan) {
				// 戻す
				bags[0][index1], bags[D-1][index2] = bags[D-1][index2], bags[0][index1]
				return
			}
			log.Println(bags[1], bags[D-1])
			cmp3, err := balanceItems(bags[1], bags[D-1], &leftZero)
			if err != nil {
				if errors.Is(err, ErrNoLeftNoCache) {
					// 戻す
					bags[0][index1], bags[D-1][index2] = bags[D-1][index2], bags[0][index1]
					return
				}
				log.Fatal(err)
			}
			if !(cmp3 == LessThan) {
				// 戻す
				bags[0][index1], bags[D-1][index2] = bags[D-1][index2], bags[0][index1]
				return
			}
			swapMinMax[1]++
			log.Println(cmp1, cmp2, cmp3)
		}()
	}
	log.Printf("loop=%d\n", loop)

	//_ = warshallFloyd(balanceCache)
	log.Println(" [試行回数, 受理回数]")
	log.Println("moveItem", moveItem)
	log.Println("swapItem", swapItem)
	log.Println("swapMinMax", swapMinMax)
	log.Println("BalanceItems size", len(balanceItemsCacheBit))
	// 残ったクエリを消費
	log.Printf("Qused=%d\n", Q-Qleft)
	log.Println("Qleft", Qleft)
	var resp string
	for Qleft > 0 {
		func() {
			for i := 0; i < N; i++ {
				for j := 0; j < N; j++ {
					if i == j {
						continue
					}
					_, err := balance(i, j, &Qleft)
					if err != nil {
						if errors.Is(err, ErrNoLeftNoCache) {
							break
						}
						log.Fatal(err)
					}
				}
			}
		}()
		warshallFloyd(balanceCache)
		if Qleft > 0 {
			fmt.Println("1 1 0 1")
			fmt.Scan(&resp)
			Qleft--
		}
	}
	fmt.Println(outputBags(N, bags))
	//log.Println(outputBags(N, bags))
	estWeight := estimate()
	bagWeight := make([]int, D)
	for i := 0; i < D; i++ {
		for j := 0; j < len(bags[i]); j++ {
			bagWeight[i] += estWeight[bags[i][j]]
		}
	}
	log.Println("bagWeight", bagWeight)
}

type Item struct {
	index  int
	weight int
}

type Bug struct {
	items  []int
	weight int
}

func (b *Bug) add(item Item) {
	b.items = append(b.items, item.index)
	b.weight += item.weight
}

func (b *Bug) reset() {
	b.items = make([]int, 0)
	b.weight = 0
}

func solverWeight(estWeight []int) []Bug {
	minWeight := 1000000000
	sumWeight := 0
	for i := 0; i < N; i++ {
		minWeight = minInt(minWeight, estWeight[i])
	}
	if minWeight < 0 {
		for i := 0; i < N; i++ {
			estWeight[i] += -minWeight * 2
		}
	}
	for i := 0; i < N; i++ {
		sumWeight += estWeight[i]
	}
	items := make([]Item, N)
	for i := 0; i < N; i++ {
		items[i] = Item{index: i, weight: estWeight[i]}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].weight > items[j].weight
	})
	log.Println(items)
	bugs := make([]Bug, D)
	for i := 0; i < N; i++ {
		bugs[0].add(items[i])
		sort.Slice(bugs, func(i, j int) bool {
			return bugs[i].weight < bugs[j].weight
		})
	}
	log.Println(bugs)
	return bugs
}

var balanceCache [100][100]Comparison

func warshallFloyd(src [100][100]Comparison) (rtn [100][100]Comparison) {
	updateCnt := 0
	// 初期化
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			rtn[i][j] = src[i][j]
		}
	}

	for k := 0; k < N; k++ {
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				if rtn[i][k] == GreaterThan && rtn[k][j] == GreaterThan {
					if rtn[i][j] == Unknown {
						updateCnt++
					} else if rtn[i][j] == LessThan {
						log.Fatal("error")
					}
					rtn[i][j] = GreaterThan
				} else if rtn[i][k] == LessThan && rtn[k][j] == LessThan {
					if rtn[i][j] == Unknown {
						updateCnt++
					} else if rtn[i][j] == GreaterThan {
						log.Fatal("error")
					}
					rtn[i][j] = LessThan
				}
			}
		}
	}
	//log.Println("updateCnt", updateCnt)
	return
}

var ErrNoLeftNoCache = errors.New("no left query, no cache")

// かごの中身を比較する
func balance(a, b int, Qleft *int) (Comparison, error) {
	//log.Println(a, b)
	if balanceCache[a][b] != Unknown {
		return balanceCache[a][b], nil
	}
	if *Qleft == 0 {
		return Unknown, ErrNoLeftNoCache
	}

	// reactive step
	fmt.Printf("1 1 %d %d\n", a, b)
	var resp string
	_, err := fmt.Scan(&resp)
	if err != nil {
		log.Fatal(err)
	}
	// ------------
	(*Qleft)--
	//log.Println(a, b, resp)
	if resp == ">" {
		balanceCache[a][b] = comparisonOperators[resp]
	} else if resp == "<" {
		balanceCache[b][a] = comparisonOperators[resp] * -1
	}
	return comparisonOperators[resp], nil
}

var balanceItemsCacheBit map[BitSet]Comparison = make(map[BitSet]Comparison)

func balanceItems(a, b []int, Qleft *int) (Comparison, error) {
	//log.Printf("%d %d \n", len(a), len(b))
	if len(a) == 0 && len(b) == 0 {
		return Equal, nil
	} else if len(a) == 0 {
		return LessThan, nil
	} else if len(b) == 0 {
		return GreaterThan, nil
	}
	// 過去の結果を使う
	hash := SetBit(a, b)
	result, exsist := balanceItemsCacheBit[hash]
	if exsist && result != Unknown {
		return result, nil
	}

	if *Qleft == 0 {
		return Unknown, ErrNoLeftNoCache
	}

	astr, err := rawString(a)
	if err != nil {
		return 0, fmt.Errorf("a []int is empty")
	}
	bstr, err := rawString(b)
	if err != nil {
		return 0, fmt.Errorf("b []int is empty")
	}
	fmt.Printf("%d %d %s %s\n", len(a), len(b), astr, bstr)
	var resp string
	_, err = fmt.Scan(&resp)
	if err != nil {
		log.Fatal(err)
	}
	(*Qleft)--
	// add hash
	balanceItemsCacheBit[hash] = comparisonOperators[resp]
	//hash = SetBit(b, a)
	//balanceItemsCacheBit[hash] = comparisonOperators[resp] * -1
	return comparisonOperators[resp], nil
}

func rawString(a []int) (str string, err error) {
	if len(a) == 0 {
		return "", fmt.Errorf("a []int is empty")
	}
	for _, v := range a {
		str += fmt.Sprintf("%d ", v)
	}
	str = str[:len(str)-1]
	return
}

func outputBags(N int, bags [][]int) string {
	ans := make([]int, N)
	for i := 0; i < len(bags); i++ {
		for _, v := range bags[i] {
			ans[v] = i
		}
	}
	str, _ := rawString(ans)
	return str
}

type BitSet [4]uint64

func (b *BitSet) Set(pos uint) {
	index := pos / 64
	offset := pos % 64
	b[index] |= 1 << offset
}

func (b *BitSet) Clear(pos uint) {
	index := pos / 64
	offset := pos % 64
	b[index] &^= (1 << offset)
}

func (b *BitSet) IsSet(pos uint) bool {
	index := pos / 64
	offset := pos % 64
	return b[index]&(1<<offset) != 0
}

func (b *BitSet) Equal(other *BitSet) bool {
	return b[0] == other[0] && b[1] == other[1] && b[2] == other[2] && b[3] == other[3]
}

// make new BitSet seed
func SetBit(a, b []int) (result BitSet) {
	for i := 0; i < len(a); i++ {
		result.Set(uint(a[i]))
	}
	for i := 0; i < len(b); i++ {
		result.Set(uint(b[i]) + 100)
	}
	return
}

func Decode(bs *BitSet) (a, b []int) {
	for i := 0; i < 100; i++ {
		if bs.IsSet(uint(i)) {
			a = append(a, i)
		}
	}
	for j := 100; j < 200; j++ {
		if bs.IsSet(uint(j)) {
			b = append(b, j-100)
		}
	}
	return
}

// FlouydWarshall法をつかう closureにはGreaterのみにする
func transitiveClosure(graph [100][100]Comparison) (closure [100][100]Comparison) {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if graph[i][j] == GreaterThan {
				closure[i][j] = GreaterThan
			}
		}
	}
	for k := 0; k < N; k++ {
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				closure[i][j] |= closure[i][k] & closure[k][j]
			}
		}
	}
	return closure
}

func transitiveReduction(graph [100][100]Comparison, closure [100][100]Comparison) [100][100]Comparison {
	var reduction = closure
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if i != j && closure[i][j] != 0 {
				for k := 0; k < N; k++ {
					if i != k && j != k && closure[i][k] != 0 && closure[k][j] != 0 {
						reduction[i][j] = 0
						break
					}
				}
			}
		}
	}
	return reduction
}

func estimateItem() (items []Item) {
	items = make([]Item, N)
	estWeight := estimate()
	for i := 0; i < N; i++ {
		items[i] = Item{index: i, weight: estWeight[i]}
	}
	return
}

func estimate() []int {
	//for i := 0; i < N; i++ {
	//log.Println(balanceCache[i][:N])
	//}

	// 一対一の比較のみで、アイテムの大きさを推定する
	//closure := transitiveClosure(balanceCache)
	//reduction := transitiveReduction(balanceCache, closure)
	//for i := 0; i < N; i++ {
	//log.Println(closure[i][:N])
	//}
	//	for i := 0; i < N; i++ {
	//log.Println(reduction[i][:N])
	//}
	var greaterList [][2]int
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			//if reduction[i][j] == GreaterThan {
			if balanceCache[i][j] == GreaterThan {
				// i > j
				greaterList = append(greaterList, [2]int{i, j})
			}
		}
	}
	//log.Println(greaterList)
	estWeight := make([]int, N)
	for i := 0; i < N; i++ {
		estWeight[i] = generateItemWeight()
	}
	for i := 0; i < 100; i++ {
		var swapCnt int
		for {
			swap := false
			for i := 0; i < len(greaterList); i++ {
				a, b := greaterList[i][0], greaterList[i][1]
				//a > b
				if estWeight[a] < estWeight[b] {
					estWeight[a], estWeight[b] = estWeight[b], estWeight[a]
					swap = true
				}
			}
			if !swap {
				break
			}
			swapCnt++
		}
		//log.Println("swapCnt", swapCnt)
		if len(balanceItemsCacheBit) == 0 {
			break
		}
		var unMatchCnt int
		delta := 100
		for i := 0; i < 10; i++ {
			for k, v := range balanceItemsCacheBit {
				a, b := Decode(&k)
				aWeight := 0
				bWeight := 0
				for i := 0; i < len(a); i++ {
					aWeight += estWeight[a[i]]
				}
				for i := 0; i < len(b); i++ {
					bWeight += estWeight[b[i]]
				}
				//log.Println(len(a), aWeight, len(b), bWeight, v)
				if aWeight > bWeight {
					// 実際は a<=b のとき
					if v != GreaterThan {
						unMatchCnt++
						for i := 0; i < len(a); i++ {
							estWeight[a[i]] -= delta
						}
						for i := 0; i < len(b); i++ {
							estWeight[b[i]] += delta
						}
					}
				} else if aWeight < bWeight {
					// 実際は a>=b のとき
					if v != LessThan {
						unMatchCnt++
						//	log.Println("unmatch", k, v)
						for i := 0; i < len(a); i++ {
							estWeight[a[i]] += delta
						}
						for i := 0; i < len(b); i++ {
							estWeight[b[i]] -= delta
						}
					}
				}
			}
		}
		//log.Println(i, "unMatchCnt", unMatchCnt)
		if unMatchCnt == 0 && swapCnt == 0 {
			break
		}
	}
	minWeight := 1000000000
	for i := 0; i < N; i++ {
		minWeight = minInt(minWeight, estWeight[i])
	}
	if minWeight < 0 {
		for i := 0; i < N; i++ {
			estWeight[i] += -minWeight * 2
		}
	}

	return estWeight
}

func generateItemWeight() int {
	lambda := math.Pow10(-5)
	maxAllowedValue := N * int(math.Pow10(5)) / D
	var rawValue float64
	for {
		rawValue = rand.ExpFloat64() / lambda
		if rawValue <= float64(maxAllowedValue) {
			break
		}
	}
	return int(math.Max(1, math.Round(rawValue)))
}

type Node struct {
	index int
	left  *Node
	right *Node
}

func merge(left, right *Node) *Node {
	cmp, err := balance(left.index, right.index, &Qleft)
	if err != nil {
		log.Fatal(err)
	}
	// 小さい方を親にする
	if cmp == GreaterThan {
		left, right = right, left
	}
	parentIndex := left.index
	return &Node{index: parentIndex, left: left, right: right}
}

func createTournament(arr []int) *Node {
	nodes := make([]*Node, len(arr))
	for i, v := range arr {
		nodes[i] = &Node{index: v}
	}
	for len(nodes) > 1 {
		tmpNodes := make([]*Node, 0)
		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				parent := merge(nodes[i], nodes[i+1])
				tmpNodes = append(tmpNodes, parent)
			} else {
				tmpNodes = append(tmpNodes, nodes[i])
			}
		}
		nodes = tmpNodes
	}
	return nodes[0]
}
func findMinimum(node *Node) int {
	if node.left == nil && node.right == nil {
		return node.index
	}
	if node.left != nil && node.index == node.left.index {
		return findMinimum(node.left)
	} else {
		return findMinimum(node.right)
	}
}

func findMinimums(n int) []int {
	items := make([]int, N)
	for i := 0; i < N; i++ {
		items[i] = i
	}
	n = minInt(n, N)
	minList := make([]int, n)
	for i := 0; i < n; i++ {
		tournament := createTournament(items)
		minList[i] = findMinimum(tournament)
		//log.Println(items)
		idx := slices.Index(items, minList[i])
		items = UnstableDelete(items, idx)
	}
	return minList
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func UnstableDelete(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func ScaleRengeBugs(bags [][]int, scale []int) int {
	for i := 0; i < len(scale); i++ {
		if slices.Contains(bags[0], scale[i]) {
			continue
		}
		if slices.Contains(bags[len(bags)-1], scale[i]) {
			continue
		}
		bags[0] = append(bags[0], scale[i])
		cmp, err := balanceItems(bags[0], bags[len(bags)-1], &Qleft)
		if err != nil {
			log.Fatal(err)
		}
		if cmp == GreaterThan {
			return scale[i]
		}
		bags[0] = bags[0][:len(bags[0])-1]
	}
	return -1
}
