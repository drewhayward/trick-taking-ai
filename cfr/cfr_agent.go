package cfr

import "fmt"

// CFRAgent ...
type CFRAgent struct {
	Strat         *Strategy
	NumIterations int
}

func (agent *CFRAgent) ClearStrategy() {
	strat := NewStrategy()
	agent.Strat = &strat
}

// Act ...
func (agent *CFRAgent) Act(state State) Action {
	var euchreState *EuchreState = state.(*EuchreState)

	if len(euchreState.ValidActions()) == 1 {
		return euchreState.ValidActions()[0]
	}
	// Abstract state
	oldTrump := euchreState.trumpSuit
	euchreState.normalizeTrump(oldTrump)
	if euchreState.trumpSuit != SPADES {
		panic("Trump abstraction failed")
	}
	key := euchreState.GetInfoSetKey()

	// Train CFR on only this information set
	for i := 0; i < agent.NumIterations; i++ {
		// Sample another state in the same info set
		sampledState, err := euchreState.SampleInfoSet()
		attempts := 0
		for err != nil {
			sampledState, err = euchreState.SampleInfoSet()
			fmt.Printf("Sample Attempt %d\n", attempts)
			attempts++
		}
		// Train on it
		// TODO: Maybe we can ignore training the positions of the other players?
		// 	This would save memory and time.
		for playerID := 0; playerID < 4; playerID++ {
			probs := make([]float64, 4)
			for p := 0; p < 4; p++ {
				probs[p] = 1.0
			}
			agent.Strat.CFR(playerID, &sampledState, probs)
		}
	}

	infoSet := agent.Strat.InfoSetMap[key]
	newStrategy := infoSet.getStateStrategy()

	action := sampleAction(newStrategy)

	// Need to unnormalize the action too
	euchrePlay := Card(action)
	euchrePlay.normalizeSuit(oldTrump)
	action = Action(euchrePlay)

	euchreState.unnormalizeTrump(oldTrump)

	agent.ClearStrategy()

	return action
}
