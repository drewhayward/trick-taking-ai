package cfr

import (
	"math/rand"
)

// RandomAgent ...
type RandomAgent struct {
}

func (agent RandomAgent) EndGame() {

}

// Act chooses from the valid actions with uniform probability
func (agent RandomAgent) Act(state State) Action {
	actions := state.ValidActions()
	actionIndex := rand.Intn(len(actions))
	return actions[actionIndex]
}
