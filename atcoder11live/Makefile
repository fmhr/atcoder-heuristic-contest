BINARY=bin/a
SRC=src/main.go
SEED=0000

build:
	go build -o $(BINARY) $(SRC)

tester: build
	./tools/target/release/tester ${BINARY} < tools/in/${SEED}.txt > out/${SEED}.txt

vis: tester
	./tools/target/release/vis tools/in/${SEED}.txt out/${SEED}.txt

LIST=0000 0001 0002 0003 0004 0005 0006 0007 0008 0009
tester10: build
	for i in $(LIST); do \
	./tools/target/release/tester ${BINARY} < tools/in/$$i.txt > tmp.txt;done
	
generateSeeds:
	seq 0 5000 > seed.txt


script: build
	go run script/main.go

setup:
	mkdir -p src
	curl -of src/main.go https://raw.githubusercontent.com/fmhr/AHC_Templates/main/main.go
	curl -of .gitignore https://raw.githubusercontent.com/fmhr/AHC_Templates/main/.gitignore
	cd src && go mod init main

downloadtools:
	curl -o tester.zip https://img.atcoder.jp/atcoder11live/f62fc84.zip
	unzip tester.zip
	rm tester.zip

# gh auth login
dlmain.go:
	gh api -H "Accept: application/vnd.github.v3.raw" \
	"https://api.github.com/repos/fmhr/AHC_Templates/contents/main.go" > src/main.go

dlgitignore:
	gh api -H "Accept: application/vnd.github.v3.raw" \
	"https://api.github.com/repos/fmhr/AHC_Templates/contents/.gitignore" > .gitignore

