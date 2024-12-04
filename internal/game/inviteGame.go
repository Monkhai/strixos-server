package game

import (
	"context"
	"log"
	"sync"

	"github.com/Monkhai/strixos-server.git/pkg/shared"
	"github.com/Monkhai/strixos-server.git/pkg/utils"
)

type InviteGameLoopOverMessage struct {
	GameID  string
	Board   *Board
	Winner  *Player
	Players [2]*Player
}

func NewEmptyInviteGame(parentCtx context.Context) *Game {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Game{
		Board:   NewBoard(),
		MsgChan: make(chan interface{}, 10),
		Ctx:     ctx,
		Cancel:  cancel,
		ID:      utils.GenerateUniqueID(),
		Mux:     &sync.RWMutex{},
	}

}

func NewInviteGame(player *Player, parentCtx context.Context) *Game {
	ctx, cancel := context.WithCancel(parentCtx)
	player.SetIsInGame(true)
	return &Game{
		Board:   NewBoard(),
		Player1: player,
		MsgChan: make(chan interface{}, 10),
		Ctx:     ctx,
		Cancel:  cancel,
		ID:      utils.GenerateUniqueID(),
		Mux:     &sync.RWMutex{},
	}
}

func (g *Game) AddFirstPlayer(p *Player) {
	g.Mux.Lock()
	g.Player1 = p
	g.Mux.Unlock()
	p.SetIsInGame(true)
}

func (g *Game) AddSecondPlayer(p *Player) {
	g.Mux.Lock()
	g.Player2 = p
	g.Mux.Unlock()
	p.SetIsInGame(true)
}

func (g *Game) InviteGameLoop(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		g.Cancel()
		g.Player1.SetIsInGame(false)
		g.Player2.SetIsInGame(false)
	}()

	currentPlayer := g.Player1
	otherPlayer := g.Player2

	log.Printf("\nGame started between %s and %s\n\n", g.Player1.Identity.ID, g.Player2.Identity.ID)

	currentPlayerStartGameMsg := g.NewGameMessage("x", currentPlayer, otherPlayer)
	currentPlayer.WriteMessage(currentPlayerStartGameMsg)
	otherPlayerStartGameMsg := g.NewGameMessage("o", currentPlayer, currentPlayer)
	otherPlayer.WriteMessage(otherPlayerStartGameMsg)

	for {
		select {
		case <-g.Ctx.Done():
			{
				log.Printf("Game between %s and %s ended\n", g.Player1.Identity.ID, g.Player2.Identity.ID)
				return
			}

		case msg := <-currentPlayer.GameMessageChan:
			{
				switch m := msg.(type) {
				case DisconnectedMessage:
					{
						log.Printf("Player %s disconnected. Ending game.\n", currentPlayer.Identity.ID)
						otherPlayer.Conn.WriteJSON(m)
						g.MsgChan <- DisconnectedMessage{Player: currentPlayer}
						return
					}

				case shared.MoveMessage:
					{
						log.Println("Move message", "Row:", m.Content.Row, "Col:", m.Content.Col)
						err := g.Board.SetCell(m.Content.Row, m.Content.Col, m.Content.Mark)
						if err != nil {
							log.Println("Error getting cell", err)
							currentPlayer.WriteMessage(shared.GenericMessage{
								Type: shared.ErrorMessageType,
								Content: map[string]any{
									"message": err.Error(),
								},
							})
							continue
						}

						g.Board.UpdateLives()
						if g.Board.CheckWin() {
							var inviteGameLoopOverMessage InviteGameLoopOverMessage = InviteGameLoopOverMessage{
								GameID:  g.ID,
								Board:   g.Board,
								Winner:  currentPlayer,
								Players: [2]*Player{currentPlayer, otherPlayer},
							}
							g.MsgChan <- inviteGameLoopOverMessage
							return
						}

						currentPlayer, otherPlayer = otherPlayer, currentPlayer
						updateMsg := g.GameUpdateMessage(currentPlayer)
						currentPlayer.WriteMessage(updateMsg)
						otherPlayer.WriteMessage(updateMsg)
						break
					}

				case shared.BaseClientMessage:
					{
						switch m.Type {
						case shared.LeaveGameMessageType:
							{
								log.Printf("Player %s left the game. Ending game.\n", currentPlayer.Identity.ID)
								g.MsgChan <- LeaveGameMessage{RequestingPlayer: currentPlayer, OtherPlayer: otherPlayer}
								return
							}
						case shared.LeaveQueueMessageType:
							{
								log.Printf("Player asked to leave game queue inside game. Ignoring. %v\n", m)
							}
						default:
							{
								log.Printf("Unknown message type: %s\n", m.Type)
							}

						}
					}

				default:
					{
						log.Println("Unkown message received", m)
					}
				}
			}

		case msg := <-otherPlayer.GameMessageChan:
			switch m := msg.(type) {
			case shared.CloseMessage:
				{
					log.Printf("Player %s disconnected. Ending game.\n", currentPlayer.Identity.ID)
					otherPlayer.Conn.WriteJSON(m)
					g.MsgChan <- DisconnectedMessage{Player: otherPlayer}
					return
				}
			case shared.MoveMessage:
				{
					log.Printf("Ignoring message from %s (not their turn): %v\n", otherPlayer.Identity.ID, m.Content)
				}
			case shared.BaseClientMessage:
				{
					switch m.Type {
					case shared.LeaveGameMessageType:
						{
							log.Printf("Player %s left the game. Ending game.\n", otherPlayer.Identity.ID)
							g.MsgChan <- LeaveGameMessage{RequestingPlayer: otherPlayer, OtherPlayer: currentPlayer}
							return
						}
					case shared.LeaveQueueMessageType:
						{
							log.Printf("Player asked to leave game queue inside game. Ignoring. %v\n", m)
						}
					default:
						{
							log.Printf("Unknown message type: %s\n", m.Type)
						}

					}
				}
			}
		}

	}
}
