package main

import(
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/aorliche/cards-ai/durak"
	"github.com/aorliche/cards-ai/server"
)

type Game struct {
	Key int
	Players []*server.Player
	// The actual game state
	State *durak.GameState
	sync.Mutex
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
	return game.State.IsOver()
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
	game.State = durak.InitGameState(n)
	// TODO start AI players
	return nil
}

func (game *Game) Join(string) error {
	return nil
}

func (game *Game) Action(string) error {
	return nil
}

func (game *Game) GetState(player int) (string, error) {
	return "", nil
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
