BINARY = bin/a.out
SRC = src/main.go

.PHONY: build
build: $(BINARY)
$(BINARY): $(SRC)
	@echo "Building $(BINARY) ..."
	go build -o $(BINARY) $(SRC)

