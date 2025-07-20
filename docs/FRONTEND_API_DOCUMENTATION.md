# 🎮 Blind Ledger - Frontend Integration Guide

## Table of Contents
1. [Overview](#overview)
2. [WebSocket Connection](#websocket-connection)
3. [Message Protocol](#message-protocol)
4. [Poker Game API](#poker-game-api)
5. [Tournament API](#tournament-api)
6. [Error Handling](#error-handling)
7. [Code Examples](#code-examples)
8. [Testing](#testing)

---

## Overview

Blind Ledger is a real-time poker tournament platform built with Go backend and WebSocket communication. This document provides everything a frontend developer needs to integrate with the backend API.

### Key Features
- **Real-time poker games** with Texas Hold'em rules
- **Tournament management** with progressive blinds
- **WebSocket-based communication** for instant updates
- **Hand evaluation** with proper poker rankings
- **Multiple table support** with automatic balancing

### Tech Stack
- **Backend:** Go 1.22+ with gorilla/websocket
- **Protocol:** JSON over WebSocket
- **Real-time:** Redis pub/sub for broadcasting
- **Storage:** In-memory with Redis backing

---

## WebSocket Connection

### Connection URL
```
ws://localhost:8080/ws/{tableId}
```

### Parameters
- `tableId`: Unique identifier for the poker table/room

### Connection Example
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/table1');

ws.onopen = function(event) {
    console.log('Connected to Blind Ledger');
    // Auto-join with player name
    sendMessage('join', { player: 'PlayerName' });
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    handleMessage(data);
};

ws.onclose = function(event) {
    console.log('Disconnected from Blind Ledger');
};
```

---

## Message Protocol

### Message Structure
All messages follow this JSON envelope format:

```typescript
interface Envelope {
    type: string;           // Message type identifier
    version: number;        // Protocol version (always 1)
    payload: object;        // Message-specific data
    timestamp?: number;     // Optional timestamp
}
```

### Inbound Message Types (Client → Server)

#### Basic Poker Actions
- `join` - Join a table
- `poker_action` - Make poker action (call, raise, fold, all_in)
- `get_state` - Request current table state

#### Tournament Management
- `tournament_create` - Create new tournament
- `tournament_register` - Register for tournament
- `tournament_start` - Start tournament
- `tournament_list` - List all tournaments
- `tournament_info` - Get tournament details

### Outbound Message Types (Server → Client)

- `update` - Table state update
- `poker_update` - Poker-specific update
- `tournament_update` - Tournament-specific update
- `error` - Error message

---

## Poker Game API

### 1. Join Table

**Send:**
```json
{
    "type": "join",
    "version": 1,
    "payload": {
        "player": "PlayerName"
    }
}
```

**Receive:**
```json
{
    "type": "update",
    "version": 1,
    "payload": {
        "state": {
            "host": "PlayerName",
            "players": [{"name": "PlayerName"}],
            "pot": 0,
            "turnIndex": 0,
            "poker_table": {
                "id": "table1",
                "players": [...],
                "community_cards": [],
                "pot": 0,
                "current_player": 0,
                "phase": "waiting",
                "small_blind": 10,
                "big_blind": 20
            }
        }
    }
}
```

### 2. Poker Actions

**Call:**
```json
{
    "type": "poker_action",
    "version": 1,
    "payload": {
        "player": "PlayerName",
        "action": "call"
    }
}
```

**Raise:**
```json
{
    "type": "poker_action",
    "version": 1,
    "payload": {
        "player": "PlayerName",
        "action": "raise",
        "amount": 50
    }
}
```

**Fold:**
```json
{
    "type": "poker_action",
    "version": 1,
    "payload": {
        "player": "PlayerName",
        "action": "fold"
    }
}
```

**All-in:**
```json
{
    "type": "poker_action",
    "version": 1,
    "payload": {
        "player": "PlayerName",
        "action": "all_in"
    }
}
```

### 3. Game State

**Request State:**
```json
{
    "type": "get_state",
    "version": 1,
    "payload": {}
}
```

**Response:**
```json
{
    "type": "update",
    "version": 1,
    "payload": {
        "state": {
            "poker_table": {
                "id": "table1",
                "players": [
                    {
                        "id": "table1_PlayerName",
                        "name": "PlayerName",
                        "stack": 1000,
                        "cards": [
                            {"suit": "hearts", "rank": "A"},
                            {"suit": "spades", "rank": "K"}
                        ],
                        "position": 0,
                        "is_active": true,
                        "has_folded": false,
                        "current_bet": 20
                    }
                ],
                "community_cards": [
                    {"suit": "hearts", "rank": "Q"},
                    {"suit": "diamonds", "rank": "J"},
                    {"suit": "clubs", "rank": "10"}
                ],
                "pot": 100,
                "current_player": 1,
                "phase": "flop",
                "small_blind": 10,
                "big_blind": 20,
                "dealer_position": 0
            }
        }
    }
}
```

### Game Phases
- `waiting` - Waiting for players
- `preflop` - Before community cards
- `flop` - First 3 community cards dealt
- `turn` - 4th community card dealt
- `river` - 5th community card dealt
- `showdown` - Revealing hands and determining winner

### Card Format
```typescript
interface Card {
    suit: "hearts" | "diamonds" | "clubs" | "spades";
    rank: "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" | "10" | "J" | "Q" | "K" | "A";
}
```

---

## Tournament API

### 1. Create Tournament

**Send:**
```json
{
    "type": "tournament_create",
    "version": 1,
    "payload": {
        "tournament_id": "weekly_tournament",
        "tournament_name": "Weekly Championship",
        "buy_in": 100,
        "tournament_type": "standard"
    }
}
```

**Tournament Types:**
- `standard` - 10-minute blind levels
- `turbo` - 5-minute blind levels

**Receive:**
```json
{
    "type": "tournament_update",
    "version": 1,
    "payload": {
        "tournament": {
            "id": "weekly_tournament",
            "config": {
                "name": "Weekly Championship",
                "buy_in": 100,
                "starting_stack": 1500,
                "max_players": 18,
                "min_players": 4,
                "blind_levels": [...],
                "max_tables_size": 6
            },
            "status": "registering",
            "players": {},
            "prize_pool": 0,
            "current_level": 0
        },
        "message": "Tournament created successfully"
    }
}
```

### 2. Register for Tournament

**Send:**
```json
{
    "type": "tournament_register",
    "version": 1,
    "payload": {
        "tournament_id": "weekly_tournament",
        "player": "PlayerName"
    }
}
```

**Receive:**
```json
{
    "type": "tournament_update",
    "version": 1,
    "payload": {
        "tournament": {...},
        "registered": true,
        "players_count": 4,
        "message": "Successfully registered for tournament"
    }
}
```

### 3. Start Tournament

**Send:**
```json
{
    "type": "tournament_start",
    "version": 1,
    "payload": {
        "tournament_id": "weekly_tournament"
    }
}
```

### 4. List Tournaments

**Send:**
```json
{
    "type": "tournament_list",
    "version": 1,
    "payload": {}
}
```

**Receive:**
```json
{
    "type": "tournament_update",
    "version": 1,
    "payload": {
        "tournaments": {
            "weekly_tournament": {...},
            "daily_turbo": {...}
        },
        "message": "Tournament list"
    }
}
```

### 5. Tournament Info

**Send:**
```json
{
    "type": "tournament_info",
    "version": 1,
    "payload": {
        "tournament_id": "weekly_tournament"
    }
}
```

**Receive:**
```json
{
    "type": "tournament_update",
    "version": 1,
    "payload": {
        "tournament": {...},
        "players_count": 12,
        "blind_level": {
            "level": 3,
            "small_blind": 25,
            "big_blind": 50,
            "ante": 0,
            "duration": "10m0s"
        },
        "message": "Tournament information"
    }
}
```

### Tournament Status Values
- `registering` - Open for registration
- `starting` - About to begin
- `active` - Tournament in progress
- `final_table` - Final table phase
- `finished` - Tournament completed
- `cancelled` - Tournament cancelled

---

## Error Handling

### Error Message Format
```json
{
    "type": "error",
    "version": 1,
    "payload": {
        "error": "Error description here"
    }
}
```

### Common Errors

#### Connection Errors
- `Table ID is required`
- `WebSocket upgrade failed`

#### Poker Errors
- `player already at table`
- `table is full`
- `not your turn`
- `invalid action: [action]`
- `invalid raise amount`

#### Tournament Errors
- `tournament [id] already exists`
- `tournament [id] not found`
- `tournament registration is closed`
- `tournament is full`
- `not enough players to start tournament`
- `buy_in must be positive`
- `player already registered`

### Error Handling Best Practices

```javascript
function handleMessage(data) {
    switch (data.type) {
        case 'error':
            showError(data.payload.error);
            break;
        case 'update':
            updateGameState(data.payload.state);
            break;
        case 'tournament_update':
            updateTournamentInfo(data.payload);
            break;
    }
}

function showError(errorMessage) {
    // Display user-friendly error message
    console.error('Game Error:', errorMessage);
    // Show toast/modal to user
}
```

---

## Code Examples

### Complete JavaScript Client

```javascript
class BlindLedgerClient {
    constructor(tableId, playerName) {
        this.tableId = tableId;
        this.playerName = playerName;
        this.ws = null;
        this.gameState = null;
    }

    connect() {
        this.ws = new WebSocket(`ws://localhost:8080/ws/${this.tableId}`);
        
        this.ws.onopen = () => {
            console.log('Connected to Blind Ledger');
            this.join();
        };
        
        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleMessage(data);
        };
        
        this.ws.onclose = () => {
            console.log('Disconnected from Blind Ledger');
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    sendMessage(type, payload) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message = {
                type: type,
                version: 1,
                payload: payload
            };
            this.ws.send(JSON.stringify(message));
        }
    }

    // Poker Actions
    join() {
        this.sendMessage('join', { player: this.playerName });
    }

    call() {
        this.sendMessage('poker_action', { 
            player: this.playerName, 
            action: 'call' 
        });
    }

    raise(amount) {
        this.sendMessage('poker_action', { 
            player: this.playerName, 
            action: 'raise', 
            amount: amount 
        });
    }

    fold() {
        this.sendMessage('poker_action', { 
            player: this.playerName, 
            action: 'fold' 
        });
    }

    allIn() {
        this.sendMessage('poker_action', { 
            player: this.playerName, 
            action: 'all_in' 
        });
    }

    // Tournament Actions
    createTournament(tournamentId, name, buyIn, type = 'standard') {
        this.sendMessage('tournament_create', {
            tournament_id: tournamentId,
            tournament_name: name,
            buy_in: buyIn,
            tournament_type: type
        });
    }

    registerTournament(tournamentId) {
        this.sendMessage('tournament_register', {
            tournament_id: tournamentId,
            player: this.playerName
        });
    }

    startTournament(tournamentId) {
        this.sendMessage('tournament_start', {
            tournament_id: tournamentId
        });
    }

    listTournaments() {
        this.sendMessage('tournament_list', {});
    }

    getTournamentInfo(tournamentId) {
        this.sendMessage('tournament_info', {
            tournament_id: tournamentId
        });
    }

    getState() {
        this.sendMessage('get_state', {});
    }

    handleMessage(data) {
        switch (data.type) {
            case 'update':
            case 'poker_update':
                this.gameState = data.payload.state;
                this.onGameStateUpdate(this.gameState);
                break;
            case 'tournament_update':
                this.onTournamentUpdate(data.payload);
                break;
            case 'error':
                this.onError(data.payload.error);
                break;
        }
    }

    // Override these methods in your implementation
    onGameStateUpdate(state) {
        console.log('Game state updated:', state);
    }

    onTournamentUpdate(data) {
        console.log('Tournament updated:', data);
    }

    onError(error) {
        console.error('Game error:', error);
    }
}

// Usage Example
const client = new BlindLedgerClient('table1', 'Alice');
client.connect();

// After connection, you can use:
// client.call();
// client.raise(50);
// client.createTournament('weekly', 'Weekly Tournament', 100);
```

### React Hook Example

```javascript
import { useState, useEffect, useRef } from 'react';

export function useBlindLedger(tableId, playerName) {
    const [gameState, setGameState] = useState(null);
    const [tournamentData, setTournamentData] = useState(null);
    const [error, setError] = useState(null);
    const [connected, setConnected] = useState(false);
    const ws = useRef(null);

    useEffect(() => {
        if (!tableId || !playerName) return;

        ws.current = new WebSocket(`ws://localhost:8080/ws/${tableId}`);
        
        ws.current.onopen = () => {
            setConnected(true);
            setError(null);
            // Auto-join
            sendMessage('join', { player: playerName });
        };
        
        ws.current.onmessage = (event) => {
            const data = JSON.parse(event.data);
            
            switch (data.type) {
                case 'update':
                case 'poker_update':
                    setGameState(data.payload.state);
                    break;
                case 'tournament_update':
                    setTournamentData(data.payload);
                    break;
                case 'error':
                    setError(data.payload.error);
                    break;
            }
        };
        
        ws.current.onclose = () => {
            setConnected(false);
        };
        
        ws.current.onerror = (error) => {
            setError('Connection error');
        };

        return () => {
            if (ws.current) {
                ws.current.close();
            }
        };
    }, [tableId, playerName]);

    const sendMessage = (type, payload) => {
        if (ws.current && ws.current.readyState === WebSocket.OPEN) {
            const message = {
                type: type,
                version: 1,
                payload: payload
            };
            ws.current.send(JSON.stringify(message));
        }
    };

    const actions = {
        call: () => sendMessage('poker_action', { player: playerName, action: 'call' }),
        raise: (amount) => sendMessage('poker_action', { player: playerName, action: 'raise', amount }),
        fold: () => sendMessage('poker_action', { player: playerName, action: 'fold' }),
        allIn: () => sendMessage('poker_action', { player: playerName, action: 'all_in' }),
        getState: () => sendMessage('get_state', {}),
        createTournament: (id, name, buyIn, type) => sendMessage('tournament_create', {
            tournament_id: id,
            tournament_name: name,
            buy_in: buyIn,
            tournament_type: type
        }),
        registerTournament: (id) => sendMessage('tournament_register', {
            tournament_id: id,
            player: playerName
        }),
        startTournament: (id) => sendMessage('tournament_start', { tournament_id: id }),
        listTournaments: () => sendMessage('tournament_list', {}),
    };

    return {
        gameState,
        tournamentData,
        error,
        connected,
        actions
    };
}
```

---

## Testing

### Manual Testing Tools

1. **Test Client:** Use the provided Go test client
```bash
cd test-client && go run main.go PlayerName table1
```

2. **Web Interface:** Open `web/poker-test.html` in browser

3. **wscat:** Command line WebSocket client
```bash
npm install -g wscat
wscat -c ws://localhost:8080/ws/table1
```

### Health Check
```bash
curl http://localhost:8080/health
# Response: OK
```

### Test Scenarios

#### Basic Poker Flow
1. Connect 2+ players to same table
2. Players automatically receive cards and blinds are posted
3. Players take turns making actions
4. Hand completes and winner is determined
5. New hand starts automatically

#### Tournament Flow
1. Create tournament
2. Register multiple players
3. Start tournament
4. Players are distributed across balanced tables
5. Play poker with progressive blinds
6. Eliminated players are tracked
7. Final table formation when ≤6 players remain

### Performance Testing
- **Concurrent connections:** Test with 10+ simultaneous WebSocket connections
- **Message throughput:** Rapid-fire poker actions
- **Tournament load:** Multiple tournaments with many players

---

## TypeScript Definitions

```typescript
// Game Types
interface Card {
    suit: 'hearts' | 'diamonds' | 'clubs' | 'spades';
    rank: '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | '10' | 'J' | 'Q' | 'K' | 'A';
}

interface PokerPlayer {
    id: string;
    name: string;
    stack: number;
    cards: Card[];
    position: number;
    is_active: boolean;
    has_folded: boolean;
    current_bet: number;
}

interface PokerTable {
    id: string;
    players: PokerPlayer[];
    community_cards: Card[];
    pot: number;
    current_player: number;
    phase: 'waiting' | 'preflop' | 'flop' | 'turn' | 'river' | 'showdown';
    small_blind: number;
    big_blind: number;
    dealer_position: number;
}

interface TableState {
    host: string;
    players: { name: string }[];
    pot: number;
    turnIndex: number;
    poker_table?: PokerTable;
    phase?: string;
}

// Tournament Types
interface BlindLevel {
    level: number;
    small_blind: number;
    big_blind: number;
    ante: number;
    duration: string;
}

interface TournamentConfig {
    name: string;
    buy_in: number;
    starting_stack: number;
    max_players: number;
    min_players: number;
    blind_levels: BlindLevel[];
    max_tables_size: number;
}

interface Tournament {
    id: string;
    config: TournamentConfig;
    status: 'registering' | 'starting' | 'active' | 'final_table' | 'finished' | 'cancelled';
    players: Record<string, any>;
    prize_pool: number;
    current_level: number;
}

// Message Types
interface GameMessage {
    type: string;
    version: number;
    payload: any;
    timestamp?: number;
}

interface PokerActionPayload {
    player: string;
    action: 'call' | 'raise' | 'fold' | 'all_in';
    amount?: number;
}

interface TournamentCreatePayload {
    tournament_id: string;
    tournament_name: string;
    buy_in: number;
    tournament_type: 'standard' | 'turbo';
}

interface TournamentRegisterPayload {
    tournament_id: string;
    player: string;
}
```

---

## Support

### Backend Server Info
- **Host:** localhost:8080
- **WebSocket Path:** /ws/{tableId}
- **Health Check:** GET /health
- **Protocol:** JSON over WebSocket

### Common Issues
1. **Connection refused:** Ensure server is running on port 8080
2. **Invalid table ID:** Use alphanumeric table IDs
3. **Message format errors:** Ensure JSON is valid and includes required fields
4. **Turn violations:** Only current player can make actions

### Debugging
- Enable WebSocket message logging in browser dev tools
- Check server logs for detailed error messages
- Use the provided test client for comparison
- Verify message format matches exactly

For additional support or questions, refer to the test client implementation and examples provided in this documentation.

---

*Generated for Blind Ledger Poker Tournament Platform*  
*Version 1.0 - Real-time Texas Hold'em with Tournament Support*