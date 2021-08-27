/*
defines basic regret matching for a two player game of RPS

Each player starts with a skewed strategy and then uses regret matching to
converge towards the equilibrium strategy
*/
package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	ROCK        = 0
	PAPER       = 1
	SCISSORS    = 2
	NUM_ACTIONS = 3
)

const NUM_PLAYERS = 2

type Player struct {
	regretSum   [NUM_ACTIONS]float64
	strategy    [NUM_ACTIONS]float64
	strategySum [NUM_ACTIONS]float64
}

func newPlayer() *Player {
	return &Player{
		regretSum:   [NUM_ACTIONS]float64{0.0, 0.0, 0.0},
		strategy:    [NUM_ACTIONS]float64{0.05, 0.05, 0.9},
		strategySum: [NUM_ACTIONS]float64{0.0, 0.0, 0.0},
	}
}

func getStrategy(player *Player) [NUM_ACTIONS]float64 {
	var normSum float64 = 0
	for a := 0; a < NUM_ACTIONS; a++ {
		if player.regretSum[a] > 0 {
			player.strategy[a] = player.regretSum[a]
		} else {
			player.strategy[a] = 0
		}

		normSum += player.strategy[a]
	}

	for a := 0; a < NUM_ACTIONS; a++ {
		if normSum > 0 {
			player.strategy[a] /= normSum
		} else {
			player.strategy[a] = 1.0 / NUM_ACTIONS
		}

		player.strategySum[a] += player.strategy[a]
	}

	return player.strategy
}

func getAction(strategy [NUM_ACTIONS]float64) int {
	var r = rand.Float64()
	a := 0
	cumProb := 0.0

	for a < NUM_ACTIONS-1 {
		cumProb += strategy[a]
		if r < cumProb {
			break
		}

		a += 1
	}

	return a
}

func train(iterations int, players []*Player) {
	var actionUtilty [NUM_ACTIONS]float64
	var actions [NUM_PLAYERS]int
	for i := 0; i < iterations; i++ {
		//  At each iteration, both players update their regrets as above
		// and then both each player computes their own new strategy based on
		// their own regret tables.

		// Sample actions for both players
		for pNum, player := range players {
			strategy := getStrategy(player)
			action := getAction(strategy)
			actions[pNum] = action
		}

		for pNum, player := range players {
			myAction := actions[pNum]
			otherAction := actions[1-pNum]

			// Set action utility according to the other players action
			actionUtilty[otherAction] = 0
			if otherAction == NUM_ACTIONS-1 {
				actionUtilty[0] = 1
			} else {
				actionUtilty[otherAction+1] = 1
			}
			if otherAction == 0 {
				actionUtilty[NUM_ACTIONS-1] = -1
			} else {
				actionUtilty[otherAction-1] = -1
			}

			for a := 0; a < NUM_ACTIONS; a++ {
				player.regretSum[a] += actionUtilty[a] - actionUtilty[myAction]
			}
		}
	}
}

func getAverageStrategy(player Player) [NUM_ACTIONS]float64 {
	var avgStrat [NUM_ACTIONS]float64
	var normSum float64 = 0
	for a := 0; a < NUM_ACTIONS; a++ {
		normSum += player.strategySum[a]
	}
	for a := 0; a < NUM_ACTIONS; a++ {
		if normSum > 0 {
			avgStrat[a] = player.strategySum[a] / normSum
		} else {
			avgStrat[a] = 1.0 / NUM_ACTIONS
		}
	}

	return avgStrat
}

func main() {
	rand.Seed(time.Now().Unix())

	var players = []*Player{newPlayer(), newPlayer()}

	train(20000, players)

	fmt.Printf("%v\n", getAverageStrategy(*players[0]))
	fmt.Printf("%v\n", getAverageStrategy(*players[1]))
}
