package ai

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/aorliche/cards-ai/durak"
)

func TestFindBestAction(t *testing.T) {
	state := &GameState{*durak.InitGameState(2)}
	log.Println(state.Trump)
	for i := 0; i < 2; i++ {
		act, ok := state.FindBestAction(i, 3, 1000)
		log.Println(i, act.ToStr(), ok)
	}
}

func TestTwoPlayerGame(t *testing.T) {
	state := &GameState{*durak.InitGameState(2)}
	var mut sync.Mutex
	loopFn := func (player int) {
		for !state.IsOver() {
			time.Sleep(200 * time.Millisecond)
			act, ok := state.FindBestAction(player, 10, 100)
			if !ok {
				continue
			}
			mut.Lock()
			acts := state.PlayerActions(player)
			for _,a := range acts {
				if a == act {
					log.Println(state.CardsInDeck, act.ToStr())
					state.TakeAction(act)
					break
				}
			}
			mut.Unlock()
		}
	}
	go loopFn(0)
	go loopFn(1)
	for !state.IsOver() {
		mut.Lock()
		a1 := state.PlayerActions(0)
		a2 := state.PlayerActions(1)
		if len(a1) == 0 && len(a2) == 0 {
			log.Println(state)
			mut.Unlock()
			break
		}
		mut.Unlock()
	}
}
