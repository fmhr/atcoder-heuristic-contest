package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func readSample(seed int) {
	filename := "tools/in/" + fmt.Sprintf("%04d", seed) + ".txt"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	read(file)
}

func BenchmarkSSS(b *testing.B) {
	result := shortestSuperstring(Words[:], 1)
	_ = result
}

func BenchmarkDpRoot(b *testing.B) {
	readSample(0)
	result, n := dpRoot("ACDGEGAWEPVATPHEPP", Point{-1, -1}, true)
	//log.Println(n)
	//log.Println(result)
	_, _ = result, n
}

// go test -bench BeamSearch -cpuprofile cpu.out -benchmem
// go tool pprof -http=":8080" cpu.out
func BenchmarkBeamSearch(b *testing.B) {
	readSample(0)
	b.ResetTimer()
	str := beamSearchOrder(Words[:], startPoint, 200*time.Millisecond)
	_ = str
}

func BenchmarkGenerateNodes(b *testing.B) {
	readSample(0)
	b.ResetTimer()
	var n Node
	for i := 0; i < 30; i++ {
		nodes := generateNodes(n, Words[:])
		n = nodes[0]
	}
}

func TestShortestMerge(t *testing.T) {
	readSample(1)
	//log.Println(len(Words))
	result := shortestMerge(Words[:])
	//for _, v := range result {
	//log.Println(v)
	//}
	//log.Println(len(result))
	_ = result
}
