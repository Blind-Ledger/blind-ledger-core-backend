package ws_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/game"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/store"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/ws"
)

func TestWebSocketFlow(t *testing.T) {
	// 1. Levantar miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	// 2. Crear Store, Manager y Hub
	redisStore := store.NewRedisStore(mr.Addr(), "", 0)
	mgr := game.NewManager()
	hub := ws.NewHub(redisStore, mgr)

	// 3. Montar router con Gorilla Mux
	router := mux.NewRouter()
	router.HandleFunc("/ws/{tableId}", ws.ServeWS(hub))

	srv := httptest.NewServer(router)
	defer srv.Close()

	// 4. Conectar dos clientes WebSocket
	url := "ws" + srv.URL[len("http"):] + "/ws/testmesa"
	dialer := websocket.DefaultDialer

	c1, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial c1 error: %v", err)
	}
	defer c1.Close()

	c2, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial c2 error: %v", err)
	}
	defer c2.Close()

	// 5. Cliente 1 envía JOIN
	join := map[string]interface{}{
		"type":    "join",
		"version": 1,
		"payload": map[string]string{"player": "X"},
	}
	if err := c1.WriteJSON(join); err != nil {
		t.Fatalf("WriteJSON join error: %v", err)
	}

	// 6. Cliente 2 lee UPDATE
	c2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var resp struct {
		Type    string
		Version int
		Payload struct{ State game.TableState }
	}
	if err := c2.ReadJSON(&resp); err != nil {
		t.Fatalf("ReadJSON error: %v", err)
	}
	// 7. Verificar estado
	if resp.Type != "update" {
		t.Errorf("expected type=update, got %q", resp.Type)
	}
	if len(resp.Payload.State.Players) != 1 || resp.Payload.State.Players[0].Name != "X" {
		t.Errorf("unexpected players: %+v", resp.Payload.State.Players)
	}
}
