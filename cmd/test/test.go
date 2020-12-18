package main

import (
	"fmt"

	"github.com/drewhayward/trick-taking-ai/cfr"
)

func main() {

	for suit := 10; suit <= 40; suit += 10 {
		for value := 1; value < 7; value++ {
			card := cfr.Card(suit + value)
			transformed := cfr.TrumpRankTransform(card, cfr.SPADES)
			fmt.Printf("Card %s, value %d\n", card.ToString(), transformed)
		}
	}
	//fmt.Println(newState)
}
