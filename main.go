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

var connections = []WebSocketConnection{}

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type WebSocketConnection struct {
	Conn *websocket.Conn
}

func (w WebSocketConnection) ReadMessages(msgChan chan Message, closeChan chan struct{}) {
	defer w.Conn.Close()
	for {
		_, msgBtyes, err := w.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Println("User closed the connetion with code 1000 (normal)")
			} else if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				log.Println("User closed the connection with code 1006 (they just left bro)")
			}
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
	MsgChan   chan Message
	CloseChan chan struct{}
	Players   []WebSocketConnection
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
	connection := WebSocketConnection{Conn: conn}
	s.Players = append(s.Players, connection)
	go connection.ReadMessages(s.MsgChan, s.CloseChan)
}

func (s *Server) ReadChannels() {
	for {
		select {
		case message := <-s.MsgChan:
			log.Println("message:", message)

		case <-s.CloseChan:
			log.Println("received closing signal")
			close(s.MsgChan)
			return
		}
	}
}

func main() {
	msgChan := make(chan Message)
	closeChan := make(chan struct{})
	defer close(closeChan)

	server := Server{MsgChan: msgChan, CloseChan: closeChan}
	go server.ReadChannels()
	http.HandleFunc("/ws", server.handler)
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	<-closeChan

	close(msgChan)
	log.Println("Server shutting down, closed message channel")
}
