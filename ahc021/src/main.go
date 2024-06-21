package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	startTime := time.Now()
	timeDistance := time.Since(startTime)
	log.Println(timeDistance)
	roadInput()
	ans := bubbleSort2(Pyramid)
	output(ans)
}

var N int = 30
var Pyramid [31][31]int

func roadInput() {
	for i := 0; i < N; i++ {
		for j := 0; j < i+1; j++ {
			fmt.Scan(&Pyramid[i][j])
		}
	}
}

func bubbleSort2(pyramid [31][31]int) [][4]int {
	var swapList [][4]int
	mapping := make(map[int][2]int)
	for i := 0; i < N; i++ {
		for j := 0; j < i+1; j++ {
			mapping[pyramid[i][j]] = [2]int{i, j}
		}
	}
	for {
		cnt := 0
		for n := 0; n < 465; n++ {
			nx := mapping[n][0]
			ny := mapping[n][1]
			for nx-1 >= 0 && ny-1 >= 0 && pyramid[nx-1][ny-1] > pyramid[nx][ny] && pyramid[nx-1][ny] > pyramid[nx][ny] {
				// 左上右上どちらにも移動できる場合,大きい方と交換する
				if pyramid[nx-1][ny-1] > pyramid[nx-1][ny] {
					pyramid[nx-1][ny-1], pyramid[nx][ny] = pyramid[nx][ny], pyramid[nx-1][ny-1]
					swapList = append(swapList, [4]int{nx, ny, nx - 1, ny - 1})
					mapping[pyramid[nx][ny]] = [2]int{nx, ny}
					mapping[pyramid[nx-1][ny-1]] = [2]int{nx - 1, ny - 1}
					ny--
					nx--
					cnt++
				} else {
					pyramid[nx-1][ny], pyramid[nx][ny] = pyramid[nx][ny], pyramid[nx-1][ny]
					swapList = append(swapList, [4]int{nx, ny, nx - 1, ny})
					mapping[pyramid[nx][ny]] = [2]int{nx, ny}
					mapping[pyramid[nx-1][ny]] = [2]int{nx - 1, ny}
					nx--
					cnt++
				}
			}
			for nx-1 >= 0 && pyramid[nx-1][ny] > pyramid[nx][ny] {
				pyramid[nx-1][ny], pyramid[nx][ny] = pyramid[nx][ny], pyramid[nx-1][ny]
				swapList = append(swapList, [4]int{nx, ny, nx - 1, ny})
				mapping[pyramid[nx][ny]] = [2]int{nx, ny}
				mapping[pyramid[nx-1][ny]] = [2]int{nx - 1, ny}
				nx--
				cnt++
			}
			for nx-1 >= 0 && ny-1 >= 0 && pyramid[nx-1][ny-1] > pyramid[nx][ny] {
				pyramid[nx-1][ny-1], pyramid[nx][ny] = pyramid[nx][ny], pyramid[nx-1][ny-1]
				swapList = append(swapList, [4]int{nx, ny, nx - 1, ny - 1})
				mapping[pyramid[nx][ny]] = [2]int{nx, ny}
				mapping[pyramid[nx-1][ny-1]] = [2]int{nx - 1, ny - 1}
				ny--
				nx--
				cnt++
			}
		}
		if cnt == 0 {
			break
		}
		//log.Println(cnt)
	}
	return swapList
}

func output(swapList [][4]int) {
	fmt.Println(minInt(100000, len(swapList)))
	for i, v := range swapList {
		if i >= 10000 {
			break
		}
		fmt.Println(v[0], v[1], v[2], v[3])
	}
	//log.Println(len(swapList))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
