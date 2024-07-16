package ai

import (
	//"log"

	"github.com/aorliche/cards-ai/search"
	"github.com/aorliche/cards-ai/durak"
)

type GameState struct {
	durak.GameState
}

func (state *GameState) NumPlayers() int {
	return len(state.Hands)
}

func (state *GameState) Eval(player int) float64 {
	// Check my win
	if state.Won[player] {
		return 200.0
	}
	// My hand
	hval := 0.0
	for _,c := range state.Hands[player] {
		if c == durak.UNK_CARD {
			continue
		} else if c.Suit() == state.Trump.Suit() {
			hval += 10.0*float64(c.Rank() + 1)
		} else {
			hval += float64(c.Rank() - 6)
		}
	}
	if state.CardsInDeck <= 3 {
		hval -= 4.0*float64(len(state.Hands[player]))
	} else {
		hval -= 2.0*float64(len(state.Hands[player]))
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
			ovals = append(ovals, 200.0)
			continue
		}
		for _,c := range h {
			if c == durak.UNK_CARD {
				v += -4.0
			} else if c.Suit() == state.Trump.Suit() {
				v += 10.0*float64(c.Rank() + 1)
			} else {
				v += float64(c.Rank() - 6)
			}
		}
		if state.CardsInDeck <= 3 {
			v -= 4.0*float64(len(h))
		} else {
			v -= 2.0*float64(len(h))
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
		st := state.Clone()
		st.TakeAction(a)
		searchActs[i] = a
		children[i] = &GameState{*st}
	}
	return searchActs, children
}

func (state *GameState) FindBestAction(player int, depth int, timeBudget int64) (durak.Action, bool) {
	st := state.Clone()
	st.Mask(player)
	state = &GameState{*st}
	iface := search.SearchItDeep(state, player, depth, timeBudget)
	act, ok := iface.(durak.Action)
	return act, ok
}

