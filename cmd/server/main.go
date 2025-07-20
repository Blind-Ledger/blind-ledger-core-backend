package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/config"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/game"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/store"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/tournament"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/ws"
)

func main() {
	cfg := config.Load()
	log.Printf("🔍 Configuración Redis → Addr=%q, DB=%d\n", cfg.RedisAddr, cfg.RedisDB)

	// Recortamos espacios y nueva línea
	redisAddr := strings.TrimSpace(cfg.RedisAddr)

	// 1. Inicializa RedisStore con la dirección saneada
	redisStore := store.NewRedisStore(redisAddr, cfg.RedisPass, cfg.RedisDB)

	// 2. Crea el Coordinador (reemplaza Manager + Hub)
	coordinator := game.NewCoordinator()
	
	// 3. Crea el Hub con el coordinador
	hub := ws.NewHub(redisStore, coordinator)

	// 4. Configura el router
	r := mux.NewRouter()
	r.HandleFunc("/ws/{tableId}", ws.ServeWS(hub))
	
	// 5. Endpoint REST para testing (opcional)
	r.HandleFunc("/api/tournaments", handleGetTournaments(coordinator)).Methods("GET")
	r.HandleFunc("/api/tournaments", handleCreateTournament(coordinator)).Methods("POST")
	
	// 6. Serve static files (para frontend)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	// 7. Arranca HTTP + WS
	port := strings.TrimSpace(cfg.HTTPPort)
	addr := ":" + port
	log.Printf("🚀 Servidor escuchando en %s", addr)
	log.Printf("📡 WebSocket endpoint: ws://localhost%s/ws/{tableId}", addr)
	log.Printf("🌐 Web interface: http://localhost%s", addr)
	
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

// REST endpoints para testing del coordinador
func handleGetTournaments(coordinator *game.Coordinator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournaments, err := coordinator.GetActiveTournaments()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
// REST endpoints para testing del coordinador
func handleGetTournaments(coordinator *game.Coordinator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tournaments, err := coordinator.GetActiveTournaments()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		// Simple JSON response for testing
		response := map[string]interface{}{
			"active_tournaments": len(tournaments),
			"tournaments":        tournaments,
		}
		
		json.NewEncoder(w).Encode(response)
	}
}

func handleCreateTournament(coordinator *game.Coordinator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TableID   string `json:"table_id"`
			Organizer string `json:"organizer"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Use default config for MVP
		config := tournament.DefaultSitAndGoConfig()
		tournament, err := coordinator.CreateTournament(req.TableID, req.Organizer, config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"tournament": tournament,
		})
	}
}