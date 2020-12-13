package cfr

import "fmt"

// Value ...
type Value uint8

func (v Value) toString() string {
	switch v {
	case ACE:
		return "Ace"
	case KING:
		return "King"
	case QUEEN:
		return "Queen"
	case JACK:
		return "Jack"
	case TEN:
		return "Ten"
	case NINE:
		return "Nine"
	}
	return "NULL_VALUE"
}

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

func (s Suit) toString() string {
	switch s {
	case SPADES:
		return "Spades"
	case CLUBS:
		return "Clubs"
	case HEARTS:
		return "Hearts"
	case DIAMONDS:
		return "Diamonds"
	}
	return "NULL_SUIT"
}

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
		return s.swapCycle().redBlackCycle()
	case HEARTS:
		return s.redBlackCycle()
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
		for v := 1; v <= 6; v++ {
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

func (c Card) toString() string {
	return fmt.Sprintf("%s of %s", c.getValue().toString(), c.getSuit().toString())
}

func getRank(c Card, ranks []Card) int {
	for idx, card := range ranks {
		if card == c {
			return idx
		}
	}
	panic("Card not in rankings")
}
