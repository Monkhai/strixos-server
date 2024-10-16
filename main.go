package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// var connections = []WebSocketConnection{}

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type WebSocketConnection struct {
	Conn *websocket.Conn
}

func (w WebSocketConnection) ReadConnectionMessage(msgChan chan Message) {
	defer w.Conn.Close()
	for {
		_, msgBtyes, err := w.Conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		err = json.Unmarshal(msgBtyes, &msg)
		if err != nil {
			break
		}
		msgChan <- msg
	}
}

type Server struct {
	MsgChan chan Message
}

func (s *Server) handleUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Unauthorized"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)

}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token != "1234" {
		s.handleUnauthorized(w)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	connection := WebSocketConnection{Conn: conn}
	connection.ReadConnectionMessage(s.MsgChan)
}

func readMessages(msgChan chan Message) {
	for message := range msgChan {
		log.Println(message)
	}
}

func main() {
	msgChan := make(chan Message)
	go readMessages(msgChan)
	server := Server{MsgChan: msgChan}
	http.HandleFunc("/ws", server.handler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
