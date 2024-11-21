package game

import "github.com/Monkhai/strixos-server.git/pkg/shared"

func (g *Game) NewGameMessage(mark, activePlayer, playerId string) shared.GenericMessage {
	return shared.GenericMessage{
		Type: shared.StartGameMessageType,
		Content: map[string]any{
			"board":        g.Board.Cells,
			"mark":         mark,
			"activePlayer": activePlayer,
			"yourId":       playerId,
		},
	}
}

func (g *Game) GameUpdateMessage(activePlayer string) shared.GenericMessage {
	return shared.GenericMessage{
		Type: shared.UpdateGameMessageType,
		Content: map[string]any{
			"board":        g.Board.Cells,
			"activePlayer": activePlayer,
		},
	}
}
func GameOverMessage(board *Board, winner string) *shared.GenericMessage {
	return &shared.GenericMessage{
		Type: shared.GameOverMessageType,
		Content: map[string]any{
			"board":  board.Cells,
			"winner": winner,
		},
	}
}
