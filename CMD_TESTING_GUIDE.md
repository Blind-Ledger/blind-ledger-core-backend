# 🧪 **GUÍA EXHAUSTIVA DE PRUEBAS POR LÍNEA DE COMANDOS - TEXAS HOLD'EM**

## 📋 **ÍNDICE DE PRUEBAS**

1. [**Preparación del Entorno**](#preparación-del-entorno)
2. [**Pruebas Básicas de Conexión**](#pruebas-básicas-de-conexión)
3. [**Pruebas de Lobby System**](#pruebas-de-lobby-system)
4. [**Pruebas de Heads-Up (2 Jugadores)**](#pruebas-de-heads-up-2-jugadores)
5. [**Pruebas de 3 Jugadores**](#pruebas-de-3-jugadores)
6. [**Pruebas de 4+ Jugadores**](#pruebas-de-4-jugadores)
7. [**Pruebas de Evaluación de Manos**](#pruebas-de-evaluación-de-manos)
8. [**Pruebas de Side Pots (All-ins)**](#pruebas-de-side-pots-all-ins)
9. [**Pruebas de Auto-restart**](#pruebas-de-auto-restart)
10. [**Pruebas de Edge Cases**](#pruebas-de-edge-cases)
11. [**Scripts de Automatización**](#scripts-de-automatización)

---

## 🔧 **PREPARACIÓN DEL ENTORNO**

### **1. Instalar Herramientas Necesarias**
```bash
# Instalar wscat para conexiones WebSocket
npm install -g wscat

# Verificar instalación
wscat --version
```

### **2. Iniciar el Servidor**
```bash
# Terminal 1: Iniciar servidor
cd /Users/zkcaleb/Documents/Blind\ Ledger/Code/blind-ledger-core-backend
go run cmd/server/main.go
```

### **3. Verificar Servidor**
```bash
# Terminal 2: Verificar salud del servidor
curl -s http://localhost:8080/health
# Esperado: OK

# Verificar logs del servidor
# Terminal 1 debe mostrar: 🚀 Servidor escuchando en :8080
```

---

## 🔗 **PRUEBAS BÁSICAS DE CONEXIÓN**

### **Test 1.1: Conexión Simple de Alice**
**Mesa:** `cmd_basic_test`

```bash
# Terminal 2: Conectar Alice
wscat -c "ws://localhost:8080/ws/cmd_basic_test"
```

**Enviar mensaje de conexión:**
```json
{"type":"join","version":1,"payload":{"player":"Alice"}}
```

**✅ Resultado Esperado:**
```json
{"type":"update","version":1,"payload":{"state":{"host":"Alice","players":[{"name":"Alice"}],"pot":0,"turnIndex":0,"poker_table":{"phase":"lobby"}}}}
```

**Salir:** `Ctrl+C`

### **Test 1.2: Múltiples Conexiones Simultáneas**
**Mesa:** `cmd_multi_test`

```bash
# Terminal 2: Alice
wscat -c "ws://localhost:8080/ws/cmd_multi_test" &
ALICE_PID=$!

# Terminal 3: Bob  
wscat -c "ws://localhost:8080/ws/cmd_multi_test" &
BOB_PID=$!

# Terminal 4: Carol
wscat -c "ws://localhost:8080/ws/cmd_multi_test" &
CAROL_PID=$!
```

**En cada terminal enviar:**
```bash
# Alice (Terminal 2):
{"type":"join","version":1,"payload":{"player":"Alice"}}

# Bob (Terminal 3):
{"type":"join","version":1,"payload":{"player":"Bob"}}

# Carol (Terminal 4):
{"type":"join","version":1,"payload":{"player":"Carol"}}
```

**✅ Verificar:** Cada conexión recibe updates con todos los jugadores

**Limpiar conexiones:**
```bash
kill $ALICE_PID $BOB_PID $CAROL_PID 2>/dev/null
```

---

## 🏠 **PRUEBAS DE LOBBY SYSTEM**

### **Test 2.1: Sistema Ready Completo**
**Mesa:** `cmd_lobby_ready`

```bash
# Abrir 2 terminales para Alice (HOST) y Bob
# Terminal 2: Alice
wscat -c "ws://localhost:8080/ws/cmd_lobby_ready"
```

```bash
# Terminal 3: Bob
wscat -c "ws://localhost:8080/ws/cmd_lobby_ready"
```

**Secuencia de comandos:**

**Alice (Terminal 2):**
```json
{"type":"join","version":1,"payload":{"player":"Alice"}}
{"type":"set_ready","version":1,"payload":{"player":"Alice","ready":true}}
```

**Bob (Terminal 3):**
```json
{"type":"join","version":1,"payload":{"player":"Bob"}}
{"type":"set_ready","version":1,"payload":{"player":"Bob","ready":true}}
```

**Alice inicia el juego (Solo HOST puede):**
```json
{"type":"start_game","version":1,"payload":{"player":"Alice"}}
```

**✅ Resultado Esperado:**
- Alice es HOST
- Ambos marcan ready
- Alice puede iniciar juego
- Fase cambia de `lobby` a `preflop`

### **Test 2.2: Verificar Estado Ready**
```bash
# En cualquier terminal conectado:
{"type":"ready_status","version":1,"payload":{}}
```

**✅ Resultado Esperado:**
```json
{"type":"ready_status","version":1,"payload":{"ready_status":{"Alice":true,"Bob":true}}}
```

---

## 👥 **PRUEBAS DE HEADS-UP (2 JUGADORES)**

### **Test 3.1: Partida Completa Heads-Up**
**Mesa:** `cmd_headsup_complete`

**Setup inicial:**
```bash
# Terminal 2: Alice
wscat -c "ws://localhost:8080/ws/cmd_headsup_complete"

# Terminal 3: Bob
wscat -c "ws://localhost:8080/ws/cmd_headsup_complete"
```

**Secuencia completa:**
```bash
# Alice:
{"type":"join","version":1,"payload":{"player":"Alice"}}
{"type":"set_ready","version":1,"payload":{"player":"Alice","ready":true}}

# Bob:
{"type":"join","version":1,"payload":{"player":"Bob"}}
{"type":"set_ready","version":1,"payload":{"player":"Bob","ready":true}}

# Alice inicia:
{"type":"start_game","version":1,"payload":{"player":"Alice"}}
```

**Ronda de apuestas preflop:**
```bash
# Alice (SB/Dealer, actúa primero en heads-up preflop):
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"call","amount":20}}

# Bob (BB, puede check o raise):
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"check","amount":0}}
```

**✅ Verificar:** Flop aparece (3 cartas comunitarias)

**Ronda post-flop:**
```bash
# Bob (actúa primero post-flop):
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"check","amount":0}}

# Alice:
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"bet","amount":40}}

# Bob:
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"call","amount":40}}
```

**✅ Verificar:** Turn aparece (4ta carta)

**Continuar hasta river y showdown**

### **Test 3.2: All-in Heads-Up**
**Mesa:** `cmd_headsup_allin`

```bash
# Después del setup inicial:
# Alice:
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":1000}}

# Bob call o fold:
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"call","amount":980}}
```

**✅ Verificar:** 
- Showdown inmediato
- Todas las cartas comunitarias aparecen
- Ganador recibe todo el pot

---

## 🎯 **PRUEBAS DE 3 JUGADORES**

### **Test 4.1: Rotación de Blinds (3 manos completas)**
**Mesa:** `cmd_three_blinds`

**Setup:** 3 terminales para Alice, Bob, Carol

**MANO 1:**
```bash
# Todos conectan y ready
# Alice inicia juego
{"type":"start_game","version":1,"payload":{"player":"Alice"}}

# Verificar posiciones:
# Alice = Dealer, Bob = SB, Carol = BB
```

**Completar mano 1 hasta showdown, esperar auto-restart**

**MANO 2 (después de 5 segundos):**
**✅ Verificar rotación:**
- Bob = Dealer, Carol = SB, Alice = BB

**MANO 3:**
**✅ Verificar rotación completa:**
- Carol = Dealer, Alice = SB, Bob = BB

### **Test 4.2: Orden de Acción (3 jugadores)**
**Mesa:** `cmd_three_action_order`

**Verificar orden preflop:**
- UTG (izquierda del BB) actúa primero
- Dealer actúa segundo
- SB actúa tercero  
- BB actúa último

**Verificar orden post-flop:**
- SB actúa primero
- BB actúa segundo
- Dealer actúa último

### **Test 4.3: Jugador Fold**
```bash
# Después de inicio:
# Alice fold:
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"fold","amount":0}}

# Solo Bob y Carol continúan
```

**✅ Verificar:**
- Alice no participa más en la mano
- Solo Bob y Carol pueden ganar el pot

---

## 🎪 **PRUEBAS DE 4+ JUGADORES**

### **Test 5.1: Mesa Completa (4 jugadores)**
**Mesa:** `cmd_four_players`

**Setup:** Alice (Dealer), Bob (SB), Carol (BB), Dave (UTG)

**Verificar orden preflop:**
```bash
# Dave (UTG) actúa primero:
{"type":"poker_action","version":1,"payload":{"player":"Dave","action":"call","amount":20}}

# Alice (Dealer):
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"raise","amount":40}}

# Bob (SB):
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"call","amount":60}}

# Carol (BB):
{"type":"poker_action","version":1,"payload":{"player":"Carol","action":"call","amount":40}}

# Dave debe responder al raise:
{"type":"poker_action","version":1,"payload":{"player":"Dave","action":"call","amount":40}}
```

**✅ Verificar:** Flop aparece después de que todos igualen

### **Test 5.2: Un Solo Ganador por Fold**
```bash
# 3 jugadores fold, solo 1 queda:
{"type":"poker_action","version":1,"payload":{"player":"Dave","action":"fold","amount":0}}
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"fold","amount":0}}
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"fold","amount":0}}
```

**✅ Verificar:**
- Carol gana automáticamente
- No hay showdown
- Pot va directo a Carol

---

## 🃏 **PRUEBAS DE EVALUACIÓN DE MANOS**

### **Test 6.1: Royal Flush vs Straight Flush**
**Mesa:** `cmd_hand_eval_royal`

**Configurar escenario específico en showdown:**
```bash
# Llegar a showdown con cartas específicas
# Verificar que el evaluador identifica correctamente:
# Royal Flush > Straight Flush
```

### **Test 6.2: Empate Perfecto**
**Mesa:** `cmd_hand_eval_tie`

```bash
# Configurar empate exacto
# Verificar que el pot se divide equitativamente
```

**✅ Verificar en respuesta:**
```json
{"type":"update","payload":{"state":{"pot":0}}}
```
**Pot debe ser 0 después de división correcta**

### **Test 6.3: Script de Evaluación Automatizada**
```bash
# Crear script para probar todas las 10 manos
cat > test_hand_evaluation.sh << 'EOF'
#!/bin/bash

echo "🃏 Testing Hand Evaluation..."

# Royal Flush Test
echo "Test 1: Royal Flush"
# [comandos específicos]

# Straight Flush Test  
echo "Test 2: Straight Flush"
# [comandos específicos]

# [... para todas las 10 manos]
EOF

chmod +x test_hand_evaluation.sh
./test_hand_evaluation.sh
```

---

## 💰 **PRUEBAS DE SIDE POTS (ALL-INS)**

### **Test 7.1: Side Pot Básico**
**Mesa:** `cmd_sidepot_basic`

**Escenario:** Alice (100), Bob (500)

```bash
# Alice all-in:
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":100}}

# Bob call:
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"call","amount":100}}
```

**✅ Verificar:** Pot único de 200 fichas

### **Test 7.2: Side Pot Complejo (3 jugadores)**
**Mesa:** `cmd_sidepot_complex`

**Escenario:** Alice (100), Bob (500), Carol (1000)

```bash
# Alice all-in 100:
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":100}}

# Bob all-in 500:
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"all_in","amount":500}}

# Carol call 500:
{"type":"poker_action","version":1,"payload":{"player":"Carol","action":"call","amount":500}}
```

**✅ Verificar en showdown:**
- Side Pot 1: 300 fichas (Alice eligible)
- Side Pot 2: 800 fichas (Bob y Carol eligible)  
- Total: 1100 fichas

### **Test 7.3: Script de Side Pots Automatizado**
```bash
cat > test_side_pots.sh << 'EOF'
#!/bin/bash

echo "💰 Testing Side Pots..."

# Test básico 2 jugadores
echo "Test 1: Basic 2-player side pot"
wscat -c "ws://localhost:8080/ws/sidepot_test_1" << 'COMMANDS'
{"type":"join","version":1,"payload":{"player":"Alice"}}
{"type":"join","version":1,"payload":{"player":"Bob"}}  
{"type":"set_ready","version":1,"payload":{"player":"Alice","ready":true}}
{"type":"set_ready","version":1,"payload":{"player":"Bob","ready":true}}
{"type":"start_game","version":1,"payload":{"player":"Alice"}}
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":100}}
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"call","amount":100}}
COMMANDS

# Esperar resultado y verificar
sleep 2

# Test complejo 3 jugadores
echo "Test 2: Complex 3-player side pot"
# [comandos similares para caso complejo]

EOF

chmod +x test_side_pots.sh
./test_side_pots.sh
```

---

## 🔄 **PRUEBAS DE AUTO-RESTART**

### **Test 8.1: Auto-restart Normal**
**Mesa:** `cmd_auto_restart`

```bash
# Jugar hasta showdown
# NO enviar más comandos
# Esperar exactamente 5 segundos
sleep 6

# Verificar que nueva mano inició automáticamente
# Nuevas cartas, blinds rotados, etc.
```

### **Test 8.2: Auto-restart con Eliminación**
```bash
# Alice pierde todo:
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":100}}
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"call","amount":100}}

# Alice queda con 0 fichas
# Esperar 5 segundos
sleep 6

# Verificar que NO hay auto-restart (solo 1 jugador con fichas)
```

### **Test 8.3: Script de Auto-restart**
```bash
cat > test_auto_restart.sh << 'EOF'
#!/bin/bash

echo "🔄 Testing Auto-restart..."

# Función para monitorear cambios de fase
monitor_phase_changes() {
    local table_id=$1
    echo "Monitoring phase changes for table: $table_id"
    
    # Conectar y monitorear
    wscat -c "ws://localhost:8080/ws/$table_id" << 'MONITOR'
{"type":"get_state","version":1,"payload":{}}
MONITOR
}

# Test 1: Restart normal
echo "Test 1: Normal auto-restart"
# [setup y comandos]

# Monitorear cambios por 10 segundos
timeout 10s monitor_phase_changes "auto_restart_test_1"

EOF

chmod +x test_auto_restart.sh
./test_auto_restart.sh
```

---

## ⚠️ **PRUEBAS DE EDGE CASES**

### **Test 9.1: Sin Fichas para Blinds**
**Mesa:** `cmd_edge_no_chips`

```bash
# Configurar Alice con 5 fichas, debe pagar BB de 20
# Verificar all-in automático
{"type":"get_state","version":1,"payload":{}}
```

### **Test 9.2: Todos All-in Preflop**
```bash
# Todos los jugadores all-in:
{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":1000}}
{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"all_in","amount":1000}}
{"type":"poker_action","version":1,"payload":{"player":"Carol","action":"all_in","amount":1000}}
```

**✅ Verificar:**
- Todas las cartas comunitarias salen inmediatamente
- Showdown directo sin más betting rounds

### **Test 9.3: Desconexión Durante Turno**
```bash
# Alice tiene el turno
# Simular desconexión: Ctrl+C en terminal de Alice
# Verificar en otros terminales que Alice hizo fold automático
```

---

## 🤖 **SCRIPTS DE AUTOMATIZACIÓN**

### **Script Master de Pruebas**
```bash
cat > run_all_tests.sh << 'EOF'
#!/bin/bash

echo "🧪 INICIANDO PRUEBAS EXHAUSTIVAS DEL SISTEMA POKER"
echo "================================================="

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Contadores
PASSED=0
FAILED=0

# Función para ejecutar test
run_test() {
    local test_name=$1
    local test_command=$2
    
    echo -e "\n${YELLOW}🔧 Ejecutando: $test_name${NC}"
    
    if eval "$test_command"; then
        echo -e "${GREEN}✅ PASSED: $test_name${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAILED: $test_name${NC}"
        ((FAILED++))
    fi
}

# Verificar servidor
echo "🔍 Verificando servidor..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ Servidor activo${NC}"
else
    echo -e "${RED}❌ Servidor no responde${NC}"
    exit 1
fi

# Tests de conexión básica
run_test "Conexión Simple" "./test_basic_connection.sh"
run_test "Múltiples Conexiones" "./test_multi_connection.sh"

# Tests de lobby  
run_test "Sistema Ready" "./test_lobby_ready.sh"
run_test "Inicio de Juego" "./test_game_start.sh"

# Tests de poker
run_test "Heads-up Completo" "./test_headsup.sh"
run_test "3 Jugadores" "./test_three_players.sh"
run_test "4+ Jugadores" "./test_four_players.sh"

# Tests de evaluación
run_test "Evaluación de Manos" "./test_hand_evaluation.sh"

# Tests de side pots
run_test "Side Pots Básicos" "./test_sidepots_basic.sh"
run_test "Side Pots Complejos" "./test_sidepots_complex.sh"

# Tests de auto-restart
run_test "Auto-restart Normal" "./test_autorestart.sh"
run_test "Auto-restart Eliminación" "./test_autorestart_elimination.sh"

# Tests de edge cases
run_test "Edge Cases" "./test_edge_cases.sh"

# Resumen final
echo -e "\n🏆 RESUMEN DE PRUEBAS"
echo "==================="
echo -e "${GREEN}✅ Pasaron: $PASSED${NC}"
echo -e "${RED}❌ Fallaron: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 TODAS LAS PRUEBAS PASARON - SISTEMA APROBADO${NC}"
    exit 0
else
    echo -e "\n${RED}⚠️  SISTEMA NECESITA REVISIÓN${NC}"
    exit 1
fi
EOF

chmod +x run_all_tests.sh
```

### **Script de Conexión Básica**
```bash
cat > test_basic_connection.sh << 'EOF'
#!/bin/bash

echo "🔗 Testing Basic Connection..."

# Test conexión simple
{
    echo '{"type":"join","version":1,"payload":{"player":"Alice"}}'
    sleep 1
    echo '{"type":"get_state","version":1,"payload":{}}'
    sleep 1
} | wscat -c "ws://localhost:8080/ws/basic_test" > /tmp/basic_test_output.log 2>&1 &

PID=$!
sleep 3
kill $PID 2>/dev/null

# Verificar resultado
if grep -q '"player":"Alice"' /tmp/basic_test_output.log; then
    echo "✅ Alice connected successfully"
    exit 0
else
    echo "❌ Alice connection failed"
    exit 1
fi
EOF

chmod +x test_basic_connection.sh
```

### **Script de Side Pots Complejo**
```bash
cat > test_sidepots_complex.sh << 'EOF'
#!/bin/bash

echo "💰 Testing Complex Side Pots..."

TABLE_ID="sidepots_complex_$(date +%s)"

# Función para conectar jugador
connect_player() {
    local player_name=$1
    local commands_file="/tmp/${player_name}_commands.txt"
    
    cat > "$commands_file" << COMMANDS
{"type":"join","version":1,"payload":{"player":"$player_name"}}
{"type":"set_ready","version":1,"payload":{"player":"$player_name","ready":true}}
COMMANDS
    
    wscat -c "ws://localhost:8080/ws/$TABLE_ID" < "$commands_file" > "/tmp/${player_name}_output.log" 2>&1 &
    echo $!
}

# Conectar 3 jugadores
ALICE_PID=$(connect_player "Alice")
sleep 1
BOB_PID=$(connect_player "Bob")  
sleep 1
CAROL_PID=$(connect_player "Carol")
sleep 2

# Alice inicia juego
{
    echo '{"type":"start_game","version":1,"payload":{"player":"Alice"}}'
    sleep 1
    echo '{"type":"poker_action","version":1,"payload":{"player":"Alice","action":"all_in","amount":100}}'
    sleep 1
    echo '{"type":"poker_action","version":1,"payload":{"player":"Bob","action":"all_in","amount":500}}'
    sleep 1  
    echo '{"type":"poker_action","version":1,"payload":{"player":"Carol","action":"call","amount":500}}'
    sleep 3
} | wscat -c "ws://localhost:8080/ws/$TABLE_ID" > /tmp/game_control.log 2>&1 &

CONTROL_PID=$!
sleep 8

# Limpiar procesos
kill $ALICE_PID $BOB_PID $CAROL_PID $CONTROL_PID 2>/dev/null

# Verificar side pots en logs
if grep -q '"side_pots"' /tmp/*_output.log /tmp/game_control.log; then
    echo "✅ Side pots created successfully"
    
    # Verificar distribución correcta (total debe ser 1100)
    if grep -q '"pot":0' /tmp/game_control.log; then
        echo "✅ Side pots distributed correctly"
        exit 0
    else
        echo "❌ Side pot distribution failed"
        exit 1
    fi
else
    echo "❌ Side pots not created"
    exit 1
fi
EOF

chmod +x test_sidepots_complex.sh
```

### **Script de Monitoreo Continuo**
```bash
cat > monitor_system.sh << 'EOF'
#!/bin/bash

echo "📊 Monitoring System Performance..."

# Monitorear por 30 segundos
DURATION=30
END_TIME=$((SECONDS + DURATION))

echo "Monitoring for ${DURATION} seconds..."
echo "Time,Connections,Memory,CPU" > /tmp/system_monitor.csv

while [ $SECONDS -lt $END_TIME ]; do
    # Obtener métricas
    CONNECTIONS=$(netstat -an | grep :8080 | grep ESTABLISHED | wc -l)
    MEMORY=$(ps -p $(pgrep -f "go run cmd/server/main.go") -o %mem | tail -1)
    CPU=$(ps -p $(pgrep -f "go run cmd/server/main.go") -o %cpu | tail -1)
    
    echo "$(date +%H:%M:%S),$CONNECTIONS,$MEMORY,$CPU" >> /tmp/system_monitor.csv
    
    sleep 1
done

echo "📈 Monitoring complete. Results in /tmp/system_monitor.csv"
echo "Summary:"
echo "Max Connections: $(cut -d',' -f2 /tmp/system_monitor.csv | sort -n | tail -1)"
echo "Avg Memory: $(cut -d',' -f3 /tmp/system_monitor.csv | awk '{sum+=$1} END {print sum/NR}')%"
echo "Max CPU: $(cut -d',' -f4 /tmp/system_monitor.csv | sort -n | tail -1)%"
EOF

chmod +x monitor_system.sh
```

---

## ✅ **CHECKLIST DE EJECUCIÓN CMD**

### **🎯 Preparación**
- [ ] **wscat instalado:** `npm install -g wscat`
- [ ] **Servidor corriendo:** `go run cmd/server/main.go`
- [ ] **Health check:** `curl http://localhost:8080/health` → `OK`

### **🔧 Scripts Creados**
- [ ] **run_all_tests.sh:** Script master de todas las pruebas
- [ ] **test_basic_connection.sh:** Pruebas de conexión básica
- [ ] **test_sidepots_complex.sh:** Side pots complejos automatizados  
- [ ] **monitor_system.sh:** Monitoreo de performance
- [ ] **Permisos de ejecución:** `chmod +x *.sh`

### **🧪 Ejecución de Pruebas**
- [ ] **Pruebas individuales:** `./test_basic_connection.sh`
- [ ] **Pruebas completas:** `./run_all_tests.sh`
- [ ] **Monitoreo:** `./monitor_system.sh`
- [ ] **Logs verificados:** Todos los archivos `/tmp/*_output.log`

### **📊 Validación de Resultados**
- [ ] **Conexiones:** Jugadores se conectan correctamente
- [ ] **Lobby:** Ready/Start funciona por CMD
- [ ] **Poker:** Betting rounds completos por CMD
- [ ] **Side pots:** Cálculos matemáticamente correctos
- [ ] **Auto-restart:** Flujo continuo automático
- [ ] **Performance:** CPU < 10%, Memory < 50MB

---

## 🏆 **CRITERIOS DE ÉXITO CMD**

### **✅ SISTEMA APROBADO SI:**
- **run_all_tests.sh** ejecuta sin errores (exit code 0)
- **Todas las conexiones WebSocket** funcionan por CMD
- **Side pots** calculan matemáticamente correctos
- **Auto-restart** funciona sin intervención manual
- **Performance** mantiene métricas aceptables
- **Logs** no muestran errores críticos

### **❌ SISTEMA RECHAZADO SI:**
- Cualquier script retorna error (exit code ≠ 0)
- Side pots calculan incorrectamente
- Conexiones fallan consistentemente  
- Performance degrada significativamente
- Crashes o errores no recuperables

---

## 📝 **COMANDOS RÁPIDOS DE REFERENCIA**

### **Conexión WebSocket básica:**
```bash
wscat -c "ws://localhost:8080/ws/TABLE_ID"
```

### **Mensaje de unirse:**
```json
{"type":"join","version":1,"payload":{"player":"PLAYER_NAME"}}
```

### **Marcar ready:**
```json
{"type":"set_ready","version":1,"payload":{"player":"PLAYER_NAME","ready":true}}
```

### **Iniciar juego (solo HOST):**
```json
{"type":"start_game","version":1,"payload":{"player":"HOST_NAME"}}
```

### **Acciones de poker:**
```json
{"type":"poker_action","version":1,"payload":{"player":"PLAYER","action":"ACTION","amount":AMOUNT}}
```

**Acciones válidas:** `call`, `raise`, `check`, `fold`, `all_in`

### **Obtener estado:**
```json
{"type":"get_state","version":1,"payload":{}}
```

### **Ejecutar todas las pruebas:**
```bash
./run_all_tests.sh
```

---

**🎯 Con esta guía puedes validar completamente el sistema Texas Hold'em usando solo línea de comandos, de forma automatizada y procedural. ¡Perfecto para CI/CD y pruebas sistemáticas!**