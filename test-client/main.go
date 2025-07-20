package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Version int         `json:"version"`
	Payload interface{} `json:"payload"`
}

type Payload struct {
	Player         string `json:"player,omitempty"`
	Amount         int    `json:"amount,omitempty"`
	Action         string `json:"action,omitempty"`
	Ready          bool   `json:"ready,omitempty"`
	TournamentID   string `json:"tournament_id,omitempty"`
	TournamentName string `json:"tournament_name,omitempty"`
	BuyIn          int    `json:"buy_in,omitempty"`
	TournamentType string `json:"tournament_type,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <player_name> [table_id]")
		fmt.Println("Example: go run main.go Alice table1")
		os.Exit(1)
	}

	playerName := os.Args[1]
	tableID := "table1"
	if len(os.Args) > 2 {
		tableID = os.Args[2]
	}

	// Conectar al WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws/" + tableID}
	fmt.Printf("🔗 Conectando a %s como %s\n", u.String(), playerName)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Canal para manejar interrupciones
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Goroutine para leer mensajes
	go func() {
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("📨 Recibido: %s\n", message)
		}
	}()

	// Unirse automáticamente
	joinMsg := Message{
		Type:    "join",
		Version: 1,
		Payload: Payload{Player: playerName},
	}
	sendMessage(c, joinMsg)

	// Leer comandos del usuario
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n📋 Comandos disponibles:")
	printHelp()

	for {
		select {
		case <-interrupt:
			log.Println("🔌 Desconectando...")
			return
		default:
			fmt.Print("\n> ")
			if !scanner.Scan() {
				return
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			if input == "help" || input == "h" {
				printHelp()
				continue
			}

			handleCommand(c, playerName, input)
		}
	}
}

func printHelp() {
	fmt.Println(`
🏁 COMANDOS DE LOBBY:
  ready                   - Marcar como listo
  not_ready              - Marcar como no listo
  start_game             - Iniciar juego (solo host)
  ready_status           - Ver estado de jugadores listos

🎮 COMANDOS DE POKER:
  call                    - Hacer call
  raise <amount>          - Hacer raise (ej: raise 50)
  fold                    - Retirarse de la mano
  all_in                  - Apostar todo

🏆 COMANDOS DE TORNEO:
  create_tournament <id> <name> <buyin> [type]  - Crear torneo
      Ejemplos: 
        create_tournament weekly WeeklyTournament 100 standard
        create_tournament turbo1 TurboEvent 50 turbo
  register <tournament_id>                      - Registrarse en torneo (ej: register weekly)
  start_tournament <tournament_id>              - Iniciar torneo (ej: start_tournament weekly)
  list_tournaments                              - Listar torneos
  tournament_info <tournament_id>               - Info de torneo (ej: tournament_info weekly)

📊 COMANDOS DE ESTADO:
  state               - Obtener estado actual
  help / h            - Mostrar ayuda
  quit / q            - Salir
`)
}

func handleCommand(c *websocket.Conn, playerName, input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	var msg Message

	switch command {
	// Lobby commands
	case "ready":
		msg = Message{
			Type:    "set_ready",
			Version: 1,
			Payload: Payload{Player: playerName, Ready: true},
		}

	case "not_ready":
		msg = Message{
			Type:    "set_ready",
			Version: 1,
			Payload: Payload{Player: playerName, Ready: false},
		}

	case "start_game":
		msg = Message{
			Type:    "start_game",
			Version: 1,
			Payload: Payload{Player: playerName},
		}

	case "ready_status":
		msg = Message{
			Type:    "ready_status",
			Version: 1,
			Payload: Payload{},
		}

	// Poker commands
	case "call":
		msg = Message{
			Type:    "poker_action",
			Version: 1,
			Payload: Payload{Player: playerName, Action: "call"},
		}

	case "raise":
		if len(parts) < 2 {
			fmt.Println("❌ Uso: raise <amount>")
			return
		}
		amount := parseInt(parts[1])
		msg = Message{
			Type:    "poker_action",
			Version: 1,
			Payload: Payload{Player: playerName, Action: "raise", Amount: amount},
		}

	case "fold":
		msg = Message{
			Type:    "poker_action",
			Version: 1,
			Payload: Payload{Player: playerName, Action: "fold"},
		}

	case "all_in":
		msg = Message{
			Type:    "poker_action",
			Version: 1,
			Payload: Payload{Player: playerName, Action: "all_in"},
		}

	case "create_tournament":
		if len(parts) < 4 {
			fmt.Println("❌ Uso: create_tournament <id> <name> <buyin> [type]")
			fmt.Println("   Ejemplo: create_tournament weekly WeeklyTournament 100 standard")
			fmt.Println("   Ejemplo: create_tournament turbo1 TurboEvent 50 turbo")
			return
		}
		tournamentType := "standard"
		if len(parts) > 4 {
			tournamentType = parts[4]
		}
		
		// Limpiar el nombre del torneo de comillas si las tiene
		tournamentName := parts[2]
		if len(tournamentName) > 0 && tournamentName[0] == '"' {
			tournamentName = tournamentName[1:]
		}
		if len(tournamentName) > 0 && tournamentName[len(tournamentName)-1] == '"' {
			tournamentName = tournamentName[:len(tournamentName)-1]
		}
		
		buyIn := parseInt(parts[3])
		if buyIn <= 0 {
			fmt.Printf("❌ Buy-in debe ser positivo, recibido: %d\n", buyIn)
			return
		}
		
		msg = Message{
			Type:    "tournament_create",
			Version: 1,
			Payload: Payload{
				TournamentID:   parts[1],
				TournamentName: tournamentName,
				BuyIn:          buyIn,
				TournamentType: tournamentType,
			},
		}
		
		fmt.Printf("🏆 Creando torneo: ID=%s, Name=%s, BuyIn=%d, Type=%s\n", 
			parts[1], tournamentName, buyIn, tournamentType)

	case "register":
		if len(parts) < 2 {
			fmt.Println("❌ Uso: register <tournament_id>")
			return
		}
		msg = Message{
			Type:    "tournament_register",
			Version: 1,
			Payload: Payload{
				TournamentID: parts[1],
				Player:       playerName,
			},
		}

	case "start_tournament":
		if len(parts) < 2 {
			fmt.Println("❌ Uso: start_tournament <tournament_id>")
			return
		}
		msg = Message{
			Type:    "tournament_start",
			Version: 1,
			Payload: Payload{TournamentID: parts[1]},
		}

	case "list_tournaments":
		msg = Message{
			Type:    "tournament_list",
			Version: 1,
			Payload: Payload{},
		}

	case "tournament_info":
		if len(parts) < 2 {
			fmt.Println("❌ Uso: tournament_info <tournament_id>")
			return
		}
		msg = Message{
			Type:    "tournament_info",
			Version: 1,
			Payload: Payload{TournamentID: parts[1]},
		}

	case "state":
		msg = Message{
			Type:    "get_state",
			Version: 1,
			Payload: Payload{},
		}

	case "quit", "q":
		fmt.Println("👋 ¡Hasta luego!")
		os.Exit(0)

	default:
		fmt.Printf("❌ Comando desconocido: %s\n", command)
		fmt.Println("💡 Escribe 'help' para ver comandos disponibles")
		return
	}

	sendMessage(c, msg)
}

func sendMessage(c *websocket.Conn, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Error marshaling: %v", err)
		return
	}

	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("❌ Error enviando: %v", err)
		return
	}

	fmt.Printf("📤 Enviado: %s\n", data)
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}