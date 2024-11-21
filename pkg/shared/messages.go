package shared

type MessageType string

type BaseMessage struct {
	Type MessageType `json:"type"`
}

type GenericMessage struct {
	Type    MessageType    `json:"type"`
	Content map[string]any `json:"content"`
}

var UnknownMessage = BaseMessage{
	Type: UnknownMessageType,
}
