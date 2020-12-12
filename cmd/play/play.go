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

	// for i := range game.Agents {
	// 	if i%2 == 0 {
	// 		game.Agents[i] = cfr.OptimalAgent{}
	// 	} else {
	// 		//game.Agents[i] = cfr.CFRAgent{Strat: &strat}
	// 		game.Agents[i] = cfr.OptimalAgent{}
	// 	}
	// }
	game.Agents[0] = cfr.RandomAgent{}
	game.Agents[1] = &cfr.CFRAgent{Strat: strat}
	game.Agents[2] = cfr.RandomAgent{}
	game.Agents[3] = &cfr.CFRAgent{Strat: strat}

	evenWins := 0
	oddWins := 0
	for i := 0; i < 100; i++ {
		state = cfr.NewEuchreState()
		game.GameState = &state
		//strat := cfr.NewStrategy()

		utilities := game.Play()
		if utilities[0] > utilities[1] {
			evenWins += 1
		} else {
			oddWins += 1

		}
		fmt.Printf("Completed game %d\n", i)
	}
	fmt.Printf("Even Score: %d, Odd score %d\n", evenWins, oddWins)

}
