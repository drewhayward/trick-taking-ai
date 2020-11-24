package cfr

type (
	Action     int
	InfoSetKey string
)

type infoSet struct {
	cumulativeStrategy map[Action]float32
	cumulativeRegret   map[Action]float32
	currentStrategy    map[Action]float32
}

func (info *infoSet) updateStrategy() {
	// TODO: Implement
}

func makeInfoSet(ValidActions []Action) *infoSet {
	info := infoSet{
		cumulativeStrategy: make(map[Action]float32),
		cumulativeRegret:   make(map[Action]float32),
		currentStrategy:    make(map[Action]float32),
	}

	for index := range ValidActions {
		info.cumulativeStrategy[ValidActions[index]] = 0.0
		info.cumulativeRegret[ValidActions[index]] = 0.0
		info.currentStrategy[ValidActions[index]] = 1.0 / float32(len(ValidActions))
	}

	return &info
}

type Strategy struct {
	infoSetMap map[InfoSetKey]*infoSet
}

func NewStrategy() Strategy {
	return Strategy{
		infoSetMap: make(map[InfoSetKey]*infoSet),
	}
}

type State interface {
	ValidActions() []Action
	// Should return a pointer to the same state object to minimize copying
	TakeAction(Action) State
	// Pass by value so the state is duplicated
	TakeActionCopy(Action) State
	IsTerminal() bool
	GetCurrentAgent() int
	GetUtility(playerID int) float32
	GetInfoSetKey() InfoSetKey
}

func (strat *Strategy) CFR(playerID int, pstate State, agentPathProbs []float32) float32 {
	state := pstate
	currentAgent := state.GetCurrentAgent()

	if state.IsTerminal() {
		return state.GetUtility(playerID)
	}

	validActions := state.ValidActions()

	if len(validActions) == 1 {
		// Pass the state by reference since we don't need to run other actions
		return strat.CFR(playerID, state.TakeAction(validActions[0]), agentPathProbs)
	}

	InfoSetKey := state.GetInfoSetKey()
	info, exists := strat.infoSetMap[InfoSetKey]
	if !exists {
		info = makeInfoSet(validActions)
		strat.infoSetMap[InfoSetKey] = info
	}

	utility := float32(0.0)
	actionUtility := make(map[Action]float32)
	for _, action := range validActions {
		actionProb := info.currentStrategy[action]

		newPathProbs := make([]float32, len(agentPathProbs))
		copy(newPathProbs, agentPathProbs)
		newPathProbs[currentAgent] = newPathProbs[currentAgent] * actionProb

		actionUtility[action] = strat.CFR(playerID, state.TakeActionCopy(action), newPathProbs)
		utility += (actionProb * actionUtility[action])
	}

	if currentAgent == playerID {
		// Find the probability of attempting to reach the current state
		nonPlayerPathProb := float32(1.0)
		for index, prob := range agentPathProbs {
			if index != playerID {
				nonPlayerPathProb *= prob
			}
		}

		for _, action := range validActions {
			info.cumulativeRegret[action] += nonPlayerPathProb * (actionUtility[action] - utility)
			info.cumulativeStrategy[action] += agentPathProbs[playerID] * info.currentStrategy[action]
		}

		info.updateStrategy()
	}

	return utility
}
