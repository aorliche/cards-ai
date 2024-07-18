package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"

    "github.com/gorilla/websocket"
)

type Request struct {
   Type string 
   Game int
   Name string
   Types []string
   // Json
   Data string
}

type Reply struct {
	Type string
	// Json
	Data string
}

type Chat struct {
	Name string
	Message string
}

type Player struct {
	Name string
	Type string
	Joined bool
	Conn *websocket.Conn
}

type Game interface {
	// Callbacks
	GetKey() int
	SetKey(int)
	HasOpenSlots() bool
	IsOver() bool
	Lock()
	Unlock()
	AddPlayer(Player)
	GetPlayers() []*Player
	Init(string) error
	Join(string) error
	Action(string) error
	// Update info for player n
	GetState(int) (string, error)
	// Terminate game on player disconnect
	Terminate()
}

// Hack
var CreateGameFunc func() Game

var games = make(map[int]Game)
var upgrader = websocket.Upgrader{} // Default options

func NextGameIdx() int {
    max := -1
    for key := range games {
        if key > max {
            max = key
        }
    }
    return max+1
}

// Update all human clients
// e.g., in response to an AI move
// Don't lock in update
func UpdatePlayers(game Game) {
	for i,p := range game.GetPlayers() {
		if p.Conn == nil {
			continue
		}
		state, err := game.GetState(i)
		if err != nil {
			log.Println("Error in GetState")
			continue
		}
		reply := Reply{Type: "Update", Data: state}
		jsn, _ := json.Marshal(reply)
		p.Conn.WriteMessage(websocket.TextMessage, jsn);
	}
}

func Socket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	var socketGame Game
	player := -1
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			if socketGame != nil && player != -1 {
				socketGame.Terminate()
				for _,p := range socketGame.GetPlayers() {
					if p.Conn != nil {
						p.Conn.Close()
					}
				}
			}
			return  
		}
		// Do we ever get any other types of messages?
		if msgType != websocket.TextMessage {
			log.Println("Not a text message")
			return
		}
		var req Request
		json.NewDecoder(bytes.NewBuffer(msg)).Decode(&req)
		log.Println(string(msg))
		switch req.Type {
			case "List" : {
				keys := make([]int, 0)
				for key,game := range games {
					// Check if has not been won and has open slots
					if game.IsOver() || !game.HasOpenSlots() {
						continue
					}
					keys = append(keys, key) 
				}
				sort.Ints(keys)
				jsn, _ := json.Marshal(keys)
				reply := Reply{Type: "List", Data: string(jsn)}
				repJsn, _ := json.Marshal(reply)
				err = conn.WriteMessage(websocket.TextMessage, repJsn)
				if err != nil {
					log.Println(err)
					continue
				}
			}
			case "New": {
				if CreateGameFunc == nil {
					log.Println("Attempted to create game without CreateGameFunc being set")
					continue
				}
				game := CreateGameFunc()
				game.SetKey(NextGameIdx())
				// Add players, and also
				// Check if we have a human player or send everything to conn 0
				player = -1
				for i,typ := range req.Types {
					game.AddPlayer(Player{
						Name: "",
						Type: typ,
						Joined: false,
						Conn: nil,
					})
					// Computer players are joined
					if typ != "Human" {
						game.GetPlayers()[i].Joined = true
					}
					if player == -1 && typ == "Human" {
						player = i
						game.GetPlayers()[i].Name = req.Name
						game.GetPlayers()[i].Conn = conn
						game.GetPlayers()[i].Joined = true
					}
				}
				if player == -1 {
					game.GetPlayers()[0].Conn = conn
				}
				// Check game state okay
				err := game.Init(req.Data)
				if err != nil {
					log.Println("Error in game init")
					log.Println(err)
					jsn, _ := json.Marshal(err.Error())
					reply := Reply{Type: "Error", Data: string(jsn)}
					repJsn, _ := json.Marshal(reply)
					conn.WriteMessage(websocket.TextMessage, repJsn);
					continue
				}
				// Not protected by mutex because game only becomes visible here
				games[game.GetKey()] = game
				// Send the game ID to the player
				jsn, _ := json.Marshal(game.GetKey())
				reply := Reply{Type: "New", Data: string(jsn)}
				repJsn, _ := json.Marshal(reply)
				conn.WriteMessage(websocket.TextMessage, repJsn);
				// For ending games on closed connections
				socketGame = game
				// Update single player
				UpdatePlayers(game)
			}
			case "Join": {
				game := games[req.Game]
				if game == nil { 
					log.Println("No such game ", req.Game)
					continue
				}
				game.Lock()
				player = -1
				for i,p := range game.GetPlayers() {
					if !p.Joined {
						player = i
						p.Name = req.Name
						p.Joined = true
						p.Conn = conn
						break
					}
				}
				game.Unlock()
				if player == -1 {
					log.Println("No open slots")
					continue
				}
				game.Lock()
				// Confirm join action by giving player a new gameId
				// (although he should already know)
				err := game.Join(req.Data)
				if err == nil {
					jsn, _ := json.Marshal(game.GetKey())
					reply := Reply{Type: "Join", Data: string(jsn)}
					repJsn, _ := json.Marshal(reply)
					conn.WriteMessage(websocket.TextMessage, repJsn);
					UpdatePlayers(game)
				}
				// For ending games on closed connections
				socketGame = game
				game.Unlock()
			}
			case "Action": {
				game := games[req.Game]
				if game == nil { 
					log.Println("No such game", req.Game)
					continue
				}
				if player == -1 || player >= len(game.GetPlayers()) {
					log.Println("Invalid player")
					continue
				}
				p := game.GetPlayers()[player]  
				if !p.Joined || p.Type != "Human" {
					log.Println("Player not joined or not human")
					continue
				}
				if game.IsOver() {
					log.Println("Game already over")
					continue
				}
				game.Lock()
				// Let the game do what it wants with the action
				err := game.Action(req.Data)
				if err != nil {
					log.Println(err)
				} else {
					UpdatePlayers(game)
				}
				game.Unlock()
			}
			case "Chat": {
				game := games[req.Game]
				if game == nil { 
					log.Println("No such game", req.Game)
					continue
				}
				if player <= -1 || player >= len(game.GetPlayers()) {
					log.Println("Chat from bad player")
					continue
				}
				myName := game.GetPlayers()[player].Name
				chat := Chat{Name: myName, Message: req.Data}
				jsnChat, _ := json.Marshal(chat)
				reply := Reply{Type: "Chat", Data: string(jsnChat)}
				jsnReply, _ := json.Marshal(reply)
				// Write message to all connected players
				game.Lock()
				for _,p := range game.GetPlayers() {
					if p.Conn != nil {
						p.Conn.WriteMessage(websocket.TextMessage, jsnReply);
					}
				}
				game.Unlock()
			}
		}
	}
}

type HFunc func (http.ResponseWriter, *http.Request)

func Headers(fn HFunc) HFunc {
    return func (w http.ResponseWriter, req *http.Request) {
        //fmt.Println(req.Method)
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers",
            "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
        fn(w, req)
    }
}
func ServeStatic(w http.ResponseWriter, req *http.Request, file string) {
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    http.ServeFile(w, req, file)
}

// Absolute paths and paths from root of webserver
func ServeLocalFiles(fsDirs []string, webDirs []string) {
	if len(fsDirs) != len(webDirs) {
		log.Fatal("Different lengths for dirs")
	}
	for i,fsDir := range fsDirs {
		webDir := webDirs[i]
        dir, err := os.Open(fsDir)
        if err != nil {
            log.Fatal(err)
        }
        files, err := dir.Readdir(0)
        if err != nil {
            log.Fatal(err)
        }
        for _,f := range files {
            log.Println(f.Name(), f.IsDir())
            if f.IsDir() {
                continue
            }
            webFile := webDir + "/" + f.Name()
            fsFile := fsDir + "/" + f.Name()
            http.HandleFunc(webFile, Headers(func (w http.ResponseWriter, req *http.Request) {ServeStatic(w, req, fsFile)}))
        }
    }
}
