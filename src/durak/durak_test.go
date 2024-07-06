package durak

import (
	"math/rand/v2"
	"testing"
	//"fmt"
)

func TestBeats(t *testing.T) {
    if Card(10).Beats(Card(11), Card(20)) {
        t.Errorf("Card(10).Beats(Card(11), Card(20))")
    }
    if Card(4).Beats(Card(17), Card(11)) {
        t.Errorf("Card(4).Beats(Card(17), Card(1))")
    }
}

func TestInitGameState(t *testing.T) {
    state := InitGameState(2)
    if state == nil {
        t.Errorf("InitGameState failed")
    }
}

func TestInitBadPlayerNumbers(t *testing.T) {
	state := InitGameState(7)
	if state != nil {
		t.Errorf("Maximum number of players is 6")
	}
	state = InitGameState(1)
	if state != nil {
		t.Errorf("Minimum number of players is 1")
	}
}

func TestAllActions(t *testing.T) {
	state := InitGameState(2) 
	acts := state.AllActions()
	if len(acts) != 6 {
		t.Errorf("%d actions instead of 6", len(acts))
	}
}

func TestRandom3PlayerGame(t *testing.T) {
	for i := 0; i<10; i++ {
		state := InitGameState(3)
		count := 0
		for !state.GameOver() {
			acts := state.AllActions()
			if len(acts) == 0 {
				t.Errorf("No actions! attacker: %v, defender: %v, cards in deck: %v won: %v hands: %v", state.Attacker, state.Defender, state.CardsInDeck, state.Won, state.Hands)
			}
			act := acts[rand.IntN(len(acts))]
			state.TakeAction(act)
			count++
			if count > 1000 {
				t.Errorf("Game too long")
			}
		}
	}
}

func TestRandom6PlayerGame(t *testing.T) {
	for i := 0; i<10; i++ {
		state := InitGameState(6)
		count := 0
		for !state.GameOver() {
			acts := state.AllActions()
			if len(acts) == 0 {
				t.Errorf("No actions! attacker: %v, defender: %v, cards in deck: %v won: %v hands: %v", state.Attacker, state.Defender, state.CardsInDeck, state.Won, state.Hands)
			}
			act := acts[rand.IntN(len(acts))]
			state.TakeAction(act)
			count++
			if count > 1000 {
				t.Errorf("Game too long")
			}
		}
	}
}
