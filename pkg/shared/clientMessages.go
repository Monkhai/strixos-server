package shared

import "github.com/Monkhai/strixos-server.git/internal/identity"

const (
	MoveMessageType             MessageType = "move"
	CloseMessageType            MessageType = "close"
	RequestGameMessageType      MessageType = "gameRequest"
	LeaveGameMessageType        MessageType = "leaveGame"
	LeaveQueueMessageType       MessageType = "leaveQueue"
	IdentityUpdateMessageType   MessageType = "updateIdentity"
	JoinInviteGameMessageType   MessageType = "joinInviteGame"
	CreateInviteGameMessageType MessageType = "createInviteGame"
	LeaveInviteGameMessageType  MessageType = "leaveInviteGame"
	UnknownMessageType          MessageType = "unknownMessage"
)

type BaseClientMessage struct {
	Type     MessageType       `json:"type"`
	Identity identity.Identity `json:"identity"`
}

type JoinInviteGameMessage struct {
	BaseClientMessage
	GameID string `json:"gameID"`
}

type LeaveInviteGameMessage struct {
	BaseClientMessage
	GameID string `json:"gameID"`
}

type MoveMessage struct {
	BaseClientMessage
	Content struct {
		Row  int    `json:"row"`
		Col  int    `json:"col"`
		Mark string `json:"mark"`
	} `json:"content"`
}

type CloseMessage struct {
	BaseClientMessage
	Identity identity.Identity `json:"identity"`
	Reason   string            `json:"reason"`
}

func RequestGameMessage(identity identity.Identity) BaseClientMessage {
	return BaseClientMessage{
		Type:     RequestGameMessageType,
		Identity: identity,
	}
}
