package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
)

// + 問題に関すること
// 常に手持ちのカード枚数とプロジェクトの数は維持される
// newCardは毎ターンランダムに生成される

// + 解法に活かせそうなこと
// Investカードがでたら、すぐに買いたい。moneyを買える程度になるまでためる
// Investのコストは、平均600 * 2^L
// カードのコスパ　カードで減らせる仕事量/カードのコスト
// コスパはLに関係ない　仕事量とコストは *2^L で増加する
// カードは必ず１枚使う
// 最初の100ターンぐらいは,隠れパラメータのXiがわからないので、ランダムにカードを生成しても精度が低いので、貪欲法がいいかも
// WorkSingleVlaue = work/price
// WorkAllValue = work/price*M(プロジェクトの数)
// 取るだけマイナスなWorkカードが存在する(コスパの悪すぎ)
// ++ project
// プロジェクトのコスパ　プロジェクトの価値/プロジェクトの仕事量
// ProjectのcospaはLに依存せず 0.3~2.8 average 1.07 median 1.0
// コスパの悪いプロジェクトは、CancelSingle, CancelAllで消したい
// 手元にCancelカードが多くてもメリットはすくない
// 選べるカードの数が多いほど有利
// 手札とプロジェクトの数は選べるカード数よりは影響が小さい

// + 評価値
// 手持ちのカードのコスパを大きくしたい
// 受理してるプロジェクトのコスパを大きくしたい
// 最終的にはmoneyの最大化。Investカードを買える手持ちは欲しい
// L(Invest カードの使用回数)を大きくしたい ただし、最終盤に大きくしても効果は薄い
// 評価値がうまく決まるならビームサーチが良さそう
// 決まらない時は、モンテカルロ法が効く
// プロジェクトの評価値を考える
// カード（コストー＞仕事量）プロジェクト（仕事量ー＞報酬）
// カードの下限コスパを1.5 x プロジェクト が1.0を上回りたい
// kが大きい時、期待値は大きい

// + パラメータ
var WorkCardCospaMin float64
var CancelProjectCospaMin float64 // projectのコスパとWorkCardCospaMinの積がこれより小さいときは、CancelSingleカードを買う
var CardWeight float64
var ProjectWeight float64
var InvestWeight float64

func getArguments() {
	flag.Float64Var(&WorkCardCospaMin, "ws", 1.4546, "WorkSingleCospaLine")
	flag.Float64Var(&CancelProjectCospaMin, "cp", 1.072, "CancelProjectCospaLine")
	flag.Parse()
}

func main() {
	log.SetFlags(log.Lshortfile)
	getArguments()
	//log.SetOutput(ioutil.Discard) // ログを出力しない
	rand.Seed(0)
	solver(readInput())
}

func readInput() (int, int, int, int, []Card, []Project) {
	var N, M, K, T int
	fmt.Scan(&N, &M, &K, &T)
	cards := make([]Card, N)
	for i := 0; i < N; i++ {
		fmt.Scan(&cards[i].Type, &cards[i].Workforce)
	}
	projects := make([]Project, M)
	for i := 0; i < M; i++ {
		fmt.Scan(&projects[i].Workload, &projects[i].Value)
	}
	log.Printf("N=%v, M=%v, K=%v, T=%v\n", N, M, K, T)

	// Kに基づいて、調整
	return N, M, K, T, cards, projects
}

type CardType int

const (
	WorkSingle CardType = iota
	WorkAll
	CancelSingle
	CancelAll
	Invest
)

var cardTypeString = []string{"WorkSingle", "WorkAll", "CancelSingle", "CancelAll", "Invest"}

type State struct {
	Money    int
	L        int
	Cards    []Card
	Projects []Project
}

func (s State) Value() (v float64) {
	v += float64(s.Money)
	for i := 0; i < len(s.Cards); i++ {
		v += float64(s.Cards[i].Workforce) * CardWeight
	}
	for i := 0; i < len(s.Projects); i++ {
		v += float64(s.Projects[i].Value) * ProjectWeight
	}
	v += float64(s.L) * InvestWeight
	return
}

// Cards[c]をProjects[p]に適用する
func (s *State) Move(ci, pi, ki int, cnt [5]int) {
	c := s.Cards[ci]
	if s.Cards[ci].Type == WorkSingle {
		s.Projects[pi].Workload -= c.Workforce
		if s.Projects[pi].Workload <= 0 {
			// プロジェクトが完了したら、報酬を得て、新しいプロジェクトを生成する
			s.Money += s.Projects[pi].Value
			s.Projects[pi] = generateProject(s.L)
		}
	} else if s.Cards[ci].Type == WorkAll {
		for i := 0; i < len(s.Projects); i++ {
			s.Projects[i].Workload -= c.Workforce
			if s.Projects[i].Workload <= 0 {
				// プロジェクトが完了したら、報酬を得て、新しいプロジェクトを生成する
				s.Money += s.Projects[i].Value
				s.Projects[i] = generateProject(s.L)
			}
		}
	} else if s.Cards[ci].Type == CancelSingle {
		s.Projects[pi] = generateProject(s.L)
	} else if s.Cards[ci].Type == CancelAll {
		for i := 0; i < len(s.Projects); i++ {
			s.Projects[i] = generateProject(s.L)
		}
	} else if s.Cards[ci].Type == Invest {
		s.L++
	}
	s.Cards[ci] = generateCard(s.L, len(s.Projects), cnt[0], cnt[1], cnt[2], cnt[3], cnt[4])
}

type Card struct {
	Index     int
	Type      CardType
	Workforce int
	Cost      int
}

func (c Card) String() string {
	return fmt.Sprintf("%d %v %v %v", c.Index, cardTypeString[c.Type], c.Workforce, c.Cost)
}

type Project struct {
	Index    int
	Workload int
	Value    int
}

func solver(N, M, K, T int, cards []Card, projects []Project) int {
	newCards := make([]Card, K)
	for i := 0; i < K; i++ {
		newCards[i].Index = i
	}
	var money, L, Lfail int
	var CardCounter [5]int
	for i := 0; i < T; i++ {
		//log.Println("turn:", i, "money:", money, "L:", L)
		// request
		useCard, toProject := selectMove(N, M, cards, projects)
		if cards[useCard].Type == Invest {
			L++
		}
		fmt.Println(useCard, toProject)
		CardCounter[cards[useCard].Type]++
		// response
		// update projects
		for i := 0; i < M; i++ {
			fmt.Scan(&projects[i].Workload, &projects[i].Value)
		}
		// update money
		fmt.Scan(&money)
		for k := 0; k < K; k++ {
			fmt.Scan(&newCards[k].Type, &newCards[k].Workforce, &newCards[k].Cost)
		}
		// select card from newCards
		selectCard := selectNewCard(i, M, K, L, money, newCards, projects, &Lfail)
		fmt.Println(selectCard)
		cards[useCard] = newCards[selectCard]
	}

	//time.Sleep(100 * time.Millisecond) // testerのScore=出力に干渉してしまうので、少し待つ
	// tools/src/bin/tester.rs を修正して、子プロセスの終了を待つように修正すれば、↑のsleepは不要
	log.Printf("money=%v, L=%v, Lfail=%d\n", money, L, Lfail)
	log.Println("CardCounter:", CardCounter)
	log.Printf("cWS=%v, cWA=%v, cCS=%v, cCA=%v, cInvest=%v\n", CardCounter[WorkSingle], CardCounter[WorkAll], CardCounter[CancelSingle], CardCounter[CancelAll], CardCounter[Invest])
	return 0
}

// selectMove はカードとプロジェクトから、カードとプロジェクトを選択する
// カードは、WorkSingle, CancelSingle の場合は、プロジェクトを選択する
// WorkAll, CancelAll, Invest の場合は、プロジェクトは選択しない(0を返す)
func selectMove(N, M int, cards []Card, projects []Project) (int, int) {
	c := rand.Intn(N)
	var m int
	switch cards[c].Type {
	case WorkSingle:
		// WorkSingleのときは、カードのworkforceが無駄にならないprojectを選ぶ
		// プロジェクトのValue/Workloadが小さいものは捨てる前提で選ばないのもあり
		// Value/Workloadが大きくて、かつ無駄にならないもの
		maxValuePerWorkload := 0.0
		maxIndex := 0
		for i := 0; i < M; i++ {
			if float64(projects[i].Value)/float64(projects[i].Workload) > maxValuePerWorkload {
				maxValuePerWorkload = float64(projects[i].Value)/float64(projects[i].Workload) + math.Min(float64(projects[i].Workload), float64(cards[c].Workforce))
				maxIndex = i
			}
		}
		return c, maxIndex
	case CancelSingle:
		// CancelSingleのときは、プロジェクトのValue/Workloadが最小のものを選ぶ
		minValuePerWorkload := math.MaxFloat64
		minIndex := 0
		for i := 0; i < M; i++ {
			if float64(projects[i].Value)/float64(projects[i].Workload) < minValuePerWorkload {
				minValuePerWorkload = float64(projects[i].Value) / float64(projects[i].Workload)
				minIndex = i
			}
		}
		return c, minIndex
	case WorkAll, CancelAll, Invest:
		m = 0
	}
	return c, m
}

func selectNewCard(turn, M, K, L, money int, cards []Card, projects []Project, Lfail *int) int {
	// 900 ターンを超えたら、money節約のためにコスト0のカードを買う
	if turn > 900 {
		return 0
	}
	// Investカードが買えれば買う
	for i := 0; i < K; i++ {
		if cards[i].Type == Invest && L < 20 {
			if cards[i].Cost <= money {
				return i
			} else {
				*Lfail++
			}
		}
	}

	var selectCard Card
	var cospa float64
	for i := K - 1; i > 0; i-- {
		if cards[i].Type == CancelSingle || cards[i].Type == CancelAll {
			continue
		}
		if cards[i].Cost > money {
			continue
		}
		if cards[i].Type == WorkAll {
			csp := (float64(cards[i].Workforce) / float64(cards[i].Cost)) * float64(M)
			if csp > WorkCardCospaMin && csp > cospa {
				cospa = csp
				selectCard = cards[i]
			}
		}
		if cards[i].Type == WorkSingle {
			csp := float64(cards[i].Workforce) / float64(cards[i].Cost)
			if csp > WorkCardCospaMin && csp > cospa {
				cospa = csp
				selectCard = cards[i]
			}
		}
	}
	// コスパの低いプロジェクトと、CancelSingleカードがあれば、CancelSingleカードを買う
	if cospa < WorkCardCospaMin {
		miniCospa := math.MaxFloat64
		// コスパの低いプロジェクトを探す
		for i := 0; i < M; i++ {
			csp := float64(projects[i].Value) / float64(projects[i].Workload)
			if csp < miniCospa {
				miniCospa = csp
			}
		}
		if miniCospa*WorkCardCospaMin < CancelProjectCospaMin && miniCospa != math.MaxFloat64 {
			for i := 0; i < K; i++ {
				if cards[i].Type == CancelSingle && cards[i].Cost <= money {
					return i
				}
			}
		}
	}
	return selectCard.Index
}

// モンテカルロ法で、最適なカードを選択する
func monteCarlo(N, M, K, t int, s State, playout int) {

}

func generateCard(L, M, x0, x1, x2, x3, x4 int) Card {
	sumX := x0 + x1 + x2 + x3 + x4
	randX := rand.Intn(sumX)
	if randX < x0 {
		w := float64(rand.Intn(50)+1) * math.Pow(2, float64(L))
		wPrime := w / math.Pow(2, float64(L))
		clampedGauss := math.Min(math.Max(gaussian(wPrime, wPrime/3), 1), 10000)
		p := clampedGauss * math.Pow(2, float64(L))
		return Card{Type: WorkSingle, Workforce: int(math.Round(w)), Cost: int(math.Round(p))}
	} else if randX < x0+x1 {
		w := float64(rand.Intn(50)+1) * math.Pow(2, float64(L))
		wPrime := w / math.Pow(2, float64(L))
		clampedGauss := math.Min(math.Max(gaussian(wPrime*float64(M), wPrime/(float64(M)*3)), 1), 10000)
		p := clampedGauss * math.Pow(2, float64(L))
		return Card{Type: WorkAll, Workforce: int(math.Round(w)), Cost: int(math.Round(p))}
	} else if randX < x0+x1+x2 {
		p := float64(rand.Intn(10)) * math.Pow(2, float64(L))
		return Card{Type: CancelSingle, Workforce: 0, Cost: int(math.Round(p))}
	} else if randX < x0+x1+x2+x3 {
		p := float64(rand.Intn(10)) * math.Pow(2, float64(L))
		return Card{Type: CancelAll, Workforce: 0, Cost: int(math.Round(p))}
	} else {
		p := float64(rand.Intn(1000-200)+200) * math.Pow(2, float64(L))
		return Card{Type: Invest, Workforce: 0, Cost: int(math.Round(p))}
	}
}

func generateProject(L int) Project {
	b := rand.Float64()*(8.0-2.0) + 2.0
	h := math.Round(math.Pow(2, b)) * math.Pow(2, float64(L))
	v := math.Round(math.Pow(2, math.Max(0, math.Min(gaussian(b, 0.5), 10.0))*math.Pow(2, float64(L))))
	return Project{Workload: int(math.Round(h)), Value: int(math.Round(v))}
}

// ガウス分布に基づいた乱数を生成する
func gaussian(meam, stddev float64) float64 {
	return rand.NormFloat64()*stddev + meam
}
