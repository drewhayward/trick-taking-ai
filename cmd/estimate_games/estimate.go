package main

import (
	"github.com/drewhayward/trick-taking-ai/cfr"
)

func count(state cfr.State) int {
	if state.IsTerminal() {
		return 1
	}

	total := 0
	for _, action := range state.ValidActions() {
		total += count(state.TakeActionCopy(action))
	}
	return total
}

func main() {
	iters := 100
	for i := 0; i < iters; i++ {
		state := cfr.NewEuchreState()
		println(count(&state))
	}
}
