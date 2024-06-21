package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	P, _ = os.Getwd()
	//testRun()
	parallelRun()
}

var P string = ""

// 今回はインタラクティブ問題なので配布のtesterを使う
var TESTER string = "./tools/target/release/tester "

func testn(n int) {
	sumScore := 0
	for i := 0; i < n; i++ {
		fmt.Print("case=", i)
		score, loop, time := run(i)
		fmt.Printf(" score=%d loop=%d time=%x\n", score, loop, time.Seconds())
		sumScore += score
	}
	fmt.Println("ALL SCORE = ", sumScore)
}

func testRun() {
	score, n, t := run(0)
	log.Printf("score=%d loop=%d time=%f \n", score, n, t.Seconds())
}

func run(seed int) (int, int, time.Duration) {
	exe := P + "/bin/a"
	inFile := fmt.Sprintf("%s/tools/in/%s.txt", P, fmt.Sprintf("%04d", seed))
	outFile := fmt.Sprintf("%s/out/%s.out", P, fmt.Sprintf("%04d", seed))
	cmdStr := TESTER + exe + " < " + inFile + " > " + outFile
	cmds := []string{"sh", "-c", cmdStr}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	start := time.Now()
	err := cmd.Start()
	if err != nil {
		log.Println(cmds)
		log.Fatal(err)
	}
	cmd.Wait()
	elapsed := time.Since(start)
	score, err := parseInt(stderr.String(), re_score, str_score)
	if err != nil {
		log.Println("seed=", seed)
		log.Println(err)
	}
	if score == 0 {
		log.Println(stderr.String())
	}
	return score, 0, elapsed
}

type Date struct {
	seed  int
	score int
	loop  int
	time  time.Duration
}

func parallelRun() {
	CORE := 4
	maxSeed := 500
	sumScore := 0
	var sumTime time.Duration
	var maxTime time.Duration
	var maxTimeSeed int
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, CORE-1)
	datas := make([]Date, 0)
	for seed := 0; seed < maxSeed; seed++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(seed int) {
			var d Date
			d.score, d.loop, d.time = run(seed)
			d.seed = seed
			mu.Lock()
			datas = append(datas, d)
			fmt.Print(".")
			//fmt.Printf("seed=%03d score=%d time=%fs \n", d.seed, d.score, d.time.Seconds())
			sumScore += d.score
			sumTime += d.time
			if maxTime < d.time {
				maxTime = d.time
				maxTimeSeed = seed
			}
			mu.Unlock()
			wg.Done()
			<-sem
		}(seed)
	}
	averageTime := sumTime.Seconds() / float64(maxSeed)
	fmt.Println()
	fmt.Printf("sum=%d maxTime=%fs(%d) averageTime=%fs\n", sumScore, maxTime.Seconds(), maxTimeSeed, averageTime)
}

var re_score = regexp.MustCompile(`Score = ([0-9]+)`)
var str_score = "Score = "

var re_loop = regexp.MustCompile(`loop=([0-9]+)`)
var str_loop = "loop="

func parseInt(src string, re *regexp.Regexp, str string) (int, error) {
	match := re.FindString(src)
	num, err := strconv.Atoi(strings.Replace(match, str, "", -1))
	if err != nil {
		log.Println(src)
		return -1, err
	}
	return num, nil
}

func vis(input string, output string) (score int) {
	vispath := P + "/tools/target/release/vis"
	cmdStr := vispath + " " + input + " " + output
	cmds := []string{"sh", "-c", cmdStr}
	var out []byte
	var err error
	out, err = exec.Command(cmds[0], cmds[1:]...).Output()
	if err != nil {
		log.Fatal(err)
	}
	outs := strings.Split(string(out), "\n")
	score, err = strconv.Atoi(outs[0])
	if err != nil {
		panic(err)
	}
	return score
}

func maxTimeDuration(t1, t2 time.Duration) time.Duration {
	if t1 > t2 {
		return t1
	}
	return t2
}
