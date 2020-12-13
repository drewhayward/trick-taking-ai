package cfr

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// EuchreAction ...
// Actions defined as the sum of the the suit and Value for
// fast calculation of play actions
type EuchreAction uint8

const (
	// Play actions
	// Diamonds
	PLAY_AD = EuchreAction(int(DIAMONDS) + int(ACE))
	PLAY_KD = EuchreAction(int(DIAMONDS) + int(KING))
	PLAY_QD = EuchreAction(int(DIAMONDS) + int(QUEEN))
	PLAY_JD = EuchreAction(int(DIAMONDS) + int(JACK))
	PLAY_TD = EuchreAction(int(DIAMONDS) + int(TEN))
	PLAY_ND = EuchreAction(int(DIAMONDS) + int(NINE))
	// Hearts
	PLAY_AH = EuchreAction(int(HEARTS) + int(ACE))
	PLAY_KH = EuchreAction(int(HEARTS) + int(KING))
	PLAY_QH = EuchreAction(int(HEARTS) + int(QUEEN))
	PLAY_JH = EuchreAction(int(HEARTS) + int(JACK))
	PLAY_TH = EuchreAction(int(HEARTS) + int(TEN))
	PLAY_NH = EuchreAction(int(HEARTS) + int(NINE))
	// Spades
	PLAY_AS = EuchreAction(int(SPADES) + int(ACE))
	PLAY_KS = EuchreAction(int(SPADES) + int(KING))
	PLAY_QS = EuchreAction(int(SPADES) + int(QUEEN))
	PLAY_JS = EuchreAction(int(SPADES) + int(JACK))
	PLAY_TS = EuchreAction(int(SPADES) + int(TEN))
	PLAY_NS = EuchreAction(int(SPADES) + int(NINE))
	// Clubs
	PLAY_AC = EuchreAction(int(CLUBS) + int(ACE))
	PLAY_KC = EuchreAction(int(CLUBS) + int(KING))
	PLAY_QC = EuchreAction(int(CLUBS) + int(QUEEN))
	PLAY_JC = EuchreAction(int(CLUBS) + int(JACK))
	PLAY_TC = EuchreAction(int(CLUBS) + int(TEN))
	PLAY_NC = EuchreAction(int(CLUBS) + int(NINE))
	// Bidding actions
	// PASS_BID      = uint8(7)
	// CALL_DIAMONDS = DIAMONDS
	// CALL_HEARTS   = HEARTS
	// CALL_SPADES   = SPADES
	// CALL_CLUBS    = CLUBS
	// Discard Actions
	// DISCARD_0 = uint8(0)
	// DISCARD_1 = uint8(1)
	// DISCARD_2 = uint8(2)
	// DISCARD_3 = uint8(3)
	// DISCARD_4 = uint8(4)
	// DISCARD_5 = uint8(5)
)

// EuchreState stores the current game state of a euchre hand
type EuchreState struct {
	playerHands [4][]Card
	shortSuited [4][]Suit
	table       []Card
	history     []Card
	kitty       []Card
	teamTricks  [2]int

	leadSuit     Suit
	trumpSuit    Suit
	lead         int
	callingTeam  int
	currentAgent int
}

// NewEuchreState ...
func NewEuchreState() EuchreState {
	rand.Seed(time.Now().UnixNano())
	var deck []Card
	for value := 1; value < 7; value++ {
		for suit := 10; suit <= 40; suit += 10 {
			deck = append(deck, Card(suit+value))
		}
	}
	deck = shuffle(deck)

	leadPlayer := rand.Intn(4)
	state := EuchreState{
		lead:         leadPlayer,
		callingTeam:  rand.Intn(2),
		currentAgent: leadPlayer,
		teamTricks:   [2]int{0, 0},
		table:        make([]Card, 0, 4),
		history:      make([]Card, 0, 24),
		kitty:        make([]Card, 0, 4),
	}

	// Deal cards
	for i := 0; i < 4; i++ {
		state.shortSuited[i] = make([]Suit, 0)
		state.playerHands[i] = make([]Card, 5)
		for c := 0; c < 5; c++ {
			state.playerHands[i][c] = deck[c+i*5]
		}
		// Sort hands
		sort.Slice(state.playerHands[i], func(j, k int) bool {
			return state.playerHands[i][j] < state.playerHands[i][k]
		})
	}
	state.kitty = append(state.kitty, deck[20:]...)
	state.trumpSuit = state.kitty[0].getSuit()

	return state
}

func (state EuchreState) SampleInfoSet() (EuchreState, error) {
	key := state.GetInfoSetKey()
	newState := state.Clone()
	newState.CheckCards()

	// Collect unknown cards
	intermediateDeck := make([]Card, 0)
	for i, hand := range state.playerHands {
		if i != newState.currentAgent {
			intermediateDeck = append(intermediateDeck, hand...)
			newState.playerHands[i] = make([]Card, 0, len(hand))
		}
	}
	intermediateDeck = append(intermediateDeck, newState.kitty[1:]...)

	intermediateDeck = shuffle(intermediateDeck)

	// Need to deal cards that can only fit in one hand to that person
	// Count number of compatible cards per hand
	type pair struct {
		count int
		index int
	}
	counts := []pair{{0, 0}, {0, 1}, {0, 2}, {0, 3}}
	for handIdx := 0; handIdx < 4; handIdx++ {
		for _, card := range intermediateDeck {
			if !inSlice(newState.shortSuited[handIdx], card.effectiveSuit(newState.trumpSuit)) {
				counts[handIdx].count++
			}
		}
	}
	sort.Slice(counts, func(j, k int) bool {
		return counts[j].count < counts[k].count
	})
	// Deal from most constrained to least
	for _, countPair := range counts {
		handIdx := countPair.index
		suits := newState.shortSuited[handIdx]
		if handIdx == newState.currentAgent {
			continue
		}
		// Advance through the shuffled cards until a valid card is found
		for j := 0; len(newState.playerHands[handIdx]) < cap(newState.playerHands[handIdx]); j++ {
			if j >= len(intermediateDeck) {
				return EuchreState{}, errors.New("Invalid shuffle")
			}
			deckCard := intermediateDeck[j]
			if deckCard != 0 && !inSlice(suits, deckCard.effectiveSuit(newState.trumpSuit)) {
				newState.playerHands[handIdx] = append(newState.playerHands[handIdx], deckCard)
				intermediateDeck[j] = 0
			}
		}
		// Keep the hands sorted
		sort.Slice(newState.playerHands[handIdx], func(j, k int) bool {
			return newState.playerHands[handIdx][j] < newState.playerHands[handIdx][k]
		})

		// Need to reshuffle to unbias the next hands deal
		intermediateDeck = shuffle(intermediateDeck)
	}

	newState.kitty = make([]Card, len(state.kitty))
	newState.kitty[0] = state.kitty[0]
	// Put remaining in kitty
	for i := 1; i < 4; i++ {
		for j := 0; j < len(intermediateDeck); j++ {
			if intermediateDeck[j] != 0 {
				newState.kitty[i] = intermediateDeck[j]
				intermediateDeck[j] = 0
				break
			}
		}
	}
	newKey := newState.GetInfoSetKey()
	if key != newKey {
		panic("Incorrect sampling, key should remain the same")
	}

	newState.CheckCards()
	return newState, nil
}

func inSlice(slice []Suit, suit Suit) bool {
	for _, item := range slice {
		if item == suit {
			return true
		}
	}
	return false
}

// Clone ...
func (state EuchreState) Clone() EuchreState {
	newState := state

	// Copy hands
	for handIdx, hand := range state.playerHands {
		newState.playerHands[handIdx] = make([]Card, len(hand))

		for cIdx, card := range hand {
			newState.playerHands[handIdx][cIdx] = card
		}

	}

	// Copy shortsuitedness
	for handIdx, suits := range state.shortSuited {
		newState.shortSuited[handIdx] = make([]Suit, len(suits))

		for sIdx, suit := range suits {
			newState.shortSuited[handIdx][sIdx] = suit
		}
	}

	newState.kitty = make([]Card, len(state.kitty))
	for cIdx, card := range state.kitty {
		newState.kitty[cIdx] = card
	}

	return newState
}

// ValidActions ...
func (state *EuchreState) ValidActions() []Action {
	hand := state.playerHands[state.currentAgent]

	// Playing the hand
	var playableActions []Action
	if state.leadSuit != 0 {
		// Follow suit if possible
		playableActions = make([]Action, 0, len(hand))
		var lastCard Card = 0
		for _, card := range hand {
			skipAction := (lastCard != 0) && (card == (lastCard + 1))
			if !skipAction && (card.effectiveSuit(state.trumpSuit) == state.leadSuit) {
				playableActions = append(playableActions, Action(card))
			}

			lastCard = card
		}

		// If no lead suit, play normally
		if len(playableActions) == 0 {
			var lastCard Card = 0
			for _, card := range hand {
				skipAction := (card == (lastCard + 1))
				if !skipAction {
					playableActions = append(playableActions, Action(card))
				}

				lastCard = card
			}
		}
	} else {
		// TODO: Fix the action reduction for bowers
		playableActions = make([]Action, 0, len(hand))
		var lastCard Card = 0
		for _, card := range hand {
			skipAction := (card == (lastCard + 1))
			if !skipAction {
				playableActions = append(playableActions, Action(card))
			}

			lastCard = card
		}
	}

	if len(playableActions) > 5 {
		panic("Too many playable card actions")
	}

	return playableActions
}

// TakeAction ...
func (state *EuchreState) TakeAction(action Action, narrate bool) State {

	if narrate {
		fmt.Println("-----")
		fmt.Printf("Current Score %d-%d\n", state.teamTricks[0], state.teamTricks[1])
		fmt.Printf("Calling Team: team %d\n", state.callingTeam)
		fmt.Printf("Trump Suit %s\n", state.trumpSuit.toString())
		fmt.Printf("Lead Suit %s\n", state.leadSuit.toString())
		fmt.Printf("Table state:\n")
		for i, card := range state.table {
			if i == 0 {
				fmt.Printf("\tPlayer %d lead the %s\n", state.lead, card.toString())
			} else {
				fmt.Printf("\tPlayer %d played the %s\n", (state.lead+i)%4, card.toString())
			}
		}

		fmt.Printf("Player %d's hand\n", state.currentAgent)
		for _, card := range state.playerHands[state.currentAgent] {
			fmt.Printf("\t%s\n", card.toString())
		}
	}

	// Playing a card
	card := Card(action)
	state.history = append(state.history, card)

	if narrate {
		fmt.Printf("Player %d plays the %s.\n", state.currentAgent, card.toString())
	}

	state.playerHands[state.currentAgent] = RemoveValue(state.playerHands[state.currentAgent], card)
	state.table = append(state.table, card)

	// Handle bower lead
	if state.leadSuit == 0 {
		if card == makeCard(state.trumpSuit.complement(), JACK) {
			state.leadSuit = state.trumpSuit
		} else {
			state.leadSuit = card.getSuit()
		}
	} else if card.effectiveSuit(state.trumpSuit) != state.leadSuit {
		// Track shortsuitedness
		present := false
		for _, suit := range state.shortSuited[state.currentAgent] {
			if state.leadSuit == suit {
				present = true
			}
		}
		if !present {
			state.shortSuited[state.currentAgent] = append(state.shortSuited[state.currentAgent], state.leadSuit)
		}
	}

	// Trick completion
	if len(state.table) == 4 {
		rankings := getRankings(state.trumpSuit, state.leadSuit)

		// Get highest card
		bestIdx := -1
		val := -1
		for idx, card := range state.table {
			rank := getRank(card, rankings)
			if rank > val {
				bestIdx = idx
				val = rank
			}
		}
		winningPlayer := (bestIdx + state.lead) % 4

		if narrate {
			fmt.Printf("Player %d wins the trick ", winningPlayer)
		}

		// Award player and reset table
		state.teamTricks[winningPlayer%2]++
		state.lead = winningPlayer
		state.currentAgent = winningPlayer

		state.table = make([]Card, 0, 4)
		state.leadSuit = 0
	} else {
		state.currentAgent = (state.currentAgent + 1) % 4
	}

	return State(state)
}

// GetCurrentAgent ...
func (state EuchreState) GetCurrentAgent() int {
	return state.currentAgent
}

// GetUtility ...
func (state *EuchreState) GetUtility(playerID int) float64 {
	playerTeam := playerID % 2
	nonCallingTeam := 1 - state.callingTeam

	points := [2]int{0, 0}
	if state.teamTricks[nonCallingTeam] > state.teamTricks[state.callingTeam] {
		points[nonCallingTeam] = 2
	} else if state.teamTricks[state.callingTeam] == 5 {
		points[state.callingTeam] = 2
	} else {
		points[state.callingTeam] = 1
	}

	return float64(points[playerTeam] - points[1-playerTeam])
}

// TakeActionCopy ...
func (state EuchreState) TakeActionCopy(action Action) State {
	clone := state.Clone()
	return clone.TakeAction(action, false)
}

// GetInfoSetKey ...
func (state EuchreState) GetInfoSetKey() InfoSetKey {
	cardStrings := ""

	for _, card := range state.playerHands[state.currentAgent] {
		cardStrings += fmt.Sprintf("%d", card)
	}
	cardStrings += "_"

	for _, card := range state.history {
		cardStrings += fmt.Sprintf("%d", card)
	}

	return InfoSetKey(cardStrings)
}

// IsTerminal ...
func (state *EuchreState) IsTerminal() bool {
	nonCallingTeam := 1 - state.callingTeam
	return (state.teamTricks[state.callingTeam] == 5) || (state.teamTricks[nonCallingTeam] == 3) || (state.teamTricks[0]+state.teamTricks[1] == 5)
}

// Ensures that no cards have been duplicated or lost
func (state EuchreState) CheckCards() {
	for value := 1; value < 7; value++ {
		for suit := 10; suit <= 40; suit += 10 {
			card := Card(suit + value)
			pass := state.checkCard(card)
			if !pass {
				panic(fmt.Sprintf("Lost %s somewhere", card.toString()))
			}
		}
	}

	for handIdx, hand := range state.playerHands {
		for _, card := range hand {
			if inSlice(state.shortSuited[handIdx], card.effectiveSuit(state.trumpSuit)) {
				panic(fmt.Sprintf("Shortsuited tracking incorrect"))
			}
		}
	}
}

// Look for a single card and make sure it appears once
func (state EuchreState) checkCard(c Card) bool {
	found := false

	for _, hand := range state.playerHands {
		for _, card := range hand {
			if card == c {
				if found {
					return false
				}
				found = true
			}
		}
	}
	for _, card := range state.kitty {
		if card == c {
			if found {
				return false
			}
			found = true
		}
	}
	for _, card := range state.history {
		if card == c {
			if found {
				return false
			}
			found = true
		}
	}

	return found
}

func (state *EuchreState) normalizeTrump(suit Suit) {

	// Hands
	for i := range state.playerHands {
		for j := range state.playerHands[i] {
			state.playerHands[i][j].normalizeSuit(suit)
		}
	}

	// History
	for i := range state.history {
		state.history[i].normalizeSuit(suit)
	}

	// Table
	for i := range state.table {
		state.table[i].normalizeSuit(suit)
	}

	// Kitty
	for i := range state.kitty {
		state.kitty[i].normalizeSuit(suit)
	}

	// Shortsuitedness
	for i := range state.shortSuited {
		for j := range state.shortSuited[i] {
			state.shortSuited[i][j] = state.shortSuited[i][j].normalizeSuit(suit)
		}
	}

	// Trump/lead
	state.trumpSuit = state.trumpSuit.normalizeSuit(suit)
	if state.leadSuit != 0 {
		state.leadSuit = state.leadSuit.normalizeSuit(suit)
	}
}

func (state *EuchreState) unnormalizeTrump(suit Suit) {
	// This operation is it's own inverse
	state.normalizeTrump(suit)
}

// RemoveValue ...
func RemoveValue(s []Card, value Card) []Card {
	// find index of value
	index := -1
	for idx, val := range s {
		if val == value {
			index = idx
			break
		}
	}
	if index == -1 {
		panic("Item to remove was not found")
	}

	return append(s[:index], s[index+1:]...)
}

func shuffle(vals []Card) []Card {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]Card, len(vals))
	perm := r.Perm(len(vals))
	for i, randIndex := range perm {
		ret[i] = vals[randIndex]
	}
	return ret
}
