package kawka

import "sync"

type wsHub struct {
	sync.RWMutex
	clients map[*wsClient]struct{}
}

func newHub() *wsHub {
	h := &wsHub{
		clients: make(map[*wsClient]struct{}),
	}
	return h
}

func (h *wsHub) Connect(client *wsClient) {
	h.Lock()
	h.clients[client] = struct{}{}
	h.Unlock()
}

func (h *wsHub) Disconnect(client *wsClient) {
	h.Lock()
	delete(h.clients, client)
	h.Unlock()
}
