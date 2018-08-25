package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

var (
	websocket = flag.Bool("websocket", true, "enable/disable websocket protocol")
)

func init() {
	flag.Parse()
}

func main() {
	opts := sockjs.DefaultOptions
	opts.Websocket = *websocket
	handler := sockjs.NewHandler("/ws", opts, wsHandler)
	http.Handle("/ws/", handler)
	http.Handle("/", http.FileServer(http.Dir("web/")))
	log.Println("Listening on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func wsHandler(session sockjs.Session) {
	log.Println("new sockjs session connected")
	for {
		msg, err := session.Recv()
		if err != nil {
			break
		}

		run := func() bool {
			return false
		}
		switch msg {
		case "vol+":
			run = func() bool {
				return vol(true)
			}
		case "vol-":
			run = func() bool {
				return vol(false)
			}
		}
		session.Send(getMessage(run()))
		continue
	}
	log.Println("sockjs session closed")
}

func vol(plus bool) bool {
	val := "5%+"
	if !plus {
		val = "5%-"
	}
	cmd := exec.Command("amixer", "sset", "Master", val)
	err := cmd.Run()

	return err == nil
}

func getMessage(res bool) string {
	if res {
		return "success"
	}
	return "fail"
}
