package main

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// build an executable just for testing. Named 'nhmlg:test' to minimize any possible collission with aliases in user space.
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"nhmlg:test": Main,
	}))
}

func TestCommands(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
