package main

import (
	"fmt"

	"github.com/drewhayward/trick-taking-ai/cfr"
)

func main() {
	state := cfr.NewEuchreState()
	newState := state.SampleInfoSet()
	fmt.Println(newState)
}
