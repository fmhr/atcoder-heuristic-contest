BINARY = a.out
SRC = main.go

.PHONY: build
build: $(BINARY)

$(BINARY): $(SRC)
	echo "Building $(BINARY) ..."
	go build -o ./$(BINARY)


.PHONY: clear
clear:
	rm -rf bin/*
	rm -rf out/*

.PHONY: seed
seed:
	seq 0 10000 > tools/seed.txt

profile: build
	./a.out -cpuprofile cpu.prof -memprofile mem.prof < tools/in/0001.txt > out.txt