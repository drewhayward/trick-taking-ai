package cfr

// CFRAgent ...
type CFRAgent struct {
	Strat         *Strategy
	NumIterations int
}

func (agent *CFRAgent) ClearStrategy() {
	strat := NewStrategy()
	agent.Strat = &strat
}

func (agent *CFRAgent) EndGame() {
	agent.ClearStrategy()
}

// Act ...
func (agent *CFRAgent) Act(state State) Action {
	var euchreState *EuchreState = state.(*EuchreState)

	if len(euchreState.ValidActions()) == 1 {
		return euchreState.ValidActions()[0]
	}
	// Abstract state
	oldTrump := euchreState.TrumpSuit
	euchreState.Normalize(oldTrump)
	if euchreState.TrumpSuit != SPADES {
		panic("Trump abstraction failed")
	}
	key := euchreState.GetInfoSetKey()

	// Train CFR on only this information set
	for i := 0; i < agent.NumIterations; i++ {
		// Sample another state in the same info set
		sampledState, _ := euchreState.SampleInfoSet()

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

	euchreState.Unnormalize(oldTrump)

	//agent.ClearStrategy()

	return action
}
