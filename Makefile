# Название бинарника после сборки
BINARY_NAME=shortener.exe

# Путь к main.go
MAIN=./cmd/shortener


build_and_run:
	go build -o $(MAIN) $(MAIN)
	./cmd/shortener/shortener.exe

# build:
# 	go build -o $(MAIN) $(MAIN)

# run:
# 	./cmd/shortener/shortener.exe
