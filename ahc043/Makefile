BINARY = bin/a.out
SRC = src/main.go

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

# コンテスト後に不要なファイルを削除
clear:
	rm -rf bin/*
	rm -rf out/*
	rm -rf $(BINARY)
	rm -rf *.prof
	rm -f [0-9][0-9][0-9][0-9].txt

seed:
	seq 0 10000 > tools/seed.txt
	cd tools && cargo run -r --bin gen seed.txt

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
	/usr/bin/time -l $(BINARY) < tools/in/$(PADDING_NUM).txt > $(PADDING_NUM).txt

# 可視化ツールを実行
run_vis:
	./tools/target/release/vis tools/in/$(PADDING_NUM).txt $(PADDING_NUM).txt
#cd tools && cargo run -r --bin vis in/$(PADDING_NUM).txt ../$(PADDING_NUM).txt

# 両方のコマンドをまとめて実行
run: build run_go run_vis

#### reactive
run_reactive:
	$(RUN_CMD) go run $(SRC) < tools/in/$(PADDING_NUM).txt > $(PADDING_NUM).txt

run_cpu:
	$(RUN_CMD) go run $(SRC) -cpuprofile cpu.prof < tools/in/$(PADDING_NUM).txt > $(PADDING_NUM).txt
