package kawka

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestWsClient(t *testing.T) {
	const prepared = ""

	handler := func(data []byte) error {
		if data == nil {
			return nil
		}
		return nil
	}

	var conn *websocket.Conn

	client := newWsClient(conn, nil, handler)
	if client == nil {
		t.Fatal("client is nil")
	}

	conn.WriteJSON(prepared)
}
