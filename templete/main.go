package main

import (
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	startTime := time.Now()
	timeDistance := time.Since(startTime)
	log.Println(timeDistance)
}
