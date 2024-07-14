package ai

import (
	"github.com/aorliche/cards-ai/search"
	"github.com/aorliche/cards-ai/durak"
)

type GameState struct {
	durak.GameState
}

func (state *GameState) NumPlayers() int {
	return len(state.Hands)
}

func (state *GameState) Eval(me int, player int) float64 {
	// Check my win
	if state.CardsInDeck == 0 && len(state.Hands[player]) == 0 {
		return float64(100)
	}
	// Mask (remember to restore)
	hands := state.Mask(me)
	// My hand
	hval := float64(0)
	for _,c := range state.Hands[player] {
		if c == durak.UNK_CARD {
			continue
		} else if c.Suit() == state.Trump.Suit() {
			hval += float64(c.Rank())
		} else {
			hval += float64(c.Rank() - 5)
		}
	}
	if state.CardsInDeck <= 3 {
		hval -= 4*len(state.Hands[player])
	}
	// Other player's hands
	ovals := make([]float64, 0)
	for i,h := range state.Hands {
		if i == player {
			continue
		}
		v := float64(0)
		// Check other player's win
		if state.CardsInDeck == 0 && len(h) == 0 {
			ovals = append(ovals, float64(100))
			continue
		}
		for _,c := range h {
			if c == durak.UNK_CARD {
				v += float64(-4)
			} else if c.Suit() == state.Trump.Suit() {
				v += float64(c.Rank())
			} else {
				v += float64(c.Rank() - 5)
			}
		}
		if state.CardsInDeck <= 3 {
			v -= 4*len(h)
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
	// Unmask
	state.Hands = sav
	return hval - worst
}

func (state *GameState) Children(me int, player int) []durak.Action, []*search.GameState {
	// Check that we don't have unknown cards on the board
	// If we do, we can't search further
	unk := false
	for i,c := range state.Plays {
		if c == durak.UNK_CARD || state.Covers[i] == durak.UNK_CARD {
			unk = true
			break
		}
	}
	if unk {
		return make([]durak.Action, 0), make([]*GameState, 0)
	}
	acts := state.PlayerActions(me, player)
	children := make([]*GameState, len(acts))
	for i,a := range acts {
		st := state.Clone()
		st.TakeAction(a)
		children[i] = st
	}
	return acts, children
}
