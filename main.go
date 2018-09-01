package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"

	"github.com/gobuffalo/packr"
	"github.com/gorilla/websocket"
	"github.com/zserge/webview"
)

var (
	upgrader = websocket.Upgrader{}
	ui       webview.WebView
	debug    = flag.Bool("debug", false, "Enable test mode")
)

func init() {
	flag.Parse()
}

type jsonMessage struct {
	Cmd string `json:"command"`
	Ok  bool   `json:"ok"`
}

type jsonError struct {
	Cmd string `json:"command"`
	Ok  bool   `json:"ok"`
}

type jsonCmd struct {
	Cmd string `json:"command"`
}

type jsonResponse interface {
	toJSON() ([]byte, error)
}

func (m jsonMessage) toJSON() ([]byte, error) {
	return json.Marshal(m)
}
func (e jsonError) toJSON() ([]byte, error) {
	return json.Marshal(e)
}

func main() {
	web := packr.NewBox("web/")
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:0", getLocalIP()))
	if err != nil {
		log.Fatal(err)
	}
	addr := ln.Addr().String()
	defer ln.Close()

	go func() {
		http.HandleFunc("/fu", ws)
		http.Handle("/", http.FileServer(web))
		log.Printf("Listening on http://%s/", addr)
		log.Fatal(http.Serve(ln, nil))
	}()

	ui = webview.New(webview.Settings{
		URL:    fmt.Sprintf("http://%s/", addr),
		Title:  "FuRemote",
		Width:  800,
		Height: 600,
		Debug:  true,
	})

	ui.Dispatch(func() {
		if *debug {
			ui.Eval("debugMe()")
		}
	})

	ui.Run()
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		ipnet, ok := address.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
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
		var req jsonCmd
		mt, request, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = json.Unmarshal(request, &req)
		if err != nil {
			break
		}

		run := func() jsonResponse {
			return jsonError{req.Cmd, false}
		}

		switch req.Cmd {
		case "vol+":
			run = vol("5%+", req.Cmd)
		case "vol-":
			run = vol("5%-", req.Cmd)
		case "volmute":
			run = vol("toggle", req.Cmd)
		}

		encoded, err := run().toJSON()
		if err != nil {
			break
		}

		err = c.WriteMessage(mt, encoded)
		if err != nil {
			break
		}
	}
}

func vol(val, command string) func() jsonResponse {
	return func() jsonResponse {
		cmd := exec.Command("amixer", "sset", "Master", val)
		err := cmd.Run()

		if err == nil {
			return jsonMessage{string(command), true}
		}
		return jsonError{string(command), false}
	}
}
