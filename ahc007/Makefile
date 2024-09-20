BINARY=a.out
SRC=src/main.go
SEED=0000

build:
	go build -o $(BINARY) src/main.go

tester:
	./$(BINARY) < ./tools/in/$(SEED).txt > $(SEED).out
	./tools/target/release/vis tools/in/$(SEED).txt $(SEED).out


LIST=0000 0001 0002 0003 0004 0005 0006 0007 0008 0009
tester10:
	for i in $(LIST); do \
	./$(BINARY) < ./tools/in/$$i.txt > $(SEED).out && \
	./tools/target/release/vis tools/in/$$i.txt $(SEED).out ;done
	

script:
	go run script/main.go

seed:
	seq 0 3000 > seed3000.txt


# https://pkg.go.dev/runtime/pprof main関数をここからとる
pprof:	
	./$(BINARY) -cpuprofile cpu.prof < tools/in/0000.txt
	pprof -http=localhost:8080 cpu.prof

.PHONY: script
