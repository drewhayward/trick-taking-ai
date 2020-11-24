package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/drewhayward/trick-taking-ai/cfr"
)

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	/* START CODE */
	state := cfr.NewEuchreState()
	for i := 0; i < 4; i++ {
		state.TakeAction(state.ValidActions()[0])
	}

	strat := cfr.NewStrategy()
	begin := time.Now().UnixNano()
	maxIter := 1
	for i := 0; i < maxIter; i++ {
		probs := make([]float32, 4)
		for p := 0; p < 4; p++ {
			probs[p] = 1.0
		}

		strat.CFR(i, &state, probs)
	}
	end := time.Now().UnixNano()
	fmt.Printf("It took %f seconds to run %d iterations\n", float64(end-begin)/float64(1e9), maxIter)

	/* END CODE */

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}