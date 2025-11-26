package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var liveClients = make(map[*websocket.Conn]bool)
var liveMu sync.Mutex

var liveUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func liveWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := liveUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	liveMu.Lock()
	liveClients[conn] = true
	liveMu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			liveMu.Lock()
			delete(liveClients, conn)
			liveMu.Unlock()
			conn.Close()
			break
		}
	}
}

func broadcastTelemetry(e Event) {
	liveMu.Lock()
	defer liveMu.Unlock()

	for conn := range liveClients {
		conn.WriteJSON(e)
	}
}
