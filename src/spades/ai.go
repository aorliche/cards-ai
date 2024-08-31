package spades

import (
	"fmt"
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
	// This used to be a problem when we incorrectly played spades
	// but had first suit left
	if hsize > len(possible) {
		fmt.Println(simulator, simulated)
		fmt.Println(state.Hands)
		fmt.Println(possible)
		fmt.Println(state.Absent[simulator][simulated])
		panic("Bad possible size")
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

// Simulate playing each card in hand
func (state *GameState) DecidePlayFirst(timeBudget int64) Action {
	hand := state.Hands[state.Attacker]
	wins := make([]int, len(hand))
	sims := 0
	start := time.Now()
	for i := 0; i < 100000; i++ {
		if time.Since(start).Milliseconds() > timeBudget {
			break
		}
		j := i % len(hand)
		if j == 0 {
			sims++
		}
		st := state.Clone()
		c := hand[j]
		act := Action{Verb: PlayVerb, Card: c, Player: state.Attacker}
		st.TakeAction(act)
		for k := 1; k < 4; k++ {
			pp := (state.Attacker+k)%4
			st.Hands[pp] = st.SimulateHand(state.Attacker, pp)
			st.TryWinTrick(pp)
		}
		if state.Tricks[state.Attacker] != st.Tricks[state.Attacker] {
			wins[j]++
		}
	}
	c := ChooseWinningCard(hand, wins, sims)
	// Decide none of the win percentages are good enough
	if c == NO_CARD {
		c = state.ChooseLowValueCard(state.Attacker)
	}
	return Action{Verb: PlayVerb, Player: state.Attacker, Card: c}
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
	// Out of possible in player actions
	possible := make([]Card, 0)
	outer:
	for _,a := range state.PlayerActions(player) {
		c := a.Card
		fmt.Println(c)
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
		c := state.ChooseLowValueWinningCard(possible)
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
			pp := (player+k+1)%4
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
	for _,a := range state.PlayerActions(player) {
		c := a.Card
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
	if card == NO_CARD {
		panic("NO_CARD in ChooseLowValueCard")
	}
	return card
}

func Includes[T comparable](haystack []T, needle T) bool {
	for _,t := range haystack {
		if t == needle {
			return true
		}
	}
	return false
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
	eval := 0
	for _,c := range hand {
		// Check how many cards higher rank than this are left
		s := c.Suit()
		n := 0
		for i := c+1; i < 13; i++ {
			cc := Card(s*13 + int(i))
			// If it's known missing from our hand then it has been played
			// Skip if we're the one holding the higher card
			if state.Absent[player][player][cc] && !Includes(hand, cc) {
				n++
			}
		}
		val := 5 - 2*n - 4*noSuit[s]
		eval += val
	}
	// Check if we're close to being out of a suit
	for s := 0; s < 4; s++ {
		if s == SUIT_SPADES {
			continue
		}
		n := 0
		for _,c := range hand {
			ss := c.Suit()
			if s == ss {
				n++
			}
		}
		if noSuit[s] == 0 {
			if n <= 3 {
				eval += 10 - 3*n;
			}
		} else if noSuit[s] == 1 {
			if n <= 2 {
				eval += 5 - 2*n;
			}
		}
	}
	return float64(eval);
}

func (state *GameState) ChooseLowValueWinningCard(possible []Card) Card {
	// Choose non-spades
	choice := NO_CARD
	for _,c := range possible {
		if c.Suit() == SUIT_SPADES {
			continue
		}
		if choice == NO_CARD || c.Rank() < choice.Rank() {
			choice = c
		}
	}
	if choice != NO_CARD {
		return choice
	}
	// Choose spades
	for _,c := range possible {
		if choice == NO_CARD || c.Rank() < choice.Rank() {
			choice = c
		}
	}
	return choice 
}

func (state *GameState) CheckAbsentCompatible() {
	for p,hand := range state.Hands {
		for _,c := range hand {
			for i := 0; i < 4; i++ {
				if state.Absent[i][p][int(c)] {
					fmt.Println(i, p, int(c))
					fmt.Println(state.Hands)
					fmt.Println(state.Absent[i][p])
					panic("not compatible")
				}
			}
		}
	}
}
