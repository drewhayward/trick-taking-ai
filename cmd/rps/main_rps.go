package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/drewhayward/trick-taking-ai/cfr"
)

func main() {
	strat := cfr.NewStrategy()

	// Load strategy file
	dataFile, err := os.Open("strategy.gob")
	if err == nil {
		stratDecoder := gob.NewDecoder(dataFile)
		err = stratDecoder.Decode(&strat.InfoSetMap)
		dataFile.Close()

		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	begin := time.Now().UnixNano()
	util := 0.0
	maxIter := 10
	for iter := 0; iter < maxIter; iter++ {
		state := cfr.NewEuchreState()
		//sampledState, _ := state.SampleInfoSet()
		trump := state.TrumpSuit
		state.Normalize(trump)
		for playerId := 0; playerId < 4; playerId++ {
			probs := make([]float64, 4)
			for p := 0; p < 4; p++ {
				probs[p] = 1.0
			}

			util += strat.CFR(playerId, &state, probs)
			fmt.Printf("There are %d info sets in the map.\n", len(strat.InfoSetMap))
		}
	}
	end := time.Now().UnixNano()

	dataFile, err = os.Create("strategy.gob")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	stratEncoder := gob.NewEncoder(dataFile)
	err = stratEncoder.Encode(strat.InfoSetMap)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	fmt.Printf("Mean time %f seconds per iteration\n", (float64(end-begin)/float64(1e9))/float64(maxIter*4))
	fmt.Printf("Average utility %f\n", util/float64(maxIter))
}
