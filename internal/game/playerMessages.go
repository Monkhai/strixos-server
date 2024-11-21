package game

import "github.com/Monkhai/strixos-server.git/pkg/shared"

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

var PlayerDisconnectedMessage = shared.GenericMessage{
	Type: PlayerDisconnected,
}

var LeaveGameMessageMessage = shared.GenericMessage{
	Type: shared.LeaveGameMessageType,
}
