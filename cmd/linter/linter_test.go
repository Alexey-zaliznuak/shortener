package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoExit(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Linter, "noexit", "noexit_utils")
}
