package durak

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

type Verb int
type Card int

var UNK_CARD = Card(-1)
var NO_CARD = Card(-2)

const (
    PlayVerb Verb = iota
    CoverVerb 
    ReverseVerb
    PassVerb
    PickUpVerb
	DeferVerb
)

var suits = []string{"clubs", "spades", "hearts", "diamonds"}
var ranks = []string{"6", "7", "8", "9", "10", "jack", "queen", "king", "ace"}
var verbs = []string{"Play", "Cover", "Reverse", "Pass", "Pick Up", "Defer"}

func CardFromRankSuit(rank int, suit int) Card {
    return Card(suit*9 + rank)
}

func (card Card) Rank() int {
    return int(card)%9
}

func (card Card) Suit() int {
    return int(card)/9
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

func (card Card) Beats(other Card, trump Card) bool {
	if card == UNK_CARD || other == UNK_CARD {
		return false
	}
    if card.Suit() == trump.Suit() && card.Suit() != other.Suit() {
        return true
    }
    return card.Rank() > other.Rank() && card.Suit() == other.Suit()
}

func GenerateDeck() []Card {
    res := make([]Card, 0)
    for suit := 0; suit < 4; suit++ {
        for rank := 0; rank < 9; rank++ {
            res = append(res, CardFromRankSuit(rank, suit))
        }
    }
    rand.Shuffle(len(res), func(i, j int) {
        res[i], res[j] = res[j], res[i]
    })
    return res
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

type Action struct {
    Player int
    Verb Verb
    Card Card
    Covering Card
}

func (a Action) IsNull() bool {
    return a.Card == 0 && a.Covering == 0
}

func (a Action) ToStr() string {
    mp := map[string]any {
        "Player": a.Player,
        "Verb": verbs[a.Verb],
        "Card": a.Card,
        "Covering": a.Covering,
    }
    jsn, _ := json.Marshal(mp)
    return string(jsn)
}

type GameState struct {
    Attacker int
    Defender int
    PickingUp bool
    Passed []bool
	Deferring []bool
    Won []bool
    Trump Card
    Plays []Card
    Covers []Card
    Hands [][]Card
	Known [][]Card
    Dir int
	Deck []Card
	CardsInDeck int
}

func (state *GameState) NumCovered() int {
    res := 0
    for _,c := range state.Covers {
        if c != NO_CARD {
            res++
        }
    }
    return res
}

// For determining who goes first
func LowestTrumpRank(trump Card, cards []Card) int {
	rank := -1
	for _,c := range cards {
		if c.Suit() == trump.Suit() {
			if rank == -1 || c.Rank() < rank {
				rank = c.Rank()
			}
		}
	}
	return rank
}

func InitGameState(nPlayers int) *GameState {
	if nPlayers < 2 || nPlayers > 6 {
		return nil
	}
	deck := GenerateDeck()
	// Deal deck to players
	hands := make([][]Card, nPlayers)
	known := make([][]Card, nPlayers)
	ci := 0
	trump := deck[len(deck)-1]
	for i := 0; i < nPlayers; i++ {
		hands[i] = make([]Card, 0)
		known[i] = make([]Card, 0)
		for j := 0; j < 6; j++ {
			hands[i] = append(hands[i], deck[ci])
			ci++
		}
	}
	// Determine who goes first
	lowRank := -1
	lowRankIdx := 0
	for i,h := range hands {
		r := LowestTrumpRank(trump, h)
		if r != -1 {
			if lowRank == -1 || r < lowRank {
				lowRank = r
				lowRankIdx = i
			}
		}
	}
    return &GameState{
        Attacker: lowRankIdx,
        Defender: (lowRankIdx+1)%len(hands), 
        PickingUp: false, 
        Passed: make([]bool, len(hands)),
		Deferring: make([]bool, len(hands)),
        Won: make([]bool, len(hands)),
        Trump: trump,
        Plays: make([]Card, 0),
        Covers: make([]Card, 0),
        Hands: hands,
		Known: known,
        Dir: 1,
		Deck: deck,
		CardsInDeck: len(deck)-ci,
    }
}

func (state *GameState) AttackerActions(player int) []Action {
    res := make([]Action, 0)
	// Deferring 
	if state.Deferring[player] {
		return res
	}
    // Player has already passed
    if state.Passed[player] {
        return res
    }
    // Player has already won
    if state.Won[player] {
        return res
    }
    if len(state.Plays) == 0 {
        // Only initial attacker may play first
        if state.Attacker != player {
            return res
        }
        for _,card := range state.Hands[player] {
            res = append(res, Action{player, PlayVerb, card, NO_CARD})
        }
        return res
    }
    // Don't allow playing more cards than defender can defend
    unmetCards := len(state.Plays) - state.NumCovered()
    if len(state.Hands[state.Defender]) - unmetCards > 0 {
        for _,card := range state.Hands[player] {
            // Allow play unknown card in search
            if card == UNK_CARD {
                res = append(res, Action{player, PlayVerb, card, NO_CARD})
                continue
            }
            // Regular cards
            for i := 0; i < len(state.Plays); i++ {
                if card.Rank() == state.Plays[i].Rank() || (state.Covers[i] != UNK_CARD && state.Covers[i] != NO_CARD && card.Rank() == state.Covers[i].Rank()) {
                    res = append(res, Action{player, PlayVerb, card, NO_CARD})
                    break
                }
            }
        }
    }
    if state.PickingUp || (state.NumCovered() == len(state.Plays) && len(state.Plays) > 0) {
        res = append(res, Action{player, PassVerb, NO_CARD, NO_CARD})
    }
    // For AI to not throw trumps away
    if !state.PickingUp && len(state.Plays) > 0 && len(state.Plays) > state.NumCovered() {
        res = append(res, Action{player, DeferVerb, NO_CARD, NO_CARD})    
    }
    return res
}

func (state *GameState) ReverseRank() int {
    if len(state.Plays) == 0 {
        return -1
    }
    rank := state.Plays[0].Rank()
    for i := 0; i < len(state.Plays); i++ {
        if rank != state.Plays[i].Rank() {
            return -1
        }
        if state.Covers[i] != NO_CARD {
            return -1
        }
    }
    return rank
}

func (state *GameState) DefenderActions(player int) []Action {
    res := make([]Action, 0)
    if state.PickingUp {
        return res
    }
    revRank := state.ReverseRank()
    // Only allow reverse when defender can potentially meet it
    if revRank != -1 && state.NumCovered() == 0 && len(state.Plays)+1 <= len(state.Hands[state.Attacker]) {
        for _,card := range state.Hands[player] {
            if card.Rank() == revRank {
                res = append(res, Action{player, ReverseVerb, card, NO_CARD})
            }
        }
    }
    for i := 0; i < len(state.Plays); i++ {
		if state.Covers[i] != NO_CARD {
			continue
		}
        for _,card := range state.Hands[player] {
            // For AI search allow cover with unknown card
            if card == UNK_CARD || card.Beats(state.Plays[i], state.Trump) {
                res = append(res, Action{player, CoverVerb, card, state.Plays[i]})
            }
        }
    }
    if len(state.Plays) > 0 && state.NumCovered() < len(state.Plays) {
        res = append(res, Action{player, PickUpVerb, NO_CARD, NO_CARD})
    }
    return res
}

func (state *GameState) PlayerActions(player int) []Action {
	var acts []Action
    if player == state.Defender {
        acts = state.DefenderActions(player)
    } else {
        acts = state.AttackerActions(player)
    }
	return acts
}

func (state *GameState) NextRole(player int) int {
    for i := 0; i < len(state.Hands); i++ {
        p := player+((i+1)*state.Dir)
        if p < 0 {
            p += len(state.Hands)
        }
        if p >= len(state.Hands) {
            p -= len(state.Hands)
        }
        if !state.Won[p] {
            return p
        }
    }
    panic("NextRole failed")
}

func (state *GameState) AllPassed() bool {
    for i := 0; i < len(state.Hands); i++ {
        if !state.Passed[i] && state.Defender != i && !state.Won[i] {
            return false
        }
    }
    return true
}

func (state *GameState) IsOver() bool {
    if state.CardsInDeck > 0 {
        return false
    }
    n := 0
    for i := 0; i < len(state.Won); i++ {
        if state.Won[i] {
			n++
		}
    }
    return n == len(state.Hands)-1
}

// This probably deals in correct order now
func (state *GameState) Deal(defender int) {
	for i := 0; i < len(state.Hands); i++ {
		p := defender+((i+1)*state.Dir)
        if p < 0 {
            p += len(state.Hands)
        }
        if p >= len(state.Hands) {
            p -= len(state.Hands)
        }
		for len(state.Hands[p]) < 6 {
			if state.CardsInDeck == 0 {
				// Check winner
				if len(state.Hands[p]) == 0 {
					state.Won[p] = true
				}
				break
			} else {
				j := len(state.Deck)-state.CardsInDeck
				state.Hands[p] = append(state.Hands[p], state.Deck[j])
				state.CardsInDeck--
			}
		}
	}
}

func (state *GameState) TakeAction(action Action) {
    switch action.Verb {
        case PlayVerb: {
            state.Plays = append(state.Plays, action.Card)
            state.Covers = append(state.Covers, NO_CARD)
            RemoveCard(&state.Hands[action.Player], action.Card)
			RemoveCard(&state.Known[action.Player], action.Card)
			// Reset deferring
            state.Deferring = make([]bool, len(state.Hands))
			// Check win passed is important for not hanging game with no actions
			if state.CardsInDeck == 0 && len(state.Hands[action.Player]) == 0 {
				state.Passed = make([]bool, len(state.Hands))
				state.Won[action.Player] = true
			}
        }
        case CoverVerb: {
            for i := 0; i < len(state.Plays); i++ {
                if action.Covering == state.Plays[i] {
                    state.Covers[i] = action.Card
                }
            }
            RemoveCard(&state.Hands[action.Player], action.Card)
			RemoveCard(&state.Known[action.Player], action.Card)
			// Reset deferring
            state.Deferring = make([]bool, len(state.Hands))
			// Check win passed is important for not hanging game with no actions
			if state.CardsInDeck == 0 && len(state.Hands[action.Player]) == 0 {
				state.Passed = make([]bool, len(state.Hands))
				state.Won[action.Player] = true
			}
        }
        case ReverseVerb: {
            state.Plays = append(state.Plays, action.Card)
            state.Covers = append(state.Covers, NO_CARD)
            RemoveCard(&state.Hands[action.Player], action.Card)
			RemoveCard(&state.Known[action.Player], action.Card)
            state.Attacker, state.Defender = state.Defender, state.Attacker
			// Reset deferring
            state.Deferring = make([]bool, len(state.Hands))
			// Change direction
            state.Dir *= -1
			// Check win passed is important for not hanging game with no actions
			if state.CardsInDeck == 0 && len(state.Hands[action.Player]) == 0 {
				state.Passed = make([]bool, len(state.Hands))
				state.Won[action.Player] = true
			}
        }
        case PickUpVerb: {
            state.PickingUp = true
			// Reset deferring
            state.Deferring = make([]bool, len(state.Hands))
        }
        case PassVerb: {
            state.Passed[action.Player] = true
            if !state.AllPassed() {
                break
            }
			defender := state.Defender
            if state.PickingUp {
				// Pick up all cards
                for i := 0; i < len(state.Plays); i++ {
                    state.Hands[state.Defender] = append(state.Hands[state.Defender], state.Plays[i])
					state.Known[state.Defender] = append(state.Known[state.Defender], state.Plays[i])
                    if state.Covers[i] != NO_CARD {
                        state.Hands[state.Defender] = append(state.Hands[state.Defender], state.Covers[i])
						state.Known[state.Defender] = append(state.Known[state.Defender], state.Covers[i])
                    }
                }
                state.Attacker = state.NextRole(state.Defender) 
                state.Defender = state.NextRole(state.Attacker)
            } else {
                state.Attacker = state.NextRole(state.Attacker) 
                state.Defender = state.NextRole(state.Attacker)
            }
            state.Plays = make([]Card, 0)
            state.Covers = make([]Card, 0)
            state.PickingUp = false
            state.Deferring = make([]bool, len(state.Hands))
            state.Passed = make([]bool, len(state.Hands))
			state.Deal(defender)
        }
        case DeferVerb: {
            state.Deferring[action.Player] = true
        }
    }
	// Check for a winner being marked as attacker or defender
	// This isn't necessary?
	if !state.IsOver() {
		for state.Won[state.Attacker] || state.Attacker == state.Defender {
			state.Attacker = state.NextRole(state.Attacker)
			state.Defender = state.NextRole(state.Attacker)
		}
		for state.Won[state.Defender] {
			state.Defender = state.NextRole(state.Defender)
		}
	}
}

func (state *GameState) Clone() *GameState {
    hands := make([][]Card, len(state.Hands))
	known := make([][]Card, len(state.Known))
    for i := 0; i < len(hands); i++ {
        hands[i] = append(make([]Card, 0), state.Hands[i]...)
        known[i] = append(make([]Card, 0), state.Known[i]...)
    }
    return &GameState {
        Defender: state.Defender,
        Attacker: state.Attacker,
        PickingUp: state.PickingUp,
        Deferring: append(make([]bool, 0), state.Deferring...),
        Passed: append(make([]bool, 0), state.Passed...),
        Won: append(make([]bool, 0), state.Won...),
        Trump: state.Trump,
        Plays: append(make([]Card, 0), state.Plays...),
        Covers: append(make([]Card, 0), state.Covers...),
        Hands: hands,
		Known: known,
        Dir: state.Dir,
		Deck: state.Deck,
		CardsInDeck: state.CardsInDeck,
    }
}

func (state *GameState) Mask(me int) {
	for i := 0; i < len(state.Hands); i++ {
		if i == me {
			continue
		}
		n := len(state.Hands[i])
		state.Hands[i] = append(make([]Card, 0), state.Known[i]...)
		for j := len(state.Hands[i]); j < n; j++ {
			state.Hands[i] = append(state.Hands[i], -1)
		}
	}
}

func (state *GameState) AllActions() []Action {
	acts := make([]Action, 0)
	for i := 0; i < len(state.Hands); i++ {
		acts = append(acts, state.PlayerActions(i)...)
	}
	return acts
}
