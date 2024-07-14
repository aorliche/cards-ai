package search

import (
	//"log"
	"math"
	"time"
)

type Action interface{}

type GameState interface {
	NumPlayers() int
	Eval(int) float64
	Children(int) ([]Action, []GameState)
}

// Iterative deepening
// Most general, no alpha-beta pruning
func SearchItDeep(state GameState, player int, depth int, timeBudget int64) Action {
    startTime := time.Now()
	actions, states := state.Children(player)
	if len(actions) == 0 {
		return nil
	}
	prevBestAction := Action(nil)
	outer:
	for d := 1; d < depth; d++ {
		best := math.Inf(-1)
		bestAction := Action(nil)
		for i := 0; i < len(actions); i++ {
			_, e, _, to := Search(states[i], player, d, startTime, timeBudget)
			// Time exceeded
			if to {
				break outer
			}
			if e > best {
				best = e
				bestAction = actions[i]
			}
		}
		prevBestAction = bestAction
	}
	return prevBestAction
}

// Returns best action
// Returns evaluation
// Returns number of states searched
// Returns whether it's been timed out
func Search(state GameState, player int, depth int, startTime time.Time, timeBudget int64) (Action, float64, int, bool) {
	if depth == 0 {
		return nil, state.Eval(player), 1, false
	}
    if time.Since(startTime).Milliseconds() > timeBudget {
        return nil, 0, 0, true
    }
	ns := 0
	best := math.Inf(-1)
	bestMult := 1
	bestAction := Action(nil)
	for i := 0; i < state.NumPlayers(); i++ {
		actions, states := state.Children(i)
		if len(actions) == 0 {
			e := state.Eval(i)
			if e > best {
				best = e
				bestAction = nil
				if i != player {
					bestMult = -1
				}
			}
			ns += 1
		} else {
			for j := 0; j < len(actions); j++ {
				_, e, n, to := Search(states[j], i, depth-1, startTime, timeBudget)
				if to {
					return nil, 0, 0, true
				}
				if e > best {
					best = e
					bestAction = actions[j]
					if i != player {
						bestMult = -1
					}
				}
				ns += n
			}
		}
	}
	return bestAction, best*float64(bestMult), ns, false
}
