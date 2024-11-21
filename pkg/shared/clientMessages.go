package shared

const (
	MoveMessageType        MessageType = "move"
	CloseMessageType       MessageType = "close"          // why do we have this>???
	RequestGameMessageType MessageType = "gameRequest"    // base message type
	LeaveGameMessageType   MessageType = "leaveGame"      // base message type
	LeaveQueueMessageType  MessageType = "leaveQueue"     // base message type
	UnknownMessageType     MessageType = "unknownMessage" // unknown
)

type MoveMessage struct {
	BaseMessage //expected to be "move"
	Content     struct {
		Row  int    `json:"row"`
		Col  int    `json:"col"`
		Mark string `json:"mark"`
	} `json:"content"`
}

type CloseMessage struct {
	BaseMessage        //exptected to be "close"
	Reason      string `json:"reason"`
}

var RequestGameMessage = BaseMessage{
	Type: RequestGameMessageType,
}
