package main

import (
	"fmt"

	"github.com/drewhayward/trick-taking-ai/cfr"
)

func main() {
	state := cfr.NewEuchreState()
	strat := cfr.NewStrategy()
	game := cfr.Game{
		GameState: &state,
		Agents:    make([]cfr.Agent, 4),
	}

	for i := range game.Agents {
		if i%2 == 0 {
			game.Agents[i] = cfr.OptimalAgent{}
		} else {
			game.Agents[i] = cfr.CFRAgent{Strat: strat}
		}
	}

	fmt.Println(game.Play())
}
