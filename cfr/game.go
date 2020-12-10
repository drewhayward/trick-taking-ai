package cfr

// Agent ...
type Agent interface {
	Act(State) Action
}

// Game ...
type Game struct {
	GameState State
	Agents    []Agent
}

// Play ...
func (game *Game) Play() []float64 {
	// While the game is not complete
	for !game.GameState.IsTerminal() {
		currentAgent := game.GameState.GetCurrentAgent()

		// Take a move according to the current agent's policy
		game.GameState.TakeAction(game.Agents[currentAgent].Act(game.GameState))
	}

	// Get resulting utility
	playerUtilities := make([]float64, 4)
	for i := 0; i < 4; i++ {
		playerUtilities[i] = game.GameState.GetUtility(i)
	}

	return playerUtilities
}
