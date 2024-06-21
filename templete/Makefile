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

setuptools:
	curl -o tester.zip https://img.atcoder.jp/ahc018/6bada50282.zip
	unzip tester.zip
	rm tester.zip

script: build
	go run script/main.go