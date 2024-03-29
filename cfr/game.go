package cfr

// Agent ...
type Agent interface {
	Act(State) Action
	EndGame()
}

// Game ...
type Game struct {
	GameState State
	Agents    []Agent
}

// Play ...
func (game *Game) Play() []float64 {
	// While the game is not complete
	turnsTaken := 0
	for !game.GameState.IsTerminal() {
		currentAgent := game.GameState.GetCurrentAgent()

		agentAction := game.Agents[currentAgent].Act(game.GameState)
		// Take a move according to the current agent's policy
		game.GameState.TakeAction(agentAction)
		turnsTaken++
	}

	for _, agent := range game.Agents {
		agent.EndGame()
	}

	// Get resulting utility
	playerUtilities := make([]float64, 4)
	for i := 0; i < 4; i++ {
		playerUtilities[i] = game.GameState.GetUtility(i)
	}

	return playerUtilities
}
