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

var TESTER string = "./tools/target/release/tester "

func testn(n int) {
	sumScore := 0
	for i := 0; i < n; i++ {
		fmt.Print("case=", i)
		d := run(i)
		fmt.Printf(" score=%d time=%x\n", d.Score, d.Time.Seconds())
		sumScore += d.Score
	}
	fmt.Println("ALL SCORE = ", sumScore)
}

func testRun() {
	d := run(0)
	log.Printf("score=%d time=%f \n", d.Score, d.Time.Seconds())
}

func run(seed int) Date {
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
	d, err := parseDates(stderr.String())
	if err != nil {
		panic(err)
	}
	d.Seed = seed
	d.Time = elapsed
	if d.Score == 0 {
		log.Println(stderr.String())
	}
	return d
}

func parseDates(stderr string) (d Date, err error) {
	d.Score, err = parseInt(stderr, re_score, str_score)
	if err != nil {
		return d, err
	}
	// d.loop, err = parseInt(stderr, re_loop, str_loop)
	// if err != nil {
	// 	return d, err
	// }
	d.Human, err = parseInt(stderr, re_human, str_human)
	if err != nil {
		return d, err
	}
	d.Cow, err = parseInt(stderr, re_cow, str_cow)
	if err != nil {
		return d, err
	}
	d.Pig, err = parseInt(stderr, re_pig, str_pig)
	if err != nil {
		return d, err
	}
	d.Rabbit, err = parseInt(stderr, re_rabbit, str_rabbit)
	if err != nil {
		return d, err
	}
	d.Dog, err = parseInt(stderr, re_dog, str_dog)
	if err != nil {
		return d, err
	}
	d.Cat, err = parseInt(stderr, re_cat, str_cat)
	if err != nil {
		return d, err
	}
	return d, err
}

type Date struct {
	Seed   int           `json:"seed"`
	Score  int           `json:"score"`
	Loop   int           `json:"loop"`
	Human  int           `json:"human"`
	Cow    int           `json:"cow"`
	Pig    int           `json:"pig"`
	Rabbit int           `json:"rabbit"`
	Dog    int           `json:"dog"`
	Cat    int           `json:"cat"`
	Time   time.Duration `json:"time"`
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
			d := run(seed)
			mu.Lock()
			datas = append(datas, d)
			fmt.Print(".")
			sumScore += d.Score
			sumTime += d.Time
			if maxTime < d.Time {
				maxTime = d.Time
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

	// json output
	f, err := os.Create("data_output.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString("seed,score,loop,human,cow,pig,rabbit,dog,cat,time\n")
	if err != nil {
		panic(err)
	}
	for _, d := range datas {
		_, err = f.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d\n", d.Seed, d.Score, d.Loop, d.Human, d.Cow, d.Pig, d.Rabbit, d.Dog, d.Cat, d.Time.Milliseconds()))
	}
}

var re_score = regexp.MustCompile(`Score = ([0-9]+)`)
var str_score = "Score = "

var re_loop = regexp.MustCompile(`loop=([0-9]+)`)
var str_loop = "loop="

// ahc008
var re_human = regexp.MustCompile(`human=([0-9]+)`)
var str_human = "human="
var re_cow = regexp.MustCompile(`cow=([0-9]+)`)
var str_cow = "cow="
var re_pig = regexp.MustCompile(`pig=([0-9]+)`)
var str_pig = "pig="
var re_rabbit = regexp.MustCompile(`rabbit=([0-9]+)`)
var str_rabbit = "rabbit="
var re_dog = regexp.MustCompile(`dog=([0-9]+)`)
var str_dog = "dog="
var re_cat = regexp.MustCompile(`cat=([0-9]+)`)
var str_cat = "cat="

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
