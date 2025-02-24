package main

import (
	"log"
	"sort"
	"strings"
	"time"
)

const (
	BEAM_WIDHT_X_ACTIONS = 20000
)

// すべての駅の場所と、それらをつなぐエッジを行動にする
func beamSearch(in Input) string {
	var cnt int // 探索空間のカウント
	// 駅の位置を選ぶ
	stations := ChooseStationPositionFast(in)
	log.Printf("stations=%v\n", len(stations))
	// 駅を繋ぐエッジを求める
	edges, _ := constructMSTRailway(in, stations)

	allAction := make([]bsAction, 0, len(stations)+len(edges))
	for _, s := range stations {
		allAction = append(allAction, bsAction{path: []Pos{s}, typ: []int{STATION}})
	}
	for _, e := range edges {
		allAction = append(allAction, bsAction{path: e.Path, typ: e.Rail})
		//log.Println(railToString(e.Rail))
	}
	log.Println("actionNum", len(allAction))
	initialState := newBsState(&in, len(allAction))
	beamWidth := BEAM_WIDHT_X_ACTIONS / len(allAction)
	log.Printf("Width=%d\n", beamWidth)
	beamWidth = intMax(beamWidth, 5)
	beamStates := make([]bsState, 0, beamWidth)
	beamStates = append(beamStates, *initialState)
	nextStates := make([]bsState, 0, beamWidth)
	bestState := initialState.Clone()
	var loop int
	var timeout bool
	for len(beamStates) > 0 {
		for i := 0; i < minInt(beamWidth, len(beamStates)); i++ {
			if beamStates[i].state.turn > 800 {
				if beamStates[i].state.score > bestState.state.score {
					bestState = beamStates[i].Clone()
				}
				continue
			}
			// DO_NOTHINGの場合
			//newState := beamStates[i].Clone()
			//err := newState.state.do(Action{Kind: DO_NOTHING}, in)
			//if err != nil {
			//panic(err)
			//}
			//nextStates = append(nextStates, *newState)
		NEWSTATE:
			for j := 0; j < len(beamStates[i].restActions); j++ {
				acts := allAction[beamStates[i].restActions[j]]
				// costの確認
				// actionを精査（駅がすでにあるなら除く)
				if len(acts.path) != len(acts.typ) {
					panic("invalid action")
				}
				// 不要なactionをのぞいたp,tを作る
				p := make([]Pos, 0, len(acts.path))
				t := make([]int, 0, len(acts.typ))
				tmp := beamStates[i].state.field // これはcloneしていないので使わない
				for k := 0; k < len(acts.path); k++ {
					if acts.typ[k] == tmp.cell[acts.path[k].Y][acts.path[k].X] {
						continue
					}
					if isRail(acts.typ[k]) && tmp.cell[acts.path[k].Y][acts.path[k].X] == STATION {
						continue
					}
					if isRail(acts.typ[k]) && isRail(tmp.cell[acts.path[k].Y][acts.path[k].X]) {
						// 両方線路で種類が違う時
						break NEWSTATE
					}
					p = append(p, acts.path[k])
					t = append(t, acts.typ[k])
				}
				if len(p) == 0 {
					continue
				}
				// 孤立するのは作らない
				isolated := true
				if len(beamStates[i].state.field.stations) > 0 {
					for l := 0; l < len(p) && isolated; l++ {
						for d := 0; d < 4; d++ {
							y, x := int(p[l].Y)+int(dy[d]), int(p[l].X)+int(dx[d])
							if y < 0 || y >= 50 || x < 0 || x >= 50 {
								continue
							}
							if checkConnec(tmp.cell[p[l].Y][p[l].X], d, true) && checkConnec(tmp.cell[y][x], d, false) {
								isolated = false
								break
							}
						}
					}
					if isolated {
						continue
					}
				}
				costMoney := calBuildCost(t) //純粋なコスト(money)
				if beamStates[i].state.money < costMoney && beamStates[i].state.income == 0 {
					// お金が足りない＋収入がない時はスキップ
					continue
				}
				// DO_NOTHINGで必要な分だけ待つ
				costTime := 0
				if beamStates[i].state.income > 0 {
					costTime = costMoney / beamStates[i].state.income // incomeを考慮したコスト
				}
				////log.Println(len(p), "DoNothing", costTime-len(p))
				//// 残されたターン数で実行できない時
				if beamStates[i].state.turn+costTime > 800 {
					continue
				}

				newState := beamStates[i].Clone()
				cnt++
				for costMoney > newState.state.money {
					// 必要なお金が貯まるまで待つ
					err := newState.state.do(Action{Kind: DO_NOTHING}, in, false)
					if err != nil {
						panic(err)
					}
					if newState.state.turn > 800 {
						log.Println("over time 800:", newState.state.turn)
						panic("over time")
					}
				}
				if newState.state.money < costMoney {
					log.Println("----------------------------")
					log.Println("actions", railToString(t))
					log.Println("costMoney", costMoney, "costTime", costTime)
					log.Println("money", newState.state.money, "income", newState.state.income)
					log.Println("turn", newState.state.turn)
					panic("not enough money")
				}
				for j := 0; j < len(p); j++ {
					last := false
					if j == len(p)-1 {
						last = true
					}
					err := newState.state.do(Action{Kind: t[j], Y: int(p[j].Y), X: int(p[j].X)}, in, last)
					if err != nil {
						log.Println("j", j)
						log.Println("actions", railToString(t))
						log.Println("posision", p[j])
						log.Println("action", t[j], p[j])
						panic(err)
					}
				}
				// delete action
				newState.restActions = append(newState.restActions[:j], newState.restActions[j+1:]...)
				nextStates = append(nextStates, *newState)
				if newState.state.score > bestState.state.score {
					bestState = newState.Clone()
				}
			}

		}
		//log.Println("nextStates", len(nextStates))
		sort.Slice(nextStates, func(i, j int) bool {
			if nextStates[i].state.score == nextStates[j].state.score {
				return nextStates[i].state.income > nextStates[j].state.income
			}
			return nextStates[i].state.score > nextStates[j].state.score
		})
		//if len(nextStates) > 0 {
		//log.Println("score:", nextStates[0].state.score, nextStates[len(nextStates)-1].state.score)
		//log.Println("income:", nextStates[0].state.income, nextStates[len(nextStates)-1].state.income)
		//}
		//log.Println("loop", loop, "beamStates", len(beamStates))
		if len(nextStates) > beamWidth {
			beamStates = nextStates[:beamWidth]
		} else {
			beamStates = nextStates
		}
		nextStates = make([]bsState, 0, beamWidth)
		loop++
		ATCODER = true
		if ATCODER {
			elpstime := time.Since(startTime)
			if elpstime > time.Millisecond*2800 {
				timeout = true
				log.Println("time out")
				break
			}
		}
	}
	log.Printf("Cnt=%v\n", cnt)
	if timeout {
		log.Printf("TO=1\n")
	} else {
		log.Printf("TO=0\n")
	}
	log.Println("bestScore", bestState.state.score, "income:", bestState.state.income, "turn:", bestState.state.turn)
	log.Println(bestState.state.field.ToString())

	sb := strings.Builder{}
	for _, act := range bestState.state.actions {
		sb.WriteString(act.String())
	}
	for i := len(bestState.state.actions); i < in.T; i++ {
		sb.WriteString("-1\n")
	}
	return sb.String()
}
