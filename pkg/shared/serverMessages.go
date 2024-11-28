package shared

import "github.com/Monkhai/strixos-server.git/internal/identity"

const (
	StartGameMessageType              MessageType = "startGame"
	UpdateGameMessageType             MessageType = "update"
	GameOverMessageType               MessageType = "gameOver"
	GameClosedMessageType             MessageType = "gameClosed"
	ErrorMessageType                  MessageType = "error"
	GameWaitingMessageType            MessageType = "gameWaiting"
	RemovedFromQueueMessageType       MessageType = "removedFromQueue"
	RemovedFromGameMessageType        MessageType = "removedFromGame"
	OpponentDisconnectedMessageType   MessageType = "opponentDisconnected"
	DisconnectedFromServerMessageType MessageType = "disconnectedFromServer"
	//register flow
	AuthIdentityMessageType MessageType = "authIdentity"
	RegisteredMessageType   MessageType = "registered"
	//invite game flow
	InviteGameCreatedMessageType MessageType = "inviteGameCreated"
)

var DisconnectedFromServerMessage = GenericMessage{
	Type: DisconnectedFromServerMessageType,
}

var OpponentDisconnectedMessage = GenericMessage{
	Type: OpponentDisconnectedMessageType,
}

var RemovedFromQueueMessage = GenericMessage{
	Type: RemovedFromQueueMessageType,
}

func InitialIdentityMessage(identity identity.InitialIdentity) GenericMessage {
	return GenericMessage{
		Type: AuthIdentityMessageType,
		Content: map[string]any{
			"identity": identity,
		},
	}
}

func InviteGameCreatedMessage(gameID string) GenericMessage {
	return GenericMessage{
		Type: InviteGameCreatedMessageType,
		Content: map[string]any{
			"gameID": gameID,
		},
	}
}

func RegistedMesage(identity *identity.Identity) GenericMessage {
	return GenericMessage{
		Type: RegisteredMessageType,
		Content: map[string]any{
			"identity": identity,
		},
	}
}

func ErrorMessage(message string) GenericMessage {
	return GenericMessage{
		Type: ErrorMessageType,
		Content: map[string]any{
			"message": message,
		},
	}
}

func GameWaitingMessage() GenericMessage {
	return GenericMessage{
		Type: GameWaitingMessageType,
	}
}

func GameClosedMessage() GenericMessage {
	return GenericMessage{
		Type: GameClosedMessageType,
	}
}

func RemovedFromGameMessage() GenericMessage {
	return GenericMessage{
		Type: RemovedFromGameMessageType,
	}
}
