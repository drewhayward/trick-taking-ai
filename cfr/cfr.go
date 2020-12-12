package cfr

import (
	"math"
	"math/rand"
)

type (
	Action     int
	InfoSetKey string
)

// InfoSet ...
type InfoSet struct {
	CumulativeStrategySum map[Action]float64
	CumulativeRegret      map[Action]float64
	CurrentStrategy       map[Action]float64
}

func (info InfoSet) getStateStrategy() map[Action]float64 {
	var norm float64 = 0.0
	stateStrategy := make(map[Action]float64)
	for action, regret := range info.CumulativeRegret {
		if regret > 0 {
			stateStrategy[action] = regret
		} else {
			stateStrategy[action] = 0
		}
		norm += stateStrategy[action]
	}

	for action := range info.CumulativeStrategySum {
		if norm > 0 {
			stateStrategy[action] = stateStrategy[action] / norm
		} else {
			stateStrategy[action] = 1.0 / float64(len(info.CumulativeRegret))
		}
	}
	return stateStrategy
}

func sampleAction(stateStrat map[Action]float64) Action {
	num := rand.Float64()
	sum := 0.0
	for action, likelihood := range stateStrat {
		sum += likelihood
		if sum > num {
			return action
		}
	}
	panic("Unable to sample action for strategy. Maybe strategy is not a valid distribution?")
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
		CumulativeStrategySum: make(map[Action]float64),
		CumulativeRegret:      make(map[Action]float64),
		CurrentStrategy:       make(map[Action]float64),
	}

	for index := range ValidActions {
		info.CumulativeStrategySum[ValidActions[index]] = 0.0
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
	TakeAction(Action, bool) State
	// Pass by value so the state is duplicated
	TakeActionCopy(Action) State
	IsTerminal() bool
	GetCurrentAgent() int
	GetUtility(playerID int) float64
	GetInfoSetKey() InfoSetKey
}

func (strat *Strategy) CFR(playerID int, state State, agentPathProbs []float64) float64 {
	return strat.cfr(playerID, state, agentPathProbs, 0)
}
func (strat *Strategy) cfr(playerID int, state State, agentPathProbs []float64, depth int) float64 {
	currentAgent := state.GetCurrentAgent()

	if state.IsTerminal() {
		return state.GetUtility(playerID)
	}

	validActions := state.ValidActions()

	if len(validActions) == 1 {
		// Pass the state by reference since we don't need to run other actions
		return strat.cfr(playerID, state.TakeAction(validActions[0], false), agentPathProbs, depth+1)
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

		actionUtility[action] = strat.cfr(playerID, state.TakeActionCopy(action), newPathProbs, depth+1)
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
			info.CumulativeStrategySum[action] += agentPathProbs[playerID] * info.CurrentStrategy[action]
		}

		info.updateStrategy()
	}

	return utility
}
