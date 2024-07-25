BINARY = a.out
SRC = main.go

# Reactive
Reactive = true
ifeq ($(Reactive), true)
	RUN_CMD = ./tools/target/release/tester ./a.out
else
	RUN_CMD = ./a.out
endif

build: $(BINARY)

$(BINARY): $(SRC)
	echo "Building $(BINARY) ..."
	go build

clear:
	rm -rf bin/*
	rm -rf out/*
	rm -rf *.out
	rm -rf *.prof

seed:
	seq 0 10000 > tools/seed.txt

profile: build
	$(RUN_CMD) -cpuprofile cpu.prof < tools/in/0000.txt > tmp.out

cpuprofile:
	go tool pprof -http=:6060 cpu.prof 	

.PHONY: build clear seed cpuprofile