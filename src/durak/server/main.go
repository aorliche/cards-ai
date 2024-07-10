package main

import(
	"errors"
	"log"
	"net/http"
	"sync"

    "github.com/gorilla/websocket"

	"github.com/aorliche/cards-ai/durak"
	"github.com/aorliche/cards-ai/server"
)

type Game struct {
	Key int
	Names []string
	Types []string
	Joined []bool
	Mutex sync.Mutex
	Conns []*websocket.Conn
	// The actual game state
	State *durak.GameState
}

func CreateGame() server.Game {
	return &Game{}
}

func (game *Game) GetKey() *int {
	return &game.Key
}

func (game *Game) GetNames() *[]string {
	return &game.Names
}

func (game *Game) GetTypes() *[]string {
	return &game.Types
}

func (game *Game) GetJoined() *[]bool {
	return &game.Joined
}

func (game *Game) GetMutex() *sync.Mutex {
	return &game.Mutex
}

func (game *Game) GetConns() *[]*websocket.Conn {
	return &game.Conns
}

func (game *Game) IsOver() bool {
	return game.State.IsOver()
}

func (game *Game) Init(string) error {
	n := len(game.Types)
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
    server.ServeLocalFiles([]string{"/home/anton/GitHub/cards-ai/static/cards/fronts"}, []string{"/cards"})
	server.CreateGameFunc = CreateGame
    http.HandleFunc("/ws", server.Socket)
    log.Fatal(http.ListenAndServe(":8000", nil))
}
