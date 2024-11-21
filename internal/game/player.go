package game

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/Monkhai/strixos-server.git/pkg/shared"
	"github.com/gorilla/websocket"
)

type Player struct {
	ID                string
	Conn              *websocket.Conn
	GameMessageChan   chan interface{}
	ServerMessageChan chan interface{}
	Ctx               context.Context
	Cancel            context.CancelFunc
	IsInGame          bool
	Mux               *sync.RWMutex
}

func NewPlayer(conn *websocket.Conn, ctx context.Context) *Player {
	id := GenerateUniqueID()
	ctx, cancel := context.WithCancel(ctx)
	return &Player{
		ID:                id,
		Conn:              conn,
		GameMessageChan:   make(chan interface{}, 10),
		ServerMessageChan: make(chan interface{}, 10),
		Ctx:               ctx,
		Cancel:            cancel,
		IsInGame:          false,
		Mux:               &sync.RWMutex{},
	}
}

func (p *Player) Listen() {
	defer close(p.GameMessageChan)

	for {
		_, msg, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Player %s disconnected gracefully\n", p.ID)
			} else {
				log.Printf("Unexpected error reading from player %s: %v\n", p.ID, err)
			}

			if p.IsInGame {
				p.GameMessageChan <- DisconnectedMessage{Player: p}
			} else {
				p.ServerMessageChan <- DisconnectedMessage{Player: p}
			}
			return
		}

		var baseMsg shared.BaseMessage
		if err := json.Unmarshal(msg, &baseMsg); err != nil {
			log.Printf("Invalid JSON message from player %s: %v\n", p.ID, err)
			continue
		}
		log.Println(p.ID, baseMsg.Type)

		switch baseMsg.Type {
		case shared.MoveMessageType:
			{
				var moveMsg shared.MoveMessage
				if err := json.Unmarshal(msg, &moveMsg); err != nil {
					log.Printf("Invalid JSON message from player %s: %v\n", p.ID, err)
					continue
				}
				p.GameMessageChan <- moveMsg
			}

		case shared.CloseMessageType:
			{
				var closeMsg shared.CloseMessage
				if err := json.Unmarshal(msg, &closeMsg); err != nil {
					log.Printf("Invalid JSON message from player %s: %v\n", p.ID, err)
					continue
				}
				closeGameMessage := shared.CloseMessage{
					BaseMessage: shared.BaseMessage{Type: "gameClosed"},
					Reason:      closeMsg.Reason,
				}
				p.GameMessageChan <- closeGameMessage
			}

		case shared.RequestGameMessageType:
			{
				log.Printf("Player %s requested a game\n", p.ID)
				p.ServerMessageChan <- shared.RequestGameMessage
			}

		case shared.LeaveGameMessageType:
			{
				var leaveGameMessage shared.BaseMessage
				if err := json.Unmarshal(msg, &leaveGameMessage); err != nil {
					log.Printf("Invalid JSON message from player %s: %v\n", p.ID, err)
					continue
				}
				p.GameMessageChan <- leaveGameMessage
			}

		case shared.LeaveQueueMessageType:
			var leaveQueueMessage shared.BaseMessage
			if err := json.Unmarshal(msg, &leaveQueueMessage); err != nil {
				log.Printf("Invalid JSON message from player %s: %v\n", p.ID, err)
				continue
			}
			{
				p.ServerMessageChan <- leaveQueueMessage
			}

		default:
			log.Printf("Unknown message type: %s\n", baseMsg.Type)
			p.GameMessageChan <- shared.UnknownMessage
		}
	}
}

func (p *Player) WriteMessage(message interface{}) error {
	err := p.Conn.WriteJSON(message)
	if err != nil {
		log.Printf("error sending message to player %v\n", err)
		return err
	}
	return nil
}

func GenerateUniqueID() string {
	length := 20
	buffer := make([]byte, length*2)
	_, err := rand.Read(buffer)
	if err != nil {
		log.Fatalf("failed to generate secure random bytes: %v", err)

	}
	encoded := base64.RawURLEncoding.EncodeToString(buffer)
	return strings.ReplaceAll(encoded[:length], "-", "")
}

func (p *Player) SetIsInGame(val bool) {
	p.Mux.Lock()
	defer p.Mux.Unlock()
	p.IsInGame = val
}
