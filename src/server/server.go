package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

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

type Game interface {
	// No data fields in golang interfaces
	GetKey() *int
	GetNames() *[]string
	GetTypes() *[]string
	GetJoined() *[]bool
	GetMutex() *sync.Mutex
	GetConns() *[]*websocket.Conn
	// Callbacks
	IsOver() bool
	Init(string) error
	Join(string) error
	Action(string) error
	// Update info for player n
	GetState(int) string
}

// Hack
var InitGameFunc func(string) (Game, error)

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
func Update(game Game) {
	for i,c := range *game.GetConns() {
		if c == nil {
			continue
		}
		reply := Reply{Type: "Update", Data: game.GetState(i)}
		jsn, _ := json.Marshal(reply)
		c.WriteMessage(websocket.TextMessage, jsn);
	}
}

func Socket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	player := -1
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return  
		}
		// Do we ever get any other types of messages?
		if msgType != websocket.TextMessage {
			log.Println("Not a text message")
			return
		}
		var req Request
		json.NewDecoder(bytes.NewBuffer(msg)).Decode(&req)
		//log.Println(string(msg))
		switch req.Type {
			case "List" : {
				keys := make([]int, 0)
				for key,game := range games {
					// Check if has not been won and has open slots
					if game.IsOver() {
						continue
					}
					joined := *game.GetJoined()
					for i := 0; i < len(joined); i++ {
						if !joined[i] {
							keys = append(keys, key) 
							break
						}
					}
				}
				// Here we write a raw array of ints
				jsn, _ := json.Marshal(keys)
				err = conn.WriteMessage(websocket.TextMessage, jsn)
				if err != nil {
					log.Println(err)
					continue
				}
			}
			case "New": {
				if player != -1 {
					log.Println("Player already joined")
					continue
				}
				player = 0
				if InitGameFunc == nil {
					log.Println("Attempted to create game without InitGameFunc being set")
					continue
				}
				game, err := InitGameFunc(req.Data)
				if err != nil {
					log.Println("Error in InitGameFunc")
					continue
				}
				*game.GetKey() = NextGameIdx()
				*game.GetNames() = make([]string, len(req.Types))
				*game.GetTypes() = append(make([]string, 0), req.Types...)
				*game.GetJoined() = make([]bool, len(req.Types))
				*game.GetConns() = make([]*websocket.Conn, len(req.Types))
				*game.GetMutex() = sync.Mutex{}
				(*game.GetNames())[0] = req.Name
				(*game.GetConns())[0] = conn
				(*game.GetJoined())[0] = true
				// Not protected by mutex because game only becomes visible here
				games[*game.GetKey()] = game
				Update(game)
			}
			case "Join": {
				if player != -1 {
					log.Println("Player already joined")
					continue
				}
				game := games[req.Game]
				if game == nil { 
					log.Println("No such game", req.Game)
					continue
				}
				game.GetMutex().Lock()
				joined := *game.GetJoined()
				for i := 1; i < len(joined); i++ {
					if !joined[i] {
						player = i
						(*game.GetNames())[i] = req.Name
						break
					}
				}
				game.GetMutex().Unlock()
				if player == -1 {
					log.Println("No open slots")
					continue
				}
				game.GetMutex().Lock()
				(*game.GetJoined())[player] = true
				(*game.GetConns())[player] = conn
				// Tell the client a player has joined, giving optional data
				game.Join(req.Data)
				Update(game)
				game.GetMutex().Unlock()
			}
			case "Action": {
				game := games[req.Game]
				if game == nil { 
					log.Println("No such game", req.Game)
					continue
				}
				if player == -1 || player > len(*game.GetJoined()) {
					log.Println("Invalid player")
					continue
				}
				if (*game.GetTypes())[player] == "Human" && !(*game.GetJoined())[player] {
					log.Println("Player not joined")
					continue
				}
				if game.IsOver() {
					log.Println("Game already over")
					continue
				}
				game.GetMutex().Lock()
				// Let the game do what it wants with the action
				game.Action(req.Data)
				Update(game)
				game.GetMutex().Unlock()
			}
			case "Chat:": {
				game := games[req.Game]
				if game == nil { 
					log.Println("No such game", req.Game)
					continue
				}
				myName := ""
				for i,c := range *game.GetConns() {
					if c == conn {
						myName = (*game.GetNames())[i]
						break
					}
				}
				if myName == "" {
					log.Println("Unknown connection")
					continue
				}
				chat := Chat{Name: myName, Message: req.Data}
				jsnChat, _ := json.Marshal(chat)
				reply := Reply{Type: "Chat", Data: string(jsnChat)}
				jsnReply, _ := json.Marshal(reply)
				// Write message to all connected players
				game.GetMutex().Lock()
				for _,c := range *game.GetConns() {
					if c != nil {
						c.WriteMessage(websocket.TextMessage, jsnReply);
					}
				}
				game.GetMutex().Unlock()
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
            //fmt.Println(f.Name(), f.IsDir())
            if f.IsDir() {
                continue
            }
            webFile := webDir + "/" + f.Name()
            fsFile := fsDir + "/" + f.Name()
            http.HandleFunc(webFile, Headers(func (w http.ResponseWriter, req *http.Request) {ServeStatic(w, req, fsFile)}))
        }
    }
}
