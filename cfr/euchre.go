package cfr

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

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
	playedCards []Card
	table       []Card
	//kitty       []Card
	teamTricks [2]int

	leadSuit     Suit
	trumpSuit    Suit
	lead         int
	callingTeam  int
	currentAgent int
}

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
		playedCards:  make([]Card, 0),
		trumpSuit:    SPADES,
		table:        make([]Card, 0, 4),
	}

	// Deal cards
	for i := 0; i < 4; i++ {
		state.playerHands[i] = make([]Card, 5)
		for c := 0; c < 5; c++ {
			state.playerHands[i][c] = deck[c+i*5]
		}
		// Sort hands
		sort.Slice(state.playerHands[i], func(j, k int) bool {
			return state.playerHands[i][j] < state.playerHands[i][k]
		})
	}

	return state
}

func (state *EuchreState) Clone() EuchreState {
	newState := *state

	for handIdx, hand := range state.playerHands {
		newState.playerHands[handIdx] = make([]Card, len(hand))

		for cIdx, card := range hand {
			newState.playerHands[handIdx][cIdx] = card
		}
	}

	return newState
}

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

func (state *EuchreState) TakeAction(action Action) State {
	// Playing a card
	card := Card(action)

	state.playerHands[state.currentAgent] = RemoveValue(state.playerHands[state.currentAgent], card)
	state.table = append(state.table, card)
	state.currentAgent = (state.currentAgent + 1) % 4

	// Handle bower lead
	if state.leadSuit == 0 {
		if card == makeCard(state.trumpSuit.complement(), JACK) {
			state.leadSuit = state.trumpSuit
		} else {
			state.leadSuit = card.getSuit()
		}
	}

	// Trick completion
	if len(state.table) == 4 {
		rankings := getRankings(state.trumpSuit, state.leadSuit)

		// Get highest card
		best_idx := -1
		val := -1
		for idx, card := range state.table {
			rank := getRank(card, rankings)
			if rank > val {
				best_idx = idx
				val = rank
			}
		}
		winningPlayer := (best_idx + state.lead) % 4

		// Award player and reset table
		state.teamTricks[winningPlayer%2]++
		state.lead = winningPlayer
		state.currentAgent = winningPlayer

		state.playedCards = append(state.playedCards, state.table...)
		state.table = make([]Card, 0, 4)
		state.leadSuit = 0
	}

	return State(state)
}

func (state EuchreState) GetCurrentAgent() int {
	return state.currentAgent
}

func (state *EuchreState) GetUtility(playerID int) float32 {
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

	return float32(points[playerTeam] - points[1-playerTeam])
}

func (state EuchreState) TakeActionCopy(action Action) State {
	clone := state.Clone()
	return clone.TakeAction(action)
}

func (state EuchreState) GetInfoSetKey() InfoSetKey {
	cardStrings := ""

	for _, card := range state.playerHands[state.currentAgent] {
		cardStrings += fmt.Sprintf("%d", card)
	}
	return InfoSetKey(fmt.Sprintf("%d", state.leadSuit) + "_" + cardStrings)
}

func (state *EuchreState) IsTerminal() bool {
	nonCallingTeam := 1 - state.callingTeam
	return (state.teamTricks[state.callingTeam] == 5) || (state.teamTricks[nonCallingTeam] == 3)
}

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
