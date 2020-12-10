package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/drewhayward/trick-taking-ai/cfr"
)

func main() {
	// var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	// var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	// flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create CPU profile: ", err)
	// 	}
	// 	defer f.Close() // error handling omitted for example
	// 	if err := pprof.StartCPUProfile(f); err != nil {
	// 		log.Fatal("could not start CPU profile: ", err)
	// 	}
	// 	defer pprof.StopCPUProfile()
	// }

	/* START CODE */

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
		for playerId := 0; playerId < 4; playerId++ {
			probs := make([]float64, 4)
			for p := 0; p < 4; p++ {
				probs[p] = 1.0
			}

			util += strat.CFR(playerId, &state, probs)
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

	/* END CODE */

	// if *memprofile != "" {
	// 	f, err := os.Create(*memprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create memory profile: ", err)
	// 	}
	// 	defer f.Close() // error handling omitted for example
	// 	runtime.GC()    // get up-to-date statistics
	// 	if err := pprof.WriteHeapProfile(f); err != nil {
	// 		log.Fatal("could not write memory profile: ", err)
	// 	}
	// }
}
