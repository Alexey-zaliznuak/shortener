package utils

import (
	"log"
	"os"
)

func f() {
	panic("no")     // want "usage of panic is not allowed"
	log.Fatal("no") // want "log.Fatal is not allowed outside main function in main package"
	os.Exit(1)      // want "os.Exit is not allowed outside main function in main package"
}
