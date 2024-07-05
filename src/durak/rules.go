package durak

import (
	"encoding/json"
	"fmt"
)

type Verb int
type Card int

var UNK_CARD = Card(-1)

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
    Won []bool
    Trump Card
    Plays []Card
    Covers []Card
    Hands [][]Card
    Dir int
}

func NumNotUnk(cards []Card) int {
    res := 0
    for _,c := range cards {
        if c != UNK_CARD {
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

func InitGameState(trump Card, hands [][]Card) *GameState {
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
        Won: make([]bool, len(hands)),
        Trump: trump,
        Plays: make([]Card, 0),
        Covers: make([]Card, 0),
        Hands: hands,
        Dir: 1,
    }
}

func (state *GameState) AttackerActions(player int) []Action {
    res := make([]Action, 0)
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
            res = append(res, Action{player, PlayVerb, card, UNK_CARD})
        }
        return res
    }
    // Don't allow playing more cards than defender can defend
    unmetCards := len(state.Plays) - NumNotUnk(state.Covers)
    if len(state.Hands[state.Defender]) - unmetCards > 0 {
        for _,card := range state.Hands[player] {
            // Allow play unknown card in search
            if card == UNK_CARD {
                res = append(res, Action{player, PlayVerb, card, UNK_CARD})
                continue
            }
            // Regular cards
            for i := 0; i < len(state.Plays); i++ {
                if card.Rank() == state.Plays[i].Rank() || (card != UNK_CARD && card.Rank() == state.Covers[i].Rank()) {
                    res = append(res, Action{player, PlayVerb, card, UNK_CARD})
                    break
                }
            }
        }
    }
    if state.PickingUp || (NumNotUnk(state.Covers) == len(state.Plays) && len(state.Plays) > 0) {
        res = append(res, Action{player, PassVerb, UNK_CARD, UNK_CARD})
    }
    // For AI to not throw trumps away
    if !state.PickingUp && len(state.Plays) > NumNotUnk(state.Covers) {
        res = append(res, Action{player, DeferVerb, UNK_CARD, UNK_CARD})    
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
        if state.Covers[i] != UNK_CARD {
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
    if revRank != -1 && NumNotUnk(state.Covers) == 0 && len(state.Plays)+1 <= len(state.Hands[state.Attacker]) {
        for _,card := range state.Hands[player] {
            if card.Rank() == revRank {
                res = append(res, Action{player, ReverseVerb, card, UNK_CARD})
            }
        }
    }
    for i := 0; i < len(state.Plays); i++ {
        for _,card := range state.Hands[player] {
            // For AI search allow cover with unknown card
            if card == UNK_CARD || (state.Covers[i] == UNK_CARD && card.Beats(state.Plays[i], state.Trump)) {
                res = append(res, Action{player, CoverVerb, card, state.Plays[i]})
            }
        }
    }
    if len(state.Plays) > 0 && NumNotUnk(state.Covers) < len(state.Plays) {
        res = append(res, Action{player, PickupVerb, UNK_CARD, UNK_CARD})
    }
    return res
}

func (state *GameState) PlayerActions(player int) []Action {
    if player == state.Defender {
        return state.DefenderActions(player)
    } else {
        return state.AttackerActions(player)
    }
}
