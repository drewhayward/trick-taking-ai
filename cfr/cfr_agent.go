package cfr

// CFRAgent ...
type CFRAgent struct {
	Strat Strategy
}

// Act ...
func (agent CFRAgent) Act(state State) Action {
	var euchreState *EuchreState = state.(*EuchreState)
	// Abstract state
	oldTrump := euchreState.trumpSuit
	euchreState.normalizeTrump(oldTrump)

	// Train CFR on only this information set
	for i := 0; i < 10; i++ {
		// Sample another state in the same info set
		sampledState := euchreState.SampleInfoSet()

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

	key := euchreState.GetInfoSetKey()
	newStrategy := agent.Strat.InfoSetMap[key].getStateStrategy()
	action := sampleAction(newStrategy)

	// Need to unnormalize the action too
	euchrePlay := Card(action)
	euchrePlay.normalizeSuit(oldTrump)
	action = Action(euchrePlay)

	euchreState.unnormalizeTrump(oldTrump)
	return action
}
