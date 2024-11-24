package game

import (
	"github.com/Monkhai/strixos-server.git/internal/identity"
	"github.com/Monkhai/strixos-server.git/pkg/shared"
)

const (
	PlayerDisconnected shared.MessageType = "playerDisconnected"
)

type DisconnectedMessage struct {
	Player *Player
}

type LeaveGameMessage struct {
	RequestingPlayer *Player
	OtherPlayer      *Player
}

type LeaveQueueMessage struct {
	Player *Player
}

type UpdateIdentityMessage struct {
	Type    shared.MessageType `json:"type"`
	Content struct {
		Identity identity.Identity `json:"identity"`
	} `json:"content"`
}

var PlayerDisconnectedMessage = shared.GenericMessage{
	Type: PlayerDisconnected,
}

var LeaveGameMessageMessage = shared.GenericMessage{
	Type: shared.LeaveGameMessageType,
}
