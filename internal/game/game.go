package game

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Monkhai/strixos-server.git/pkg/shared"
	"github.com/Monkhai/strixos-server.git/pkg/utils"
)

type Game struct {
	ID      string
	Board   *Board
	Player1 *Player
	Player2 *Player
	MsgChan chan interface{}
	Ctx     context.Context
	Cancel  context.CancelFunc
	Mux     *sync.RWMutex
}

func NewGame(players [2]*Player, parentCtx context.Context) *Game {
	ctx, cancel := context.WithCancel(parentCtx)
	players[0].SetIsInGame(true)
	players[1].SetIsInGame(true)
	return &Game{
		Mux:     &sync.RWMutex{},
		Board:   NewBoard(),
		Player1: players[0],
		Player2: players[1],
		MsgChan: make(chan interface{}, 10),
		Ctx:     ctx,
		Cancel:  cancel,
		ID:      utils.GenerateUniqueID(),
	}
}

func (g *Game) GameLoop(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		g.Cancel()
		g.Player1.SetIsInGame(false)
		g.Player2.SetIsInGame(false)
	}()

	currentPlayer := g.Player1
	otherPlayer := g.Player2

	log.Printf("\nGame started between %s and %s\n\n", g.Player1.Identity.ID, g.Player2.Identity.ID)

	// start game for the Players and tell them who they are and who is the next player
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
							currentPlayer.WriteMessage(*GameOverMessage(g.Board, currentPlayer))
							otherPlayer.WriteMessage(*GameOverMessage(g.Board, currentPlayer))
							log.Printf("Game over. Player %s won.\n", currentPlayer.Identity.ID)
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
					fmt.Printf("Player %s disconnected. Ending game.\n", currentPlayer.Identity.ID)
					otherPlayer.Conn.WriteJSON(m)
					g.MsgChan <- DisconnectedMessage{Player: otherPlayer}
					return
				}
			case shared.MoveMessage:
				{
					fmt.Printf("Ignoring message from %s (not their turn): %v\n", otherPlayer.Identity.ID, m.Content)
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
