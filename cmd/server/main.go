package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/config"
	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/game"
	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/store"
	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/ws"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()
	log.Printf("🔍 Configuración Redis → Addr=%q, DB=%d\n", cfg.RedisAddr, cfg.RedisDB)

	// Recortamos espacios y nueva línea
	redisAddr := strings.TrimSpace(cfg.RedisAddr)

	// 1. Inicializa RedisStore con la dirección saneada
	redisStore := store.NewRedisStore(redisAddr, cfg.RedisPass, cfg.RedisDB)

	// 2. Crea el Hub
	gameMgr := game.NewManager()
	hub := ws.NewHub(redisStore, gameMgr)

	// 3. Configura el router
	r := mux.NewRouter()
	r.HandleFunc("/ws/{tableId}", ws.ServeWS(hub))

	// 4. Servir archivos estáticos del directorio web
	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(http.Dir("./web/"))))

	// 5. Endpoint de salud
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// 6. Arranca HTTP + WS
	port := strings.TrimSpace(cfg.HTTPPort)
	addr := ":" + port
	log.Printf("🚀 Servidor escuchando en %s", addr)
	log.Printf("📡 WebSocket: ws://localhost%s/ws/{tableId}", addr)
	log.Printf("💚 Health check: http://localhost%s/health", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
