BINARY=bin/a
SRC=src/main.go
SEED=0000

build:
	go build -o $(BINARY) $(SRC)

example: build
	${BINARY} < tools/in/${SEED}.txt > out/${SEED}.txt

tester: example	
	./tools/target/release/vis tools/in/${SEED}.txt out/${SEED}.txt

LIST=0000 0001 0002 0003 0004 0005 0006 0007 0008 0009
tester10: build
	for i in $(LIST); do \
	./$(BINARY) < tools/in/$$i.txt > tmp.txt && \
	./tools/target/release/vis tools/in/$$i.txt tmp.txt ;done
	
generateSeeds:
	seq 0 3000 > seed.txt

setuptools:
	curl -O https://img.atcoder.jp/ahc017/a7a9e57e.zip
	unzip a7a9e57e.zip
	rm a7a9e57e.zip

script: build
	go run script/main.go