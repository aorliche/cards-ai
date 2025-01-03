package main

import(
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aorliche/cards-ai/spades"
	"github.com/aorliche/cards-ai/server"
)

// To give player index during update
// As well as player names
type GameState struct {
	spades.GameState
	Player int
	Names []string
	Actions []spades.Action
}

type Game struct {
	sync.Mutex
	Key int
	Players []*server.Player
	// The actual game state
	State *GameState
	Terminated bool
}

func CreateGame() server.Game {
	return &Game{Players: make([]*server.Player, 0)}
}

func (game *Game) Terminate() {
	game.Terminated = true
}

func (game *Game) GetKey() int {
	return game.Key
}

func (game *Game) SetKey(key int) {
	game.Key = key
}

func (game *Game) HasOpenSlots() bool {
	for _,p := range game.Players {
		if p.Type == "Human" && !p.Joined {
			return true
		}
	}
	return false
}

func (game *Game) IsOver() bool {
	return game.State.IsOver() || game.Terminated
}
	
func (game *Game) AddPlayer(player server.Player) {
	game.Players = append(game.Players, &player)
}

func (game *Game) GetPlayers() []*server.Player {
	return game.Players
}

func (game *Game) Init(string) error {
	n := len(game.Players)
	if n != 4 {
		return errors.New("Bad number of players for Spades")
	}
	game.State = &GameState{*spades.InitGameState(), 0, nil, nil}
	// AI Logic
	aiFunc := func (player int) {
		for !game.IsOver() {
			time.Sleep(200 * time.Millisecond)
			game.Lock()
			st := game.State.Clone()
			game.Unlock()
			if game.IsOver() {
				break
			}
			var act spades.Action
			if st.Bids[player] == -1 && len(st.PlayerActions(player)) > 0 {
				b := st.DecideBids(player, 100)
				// Computers are conservative
				if b > 0 {
					b--
				}
				act = spades.Action{Verb: spades.BidVerb, Player: player, Bid: b, Card: spades.NO_CARD}
			} else if st.Trick[0] == spades.NO_CARD && st.Attacker == player {
				if st.PrevTrick[0] != spades.NO_CARD {
					time.Sleep(2000 * time.Millisecond)
				}
				act = st.DecidePlayFirst(100)
			} else {
				i := (player + 4 - st.Attacker) % 4
				j := (i + 4 - 1) % 4
				if st.Trick[i] == spades.NO_CARD && st.Trick[j] != spades.NO_CARD {
					act = st.DecidePlayNotFirst(100)
				}
			}
			game.Lock()
			acts := game.State.PlayerActions(player)
			for _,a := range acts {
				if a == act {
					game.State.TakeAction(act)
					sumTricks := 0
					for i := 0; i < 4; i++ {
						sumTricks += game.State.Tricks[i]
					}
					log.Println(sumTricks, act.ToStr())
					server.UpdatePlayers(game)
					break
				}
			}
			game.Unlock()
		}
	}
	// Start AI players
	for i,p := range game.Players {
		if p.Type == "Computer" {
			go aiFunc(i)
		}
	}
	return nil
}

func (game *Game) Join(string) error {
	return nil
}

// No need to lock in here since this is done in server code
func (game *Game) Action(data string) error {
	var act spades.Action
	json.NewDecoder(bytes.NewBuffer([]byte(data))).Decode(&act)
	actions := game.State.PlayerActions(act.Player)
	for _,a := range actions {
		if a == act {
			game.State.TakeAction(act)
			return nil
		}
	}
	return errors.New("Invalid action")
}

func (game *Game) GetState(player int) (string, error) {
	// Set player
	game.State.Player = player
	// Add player names
	game.State.Names = make([]string, len(game.Players))
	for i,p := range game.Players {
		if p.Type == "Human" {
			game.State.Names[i] = p.Name
		} else {
			game.State.Names[i] = p.Type
		}
	}
	// Get player actions
	game.State.Actions = game.State.PlayerActions(player)
	data, err := json.Marshal(*game.State)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
    log.SetFlags(0)
    server.ServeLocalFiles([]string{
		"/home/anton/GitHub/cards-ai/static/cards/fronts",
		"/home/anton/GitHub/cards-ai/static/cards/backs",
		"/home/anton/GitHub/cards-ai/static/images",
		"/home/anton/GitHub/cards-ai/static",
		"/home/anton/GitHub/cards-ai/static/js",
		"/home/anton/GitHub/cards-ai/static/css",
	}, 
	[]string{
		"/cards/fronts",
		"/cards/backs",
		"/images",
		"",
		"/js",
		"/css",
	})
	server.CreateGameFunc = CreateGame
    http.HandleFunc("/ws", server.Socket)
    log.Fatal(http.ListenAndServe(":8004", nil))
}
