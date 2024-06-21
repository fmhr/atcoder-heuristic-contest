package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	P, _ = os.Getwd()
	if _, err := os.Stat(P + "/tools"); os.IsNotExist(err) {
		log.Fatal("tools not exists")
	}
	if _, err := os.Stat(P + "/out"); os.IsNotExist(err) {
		os.Mkdir(P+"/out", os.ModePerm)
	}
	//testRun()
	parallelRun()
}

var P string = ""

// 	./tools/target/release/tester tools/in/0000.txt ./solver > out.txt
func run(seed int) (int, int, error) {
	tester := P + "/tools/target/release/tester"
	solver := P + "/solver"
	inFile := fmt.Sprintf("%s/in/%s.txt", P, fmt.Sprintf("%04d", seed))
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		return 0, 0, err
	}
	outFile := fmt.Sprintf("%s/out/%s.out", P, fmt.Sprintf("%04d", seed))
	cmdStr := tester + " " + inFile + " " + solver + " > " + outFile
	//cmdStr := exe + " < " + inFile + " > " + outFile
	cmds := []string{"sh", "-c", cmdStr}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		log.Println(cmds)
		log.Fatal(err)
	}
	// TLE対策
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(10 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Println("failed to kill prosess:", err)
		}
		cmd.Wait()
	case err := <-done:
		if err != nil {
			log.Fatalf("process finished with err = %v\n", err)
		}
	}
	// cmd.Wait()
	score := parseScore(stderr.String())
	turn := parseTurn(stderr.String())
	if score == 0 {
		log.Println(stderr.String())
	}
	//loop := parseLoop(stderr.String())
	return score, turn, nil
}

type Date struct {
	seed  int
	score int
	time  int
	turn  int
}

func parallelRun() {
	cpus := runtime.NumCPU()
	maxSeed := 100
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, cpus)
	datas := make([]Date, 0)
	sumScore := 0
	for seed := 0; seed < maxSeed; seed++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(seed int) {
			startTime := time.Now()
			var d Date
			var err error
			d.score, d.turn, err = run(seed)
			elapsed := time.Since(startTime)
			d.seed = seed
			mu.Lock()
			datas = append(datas, d)
			// fmt.Print(".")
			if err != nil {
				fmt.Printf("seed=%d %s\n", d.seed, err)
			} else {
				fmt.Printf("seed=%d score=%d time=%v switch=%d\n", d.seed, d.score, elapsed, d.turn)
			}
			sumScore += d.score
			mu.Unlock()
			wg.Done()
			<-sem
		}(seed)
	}
	fmt.Printf("SCORE=%d\n", sumScore)
}

func parseScore(s string) int {
	ms := `Score = ([0-9]+)`
	re := regexp.MustCompile(ms)
	ma := re.FindString(s)
	score, err := strconv.Atoi(strings.Replace(ma, "Score = ", "", -1))
	if err != nil {
		log.Println(score)
		log.Println(ma)
	}
	return score
}

func parseTime(s string) int {
	ms := `time=([0-9]+)`
	re := regexp.MustCompile(ms)
	ma := re.FindString(s)
	n, err := strconv.Atoi(strings.Replace(ma, "loop=", "", -1))
	if err != nil {
		log.Println(n)
	}
	return n
}

func parseTurn(s string) int {
	ms := `turn=([0-9]+)`
	re := regexp.MustCompile(ms)
	ma := re.FindString(s)
	n, err := strconv.Atoi(strings.Replace(ma, "turn=", "", -1))
	if err != nil {
		log.Println(n)
	}
	return n
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
