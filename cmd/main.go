package main

import (
	"fmt"
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
	// clone := state.Clone()
	// fmt.Println(clone)
	// for i := 0; i < 4; i++ {
	// 	state.TakeAction(state.ValidActions()[0])
	// }

	strat := cfr.NewStrategy()
	begin := time.Now().UnixNano()
	maxIter := 100
	for i := 0; i < maxIter; i++ {
		state := cfr.NewEuchreState()
		probs := make([]float32, 4)
		for p := 0; p < 4; p++ {
			probs[p] = 1.0
		}

		strat.CFR(i, &state, probs)
	}
	end := time.Now().UnixNano()
	fmt.Printf("Mean time %f seconds per iteration\n", (float64(end-begin)/float64(1e9))/float64(maxIter))

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
