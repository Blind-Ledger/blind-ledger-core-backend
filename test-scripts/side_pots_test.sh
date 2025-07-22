#!/bin/bash

echo "🧪 PRUEBA EXHAUSTIVA DE SIDE POTS - TEXAS HOLD'EM"
echo "================================================="

# Configuración
TABLE_ID="sidepots_test"
WS_URL="ws://localhost:8080/ws/${TABLE_ID}"

echo "📋 ESCENARIO DE PRUEBA:"
echo "- Alice: 100 chips"
echo "- Bob: 500 chips" 
echo "- Carol: 1000 chips"
echo "- All-ins múltiples para crear side pots"
echo ""

# Función para enviar mensaje WebSocket
send_ws_message() {
    local message="$1"
    local connection="$2"
    echo "📤 Enviando ($connection): $message"
    echo "$message" | wscat -c "$WS_URL" -w 1 &
}

echo "🔄 PASO 1: Conectar jugadores..."
sleep 2

# Conectar Alice
send_ws_message '{"type":"join","version":1,"payload":{"player":"Alice"}}' "Alice"
sleep 1

# Conectar Bob
send_ws_message '{"type":"join","version":1,"payload":{"player":"Bob"}}' "Bob"
sleep 1

# Conectar Carol
send_ws_message '{"type":"join","version":1,"payload":{"player":"Carol"}}' "Carol"
sleep 2

echo "🔄 PASO 2: Marcar todos como ready..."

# Marcar Alice como ready
send_ws_message '{"type":"set_ready","version":1,"payload":{"player":"Alice","ready":true}}' "Alice"
sleep 1

# Marcar Bob como ready
send_ws_message '{"type":"set_ready","version":1,"payload":{"player":"Bob","ready":true}}' "Bob"
sleep 1

# Marcar Carol como ready
send_ws_message '{"type":"set_ready","version":1,"payload":{"player":"Carol","ready":true}}' "Carol"
sleep 2

echo "🔄 PASO 3: Iniciar juego (Alice es host)..."
send_ws_message '{"type":"start_game","version":1,"payload":{"player":"Alice"}}' "Alice"
sleep 3

echo "🔄 PASO 4: Crear scenario de side pots..."
echo "Alice hace all-in (100), Bob hace all-in (500), Carol call (500)"

# Alice all-in (100 chips)
send_ws_message '{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":100}}' "Alice"
sleep 1

# Bob all-in (500 chips)
send_ws_message '{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"all_in","amount":500}}' "Bob"
sleep 1

# Carol call (500 chips)
send_ws_message '{"type":"poker_action","version":1,"payload":{"player":"Carol","action":"call","amount":500}}' "Carol"
sleep 2

echo "🔄 PASO 5: Obtener estado final para verificar side pots..."
send_ws_message '{"type":"get_state","version":1,"payload":{}}' "Check"

echo ""
echo "✅ PRUEBA COMPLETADA"
echo "📊 Verifica los logs del servidor para:"
echo "   - Side pots creados correctamente"
echo "   - Pot principal: 100 (Alice) + 500 (Bob) + 500 (Carol) = 1100"
echo "   - Side pot 1: 300 (Alice eligible: 100x3)"
echo "   - Side pot 2: 800 (Bob+Carol eligible: 400x2)"
echo ""