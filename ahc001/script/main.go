package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	log.SetFlags(log.Lshortfile)
	P, _ = os.Getwd()
	testn(10)
}

var P string = ""

func testn(n int) {
	inputpaths, names := dirwalk(P + "/tools/in")
	exe := P + "/solver"
	out := strings.Replace(names[0], "txt", "out", 1)
	sumScore := 0
	for i := 0; i < n; i++ {
		fmt.Print("case=", i)
		score, loop := run(exe, inputpaths[i], out)
		fmt.Printf(" score=%d loop=%d\n", score, loop)
		sumScore += score
	}
	fmt.Println("ALL SCORE = ", sumScore)
}

func testRun() {
	inputpaths, names := dirwalk(P + "/tools/in")
	exe := P + "/solver"
	out := strings.Replace(names[50], "txt", "out", 1)
	score, n := run(exe, inputpaths[50], out)
	log.Println(n)
	vscore := vis(inputpaths[50], out)
	log.Printf("score=%d visscore=%d\n", score, vscore)
}

func run(exe string, in string, out string) (int, int) {
	cmdStr := exe + "<" + in + ">" + out
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
	score := parseScore(stderr.String())
	if score == 0 {
		log.Println(stderr.String())
	}
	loop := parseLoop(stderr.String())
	return score, loop
}

func parallelRun() {

}

func parseScore(s string) int {
	ms := `score=([0-9]+)`
	re := regexp.MustCompile(ms)
	ma := re.FindString(s)
	score, err := strconv.Atoi(strings.Replace(ma, "score=", "", -1))
	if err != nil {
		log.Println(score)
	}
	return score
}

func parseLoop(s string) int {
	ms := `loop=([0-9]+)`
	re := regexp.MustCompile(ms)
	ma := re.FindString(s)
	n, err := strconv.Atoi(strings.Replace(ma, "loop=", "", -1))
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

func dirwalk(dir string) (paths []string, names []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		names = append(names, file.Name())
		paths = append(paths, filepath.Join(dir, file.Name()))
	}
	return
}
