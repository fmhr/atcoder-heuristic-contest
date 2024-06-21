package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/fmhr/fj"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func sendWorker(config fj.Config, username string) {
	seeds := make([]int, 10)
	for i := 0; i < 10; i++ {
		seeds[i] = i
	}
	RunParallel(&config, seeds, username)
}

func RunParallel(cnf *fj.Config, seeds []int, username string) {
	// 並列実行数の設定
	concurrentNum := cnf.ConcurrentRequests
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrentNum)
	datas := make([]*orderedmap.OrderedMap[string, any], 0, len(seeds))
	errorChan := make(chan string, len(seeds))
	errorSeedChan := make(chan int, len(seeds))

	// Ctrl+Cで中断したときに、現在実行中のseedを表示する
	var currentlyRunningSeed sync.Map
	var datasMutex sync.Mutex
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigCh
		log.Printf("signal received.")
		os.Exit(1)
	}()
	for _, seed := range seeds {
		wg.Add(1)
		sem <- struct{}{}
		currentlyRunningSeed.Store(seed, true)
		time.Sleep(5 * time.Millisecond)
		go func(seed int) {
			data, err := fj.RunSelector(cnf, seed)
			if err != nil {
				errorChan <- fmt.Sprintf("Run error: seed=%d %v\n", seed, err)
				errorSeedChan <- seed
			}
			// 後処理
			datasMutex.Lock()
			datas = append(datas, data) // 結果を追加
			currentlyRunningSeed.Delete(seed)
			datasMutex.Unlock()
			wg.Done()
			<-sem
		}(seed)
	}
	wg.Wait()
	fmt.Fprintf(os.Stderr, "\n") // Newline after progress bar
	close(errorChan)
	close(errorSeedChan)
	for err := range errorChan {
		log.Println(err)
	}
	errSeeds := make([]int, 0, len(errorSeedChan))
	for seed := range errorSeedChan {
		errSeeds = append(errSeeds, seed)
	}
	sumScore := 0.0
	logScore := 0.0
	for i := 0; i < len(datas); i++ {
		_, ok := datas[i].Get("seed")
		if !ok {
			log.Println("seed not found")
			continue
		}
		// error のときnilになる
		if datas[i] != nil {
			score, ok := datas[i].Get("Score")
			if !ok {
				log.Println("Score not found")
				continue
			}
			if score == -1 {
				continue
			}
			sumScore += score.(float64)
			logScore += math.Max(0, math.Log(score.(float64)))
			if !ok {
				log.Println("seed not found")
				continue
			}
		}
	}

	// timeがあれば、平均と最大を表示
	_, exsit := datas[0].Get("time")
	timeNotFound := 0
	if exsit {
		sumTime := 0.0
		maxTime := 0.0
		for i := 0; i < len(datas); i++ {
			if datas[i] == nil {
				log.Println("skip seed=", i)
				continue
			}
			if t, ok := datas[i].Get("time"); !ok {
				//log.Printf("seed:%d time not found", i)
				timeNotFound++
			} else {
				sumTime += t.(float64)
				maxTime = math.Max(maxTime, t.(float64))
			}
		}
		sumTime /= float64(len(datas) - len(errSeeds) - timeNotFound)
		fmt.Fprintf(os.Stderr, "avarageTime=%.2f  maxTime=%.2f\n", sumTime, maxTime)
	}
	avarageScore := sumScore / float64(len(datas)-len(errSeeds))
	p := message.NewPrinter(language.English)
	p.Fprintf(os.Stderr, "(Score)sum=%.2f avarage=%.2f log=%.2f\n", sumScore, avarageScore, logScore)
	// standingのためのjsonを出力
	standingJson(datas, username, cnf.Contest)
}

func standingJson(datas []*orderedmap.OrderedMap[string, any], username string, contest string) error {
	// stainding/ahc032/ がなければ作成
	dir := "standing/" + contest
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			log.Println(err)
			return err
		}
	}

	scoreData := make([]string, 1001)
	scoreData[0] = username
	// standingのためのjsonを出力
	for i := 0; i < len(datas); i++ {
		seed, ok := datas[i].Get("seed")
		if !ok {
			log.Println("seed not found")
			continue
		}
		seedInt := int(seed.(float64))
		score, ok := datas[i].Get("Score")
		scoreInt := int(score.(float64))
		if !ok {
			log.Println("Score not found")
			continue
		}
		scoreData[seedInt+1] = fmt.Sprintf("%d", scoreInt)
	}

	file, err := os.OpenFile(dir+"/result.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	// nodataのときは0にする
	for i := 0; i < len(scoreData); i++ {
		if scoreData[i] == "" {
			scoreData[i] = "0"
		}
	}

	// csvファイルに書き込み
	writer := csv.NewWriter(file)
	if err := writer.Write(scoreData); err != nil {
		log.Println(err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Println(err)
	}
	return nil
}
