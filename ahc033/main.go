package main

import (
	"fmt"
	"log"
	"strings"
)

// クレーン同士は重なったりすれ違ったりできない
// 小クレーン
// コンテナのない場所へはいつでも移動できる
// コンテナを持っている時に、コンテナが置いてある場所へは移動できない
// 大クレーン
// コンテナを持っている時でも、コンテナが置いてある場所に移動できる

// 搬出口
// 0,1,2,3,4
// 5,6,7,8,9
// 10,11,12,13,14
// 15,16,17,18,19

// 今すぐ搬出できるものは搬出する
// 搬出順番が来ているものが奥にあれば手前のものを搬出する
// 搬出順番持ちのものは邪魔にならない場所に移動する

// コンテナの状態
// 搬入待ち
// 盤上にある
// 1. 搬入直後(これを動かさないと次の搬入がない状態)
// 2. 搬出待ち番号０　搬出口に持っていける状態
// 3. 搬出待ち番号１　搬出待ち状態
// クレーンで吊り下げられている状態

// クレーンの行動
// 目的のコンテナを目指す途中
// 目的のコンテナの上にあり、持ち上げる(PickUp)
// コンテナを持っている
// 1. 搬出口に持っていく
// 2. 搬出順番を待つ
// 3. 邪魔にならない場所にコンテナを移動させる
// 4. コンテナを下す

// 5クレーンのランダム行動をつかった、ビームサーチは探索が広すぎる？
// 最小で45ターン
// 1クレーンずつ動かす？
// 45*5=225ターン?
// 全体の評価関数も難しいか？
// 役割分担
// ルート決め打ち
// 回転方式

func main() {
	in := input()
	solver(in)
}

type Input struct {
	Num       int
	Container [5][5]int32
}

type Action uint8

const (
	PickUp Action = iota + 1
	Release
	Up    // 3
	Down  // 4
	Left  // 5
	Right // 6
	DoNothing
	Bomb
)

var DIJ [4][2]int32 = [4][2]int32{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
var DIR [4]string = [4]string{"U", "D", "L", "R"}
var ACT []string = []string{"", "P", "Q", "U", "D", "L", "R", ".", "B"}

func input() (in Input) {
	fmt.Scan(&in.Num)
	for i := 0; i < in.Num; i++ {
		for j := 0; j < in.Num; j++ {
			fmt.Scan(&in.Container[i][j])
		}
	}
	return
}

func solver(in Input) {
	_ = in

}

type State struct {
	incontainer  [5][5]int32 // 入荷待ちコンテナ
	outcontainer [5][5]int32 // 搬出済みコンテナ
	board        [5][5]int32 //
	pos          [5][3]int32 // x, y, box クレーンの座標と持っている箱
	turn         int32
}

func newState(in Input) (s State) {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			s.incontainer[i][j] = in.Container[i][j]
		}
	}
	for i := 0; i < 5; i++ {
		s.pos[i][0], s.pos[i][1], s.pos[i][2] = 0, 0, 0
	}
	return
}

func (s *State) Do(acts [5]Action) {
	var to [5][3]int32
	for i := 0; i < 5; i++ {
		y, x, z := s.pos[i][0], s.pos[i][1], s.pos[i][2]
		switch acts[i] {
		case DoNothing:
		case PickUp:
			if x == -1 {
				log.Printf("Error: Crane %d is already bombed\n", i)
			} else if z != 0 {
				log.Printf("Error: Crane %d is already holding a box\n", i)
			} else if s.board[x][y] == 0 {
				log.Printf("Error: Crane %d is trying to pick up a box from empty space\n", i)
			} else {
				z = s.board[y][x]
				s.board[y][x] = 0
			}
		case Release:
			if x == -1 {
				log.Printf("Error: Crane %d is already bombed\n", i)
			} else if z == 0 {
				log.Printf("Error: Crane %d is not holding a box\n", i)
			} else if s.board[y][x] != 0 {
				log.Printf("Error: Crane %d is trying to release a box to non-empty space\n", i)
			} else {
				s.board[y][x] = z
				z = 0
			}
		case Up, Down, Left, Right:
			if x == -1 {
				log.Printf("Error: Crane %d is already bombed\n", i)
			} else {
				y, x = y+DIJ[acts[i]-Up][0], x+DIJ[acts[i]-Up][1]
				if y < 0 || y >= 5 || x < 0 || x >= 5 {
					log.Printf("Error: Crane %d is trying to move out of the board\n", i)
				} else if i > 0 && z != 0 && s.board[y][x] != 0 {
					log.Printf("Error: Crane %d is trying to move to a space with a box\n", i)
				} else {
					s.pos[i][0], s.pos[i][1] = y, x
				}
			}
		case Bomb:
			if x == -1 {
				log.Printf("Error: Crane %d is already bombed\n", i)
			} else if z != 0 {
				log.Printf("Error: Crane %d is holding a box\n", i)
			} else {
				s.pos[i][0], s.pos[i][1] = -1, -1
			}
		}
		to[i][0], to[i][1], to[i][2] = y, x, z
	}
	// クレーン同士の衝突判定
	// クレーン１は大クレーンなので衝突しない
	for i := 1; i < 5-1; i++ {
		if to[i][0] == -1 {
			continue
		}
		for j := i + 1; j < 5; j++ {
			if to[i][0] == to[j][0] && to[i][1] == to[j][1] {
				log.Printf("Error: Crane %d and %d are in the same place\n", i, j)
			}
			// クレーンの入れ替えがある場合もNG
			if s.pos[i][0] == to[j][0] && s.pos[i][1] == to[j][1] && s.pos[j][0] == to[i][0] && s.pos[j][1] == to[i][1] {
				log.Printf("Error: Crane %d and %d are swapping\n", i, j)
			}
		}
	}
}

func outputActions(acts []Action) {
	var sb strings.Builder
	for _, a := range acts {
		sb.WriteString(ACT[a])
	}
	fmt.Println(sb.String())
}
