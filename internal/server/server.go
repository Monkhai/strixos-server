package server

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Monkhai/strixos-server.git/internal/game"
	"github.com/Monkhai/strixos-server.git/pkg/shared"
	"github.com/gorilla/websocket"
)

type Server struct {
	Queue *PlayerQueue
	Mux   *sync.RWMutex
	Ctx   *context.Context
	Wg    *sync.WaitGroup
}

func NewServer(ctx *context.Context, wg *sync.WaitGroup) *Server {
	q := NewPlayerQueue()
	mux := &sync.RWMutex{}
	return &Server{
		Queue: q,
		Mux:   mux,
		Ctx:   ctx,
		Wg:    wg,
	}
}

func (s *Server) WsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error creating the ws connection: %s", err)
	}

	player := game.NewPlayer(conn, *s.Ctx)
	s.AddPlayer(player, s.Wg)
}

func (s *Server) AddPlayer(p *game.Player, wg *sync.WaitGroup) {
	log.Printf("New connection with player %s\n", p.ID)
	wg.Add(1)
	go s.ListenToPlayerMessages(p, wg)
	go p.Listen()
}

func (s *Server) HandleRequestGame(p *game.Player) {
	log.Printf("Player %s disconnected\n", p.ID)
	p.WriteMessage(shared.GameWaitingMessage())
	s.Queue.Enqueue(p)
}

func (s *Server) HandleLeaveQueueRequest(p *game.Player) {
	log.Printf("Player %s left the queue\n", p.ID)
	p.WriteMessage(shared.RemovedFromQueueMessage)
	s.Queue.RemovePlayer(p)
}

func (s *Server) HandleLeaveGameRequest(requester, otherPlayer *game.Player) {
	log.Printf("Player %s left the game\n", requester.ID)
	requester.WriteMessage(shared.RemovedFromGameMessage())
	otherPlayer.WriteMessage(shared.GameClosedMessage())
	requester.SetIsInGame(false)
	otherPlayer.SetIsInGame(false)
}

func (s *Server) QueueLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("QUEUE LOOP STARTING")

	for {
		select {
		case <-ctx.Done():
			log.Println("QUEUE LOOP DONE")
			return
		default:
			s.Queue.printQueue()
			players, hasPlayers := s.Queue.GetTwoPlayers()
			if !hasPlayers {
				log.Println("Not enough players to start a game. Waiting...")
			} else {
				game := game.NewGame(players)
				go game.GameLoop()
				wg.Add(1)
				go s.ListenToGameMessages(game, wg)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// listen to messages so we can shut down games when needed or in the future do more stuff
func (s *Server) ListenToGameMessages(g *game.Game, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-g.Ctx.Done():
			{
				log.Printf("Game between %s and %s ended\n", g.Player1.ID, g.Player2.ID)
				return
			}
		case msg := <-g.MsgChan:
			{
				switch m := msg.(type) {
				case game.LeaveGameMessage:
					{
						log.Printf("Player %s left the game\n", m.RequestingPlayer.ID)
						s.HandleLeaveGameRequest(m.RequestingPlayer, m.OtherPlayer)
					}
				case game.DisconnectedMessage:
					{
						log.Printf("Player %s disconnected. Ending game.\n", m.Player.ID)
						var otherPlayer *game.Player
						if m.Player.ID == g.Player1.ID {
							otherPlayer = g.Player2
						} else {
							otherPlayer = g.Player1
						}
						s.HandleLeaveGameRequest(m.Player, otherPlayer)
					}
				}
			}
		}
	}
}

func (s *Server) ListenToPlayerMessages(p *game.Player, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-p.Ctx.Done():
			{
				log.Printf("Player %s context done\n", p.ID)
				return
			}
		case msg := <-p.ServerMessageChan:
			{
				switch m := msg.(type) {
				case game.DisconnectedMessage:
					{
						log.Printf("Player %s disconnected\n", p.ID)
						s.Queue.RemovePlayer(p)
					}

				case shared.BaseMessage:
					{
						log.Printf("Player %s sent a message of type %s\n", p.ID, m.Type)
						switch m.Type {
						case shared.LeaveQueueMessageType:
							{
								s.HandleLeaveQueueRequest(p)
							}
						case shared.LeaveGameMessageType:
							{
								log.Printf("Player %s asked to leave the game but he is not in a game right now!", p.ID)
							}
						case shared.RequestGameMessageType:
							{
								s.HandleRequestGame(p)
							}
						}
					}

				}
			}
		}
	}
}
