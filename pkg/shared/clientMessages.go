package shared

import "github.com/Monkhai/strixos-server.git/internal/identity"

const (
	MoveMessageType        MessageType = "move"
	CloseMessageType       MessageType = "close"          // why do we have this>???
	RequestGameMessageType MessageType = "gameRequest"    // base message type
	LeaveGameMessageType   MessageType = "leaveGame"      // base message type
	LeaveQueueMessageType  MessageType = "leaveQueue"     // base message type
	UnknownMessageType     MessageType = "unknownMessage" // unknown
	UpdateIdentityType     MessageType = "updateIdentity" // unknown
)

type BaseClientMessage struct {
	Type     MessageType       `json:"type"`
	Identity identity.Identity `json:"identity"`
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
