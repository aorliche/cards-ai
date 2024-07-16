package search

import (
	//"log"
	"math"
	"time"

	//"github.com/aorliche/cards-ai/durak"
)

type Action interface{}

type GameState interface {
	NumPlayers() int
	Eval(int) float64
	Children(int) ([]Action, []GameState)
	IsOver() bool
}

// Iterative deepening
// Most general, no alpha-beta pruning
func SearchItDeep(state GameState, player int, depth int, timeBudget int64) Action {
    startTime := time.Now()
	best := Action(nil)
	for d := 0; d < depth; d++ {
		act, _, _, to := Search(state, player, d, startTime, timeBudget, true)
		if to {
			break
		}
		if act != nil {
			best = act
		}
	}
	return best
}

// Returns best action
// Returns evaluation
// Returns number of states searched
// Returns whether it's been timed out
func Search(state GameState, player int, depth int, startTime time.Time, timeBudget int64, playerOnly bool) (Action, float64, int, bool) {
	if depth == 0 {
		return nil, state.Eval(player), 1, false
	}
    if time.Since(startTime).Milliseconds() > timeBudget {
        return nil, 0, 0, true
    }
	if state.IsOver() {
		return nil, state.Eval(player), 1, false
	}
	ns := 0
	best := math.Inf(-1)
	bestAction := Action(nil)
	for i := 0; i < state.NumPlayers(); i++ {
		if playerOnly && i != player {
			continue
		}
		actions, states := state.Children(i)
		for j := 0; j < len(actions); j++ {
			_, e, n, to := Search(states[j], i, depth-1, startTime, timeBudget, false)
			if to {
				return nil, 0, 0, true
			}
			if e > best {
				best = e
				bestAction = actions[j]
			}
			ns += n
		}
	}
	return bestAction, best, ns, false
}
