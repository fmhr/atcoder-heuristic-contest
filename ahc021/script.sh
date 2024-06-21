#!/bin/bash

SEED=0000
AHCNumber=21
BINARY=bin/a
SRC=src/main.go

build() {
	go build -o ${BINARY} ${SRC}
}

test() {
    echo hello
    if [ ! -e "src/main.go" ]; then
        echo "src/main.go not found"
        exit 1
    fi
}

# 1つのテストケースを実行する
tester() {
	build
	${BINARY} < tools/in/${SEED}.txt > out/${SEED}.txt
	./tools/target/release/vis tools/in/${SEED}.txt out/${SEED}.txt
}

# visulizerが別に必要な時
vis() {
	tester
	./tools/target/release/vis tools/in/${SEED}.txt out/${SEED}.txt
}

# 1~10までのテストケースを実行する
tester10() {
	build
	for i in {1..10}
    do
        formatted_number=$(printf "%04d" $i)
		echo SEED=$formatted_number
		${BINARY} < tools/in/${formatted_number}.txt > out/${formatted_number}.txt
		./tools/target/release/vis tools/in/${formatted_number}.txt out/${formatted_number}.txt
	done
}

# toolsで使うテストケースのseed.txtを生成する
generateSeeds() {
	seq 0 5000 > seed.txt
}

# 並列実行用のスクリプトを実行する
script() {
	build
	go run script/main.go
}

setup() {
	mkdir -p src
	curl -of src/main.go https://raw.githubusercontent.com/fmhr/AHC_Templates/main/main.go
	curl -of .gitignore https://raw.githubusercontent.com/fmhr/AHC_Templates/main/.gitignore
	cd src && go mod init main
}

downloadtools() {
    value=$(grep "^$AHCNumber=" toolURLs.txt | cut -d= -f2)
    echo $value
	curl -o tester.zip $value
	#curl -o tester.zip https://img.atcoder.jp/atcoder11live/f62fc84.zip
	unzip tester.zip
	rm tester.zip
}

dlmain_go() {
	gh api -H "Accept: application/vnd.github.v3.raw" \
	"https://api.github.com/repos/fmhr/AHC_Templates/contents/main.go" > src/main.go
}

dlgitignore() {
	gh api -H "Accept: application/vnd.github.v3.raw" \
	"https://api.github.com/repos/fmhr/AHC_Templates/contents/.gitignore" > .gitignore
}

$@