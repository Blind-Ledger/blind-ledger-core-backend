package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/config"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/store"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/ws"
)

func main() {
	cfg := config.Load()
	log.Printf("🔍 Configuración Redis → Addr=%q, DB=%d\n", cfg.RedisAddr, cfg.RedisDB)

	// Recortamos espacios y nueva línea
	redisAddr := strings.TrimSpace(cfg.RedisAddr)

	// 1. Inicializa RedisStore con la dirección saneada
	redisStore := store.NewRedisStore(redisAddr, cfg.RedisPass, cfg.RedisDB)

	// 2. Crea el Hub
	hub := ws.NewHub(redisStore)

	// 3. Configura el router
	r := mux.NewRouter()
	r.HandleFunc("/ws/{tableId}", ws.ServeWS(hub))

	// 4. Arranca HTTP + WS
	port := strings.TrimSpace(cfg.HTTPPort) // elimina espacios o saltos de línea
	addr := ":" + port                      // ahora es seguro: ":8080"
	log.Printf("Servidor escuchando en %s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
