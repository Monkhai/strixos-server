package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Monkhai/strixos-server.git/internal/game"
	"github.com/Monkhai/strixos-server.git/internal/identity"
	"github.com/Monkhai/strixos-server.git/pkg/shared"
	"github.com/gorilla/websocket"
)

type Server struct {
	Queue             *PlayerQueue
	Mux               *sync.RWMutex
	Ctx               *context.Context
	Wg                *sync.WaitGroup
	IdentityManager   *identity.IdentityManager
	InviteGameManager *game.InviteGameManager
}

func NewServer(ctx *context.Context, wg *sync.WaitGroup) *Server {
	return &Server{
		Ctx:               ctx,
		Wg:                wg,
		Queue:             NewPlayerQueue(),
		Mux:               &sync.RWMutex{},
		IdentityManager:   identity.NewIdentityManager(),
		InviteGameManager: game.NewInviteGameManager(),
	}
}

func (s *Server) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error creating the ws connection: %s", err)
	}
	s.AddPlayer(conn, s.Wg)
}

func (s *Server) AddPlayer(conn *websocket.Conn, wg *sync.WaitGroup) {
	i := s.IdentityManager.RegisterIdentity()
	p := game.NewPlayer(i, conn, *s.Ctx)
	log.Printf("New connection with player %s\n", p.Identity.ID)

	p.WriteMessage(shared.InitialIdentityMessage(identity.InitialIdentity{
		ID:     i.ID,
		Secret: i.Secret,
	}))
	log.Println("Identity sent to player", p.Identity.ID)

	//read auth message
	//TODO: wrap in a for loop in case the message is wrong
	_, msg, err := conn.ReadMessage()
	if err != nil {
		defer s.IdentityManager.IdentitiesMap.RemoveIdentity(p.Identity.ID)
		//error is 1000
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("Connection disconnected before finished auth")
			return
		}
		log.Printf("error reading message from player %s: %s", p.Identity.ID, err)
		return
	}
	var m game.UpdateIdentityMessage
	err = json.Unmarshal(msg, &m)
	if err != nil {
		log.Printf("error unmarshalling message from player %s: %s", p.Identity.ID, err)
		return
	}

	if m.Type != shared.IdentityUpdateMessageType {
		log.Printf("Player %s sent an unkown message type: %s\n", p.Identity.ID, m.Type)
		return
	}

	valid := s.IdentityManager.IdentitiesMap.UpdateIdentity(m.Content.Identity)
	if !valid {
		log.Printf("Player %s failed to update during initial auth\n", p.Identity.ID)
		return
	}

	p.UpdateIdentity(m.Content.Identity)
	p.WriteMessage(shared.RegistedMesage(p.Identity))

	log.Println("Identity updated for player", p.Identity.ID)

	wg.Add(2)
	go s.ListenToPlayerMessages(p, wg)
	go p.Listen(wg, s.IdentityManager.IdentitiesMap.ValidateIdentity)
}

func (s *Server) HandleRequestGame(p *game.Player) {
	log.Printf("Player %s disconnected\n", p.Identity.ID)
	p.WriteMessage(shared.GameWaitingMessage())
	s.Queue.Enqueue(p)
}

func (s *Server) HandleLeaveQueueRequest(p *game.Player) {
	log.Printf("Player %s left the queue\n", p.Identity.ID)
	p.WriteMessage(shared.RemovedFromQueueMessage)
	s.Queue.RemovePlayer(p)
}

func (s *Server) HandleLeaveGameRequest(requester, otherPlayer *game.Player) {
	log.Printf("Player %s left the game\n", requester.Identity.ID)
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
			{
				log.Println("QUEUE LOOP DONE")
				return
			}
		default:
			{
				players, hasPlayers := s.Queue.GetTwoPlayers()
				if !hasPlayers {
					// log.Println("Not enough players to start a game. Waiting...")
				} else {
					log.Println("Starting a game between", players[0].Identity.ID, "and", players[1].Identity.ID)
					game := game.NewGame(players, ctx)
					wg.Add(2)
					go game.GameLoop(wg)
					go s.ListenToGameMessages(game, wg)
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (s *Server) ListenToGameMessages(g *game.Game, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-g.Ctx.Done():
			{
				log.Printf("Game between %s and %s ended\n", g.Player1.Identity.ID, g.Player2.Identity.ID)
				return
			}
		case msg := <-g.MsgChan:
			{
				switch m := msg.(type) {
				case game.LeaveGameMessage:
					{
						log.Printf("Player %s left the game\n", m.RequestingPlayer.Identity.ID)
						s.HandleLeaveGameRequest(m.RequestingPlayer, m.OtherPlayer)
					}
				case game.DisconnectedMessage:
					{
						log.Printf("Player %s disconnected. Ending game.\n", m.Player.Identity.ID)
						var otherPlayer *game.Player
						if m.Player.Identity.ID == g.Player1.Identity.ID {
							otherPlayer = g.Player2
						} else {
							otherPlayer = g.Player1
						}
						s.HandleLeaveGameRequest(m.Player, otherPlayer)
					}
				case game.InviteGameLoopOverMessage:
					{
						log.Printf("Invite Game between %s and %s ended\n", g.Player1.Identity.ID, g.Player2.Identity.ID)
						s.InviteGameManager.RemoveGame(m.GameID)
						g := game.NewEmptyInviteGame(*s.Ctx)
						s.InviteGameManager.AddGame(g)
						for _, p := range m.Players {
							p.WriteMessage(game.InviteGameOverMessage(m.Board, m.Winner, g.ID))
						}
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
				log.Printf("Player %s context done\n", p.Identity.ID)
				return
			}
		case msg := <-p.ServerMessageChan:
			{
				switch m := msg.(type) {
				case game.DisconnectedMessage:
					{
						log.Printf("Player %s disconnected\n", p.Identity.ID)
						s.Queue.RemovePlayer(p)
						s.IdentityManager.IdentitiesMap.RemoveIdentity(p.Identity.ID)
					}

				case game.UpdateIdentityMessage:
					{
						valid := s.IdentityManager.IdentitiesMap.UpdateIdentity(m.Content.Identity)
						if !valid {
							log.Printf("Player %s tried to update an identity that does not exist\n", p.Identity.ID)
							log.Printf("message content: ID: %s Secret:%s\n", m.Content.Identity.ID, m.Content.Identity.Secret)
						}
					}

				case shared.JoinInviteGameMessage:
					{
						log.Printf("Player %s asked to join a game with id %s\n", p.Identity.ID, msg.(shared.JoinInviteGameMessage).GameID)
						typedMsg, valid := msg.(shared.JoinInviteGameMessage)
						if !valid {
							log.Printf("Player %s sent a message of type %s but it is not a JoinInviteGameMessage\n", p.Identity.ID, m.Type)
						}
						game, found := s.InviteGameManager.GetGame(typedMsg.GameID)
						if !found {
							log.Printf("Player %s asked to join a game with id %s but the game was not found\n", p.Identity.ID, typedMsg.GameID)
						}

						if game.Player1 != nil && game.Player2 != nil {
							log.Printf("Player %s asked to join a game with id %s but the game is full\n", p.Identity.ID, typedMsg.GameID)
							return
						}

						if game.Player1 != nil {
							game.AddSecondPlayer(p)
							wg.Add(2)
							go game.InviteGameLoop(wg)
							go s.ListenToGameMessages(game, wg)
						} else {
							game.AddFirstPlayer(p)
						}
					}
				case shared.LeaveInviteGameMessage:
					{
						log.Printf("Player %s asked to leave a game with id %s\n", p.Identity.ID, msg.(shared.LeaveInviteGameMessage).GameID)
						typedMsg, valid := msg.(shared.LeaveInviteGameMessage)
						if !valid {
							log.Printf("Player %s sent a message of type %s but it is not a LeaveInviteGameMessage\n", p.Identity.ID, m.Type)
						}
						s.InviteGameManager.RemoveGame(typedMsg.GameID)
					}
				case shared.BaseClientMessage:
					{
						log.Printf("Player %s sent a message of type %s\n", p.Identity.ID, m.Type)
						switch m.Type {
						case shared.LeaveQueueMessageType:
							{
								s.HandleLeaveQueueRequest(p)
							}
						case shared.LeaveGameMessageType:
							{
								log.Printf("Player %s asked to leave the game but he is not in a game right now!", p.Identity.ID)
							}
						case shared.RequestGameMessageType:
							{
								s.HandleRequestGame(p)
							}
						case shared.CreateInviteGameMessageType:
							{
								game := game.NewInviteGame(p, *s.Ctx)
								s.InviteGameManager.AddGame(game)
								log.Printf("Player %s created a game with id %s\n", p.Identity.ID, game.ID)
								p.WriteMessage(shared.InviteGameCreatedMessage(game.ID))
							}
						}
					}

				default:
					{
						log.Println("Unkown message received", m)
					}

				}
			}
		}
	}
}
