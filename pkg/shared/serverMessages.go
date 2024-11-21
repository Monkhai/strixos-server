package shared

const (
	StartGameMessageType   MessageType = "startGame"
	UpdateGameMessageType  MessageType = "update"
	GameOverMessageType    MessageType = "gameOver"
	GameClosedMessageType  MessageType = "gameClosed"
	ErrorMessageType       MessageType = "error"
	GameWaitingMessageType MessageType = "gameWaiting"
	RemovedFromQueueType   MessageType = "removedFromQueue"
	RemovedFromGameType    MessageType = "removedFromGame"
	OpponentDisconnected   MessageType = "opponentDisconnected"
)

var OpponentDisconnectedMessage = GenericMessage{
	Type: OpponentDisconnected,
}

var RemovedFromQueueMessage = GenericMessage{
	Type: RemovedFromQueueType,
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
		Type: RemovedFromGameType,
	}
}
