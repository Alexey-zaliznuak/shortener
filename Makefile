# Название бинарника после сборки
BINARY_NAME=shortener

# Путь к main.go
MAIN=./cmd/shortener


build_and_run:
	go build -o $(MAIN) $(MAIN)
	./cmd/shortener/shortener

build:
	go build -o $(MAIN) $(MAIN)

test:
	go test -v ./...

run:
	./cmd/shortener/shortener.exe

st:
	./shortenertest-windows-amd64 -test.v -test.run=^TestIteration1 -binary-path=cmd/shortener/shortener.exe
