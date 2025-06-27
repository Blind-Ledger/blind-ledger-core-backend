package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/config"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/store"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/ws"
)

func main() {
	cfg := config.Load()

	// 1. Inicializa RedisStore
	redisStore := store.NewRedisStore(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)

	// 2. Crea el Hub
	hub := ws.NewHub(redisStore)

	// 3. Configura el router
	r := mux.NewRouter()
	r.HandleFunc("/ws/{tableId}", ws.ServeWS(hub))

	// 4. Arranca HTTP = WS
	log.Printf("Servidor escuchando en :%s", cfg.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.HTTPPort, r); err != nil {
		log.Fatal(err)
	}
}
