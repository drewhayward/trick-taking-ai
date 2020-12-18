package cfr

import (
	"math"
)

// OptimalAgent ...
type OptimalAgent struct {
}

func (agent OptimalAgent) EndGame() {

}

// Act calculates the best action for an optimal player who
// can observe all cards/outcomes
func (agent OptimalAgent) Act(state State) Action {
	_, action := minimax(state, state.GetCurrentAgent())
	return action
}

func minimax(state State, maximizingPlayer int) (float64, Action) {
	if state.IsTerminal() {
		return state.GetUtility(maximizingPlayer), Action(0)
	}

	currentTeam := state.GetCurrentAgent() % 2
	maxTeam := maximizingPlayer % 2
	if currentTeam == maxTeam {
		value := math.Inf(-1)
		bestAction := Action(0)
		for _, action := range state.ValidActions() {
			actionValue, _ := minimax(state.TakeActionCopy(action), maximizingPlayer)
			if actionValue > value {
				value = actionValue
				bestAction = action
			}
		}

		return value, bestAction
	} else {
		value := math.Inf(1)
		worstAction := Action(0)
		for _, action := range state.ValidActions() {
			actionValue, _ := minimax(state.TakeActionCopy(action), maximizingPlayer)
			if actionValue < value {
				value = actionValue
				worstAction = action
			}
		}

		return value, worstAction
	}
}
