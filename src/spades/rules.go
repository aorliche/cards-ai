package spades

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

type Verb int
type Card int
type Stack [4]Card

var UNK_CARD = Card(-1)
var NO_CARD = Card(-2)
var SUIT_SPADES = 1

const (
	BidVerb Verb = iota
    PlayVerb 
)

type Action struct {
	Verb Verb
	Card Card
	Bid int
}

type GameState struct {
	Hands [4][]Card
	Bids [4]int
	Tricks [4]int
	Attacker int
	Stack Stack
}

var suits = []string{"clubs", "spades", "hearts", "diamonds"}
var ranks = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "jack", "queen", "king", "ace"}
var verbs = []string{"Bid", "Play"}

func CardFromRankSuit(rank int, suit int) Card {
    return Card(suit*13 + rank)
}

func (card Card) Rank() int {
    return int(card)%13
}

func (card Card) Suit() int {
    return int(card)/13
}

func (card Card) RankStr() string {
    return ranks[card.Rank()]
}

func (card Card) SuitStr() string {
    return suits[card.Suit()]
}

func (card Card) ToStr() string {
    return fmt.Sprintf("%s of %s", card.RankStr(), card.SuitStr()) 
}

func (card Card) Beats(other Card, firstSuit int) bool {
	if card.Suit() == SUIT_SPADES && other.Suit() != SUIT_SPADES {
		return true
	}
	if card.Suit() != SUIT_SPADES && other.Suit() == SUIT_SPADES {
		return false
	}
	if card.Suit() == firstSuit && other.Suit() != firstSuit {
		return true
	}
	if card.Suit() == firstSuit && other.Suit() == firstSuit {
		return card.Rank() > other.Rank()
	}
	return false
}

func InitStack() Stack {
	return Stack([4]Card{NO_CARD, NO_CARD, NO_CARD, NO_CARD})
}

func (stack Stack) Winner() int {
	suit := stack[0].Suit()
	for i,card := range stack {
		beats := true
		for _,c := range stack {
			if c == card {
				continue
			}
			if c == NO_CARD {
				return -1
			}
			if !card.Beats(c, suit) {
				beats = false
				break
			}
		}
		if beats {
			return i
		}
	}
	return -1
}

func (stack Stack) Play(card Card) error {
	if stack[3] != NO_CARD {
		return errors.New("Stack already has 4 cards")
	}
	for i := 0; i < 4; i++ {
		if stack[i] == NO_CARD {
			stack[i] = card
			return nil
		}
	}
	// Unnecessary
	return nil
}

func InitGameState() *GameState {
	hands := [4][]Card{}
	for i := 0; i < 4; i++ {
		hands[i] = make([]Card, 13)
	}
	deck := rand.Perm(52)
	for i := 0; i < 52; i++ {
		player := i / 13
		hands[player][i%13] = Card(deck[i])
	}
	bids := [4]int{-1,-1,-1,-1}
	tricks := [4]int{}
	return &GameState{
		Hands: hands,
		Bids: bids,
		Tricks: tricks,
		Attacker: 0,
		Stack: InitStack(),
	}
}

func (state *GameState) PlayerActions(player int) []Action {
	acts := make([]Action, 0)
	// This player has to bid
	if state.bids[player] == -1 {
		for i := 0; i <= 13; i++ {
			acts.append(Action{Verb: BidVerb, Card: NO_CARD, Bid: i})
		}
		return acts
	}
	// Other players have to bid
	for i := 0; i < 4; i++ {
		if state.bids[i] == -1 {
			return acts
		}
	}
	// Play cards
	for i := 0; i < 4; i++ {
		j := (state.Attacker + i) % 4
		if state.Stack[j] == NO_CARD {
			// Not player's turn
			if j != player {
				return acts
			}
			firstSuit := state.Stack[0].Suit()
			// Check whether we have suit
			hasSuit := false
			for _,c := range state.Hands[player] {
				if c.Suit() == firstSuit {
					hasSuit = true
					break
				}
			}
			// Add playable cards
			for _,c := range state.Hands[player] {
				if c.Suit() == firstSuit || !hasSuit {
					acts = append(Action{Verb: PlayVerb, Card: c})
				}
			}
		}
	}
	return acts
}
