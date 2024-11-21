package game

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Monkhai/strixos-server.git/pkg/shared"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func TestPlayer_Listen(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}
		player := NewPlayer(conn, func(*Player) {})
		//TODO: add the functions. Check if these functions are the correct approach
		go player.Listen()

		// Allow some time for the client to connect and send messages
		time.Sleep(1 * time.Second)

		//test move message
		select {
		case receivedMoveMsg := <-player.GameMessageChan:
			if msg, ok := receivedMoveMsg.(shared.MoveMessage); !ok || msg.Content.Row != 0 || msg.Content.Col != 1 || msg.Content.Mark != "x" {
				t.Errorf("expected move message, got %v", receivedMoveMsg)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for move message")
		}

		//test wrong message
		select {
		case receivedUnknownMessage := <-player.GameMessageChan:
			if msg, ok := receivedUnknownMessage.(shared.BaseMessage); !ok || msg.Type != "unknownMessage" {
				t.Errorf("expected unknown message, got %v ok = %v", receivedUnknownMessage, ok)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for unknown message")
		}

		//test close message
		select {
		case receivedCloseMsg := <-player.GameMessageChan:
			if msg, ok := receivedCloseMsg.(shared.CloseMessage); !ok || msg.Reason != "disconnected" {
				t.Errorf("expected close message, got %v", receivedCloseMsg)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for close message")
		}
	}))
	defer server.Close()

	// Create a test client
	url := "ws" + server.URL[4:] // Change http to ws
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send a move message
	moveMsg := shared.MoveMessage{
		BaseMessage: shared.BaseMessage{Type: "move"},
		Content: struct {
			Row  int    `json:"row"`
			Col  int    `json:"col"`
			Mark string `json:"mark"`
		}{
			Row:  0,
			Col:  1,
			Mark: "x",
		},
	}
	if err := conn.WriteJSON(moveMsg); err != nil {
		t.Fatalf("Failed to send move message: %v", err)
	}

	randomMessage := map[string]any{
		"unexpected_field": "unexpected_value",
		"another_field":    12345,
	}
	if err := conn.WriteJSON(randomMessage); err != nil {
		t.Fatalf("Failed to send unknown message: %v", err)
	}

	// Send a close message
	closeMsg := shared.CloseMessage{
		BaseMessage: shared.BaseMessage{Type: "close"},
		Reason:      "disconnected",
	}
	if err := conn.WriteJSON(closeMsg); err != nil {
		t.Fatalf("Failed to send close message: %v", err)
	}

	// Allow some time for messages to be processed
	time.Sleep(3 * time.Second)
}

func TestPlayer_WriteMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}

		// Read the message from the connection
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		// Unmarshal the message to verify its content
		var receivedMsg shared.GenericMessage
		if err := json.Unmarshal(msg, &receivedMsg); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		// Check the message content
		expectedMsg := shared.GenericMessage{
			Type: "test",
			Content: map[string]interface{}{
				"message": "hello",
			},
		}
		if receivedMsg.Type != expectedMsg.Type || !equalMaps(receivedMsg.Content, expectedMsg.Content) {
			t.Errorf("expected message %v, got %v", expectedMsg, receivedMsg)
		}
	}))
	defer server.Close()

	url := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	player := NewPlayer(conn, func(*Player) {})

	msg := shared.GenericMessage{
		Type: "test",
		Content: map[string]interface{}{
			"message": "hello",
		},
	}

	if err := player.WriteMessage(msg); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	time.Sleep(1 * time.Second)
}

func equalMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || v != bv {
			return false
		}
	}
	return true
}
