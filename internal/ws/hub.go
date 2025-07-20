package ws

import (
	"log"
	"sync"

	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/game"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/store"
)

type Hub struct {
	store       store.Store
	coordinator *game.Coordinator // New: use coordinator instead of manager
	mu          sync.RWMutex
	clients     map[string]map[*Connection]bool
	subscribed  map[string]bool
}

func NewHub(s store.Store, coordinator *game.Coordinator) *Hub {
	hub := &Hub{
		store:       s,
		coordinator: coordinator,
		clients:     make(map[string]map[*Connection]bool),
		subscribed:  make(map[string]bool),
	}

	// Set up coordinator event forwarding
	go hub.forwardCoordinatorEvents()

	return hub
}

func (h *Hub) forwardCoordinatorEvents() {
	for event := range h.coordinator.GetWSEventChannel() {
		log.Printf("📡 Forwarding coordinator event to channel %s", event.Channel)
		h.broadcast(event.Channel, event.Data)
	}
}

func (h *Hub) Register(channel string, c *Connection) {
	h.mu.Lock()
	if h.clients[channel] == nil {
		h.clients[channel] = make(map[*Connection]bool)
	}
	h.clients[channel][c] = true

	// Si es la primera conexión de este canal, arrancamos la suscripción Redis
	if !h.subscribed[channel] {
		h.subscribed[channel] = true
		go h.runSubscriber(channel)
	}
	h.mu.Unlock()

	log.Printf("✅ Client registered to channel %s (total: %d)", channel, len(h.clients[channel]))
}

func (h *Hub) Unregister(channel string, c *Connection) {
	h.mu.Lock()
	if conns, ok := h.clients[channel]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, channel)
			// opcional: cancelar la suscripción Redis si ya no hay clientes
		}
	}
	h.mu.Unlock()

	log.Printf("🔌 Client unregistered from channel %s", channel)
}

func (h *Hub) runSubscriber(channel string) {
	msgs, err := h.store.Subscribe(channel)
	if err != nil {
		log.Printf("❌ ERROR suscribiéndome a %s: %v\n", channel, err)
		return
	}
	log.Printf("✔️ Suscrito al canal Redis %s\n", channel)
	for msg := range msgs {
		log.Printf("◀ Received from Redis [%s]: %s\n", msg.Channel, string(msg.Data))
		h.broadcast(msg.Channel, msg.Data)
	}
}

func (h *Hub) Broadcast(channel string, data []byte) error {
	log.Printf("▶ Publish to Redis [%s]: %s\n", channel, string(data))
	return h.store.Publish(store.Message{Channel: channel, Data: data})
}

func (h *Hub) broadcast(channel string, data []byte) {
	h.mu.RLock()
	conns := h.clients[channel]
	log.Printf("✨ Broadcast to %d conn(s) on %q: %s\n", len(conns), channel, string(data))
	for c := range conns {
		log.Printf("   → Enviando a %p\n", c) // identifica la conexión
		c.send(data)
	}
	h.mu.RUnlock()
}
