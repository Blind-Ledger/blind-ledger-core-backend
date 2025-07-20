# Pruebas con wscat

Si prefieres usar wscat (WebSocket command line client):

## Instalación
```bash
npm install -g wscat
```

## Comandos de Ejemplo

### 1. Conectar y unirse a mesa
```bash
wscat -c ws://localhost:8080/ws/table1

# Enviar:
{"type":"join","version":1,"payload":{"player":"Alice"}}
```

### 2. Crear Torneo
```bash
# En la misma conexión:
{"type":"tournament_create","version":1,"payload":{"tournament_id":"test1","tournament_name":"Test Tournament","buy_in":100,"tournament_type":"standard"}}
```

### 3. Registrarse en Torneo
```bash
{"type":"tournament_register","version":1,"payload":{"tournament_id":"test1","player":"Alice"}}
```

### 4. Acciones de Poker
```bash
# Call
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"call"}}

# Raise
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"raise","amount":50}}

# Fold
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"fold"}}

# All-in
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in"}}
```

### 5. Comandos de Estado
```bash
# Ver estado actual
{"type":"get_state","version":1,"payload":{}}

# Listar torneos
{"type":"tournament_list","version":1,"payload":{}}

# Info de torneo específico
{"type":"tournament_info","version":1,"payload":{"tournament_id":"test1"}}
```

## Ejemplo de Sesión Completa

1. **Terminal 1 (Alice)**:
```bash
wscat -c ws://localhost:8080/ws/table1
{"type":"join","version":1,"payload":{"player":"Alice"}}
{"type":"tournament_create","version":1,"payload":{"tournament_id":"weekly","tournament_name":"Weekly Tournament","buy_in":100,"tournament_type":"standard"}}
{"type":"tournament_register","version":1,"payload":{"tournament_id":"weekly","player":"Alice"}}
```

2. **Terminal 2 (Bob)**:
```bash
wscat -c ws://localhost:8080/ws/table1
{"type":"join","version":1,"payload":{"player":"Bob"}}
{"type":"tournament_register","version":1,"payload":{"tournament_id":"weekly","player":"Bob"}}
```

3. **Continuar agregando más jugadores...**

4. **Iniciar torneo (cualquier terminal)**:
```bash
{"type":"tournament_start","version":1,"payload":{"tournament_id":"weekly"}}
```

5. **Jugar poker**:
```bash
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"call"}}
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"raise","amount":50}}
```