package main

import(
	"log"
	"net/http"

	"github.com/aorliche/cards-ai/server"
)

func main() {
    log.SetFlags(0)
    server.ServeLocalFiles([]string{"/home/anton/GitHub/cards-ai/static/cards/fronts"}, []string{"/cards"})
    //http.HandleFunc("/ws", Socket)
    log.Fatal(http.ListenAndServe(":8000", nil))
}
