BINARY = a.out
SRC = main.go

# Reactive
Reactive = false
ifeq ($(Reactive), true)
	RUN_CMD = ./tools/target/release/tester ./$(BINARY)
else
	RUN_CMD = ./$(BINARY)
endif

build: $(BINARY)

$(BINARY): $(SRC)
	echo "Building $(BINARY) ..."
	go build -o $(BINARY) $(SRC)

clear:
	rm -rf bin/*
	rm -rf out/*
	rm -rf $(BINARY)
	rm -rf *.prof

seed:
	seq 0 10000 > tools/seed.txt

profile: build
	$(RUN_CMD) -cpuprofile cpu.prof < tools/in/0000.txt > tmp.out
# flag.Pars() をわすれない

cpuprofile:
	go tool pprof -http=:6060 cpu.prof 	
