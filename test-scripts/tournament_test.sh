#!/bin/bash

# Script para probar torneos de poker automáticamente
# Uso: ./tournament_test.sh

echo "🏆 === PRUEBA DE TORNEOS DE POKER === 🏆"
echo ""

# Función para hacer requests HTTP
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -n "$data" ]; then
        curl -s -X $method "http://localhost:8080$endpoint" \
             -H "Content-Type: application/json" \
             -d "$data"
    else
        curl -s -X $method "http://localhost:8080$endpoint"
    fi
}

# Verificar que el servidor esté corriendo
echo "🔍 Verificando servidor..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Servidor no está corriendo en puerto 8080"
    echo "💡 Ejecuta: go run cmd/server/main.go"
    exit 1
fi
echo "✅ Servidor está corriendo"
echo ""

# Probar health endpoint
echo "📊 Probando health check..."
HEALTH=$(make_request GET "/health")
echo "Health response: $HEALTH"
echo ""

echo "🎮 Para probar torneos con WebSocket:"
echo "1. Abre varias terminales"
echo "2. En cada una ejecuta: cd test-client && go run main.go <player_name>"
echo "3. Sigue los pasos del escenario de prueba"
echo ""

echo "📋 === ESCENARIO DE PRUEBA SUGERIDO === 📋"
echo ""
echo "🎯 TERMINAL 1 (Alice):"
echo "   cd test-client && go run main.go Alice"
echo "   > create_tournament weekly \"Weekly Tournament\" 100 standard"
echo "   > register weekly"
echo ""
echo "🎯 TERMINAL 2 (Bob):"
echo "   cd test-client && go run main.go Bob"  
echo "   > register weekly"
echo ""
echo "🎯 TERMINAL 3 (Charlie):"
echo "   cd test-client && go run main.go Charlie"
echo "   > register weekly"
echo ""
echo "🎯 TERMINAL 4 (Diana):"
echo "   cd test-client && go run main.go Diana"
echo "   > register weekly"
echo ""
echo "🎯 DE VUELTA EN TERMINAL 1 (Alice):"
echo "   > start_tournament weekly"
echo "   > state"
echo ""
echo "🎮 AHORA JUEGA POKER:"
echo "   - Cada jugador puede hacer: call, raise 50, fold, all_in"
echo "   - Observa cómo avanzan las rondas y se distribuyen los pots"
echo "   - Los blinds subirán automáticamente cada 10 minutos"
echo ""
echo "📊 COMANDOS ÚTILES EN CUALQUIER TERMINAL:"
echo "   > list_tournaments     - Ver todos los torneos"
echo "   > tournament_info weekly - Ver estado del torneo"
echo "   > state                - Ver estado de la mesa actual"
echo ""

echo "✅ ¡Listo para probar!"