BINARY = bin/a.out
SRC = main.go

# Reactive
Reactive = true
ifeq ($(Reactive), true)
	RUN_CMD = ./tools/target/release/tester
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
	$(RUN_CMD) -cpuprofile cpu.prof < tools/in/0008.txt > 8.out
# flag.Pars() をわすれない

cpuprofile:
	go tool pprof -http=:6060 cpu.prof 	

test:
	go test -v .


# NUMはデフォルトで1に設定
SEED ?= 0

# NUMを4桁にゼロパディングするための変数
PADDING_NUM := $(shell printf "%04d" $(SEED))

# Goのプログラムを実行してCPUプロファイルを作成し、結果を出力
run_go:
	go run main.go -cpuprofile cpu.prof < tools/in/$(PADDING_NUM).txt > tools/out/$(PADDING_NUM).txt

# 可視化ツールを実行
run_vis:
	cd tools && cargo run -r --bin vis in/$(PADDING_NUM).txt out/$(PADDING_NUM).txt

# 両方のコマンドをまとめて実行
run: run_go run_vis

#### reactive
run_reactive:
	$(RUN_CMD) go run main.go < tools/in/$(PADDING_NUM).txt > $(PADDING_NUM).txt
