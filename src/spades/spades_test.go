package spades 

import (
	assert "gotest.tools/v3/assert"
	"testing"
)

func TestBeats(t *testing.T) {
	if Card(1).Beats(Card(2), 0) {
		t.Errorf("%v beats %v", Card(1), Card(2))
	}
	if !Card(2).Beats(Card(1), 0) {
		t.Errorf("%v doesn't beat %v", Card(2), Card(1))
	}
	if Card(2).Beats(Card(13), 0) {
		t.Errorf("%v beats %v", Card(2), Card(13))
	}
	if !Card(13).Beats(Card(2), 0) {
		t.Errorf("%v doesn't beat %v", Card(13), Card(2))
	}
	if Card(2).Beats(Card(18), 0) {
		t.Errorf("%v beats %v", Card(2), Card(18))
	}
}

func TestTrickWinner(t *testing.T) {
	s := Trick{Card(1), Card(2), Card(3), Card(4)}
	if s.Winner() != 3 {
		t.Errorf("winner %v", s.Winner())
	}
	s = Trick{Card(1), NO_CARD, Card(3), Card(10)}
	if s.Winner() != -1 {
		t.Errorf("winner not -1")
	}
	s = Trick{Card(20), Card(21), Card(7), Card(8)}
	if s.Winner() != 1 {
		t.Errorf("winner not 1")
	}
}

func TestOneAction(t *testing.T) {
	state := InitGameState()
	state.Bids = [4]int{1,1,1,1}
	acts := state.PlayerActions(0)
	if len(acts) != 13 {
		t.Errorf("first player doesn't have 13 actions has %v", len(acts))
	}
	assert.Assert(t, len(state.Hands[0]) == 13, "len hand = %v", len(state.Hands[0]))
	cop := state.Clone()
	state.TakeAction(acts[0])
	if state.Trick[0] != acts[0].Card {
		t.Errorf("Card not put on the stack")
	}
	if len(state.Hands[0]) != 12 {
		t.Errorf("first player doesn't have 12 cards has %v", len(state.Hands[0]))
	}
	assert.Assert(t, cop.Trick[0] != state.Trick[0], "Clone stacks alias")
}

func TestGameToCompletion(t *testing.T) {
	state := InitGameState()
	count := 0
	for !state.IsOver() && count < 1000 {
		acts := state.CurrentActions()
		if len(acts) == 0 {
			t.Errorf("%v", state)
		}
		state.TakeAction(acts[0])
		count++
	}
	assert.Assert(t, count != 1000, "Game never finished")
	out:
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 52; k++ {
				if !state.Absent[i][j][k] {
					t.Errorf("Some cards not marked absent: %v", state.Absent)
					break out
				}
			}
		}
	}
}
