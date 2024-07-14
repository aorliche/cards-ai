package search

import (
	"math"
	"time"
)

type GameState interface {
	NumPlayers() int
	// Me, Other Player
	Eval(int, int) float64
	Children(int, int) []interface{}, []*GameState
}

// Iterative deepening
// Most general, no alpha-beta pruning
func (state *GameState) SearchItDeep(me int, player int, depth int, timeBudget int64) interface{} {
    startTime := time.Now()
	actions, states := state.Children(me, player)
	if len(actions) == 0 {
		return nil
	}
	pn := 0
	prevBestAction := interface{}
	for d := 1; d < depth; d++ {
		best := math.Inf(-1)
		bestAction := interface{}
		ns := 0
		for i := 0; i < len(actions); i++ {
			_, e, n, to := states[i].Search(me, player, d, startTime, timeBudget)
			// Time exceeded
			if to {
				break
			}
			if e > best {
				best = e
				bestAction = actions[i]
			}
			ns += n
		}
		// We get no further with further depth
		if ns == pn {
			break
		}
		pn = ns
		prevBestAction = bestAction
	}
	return prevBestAction
}

// Returns best action
// Returns evaluation
// Returns number of states searched
// Returns whether it's been timed out
func (state *GameState) Search(me int, player int, depth int, startTime time.Time, timeBudget int64) (interface{}, float64, int, bool) {
	if depth == 0 {
		return nil, state.Eval(me, player), 1, false
	}
    if time.Since(startTime).Milliseconds() > timeBudget {
        return nil, 0, 0, true
    }
	ns := 0
	best := math.Inf(-1)
	bestAction := interface{}
	for i := 0; i < state.NumPlayers(); i++ {
		actions, states := state.Children(me, i)
		if len(actions) == 0 {
			e := state.Eval(me, i)
			if i != player {
				e *= -1
			}
			if e > best {
				best = e
				bestAction = nil
			}
			ns += 1
		} else {
			for j := 0; j < len(actions); j++ {
				_, e, n, to := states[j].Search(me, i, depth-1, startTime, timeBudget)
				if to {
					return nil, 0, 0, true
				}
				if i != player {
					e *= -1
				}
				if e > best {
					best = e
					bestAction = actions[j]
				}
				ns += n
			}
		}
	}
	return bestAction, best, ns, false
}
