package spades 

import (
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

func TestStackWinner(t *testing.T) {
	s := Stack{Card(1), Card(2), Card(3), Card(4)}
	if s.Winner() != 3 {
		t.Errorf("winner %v", s.Winner())
	}
	s = Stack{Card(1), NO_CARD, Card(3), Card(10)}
	if s.Winner() != -1 {
		t.Errorf("winner not -1")
	}
	s = Stack{Card(20), Card(21), Card(7), Card(8)}
	if s.Winner() != 1 {
		t.Errorf("winner not 1")
	}
	//s = Stack{Card(
}
