package cfr

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// Value ...
type Value uint8

// Card Enumerations
const (
	ACE   = Value(6)
	KING  = Value(5)
	QUEEN = Value(4)
	JACK  = Value(3)
	TEN   = Value(2)
	NINE  = Value(1)
)
const (
	DIAMONDS = Suit(10)
	HEARTS   = Suit(20)
	SPADES   = Suit(30)
	CLUBS    = Suit(40)
)

// Suit ...
type Suit uint8

func (s Suit) complement() Suit {
	switch s {
	case SPADES:
		return CLUBS
	case CLUBS:
		return SPADES
	case HEARTS:
		return DIAMONDS
	case DIAMONDS:
		return HEARTS
	default:
		panic("Suit is not valid")
	}
}

func (s Suit) swapCycle() Suit {
	switch s {
	case DIAMONDS:
		return s + 10
	case HEARTS:
		return s - 10
	case SPADES:
		return s + 10
	case CLUBS:
		return s - 10
	}
	panic("Attempted to cycle null suit")
}

func (s Suit) redBlackCycle() Suit {
	switch s {
	case DIAMONDS:
		return s + 30
	case HEARTS:
		return s + 10
	case SPADES:
		return s - 10
	case CLUBS:
		return s - 30
	}
	panic("Attempted to cycle null suit")
}

func (s Suit) normalizeSuit(suit Suit) Suit {
	switch suit {
	case DIAMONDS:
		return s.redBlackCycle()
	case HEARTS:
		return s.redBlackCycle().swapCycle()
	case CLUBS:
		return s.swapCycle()
	case SPADES:
		return s
	}
	panic("Attempting to set a null suit")
}

// Card ...
type Card uint8

func makeCard(suit Suit, value Value) Card {
	return Card(int(suit) + int(value))
}

func (c *Card) effectiveSuit(trumpSuit Suit) Suit {
	if (c.getValue() == JACK) && (trumpSuit == c.getSuit().complement()) {
		return trumpSuit
	} else {
		return c.getSuit()
	}
}

func (c *Card) getSuit() Suit {
	return Suit((int(*c) / 10) * 10)
}

func (c *Card) getValue() Value {
	return Value(int(*c) % 10)
}

func (c *Card) normalizeSuit(suit Suit) {
	*c = Card(int(c.getSuit().normalizeSuit(suit)) + int(c.getValue()))
}

func getRankings(trumpSuit Suit, leadSuit Suit) []Card {
	rightBower := makeCard(trumpSuit, JACK)
	leftBower := makeCard(trumpSuit.complement(), JACK)

	ranks := make([]Card, 0, 24)
	// Add off suits
	for s := 10; s <= 40; s += 10 {
		if (Suit(s) != trumpSuit) && (Suit(s) != leadSuit) {
			for v := 6; v > 0; v-- {
				card := makeCard(Suit(s), Value(v))
				if card != leftBower {
					ranks = append(ranks, card)
				}
			}
		}
	}

	// Add lead suit
	if leadSuit != trumpSuit {
		for v := 6; v > 0; v-- {
			card := makeCard(leadSuit, Value(v))
			if card != leftBower {
				ranks = append(ranks, card)
			}
		}
	}

	// Add trump
	for v := 6; v > 0; v-- {
		card := makeCard(trumpSuit, Value(v))
		if card != rightBower {
			ranks = append(ranks, card)
		}
	}

	// Bowers on top
	ranks = append(ranks, leftBower)
	ranks = append(ranks, rightBower)

	return ranks
}

func getRank(c Card, ranks []Card) int {
	for idx, card := range ranks {
		if card == c {
			return idx
		}
	}
	panic("Card not in rankings")
}

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
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

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

func (state EuchreState) SampleInfoSet() EuchreState {
	key := state.GetInfoSetKey()
	newState := state.Clone()

	// Collect unknown cards
	intermediateDeck := make([]Card, 0)
	for i, hand := range newState.playerHands {
		if i != newState.currentAgent {
			intermediateDeck = append(intermediateDeck, hand...)
		}
	}
	intermediateDeck = append(intermediateDeck, newState.kitty[1:]...)

	rand.Shuffle(len(intermediateDeck), func(i, j int) { intermediateDeck[i], intermediateDeck[j] = intermediateDeck[j], intermediateDeck[i] })
	// Redeal respecting shortsuitedness
	// Need to deal most-restrictive hands first
	for i := 3; i >= 0; i-- {
		for handIdx, suits := range state.shortSuited {
			if len(suits) == i && handIdx != state.currentAgent {
				// Advance through the shuffled cards until a valid card is found
				currentPos := 0
				for j := 0; currentPos < len(newState.playerHands[handIdx]); j++ {
					deckCard := intermediateDeck[j]
					if deckCard != 0 && !inSlice(suits, deckCard.effectiveSuit(newState.trumpSuit)) {
						newState.playerHands[handIdx][currentPos] = deckCard
						intermediateDeck[j] = 0
						currentPos++
					}
				}
				// Keep the hands sorted
				sort.Slice(newState.playerHands[handIdx], func(j, k int) bool {
					return newState.playerHands[handIdx][j] < newState.playerHands[handIdx][k]
				})

				// Need to reshuffle to unbias the next hands deal
				rand.Shuffle(len(intermediateDeck), func(i, j int) { intermediateDeck[i], intermediateDeck[j] = intermediateDeck[j], intermediateDeck[i] })
			}
		}
	}
	newKey := newState.GetInfoSetKey()
	if key != newKey {
		panic("Incorrect sampling, key should remain the same")
	}

	return newState
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
func (state *EuchreState) Clone() EuchreState {
	newState := *state

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
func (state *EuchreState) TakeAction(action Action) State {
	// Playing a card
	card := Card(action)
	state.history = append(state.history, card)

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
		state.shortSuited[state.currentAgent] = append(state.shortSuited[state.currentAgent], state.leadSuit)
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
	return clone.TakeAction(action)
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
