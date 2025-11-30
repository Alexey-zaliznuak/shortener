package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("ok")        // разрешено — OK, без want
	os.Exit(0)             // разрешено — OK
	panic("bad")           // want "usage of panic is not allowed"
}

func helper() {
	log.Fatal("no")        // want "log.Fatal is not allowed outside main function in main package"
	os.Exit(1)             // want "os.Exit is not allowed outside main function in main package"
	panic("also bad")      // want "usage of panic is not allowed"
}
