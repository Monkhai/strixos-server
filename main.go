package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

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

func (w WebSocketConnection) ReadMessages(msgChan chan Message, closeChan chan struct{}, wg *sync.WaitGroup) {
	defer w.Conn.Close()
	defer wg.Done()
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
			newMessage, _ := json.Marshal(map[string]string{"error": "bad message"})
			w.Conn.WriteMessage(1, newMessage)
		} else {
			msgChan <- msg
		}
	}
}

type Server struct {
	MsgChan   chan Message
	CloseChan chan struct{}
	Players   []WebSocketConnection
	Wg        *sync.WaitGroup
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

	s.Wg.Add(1)
	go connection.ReadMessages(s.MsgChan, s.CloseChan, s.Wg)
}

func (s *Server) ReadChannels() {
	for {
		select {
		case message := <-s.MsgChan:
			log.Println("message:", message)

		case <-s.CloseChan:
			log.Println("received closing signal")
			return
		}
	}
}

func main() {
	msgChan := make(chan Message)
	closeChan := make(chan struct{})
	wg := &sync.WaitGroup{}

	server := Server{MsgChan: msgChan, CloseChan: closeChan, Wg: wg, Players: []WebSocketConnection{}}
	go server.ReadChannels()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.handler)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shut down: %v", err)
	}

	close(closeChan)
	for _, conn := range server.Players {
		log.Println("Forcefully closing WebSocket connection")
		conn.Conn.Close()
	}
	wg.Wait()
	close(msgChan)

	log.Println("Server stopped nicely")
}
