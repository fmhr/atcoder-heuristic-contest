BINARY = a.out
SRC = main.go

.PHONY: build
build: $(BINARY)

$(BINARY): $(SRC)
	echo "Building $(BINARY) ..."
	go build -o ./$(BINARY)


clear:
	rm -rf bin/*
	rm -rf out/*

