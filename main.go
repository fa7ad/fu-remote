package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"

	"github.com/gorilla/websocket"
)

var (
	port     = flag.String("port", "8080", "http port for the server")
	upgrader = websocket.Upgrader{}
)

func init() {
	flag.Parse()
}

func main() {
	log.SetFlags(0)
	http.HandleFunc("/fu", ws)
	http.Handle("/", http.FileServer(http.Dir("web/")))
	log.Printf("Listening on http://localhost:%s/", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	log.Println("New client connect")
	defer log.Println("Client disconnected")
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		run := func() bool {
			return false
		}

		switch string(message) {
		case "vol+":
			run = vol("5%+")
		case "vol-":
			run = vol("5%-")
		case "mute":
			run = vol("toggle")
		}

		err = c.WriteMessage(mt, getMessage(run()))
		if err != nil {
			break
		}
	}
}

func vol(val string) func() bool {
	return func() bool {
		cmd := exec.Command("amixer", "sset", "Master", val)
		err := cmd.Run()

		return err == nil
	}
}

func getMessage(res bool) []byte {
	if res {
		return []byte("success")
	}
	return []byte("fail")
}
