package cfr

import "math"

type (
	Action     int
	InfoSetKey string
)

type InfoSet struct {
	CumulativeStrategy map[Action]float64
	CumulativeRegret   map[Action]float64
	CurrentStrategy    map[Action]float64
}

func (info *InfoSet) updateStrategy() {
	normalizingSum := 0.0
	for action, regret := range info.CumulativeRegret {
		info.CurrentStrategy[action] = math.Max(regret, 0)
		normalizingSum += info.CurrentStrategy[action]
	}
	for action := range info.CumulativeRegret {
		if normalizingSum > 0 {
			info.CurrentStrategy[action] /= normalizingSum
		} else {
			info.CurrentStrategy[action] = 1.0 / float64(len(info.CumulativeRegret))
		}
	}
}

func makeInfoSet(ValidActions []Action) InfoSet {
	info := InfoSet{
		CumulativeStrategy: make(map[Action]float64),
		CumulativeRegret:   make(map[Action]float64),
		CurrentStrategy:    make(map[Action]float64),
	}

	for index := range ValidActions {
		info.CumulativeStrategy[ValidActions[index]] = 0.0
		info.CumulativeRegret[ValidActions[index]] = 0.0
		info.CurrentStrategy[ValidActions[index]] = 1.0 / float64(len(ValidActions))
	}

	return info
}

type Strategy struct {
	InfoSetMap map[InfoSetKey]InfoSet
}

func NewStrategy() Strategy {
	return Strategy{
		InfoSetMap: make(map[InfoSetKey]InfoSet),
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
	GetUtility(playerID int) float64
	GetInfoSetKey() InfoSetKey
}

func (strat *Strategy) CFR(playerID int, state State, agentPathProbs []float64) float64 {
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
	info, exists := strat.InfoSetMap[InfoSetKey]
	if !exists {
		info = makeInfoSet(validActions)
		strat.InfoSetMap[InfoSetKey] = info
	}

	utility := float64(0.0)
	actionUtility := make(map[Action]float64)
	for _, action := range validActions {
		actionProb := info.CurrentStrategy[action]

		newPathProbs := make([]float64, len(agentPathProbs))
		copy(newPathProbs, agentPathProbs)
		newPathProbs[currentAgent] = newPathProbs[currentAgent] * actionProb

		actionUtility[action] = strat.CFR(playerID, state.TakeActionCopy(action), newPathProbs)
		utility += (actionProb * actionUtility[action])
	}

	if currentAgent == playerID {
		// Find the probability of attempting to reach the current state
		nonPlayerPathProb := float64(1.0)
		for index, prob := range agentPathProbs {
			if index != playerID {
				nonPlayerPathProb *= prob
			}
		}

		for _, action := range validActions {
			info.CumulativeRegret[action] += nonPlayerPathProb * (actionUtility[action] - utility)
			info.CumulativeStrategy[action] += agentPathProbs[playerID] * info.CurrentStrategy[action]
		}

		info.updateStrategy()
	}

	return utility
}
