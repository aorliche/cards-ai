package spades

import (
	"fmt"
	"math/rand/v2"
)

type Verb int
type Card int
type Trick [4]Card

var UNK_CARD = Card(-1)
var NO_CARD = Card(-2)
var SUIT_SPADES = 1

const (
	BidVerb Verb = iota
    PlayVerb 
)

type Action struct {
	Verb Verb
	Player int
	Card Card
	Bid int
}

type GameState struct {
	Hands [4][]Card
	Bids [4]int
	Tricks [4]int
	Attacker int
	Trick Trick
	// Known absent cards (according to each player)
	Absent [4][4][52]bool
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
	if card == NO_CARD {
		return false
	}
	if other == NO_CARD {
		return true
	}
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

func InitTrick() Trick {
	return Trick([4]Card{NO_CARD, NO_CARD, NO_CARD, NO_CARD})
}

func (stack *Trick) Winner() int {
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
	st := &GameState{
		Hands: hands,
		Bids: bids,
		Tricks: tricks,
		Attacker: 0,
		Trick: InitTrick(),
	}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i == j {
				continue
			}
			for _,c := range st.Hands[i] {
				st.Absent[i][j][int(c)] = true
			}
		}
	}
	return st
}

func (state *GameState) Clone() *GameState {
	hands := [4][]Card{}
	for i := 0; i < 4; i++ {
		hands[i] = append(make([]Card, 0), state.Hands[i]...)
	}
	return &GameState{
		Hands: hands,
		Bids: state.Bids,
		Tricks: state.Tricks,
		Attacker: state.Attacker,
		Trick: state.Trick,
		Absent: state.Absent,
	}
}

func (state *GameState) PlayerActions(player int) []Action {
	acts := make([]Action, 0)
	// Bidding in order
	for i := 0; i < 4; i++ {
		j := (state.Attacker + i) % 4
		if state.Bids[j] == -1 {
			if player == j {
				for k := 0; k <= 13; k++ {
					acts = append(acts, Action{Verb: BidVerb, Player: player, Card: NO_CARD, Bid: k})
				}
			}
			return acts
		}
	}
	// Play cards
	for i := 0; i < 4; i++ {
		j := (state.Attacker + i) % 4
		if state.Trick[i] == NO_CARD {
			// Not player's turn
			if j != player {
				return acts
			}
			// No cards played yet
			if i == 0 {
				for _,c := range state.Hands[player] {
					acts = append(acts, Action{Verb: PlayVerb, Player: player, Card: c})
				}
			} else {
				firstSuit := state.Trick[0].Suit()
				// Check whether we have suit
				haveSuit := false
				for _,c := range state.Hands[player] {
					if c.Suit() == firstSuit {
						haveSuit = true
						break
					}
				}
				// Add playable cards
				for _,c := range state.Hands[player] {
					if c.Suit() == firstSuit || !haveSuit {
						acts = append(acts, Action{Verb: PlayVerb, Player: player, Card: c})
					}
				}
			}
		}
	}
	return acts
}

func RemoveCard(cards *[]Card, c Card) bool {
    for i,card := range *cards {
        if card == c {
            (*cards)[i] = (*cards)[len(*cards)-1]
            *cards = (*cards)[:len(*cards)-1]
            return true
        }
    }
    return false
}

func (state *GameState) TakeAction(act Action) {
	if act.Verb == BidVerb {
		state.Bids[act.Player] = act.Bid
		return
	} 
	RemoveCard(&state.Hands[act.Player], act.Card)
	// Check suit gone
	if state.Trick[0] != NO_CARD {
		if act.Card.Suit() != state.Trick[0].Suit() {
			for i := 0; i < 4; i++ {
				for j := 0; j < 13; j++ {
					k := j + 13*act.Card.Suit()
					state.Absent[i][act.Player][k] = true
				}
			}
		}
	}
	// Absent the card from everyone
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state.Absent[i][j][int(act.Card)] = true
		}
	}
	for i := 0; i < 4; i++ {
		if state.Trick[i] == NO_CARD {
			state.Trick[i] = act.Card
			break
		}
	}
	// Wrap up trick
	if state.Trick[3] != NO_CARD {
		win := state.Trick.Winner()
		win = (win + state.Attacker) % 4
		state.Tricks[win]++
		state.Attacker = win
		state.Trick = InitTrick()
	}
}

func (state *GameState) IsOver() bool {
	for i := 0; i < 4; i++ {
		if len(state.Hands[i]) != 0 {
			return false
		}
	}
	return true
}

func (state *GameState) CurrentActions() []Action {
	acts := make([]Action, 0)
	for i := 0; i < 4; i++ {
		acts = append(acts, state.PlayerActions(i)...)
	}
	return acts
}
