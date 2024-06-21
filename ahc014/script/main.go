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
)

func main() {
	log.SetFlags(log.Lshortfile)
	P, _ = os.Getwd()
	parallelRun()
	//testRun()
}

var P string = ""

func testRun() {
	d := normalRun(1)
	log.Printf("score=%d cnt=%d\n", d["score"].(int), d["cnt"].(int))
}

func normalRun(seed int) map[string]interface{} {
	data := map[string]interface{}{}
	exe := P + "/bin/a"
	inFile := fmt.Sprintf("%s/tools/in/%s.txt", P, fmt.Sprintf("%04d", seed))
	outFile := fmt.Sprintf("%s/out/%s.txt", P, fmt.Sprintf("%04d", seed))
	cmdStr := exe + " < " + inFile + " > " + outFile
	cmds := []string{"sh", "-c", cmdStr}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		log.Println(cmds)
		log.Fatal(err)
	}
	cmd.Wait()
	score, err := parseInt(stderr.String(), re_score, str_score)
	if err != nil {
		log.Println("seed=", seed, err)
	}
	if score == 0 {
		log.Println(stderr.String())
	}
	data["score"] = score
	num, err := parseInt(stderr.String(), re_num, str_num)
	if err != nil {
		log.Println("num=", seed, err)
	}
	if score == 0 {
		log.Println(stderr.String())
	}
	data["num"] = num
	t, err := parseInt(stderr.String(), re_time, str_time)
	if err != nil {
		log.Println(stderr.String())
		log.Fatal(err)
	}
	data["time"] = t

	fmt.Printf("seed:%3d score=%d num=%3d time=%4d\n",
		seed, data["score"].(int), data["num"].(int), data["time"].(int))
	return data
}

func parallelRun() {
	CORE := 4
	maxSeed := 100
	sumScore := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, CORE)
	for seed := 0; seed < maxSeed; seed++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(seed int) {
			d := normalRun(seed)
			mu.Lock()
			sumScore += d["score"].(int)
			mu.Unlock()
			wg.Done()
			<-sem
		}(seed)
	}
	wg.Wait()
	fmt.Println("sum=", sumScore)
}

var re_score = regexp.MustCompile(`score=([0-9]+)`)
var str_score = "score="

var re_num = regexp.MustCompile(`num=([0-9]+)`)
var str_num = "num="

var re_time = regexp.MustCompile(`time=([0-9]+)`)
var str_time = "time="

func parseInt(src string, re *regexp.Regexp, str string) (int, error) {
	match := re.FindString(src)
	num, err := strconv.Atoi(strings.Replace(match, str, "", -1))
	if err != nil {
		log.Println(err, str)
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
