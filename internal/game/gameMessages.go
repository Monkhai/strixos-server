package game

import (
	"github.com/Monkhai/strixos-server.git/pkg/shared"
)

func (g *Game) NewGameMessage(mark string, activePlayer, opponent *Player) shared.GenericMessage {
	return shared.GenericMessage{
		Type: shared.StartGameMessageType,
		Content: map[string]any{
			"board":        g.Board.Cells,
			"mark":         mark,
			"activePlayer": activePlayer.Identity.GetSafeIdentity(),
			"opponent":     opponent.Identity.GetSafeIdentity(),
			"gameID":       g.ID,
		},
	}
}

func (g *Game) GameUpdateMessage(activePlayer *Player) shared.GenericMessage {
	return shared.GenericMessage{
		Type: shared.UpdateGameMessageType,
		Content: map[string]any{
			"board":        g.Board.Cells,
			"activePlayer": activePlayer.Identity.GetSafeIdentity(),
		},
	}
}

func GameOverMessage(board *Board, winner *Player) *shared.GenericMessage {
	return &shared.GenericMessage{
		Type: shared.GameOverMessageType,
		Content: map[string]any{
			"board":  board.Cells,
			"winner": winner.Identity.GetSafeIdentity(),
		},
	}
}

func InviteGameOverMessage(board *Board, winner *Player, newGameID string) *shared.GenericMessage {
	return &shared.GenericMessage{
		Type: shared.InviteGameOverMessageType,
		Content: map[string]any{
			"board":     board.Cells,
			"winner":    winner.Identity.GetSafeIdentity(),
			"newGameID": newGameID,
		},
	}
}
