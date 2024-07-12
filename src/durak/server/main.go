package main

import(
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/aorliche/cards-ai/durak"
	"github.com/aorliche/cards-ai/server"
)

// To give player index during update
// As well as player names
type GameState struct {
	State *durak.GameState
	Player int
	Names []string
	Actions []durak.Action
}

type Game struct {
	sync.Mutex
	Key int
	Players []*server.Player
	// The actual game state
	State GameState
}

func CreateGame() server.Game {
	return &Game{Players: make([]*server.Player, 0)}
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
	return game.State.State.IsOver()
}
	
func (game *Game) AddPlayer(player server.Player) {
	game.Players = append(game.Players, &player)
}

func (game *Game) GetPlayers() []*server.Player {
	return game.Players
}

func (game *Game) Init(string) error {
	n := len(game.Players)
	if n < 2 || n > 6 {
		return errors.New("Bad number of players for Durak")
	}
	game.State.State = durak.InitGameState(n)
	// TODO start AI players
	return nil
}

func (game *Game) Join(string) error {
	return nil
}

func (game *Game) Action(data string) error {
	var act durak.Action
	json.NewDecoder(bytes.NewBuffer([]byte(data))).Decode(&act)
	actions := game.State.State.PlayerActions(act.Player)
	for _,a := range actions {
		if a == act {
			game.State.State.TakeAction(act)
			return nil
		}
	}
	return errors.New("Invalid action")
}

func (game *Game) GetState(player int) (string, error) {
	hands := game.State.State.Mask(player)
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
	game.State.Actions = game.State.State.PlayerActions(player)
	data, err := json.Marshal(game.State)
	game.State.State.Hands = hands
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
    log.Fatal(http.ListenAndServe(":8000", nil))
}
