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
	Debug(int) []int
}

// Iterative deepening
// Most general, no alpha-beta pruning
func SearchItDeep(state GameState, player int, depth int, timeBudget int64) (Action, float64) {
    startTime := time.Now()
	best := Action(nil)
	bestState := GameState(nil)
	for d := 0; d < depth; d++ {
		act, state, _, to := Search(state, player, d, startTime, timeBudget, true)
		if to {
			break
		}
		if act != nil {
			best = act
			bestState = state
		}
	}
	e := 0.0
	if bestState != nil {
		e = bestState.Eval(player)
	}
	return best, e
}

// Returns best action
// Returns state for best action
// Returns number of states searched
// Returns whether it's been timed out
func Search(state GameState, player int, depth int, startTime time.Time, timeBudget int64, playerOnly bool) (Action, GameState, int, bool) {
	if depth == 0 {
		return nil, state, 1, false
	}
    if time.Since(startTime).Milliseconds() > timeBudget {
        return nil, nil, 0, true
    }
	if state.IsOver() {
		return nil, state, 1, false
	}
	ns := 0
	best := math.Inf(-1)
	bestAction := Action(nil)
	bestState := GameState(nil)
	allEmpty := true
	for i := 0; i < state.NumPlayers(); i++ {
		if playerOnly && i != player {
			continue
		}
		actions, states := state.Children(i)
		if len(actions) > 0 {
			allEmpty = false
		}
		for j := 0; j < len(actions); j++ {
			_, s, n, to := Search(states[j], i, depth-1, startTime, timeBudget, false)
			if to {
				return nil, nil, 0, true
			}
			if s == nil {
				continue
			}
			e := s.Eval(i)
			if e > best {
				best = e
				bestAction = actions[j]
				bestState = s
			}
			ns += n
		}
	}
	if allEmpty {
		return nil, state, 1, false
	}
	return bestAction, bestState, ns, false
}
