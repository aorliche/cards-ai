package main 

import (
	"log"
	"sync"
	"time"

	"github.com/aorliche/cards-ai/durak"
	"github.com/aorliche/cards-ai/durak/ai"
)

var winBonus = []float64{500.0}
var trumpBonus = []float64{15.0, 25.0, 35.0}
var unknown = []float64{2.0}
var cardsCutoff = []int{3}
var smallDeck = []float64{15.0, 20.0, 25.0}
var bigDeck = []float64{2.0, 5.0, 10.0}

var nSimulGames = 10
var nBatch = 3

func main() {
	startGame := func (params *ai.EvalParams) *ai.GameState {
		// Init game state
		var mutex sync.Mutex
		state := &ai.GameState{*durak.InitGameState(2), nil}
		stime := time.Now()
		// AI Logic
		aiFunc := func (player int) {
			for !state.IsOver() {
				if time.Since(stime) > 600*time.Second {
					log.Println("Timeout")
					break
				}
				time.Sleep(200 * time.Millisecond)
				mutex.Lock()
				ps := params
				if player == 0 {
					ps = nil
				}
				st := &ai.GameState{*state.Clone(), ps}
				mutex.Unlock()
				act, ok := st.FindBestAction(player, 12, 2000)
				if !ok {
					continue
				}
				mutex.Lock()
				acts := state.PlayerActions(player)
				for _,a := range acts {
					if a == act {
						log.Println(state.CardsInDeck, act.ToStr())
						state.TakeAction(act)
						break
					}
				}
				mutex.Unlock()
			}
		}
		go aiFunc(0)
		go aiFunc(1)
		return state
	}
	winners := make(map[int]int)
	for a := 0; a < 27; a++ {
		a0 := a % 3
		a1 := (a/3) % 3
		a2 := (a/9) % 3
		params := &ai.EvalParams{
			winBonus[0],
			trumpBonus[a0],
			unknown[0],
			cardsCutoff[0],
			smallDeck[a1],
			bigDeck[a2],
		}
		w := 0
		for b := 0; b < nBatch; b++ {
			states := make([]*ai.GameState, nSimulGames)
			for i := 0; i < nSimulGames; i++ {
				states[i] = startGame(params)
			}
			numActive := func () int {
				n := 0
				for i := 0; i < len(states); i++ {
					if !states[i].IsOver() {
						n++
					}
				}
				return n
			}
			for numActive() > 0 {
				log.Println(numActive(), "still active")
				log.Println(winners)
				time.Sleep(200 * time.Millisecond)
			}
			// Count winners
			for i := 0; i < len(states); i++ {
				if states[i].Won[0] {
					w++
				}
			}
			winners[a] = w
			log.Println(winners)
		}
	}
}
