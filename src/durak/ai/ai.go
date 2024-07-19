package ai

import (
	//"log"

	"github.com/aorliche/cards-ai/search"
	"github.com/aorliche/cards-ai/durak"
)

type EvalParams struct {
	WinBonus float64
	TrumpBonus float64
	NonTrumpPenalty float64
	CardsInDeckCutoff int
	SmallDeckHandPenalty float64
	BigDeckHandPenalty float64
	UnknownCardPenalty float64
}

type GameState struct {
	durak.GameState
	Params *EvalParams
}

var DefaultEvalParams = EvalParams{
	500.0, 10.0, 6.0, 3, 10.0, 2.0, 4.0,
}

func (state *GameState) Clone2() *GameState {
	return &GameState{*state.Clone(), state.Params}
}

func (state *GameState) NumPlayers() int {
	return len(state.Hands)
}

func (state *GameState) Eval(player int) float64 {
	params := state.Params
	if params == nil {
		params = &DefaultEvalParams
	}
	// Check my win
	if state.Won[player] {
		return params.WinBonus
	}
	// My hand
	hval := 0.0
	for _,c := range state.Hands[player] {
		if c == durak.UNK_CARD {
			continue
		} else if c.Suit() == state.Trump.Suit() {
			hval += params.TrumpBonus + float64(c.Rank())
		} else {
			hval += float64(c.Rank()) - params.NonTrumpPenalty
		}
	}
	if state.CardsInDeck <= params.CardsInDeckCutoff {
		hval -= params.SmallDeckHandPenalty*float64(len(state.Hands[player]))
	} else {
		hval -= params.BigDeckHandPenalty*float64(len(state.Hands[player]))
	}
	// Other player's hands
	ovals := make([]float64, 0)
	for i,h := range state.Hands {
		if i == player {
			continue
		}
		v := 0.0
		// Check other player's win
		if state.Won[i] {
			ovals = append(ovals, params.WinBonus)
			continue
		}
		for _,c := range h {
			if c == durak.UNK_CARD {
				v += -params.UnknownCardPenalty
			} else if c.Suit() == state.Trump.Suit() {
				v += params.TrumpBonus + float64(c.Rank())
			} else {
				v += float64(c.Rank()) - params.NonTrumpPenalty
			}
		}
		if state.CardsInDeck <= params.CardsInDeckCutoff {
			v -= params.SmallDeckHandPenalty*float64(len(state.Hands[player]))
		} else {
			v -= params.BigDeckHandPenalty*float64(len(state.Hands[player]))
		}
		ovals = append(ovals, v)
	}
	// We only want to not be the worst
	worst := ovals[0]
	for _,v := range ovals {
		if v < worst {
			worst = v
		}
	}
	return hval - worst
}

func (state *GameState) Children(player int) ([]search.Action, []search.GameState) {
	// Check that we don't have unknown cards on the board
	// If we do, we can't search further
	for i,c := range state.Plays {
		if c == durak.UNK_CARD || state.Covers[i] == durak.UNK_CARD {
			return make([]search.Action, 0), make([]search.GameState, 0)
		}
	}
	// Check that the game isn't over
	// If it is, we can't search further
	if state.IsOver() {
		return make([]search.Action, 0), make([]search.GameState, 0)
	}
	acts := state.PlayerActions(player)
	searchActs := make([]search.Action, len(acts))
	children := make([]search.GameState, len(acts))
	for i,a := range acts {
		st := state.Clone2()
		st.TakeAction(a)
		searchActs[i] = a
		children[i] = st
	}
	return searchActs, children
}

func (state *GameState) FindBestAction(player int, depth int, timeBudget int64) (durak.Action, bool) {
	st := state.Clone()
	st.Mask(player)
	state = &GameState{*st, state.Params}
	iface, _ := search.SearchItDeep(state, player, depth, timeBudget)
	act, ok := iface.(durak.Action)
	return act, ok
}

