package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	PASS        = 0
	BET         = 1
	NUM_ACTIONS = 2
)

var nodeMap = make(map[string]*Node)

type Node struct {
	infoSet     string
	regretSum   [NUM_ACTIONS]float64
	strategy    [NUM_ACTIONS]float64
	strategySum [NUM_ACTIONS]float64
}

func newNode() *Node {
	return &Node{
		regretSum:   [NUM_ACTIONS]float64{0, 0},
		strategy:    [NUM_ACTIONS]float64{0.5, 0.5},
		strategySum: [NUM_ACTIONS]float64{0, 0},
	}
}

func (node *Node) getStrategy(realizationWeight float64) [NUM_ACTIONS]float64 {
	normSum := 0.0
	for a := 0; a < NUM_ACTIONS; a++ {
		if node.regretSum[a] > 0 {
			node.strategy[a] = node.regretSum[a]
		} else {
			node.strategy[a] = 0
		}

		normSum += node.strategy[a]
	}

	for a := 0; a < NUM_ACTIONS; a++ {
		if normSum > 0 {
			node.strategy[a] /= normSum
		} else {
			node.strategy[a] = 1.0 / NUM_ACTIONS
		}

		node.strategySum[a] += realizationWeight * node.strategy[a]
	}

	return node.strategy
}

func (node Node) getAverageStrategy() [NUM_ACTIONS]float64 {
	var avgStrategy [NUM_ACTIONS]float64
	normSum := 0.0
	for a := 0; a < NUM_ACTIONS; a++ {
		normSum += node.strategySum[a]
	}

	for a := 0; a < NUM_ACTIONS; a++ {
		if normSum > 0 {

			avgStrategy[a] = node.strategySum[a] / normSum
		} else {
			avgStrategy[a] = 1.0 / NUM_ACTIONS
		}
	}

	return avgStrategy
}

func cfr(cards []int, history string, p0 float64, p1 float64) float64 {
	plays := len(history)
	player := plays % 2
	opponent := 1 - player

	if plays > 1 {
		terminalPass := history[plays-1] == 'p'
		doubleBet := history[plays-2] == 'b' && history[plays-1] == 'b'
		isPlayerCardHigher := cards[player] > cards[opponent]
		if terminalPass {
			if history == "pp" {
				if isPlayerCardHigher {
					return 1
				}
				return -1
			} else {
				return 1
			}
		} else if doubleBet {
			if isPlayerCardHigher {
				return 2
			}
			return -2
		}
	}
	infoSet := fmt.Sprint(cards[player]) + history

	node, exists := nodeMap[infoSet]
	if !exists {
		node = newNode()
		fmt.Printf("New Node, init strat %v\n", node.strategy)
		node.infoSet = infoSet
		nodeMap[infoSet] = node
	}

	var strat [NUM_ACTIONS]float64
	if player == 0 {
		strat = node.getStrategy(p0)
	} else {
		strat = node.getStrategy(p1)
	}
	var util [NUM_ACTIONS]float64
	nodeUtil := 0.0
	for a := 0; a < NUM_ACTIONS; a++ {
		var nextHistory string
		if a == 0 {
			nextHistory = history + "p"
		} else {
			nextHistory = history + "b"
		}
		if player == 0 {
			util[a] = -cfr(cards, nextHistory, p0*strat[a], p1)
		} else {
			util[a] = -cfr(cards, nextHistory, p0, p1*strat[a])
		}
		nodeUtil += strat[a] * util[a]
	}

	for a := 0; a < NUM_ACTIONS; a++ {
		regret := util[a] - nodeUtil
		if player == 0 {
			node.regretSum[a] += p1 * regret
		} else {
			node.regretSum[a] += p0 * regret
		}
	}

	return nodeUtil
}

func train(iterations int) {
	cards := []int{1, 2, 3}
	util := 0.0
	BATCH := 100000
	batchUtil := 0.0
	for i := 0; i < iterations; i++ {
		if i%BATCH == 0 {
			// fmt.Printf("Iteration %d: average util %f\n", i, batchUtil/float64(BATCH))
			_, ok := nodeMap["3"]
			if ok {
				fmt.Printf("%0.6f\t", nodeMap["1"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["2"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["3"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["1p"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["2p"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["3p"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["1b"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["2b"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["3b"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["1pb"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\t", nodeMap["2pb"].getAverageStrategy()[0])
				fmt.Printf("%0.6f\n", nodeMap["3pb"].getAverageStrategy()[0])
			}
		}
		// Shuffle cards
		for c1 := len(cards) - 1; c1 > 0; c1-- {
			c2 := rand.Intn(c1 + 1)
			tmp := cards[c1]
			cards[c1] = cards[c2]
			cards[c2] = tmp
		}

		iterUtil := cfr(cards, "", 1, 1)
		util += iterUtil
		batchUtil += iterUtil
	}
	fmt.Printf("Average Game Value :%f\n", util/float64(iterations))

	// for _, node := range nodeMap {
	// 	fmt.Printf("Info Set: %s, Strat: %v\n", node.infoSet, node.getAverageStrategy())
	// }
}

func main() {
	rand.Seed(time.Now().Unix())
	train(15000000)
}
