# Название бинарника после сборки
BINARY_NAME=shortener

# Путь к main.go
MAIN=./cmd/shortener


build_and_run:
	go build -o $(MAIN) $(MAIN)
	./cmd/shortener/shortener

build:
	go build -o $(MAIN) $(MAIN)
b:
	$(MAKE) build


test:
	go test -v ./...
t:
	$(MAKE) test


run:
	./cmd/shortener/shortener.exe
r:
	$(MAKE) run


st:
	./shortenertest-windows-amd64 -test.v -test.run=^TestIteration -binary-path=cmd/shortener/shortener.exe
