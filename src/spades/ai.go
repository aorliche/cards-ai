package spades

import (
	"math/rand/v2"
	"time"
)

func (state *GameState) SimulateHand(simulator int, simulated int) []Card {
	hsize := len(state.Hands[simulated])
	possible := make([]Card, 0)
	for i := 0; i < 52; i++ {
		if !state.Absent[simulator][simulated][i] {
			possible = append(possible, Card(i))
		}
	}
	idcs := rand.Perm(len(possible))
	hand := make([]Card, hsize)
	for i := 0; i < hsize; i++ {
		hand[i] = possible[idcs[i]]
	}
	return hand
}

// Assume first card in trick has been played
// Player hand has been populated with simulated cards
func (state *GameState) TryWinTrick(player int) {
	acts := state.PlayerActions(player) 
	won := false
	var act Action
	outer:
	for _,a := range acts {
		c := a.Card
		for i := 0; i < 4; i++ {
			if !c.Beats(state.Trick[i], state.Trick[0].Suit()) {
				continue outer
			}
		}
		won = true
		act = a
	}
	// Random card for best simulation
	// Can also get low value card, but may be too expensive for loop
	if !won {
		act = acts[rand.IntN(len(acts))]
	}
	state.TakeAction(act)
}

func (state *GameState) DecidePlayNotFirst(timeBudget int64) Action {
	player := state.Attacker
	for i := 0; i < 4; i++ {
		if state.Trick[i] == NO_CARD {
			player = (player+i)%4
			break
		}
	}
	// Find cards that win the trick so far
	possible := make([]Card, 0)
	outer:
	for _,c := range state.Hands[player] {
		for i := 0; i < 4; i++ {
			if !c.Beats(state.Trick[i], state.Trick[0].Suit()) {
				continue outer
			}
		}
		possible = append(possible, c)
	}
	// Not possible to win the trick
	// Choose card to get rid of
	if len(possible) == 0 {
		c := state.ChooseLowValueCard(player)
		return Action{Verb: PlayVerb, Player: player, Card: c}
	}
	// We're the last to play and can just win the trick
	if (player + 1)%4 == state.Attacker {
		c := state.ChooseLowValueWinningCard(player)
		return Action{Verb: PlayVerb, Player: player, Card: c}
	}
	// For each winning card, simulate the probability of it winning the trick
	wins := make([]int, len(possible))
	sims := 0
	start := time.Now()
	for i := 0; i < 100000; i++ {
		if time.Since(start).Milliseconds() > timeBudget {
			break
		}
		j := i % len(possible)
		if j == 0 {
			sims++
		}
		st := state.Clone()
		c := possible[j]
		act := Action{Verb: PlayVerb, Card: c, Player: player}
		st.TakeAction(act)
		nPlayersSim := 0
		for k := 0; k < 4; k++ {
			if st.Trick[k] == NO_CARD {
				nPlayersSim++
			}
		}
		for k := 0; k < nPlayersSim; k++ {
			pp := (player+i+1)%4
			st.Hands[pp] = st.SimulateHand(player, pp)
			st.TryWinTrick(pp)
		}
		if state.Tricks[player] != st.Tricks[player] {
			wins[j]++
		}
	}
	c := ChooseWinningCard(possible, wins, sims)
	// Decide none of the win percentages are good enough
	if c == NO_CARD {
		c = state.ChooseLowValueCard(player)
	}
	return Action{Verb: PlayVerb, Player: player, Card: c}
}

// Simple: choose lowest value winning card over 50% win rate
// Even simpler: choose lowest win rate over 50%
func ChooseWinningCard(cards []Card, wins []int, sims int) Card {
	wmin := -1.0
	c := NO_CARD
	for i,w := range wins {
		fct := float64(w) / float64(sims) 
		if fct < 0.5 {
			continue
		}
		if c == NO_CARD || fct < wmin {
			wmin = fct
			c = cards[i]
		} 
	}
	return c
}
