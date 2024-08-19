package spades

import (
	"math/rand/v2"
	"sort"
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

func (state *GameState) ChooseLowValueCard(player int) Card {
	card := NO_CARD
	val := -1.0
	hand := make([]Card, len(state.Hands[player])-1)
	for _,c := range state.Hands[player] {
		count := 0
		for i := 0; i < len(state.Hands[player]); i++ {
			cc := state.Hands[player][i]
			if cc == c {
				continue
			}
			hand[count] = cc
			count++
		}
		v := state.EvalHand(player, hand)
		if card == NO_CARD || val > v {
			val = v
			card = c
		}
	}
	return card
}

func (state *GameState) EvalHand(player int, hand []Card) float64 {
	// See which players have gotten rid of which suits
	noSuit := [4]int{}
	for i := 0; i < 4; i++ {
		if i == player {
			continue
		}
		outer:
		for suit := 0; suit < 4; suit++ {
			if suit == SUIT_SPADES {
				continue
			}
			for j := 0; j < 13; j++ {
				card := suit*13 + j
				if !state.Absent[player][i][card] {
					continue outer
				}
			}
			noSuit[suit]++
		}
	}
	// Evaluations
	evals := make([]float64, len(hand))
	for _,c := range hand {
		// Check how many higher rank cards than this are left
		s := c.Suit()
		n := 0
		for i := c+1; i < 13; i++ {
			cc := Card(s*13 + i)
			if state.Absent[player][player][cc] {
				n++
			}
		}

	}
	// Check for being close to being rid of a suit 
	// Ignore King or highter
	suits := [4]int{}[:]
	for _,c := range state.Hands[player] {
		suits[c.Suit()]++
	}
	sort.Ints(suits)
	// Check if we can be first out of a suit
	outer:
	for suit := 0; suit < 4; suit++ {
		if suit == SUIT_SPADES || suits[suit] == 0 {
			continue
		}
		for i := 0; i < 4; i++ {
			if i == player {
				continue
			}
			if noSuit[i][suit] {
				continue outer
			}
		}
		if suits[suit] >= 3 {
			break
		}
		minCard := 14
		for _,c := range state.Hands[player] {
			if c.Suit() == suit && int(c) < minCard {
				minCard = int(c)
			}
		}
		return Card(minCard)
	}
	// Get rid of any low rank cards

}
