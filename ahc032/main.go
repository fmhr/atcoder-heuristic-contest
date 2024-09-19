package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"sort"
	"time"
	"unsafe"
)

// ./a.out -cpuprofile cpu.prof < tools/in/0000.txt > out.txt
// ./a.out -cpuprofile cpu.prof -memprofile mem.prof < tools/in/0000.txt > out.txt
// go tool pprof -http=localhost:8888 main cpu.prof
// go tool pprof -http=localhost:8888 main mem.prof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

// StartCPUProfile は、CPUプロファイルを開始する
func StartCPUProfile() func() {
	if *cpuprofile == "" {
		return func() {}
	}
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		log.Fatal("could not start CPU profile: ", err)
	}

	return func() {
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			log.Fatal("could not stop CPU profile: ", err)
		}
	}
}

// writeMemProfile は、メモリプロファイルを書き込む
func writeMemProfile() {
	if *memprofile == "" {
		return
	}
	f, err := os.Create(*memprofile)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}

func flagCheck() {
	flag.Parse()
	if _, atcoder := os.LookupEnv("ATCODER"); atcoder {
		log.SetOutput(io.Discard)
		return
	}

	log.SetFlags(log.Lshortfile)
	//runtime.GOMAXPROCS(1) // 並列処理を抑制
	//debug.SetGCPercent(2000) // GCを抑制 2000% に設定
	//debug.SetGCPercent(-1) // GCを停止
	rand.Seed(1) // 乱数のシードを固定することで、デバッグ時に再現性を持たせる
}

func main() {
	flagCheck()
	if *cpuprofile != "" {
		stopCPUProfile := StartCPUProfile()
		defer stopCPUProfile()
	}
	if *memprofile != "" {
		defer writeMemProfile()
	}
	// --- start
	startTime := time.Now()
	//solve(input)
	beamSearch()
	elapseTime := time.Since(startTime)
	log.Printf("time=%f", float64(elapseTime)/float64(time.Second))
}

const (
	N = 9
	M = 20
	K = 81
)

type Input struct {
	N, M, K int
	board   [9][9]int32
	stamps  [20][9]int32
}

const (
	MOD = 998244353
)

func read() (in Input) {
	fmt.Scan(&in.N, &in.M, &in.K)
	for i := 0; i < in.N; i++ {
		for j := 0; j < in.N; j++ {
			fmt.Scan(&in.board[i][j])
		}
	}
	for i := 0; i < in.M; i++ {
		for j := 0; j < 9; j++ {
			fmt.Scan(&in.stamps[i][j])
		}
	}
	log.Printf("N=%d M=%d K=%d\n", in.N, in.M, in.K)
	return in
}

type State struct {
	score   int
	board   [9][9]int32
	actNode *TraceNode[Act]
}

func (s State) Less(other State) bool {
	return s.score < other.score
}

func (s State) outputAns() {
	root := (s.actNode).Get(nil)
	log.Println(root)
	fmt.Println(len(root))
	for i := 0; i < len(root); i++ {
		fmt.Println(root[i].n, root[i].p, root[i].q)
	}
}

type Act struct {
	n, p, q int16
}

func score(a [9][9]int32) (score int) {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			score += int(a[i][j] % MOD)
		}
	}
	return score
}

func (s State) Clone() (newState State) {
	newState.score = s.score
	newState.actNode = s.actNode
	newState.board = s.board
	return newState
}

// 新しい有効なポイントを追加する
func (s *State) Do(stm [9]int32, p, q int, newActive [][2]int) {
	for i := 0; i < 9; i++ {
		s.board[p+i/3][q+i%3] += stm[i]
		s.board[p+i/3][q+i%3] %= MOD
	}
	for i := 0; i < len(newActive); i++ {
		s.score += int(s.board[p+newActive[i][0]][q+newActive[i][1]])
	}
}

// すでに追加済みのポイントに対して再度行動を行う
func (s *State) ReDo(stm [9]int32, p, q int, newActive [][2]int) {
	for i := 0; i < len(newActive); i++ {
		s.score -= int(s.board[p+newActive[i][0]][q+newActive[i][1]])
	}
	for i := 0; i < 9; i++ {
		s.board[p+i/3][q+i%3] += stm[i]
		s.board[p+i/3][q+i%3] %= MOD
	}
	for i := 0; i < len(newActive); i++ {
		s.score += int(s.board[p+newActive[i][0]][q+newActive[i][1]])
	}
}

// スタンプを取り消す
func (s *State) Undo(stm [9]int32, p, q int) {
	for i := 0; i < 9; i++ {
		s.board[p+i/3][q+i%3] += MOD // マイナスになるのを避けるために一度足す
		s.board[p+i/3][q+i%3] -= stm[i]
		s.board[p+i/3][q+i%3] %= MOD
	}
}

var DoubleStamp [400][9]int32
var TripleStamp [8000][9]int32

// generateAction は、次の行動を生成する. 1~3回行動
func generateAction(in Input) {
	index2 := 0
	index3 := 0
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			for m := 0; m < 9; m++ {
				DoubleStamp[index2][m] = (in.stamps[i][m] + in.stamps[j][m]) % MOD
			}
			index2++
			for k := 0; k < 20; k++ {
				for l := 0; l < 9; l++ {
					TripleStamp[index3][l] = (in.stamps[i][l] + in.stamps[j][l] + in.stamps[k][l]) % MOD
				}
				index3++
			}
		}
	}
}

func (s *State) generateNexts(in Input, y, x int, stmCnt int, actTree *Trace[Act], nexts *BeamSearchStates[State], newActive [][2]int) {
	new := s.Clone()
	preScore := new.score
	//nexts.Insert(new) // なにもしない
	for i := 0; i <= M; i++ {
		if i == M {
			// 何もしない
			new.Do([9]int32{}, y, x, newActive)
			//new.actNode = actTree.Add(Act{n: -1, p: int16(y), q: int16(x)}, s.actNode.Parent)
			nexts.Insert(new)
			new.score = preScore
			new.actNode = s.actNode
			continue
		}
		new.Do(in.stamps[i], y, x, newActive)
		new.actNode = actTree.Add(Act{n: int16(i), p: int16(y), q: int16(x)}, s.actNode)
		nexts.Insert(new)
		// 2回行動
		if stmCnt > 1 && i < M {
			preScore2 := new.score
			preNode := new.actNode
			for j := 0; j < M; j++ {
				new.ReDo(in.stamps[j], y, x, newActive)
				new.actNode = actTree.Add(Act{n: int16(j), p: int16(y), q: int16(x)}, preNode)
				nexts.Insert(new)
				new.Undo(in.stamps[j], y, x)
				new.score = preScore2
			}
		}
		new.Undo(in.stamps[i], y, x)
		new.score = preScore
		new.actNode = s.actNode
	}
	// TODO 更新した点が９０点以下の時は、もう一度行動する?
}

func generateNextFullState(states []State, in Input, y, x int, actTree *Trace[Act], nexts *BeamSearchStates[State], newActive [][2]int) {
	var stmpCnt int = 1
	if y >= 6 || x >= 6 {
		stmpCnt = 2
	}
	for i := 0; i < len(states); i++ {
		states[i].generateNexts(in, y, x, stmpCnt, actTree, nexts, newActive)
	}
}

// beam search
const (
	beamWidth = 1000
)

// beamSearch はビームサーチを行う。スコアはそれまでに最適化したマスに対するスコア
func beamSearch() {
	var newActive [4][][2]int
	newActive[0] = append(newActive[0], [2]int{0, 0})
	newActive[1] = append(newActive[1], [2]int{0, 0}, [2]int{0, 1}, [2]int{0, 2})
	newActive[2] = append(newActive[2], [2]int{0, 0}, [2]int{1, 0}, [2]int{2, 0})
	newActive[3] = append(newActive[3], [2]int{0, 0}, [2]int{0, 1}, [2]int{0, 2}, [2]int{1, 0}, [2]int{1, 1}, [2]int{1, 2}, [2]int{2, 0}, [2]int{2, 1}, [2]int{2, 2})
	currentBeam := NewBeamSearccArray[State]()
	nextState := NewBeamSearccArray[State]()
	in := read()
	currentBeam.Insert(State{score: 0, board: in.board})
	loops := generateLoop() // グリッドを回る順番
	loops = append(loops, [2]int{N - 3, N - 3})
	actTree := NewTrace[Act]()
	currentBeam.array[0].actNode = nil
	for _, ij := range loops {
		y := ij[0]
		x := ij[1]
		if y < 6 && x < 6 {
			// 1マスに対する最適化
			generateNextFullState(currentBeam.List(), in, y, x, actTree, nextState, newActive[0])
		} else if x == 6 && y < 6 {
			// スタンプの上３マスに対する最適化
			generateNextFullState(currentBeam.List(), in, y, x, actTree, nextState, newActive[1])
		} else if y == 6 && x < 6 {
			// スタンプの左３マスに対する最適化
			generateNextFullState(currentBeam.List(), in, y, x, actTree, nextState, newActive[2])
		} else if y == 6 && x == 6 {
			// 最後の9*9マスに対する最適化
			// TODO　ここだけ別にゲーム木をつくれる？
			generateNextFullState(currentBeam.List(), in, y, x, actTree, nextState, newActive[3])
		} else {
			panic("invalid")
		}
		currentBeam, nextState = nextState, currentBeam
		nextState.Reset()
	}
	currentBeam.array[0].outputAns()
	log.Printf("stmps=%d", len(currentBeam.array[0].actNode.Get(nil)))
	log.Printf("score=%d", score(currentBeam.array[0].board))
}

// generateLoop はループを生成する
// n*n　のグリッドに対して、左上から右下に向かってループを生成する
func generateLoop() (ij [][2]int) {
	for s := 0; s < N-3; s++ {
		for j := s; j < N-2; j++ {
			ij = append(ij, [2]int{s, j})
		}
		for i := s + 1; i < N-2; i++ {
			ij = append(ij, [2]int{i, s})
		}
	}
	return ij
}

// Trace ビームサーチにおいて、行動履歴のコピーが重いので、Stateとは別に行動履歴を木構造で持つ
type TraceNode[T any] struct {
	Value  T
	Parent *TraceNode[T]
}

type Trace[T any] struct {
	Root *TraceNode[T]
}

func NewTrace[T any]() *Trace[T] {
	return &Trace[T]{Root: nil}
}

// Add は、ノードを追加する parentがnilのときはrootになる
func (t *Trace[T]) Add(value T, parent *TraceNode[T]) *TraceNode[T] {
	node := &TraceNode[T]{Value: value, Parent: parent}
	if parent == nil {
		t.Root = node
	}
	return node
}

// Get は、ノードからルートまでの値を取得する rootのときはnilを入れる
func (n *TraceNode[T]) Get(root *TraceNode[T]) []T {
	var rtn []T
	for n != root {
		rtn = append(rtn, n.Value)
		n = n.Parent
	}
	reverse(rtn)
	return rtn
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// ------------------------------------------------------------------------------------------
// HasLess は、比較関数を持つインターフェース
type HasLess[T any] interface {
	Less(other T) bool
}

const (
	arraySize = beamWidth * 2
)

// BeamSearchStates はビームサーチで使用する固定サイズの配列を管理します。
// 配列は最大で beamWidth * 2 の要素を持ち、配列が満杯になるとソートされ、
// スコアが高い上位 beamWidth の要素が保持されます。
type BeamSearchStates[T HasLess[T]] struct {
	array     [arraySize]T
	size      int
	threshold *T // 閾値となる最小スコアをもつ要素へのポインタ
}

// NewBeamSearccArray は、新しいビームサーチ用の配列を作成します。
func NewBeamSearccArray[T HasLess[T]]() *BeamSearchStates[T] {
	return &BeamSearchStates[T]{size: 0}
}

// CanInsert は要素 t が配列に追加可能かどうかを判断します。
// minimumScore が nil の場合、または t のスコアが minimumScore よりも高い場合に追加可能です。
func (bsa *BeamSearchStates[T]) CanInsert(t T) bool {
	return (bsa.threshold == nil || !t.Less(*bsa.threshold))
}

// Insert は要素 t を配列に追加します。
// 配列が最大容量に達すると、配列はソートされ、スコアが低い半分が削除されます。
// 新しい最小スコアが更新されます。
func (bsa *BeamSearchStates[T]) Insert(t T) {
	if bsa.CanInsert(t) {
		bsa.array[bsa.size] = t
		bsa.size++
		if bsa.size == arraySize {
			sort.Slice(bsa.array[:], func(i, j int) bool {
				return !bsa.array[i].Less(bsa.array[j])
			})
			bsa.size = beamWidth // 配列サイズを半分に減らす
			bsa.threshold = &bsa.array[bsa.size-1]
		}
	}
}

// List は配列の要素をスコアの高い順にソートして返します。
func (bsa *BeamSearchStates[T]) List() []T {
	sort.Slice(bsa.array[:bsa.size], func(i, j int) bool {
		return !bsa.array[i].Less(bsa.array[j])
	})
	return bsa.array[:bsa.size]
}

// Initialize は配列を初期化します。
func (bsa *BeamSearchStates[T]) Initialize() {
	bsa.array = [arraySize]T{}
	bsa.size = 0
	bsa.threshold = nil
}

// Reset は配列はそのままに、サイズと閾値をリセットします。
// Initializeよりも速いはず
func (bsa *BeamSearchStates[T]) Reset() {
	bsa.size = 0
	bsa.threshold = nil
}

func Copy[T HasLess[T]](dst, src *BeamSearchStates[T]) {
	copy(dst.array[:], src.array[:])
	dst.size = src.size
	if src.threshold != nil {
		// threshold が nil でない場合、ポインタの位置からインデックスを計算してコピーする
		minScoreIndex := int((uintptr(unsafe.Pointer(src.threshold)) - uintptr(unsafe.Pointer(&src.array[0]))) / unsafe.Sizeof(src.array[0]))
		dst.threshold = &dst.array[minScoreIndex]
	} else {
		dst.threshold = nil
	}
}

// ------------------------------------------------------------------------------------------
// binaryTreeでindexを管理して、
// 二つの[T]beamWidthの配列を持ちBeamSsarchを作る

type Item[T any] struct {
	value *T
	less  func(T, T) bool // 比較関数
}

type PriorityQueue[T any] struct {
	arr  [beamWidth]Item[T] // arrayのindexを管理
	size int
}

func (pq *PriorityQueue[T]) Insert(value T) {
	if pq.size < beamWidth {
		pq.arr[pq.size] = Item[T]{value: &value, less: pq.arr[0].less}
		pq.size++
		pq.heapifyUp()
	} else if pq.arr[0].less(value, *pq.arr[0].value) {
		pq.arr[0].value = &value

	}
}

func (pq *PriorityQueue[T]) heapifyUp() {
	index := pq.size - 1
	for index > 0 {
		parent := (index - 1) / 2
		if pq.arr[parent].less(*pq.arr[index].value, *pq.arr[parent].value) {
			pq.arr[parent], pq.arr[index] = pq.arr[index], pq.arr[parent]
			index = parent
		} else {
			break
		}
	}
}
