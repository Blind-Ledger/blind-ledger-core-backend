# 🃏 Guía de Pruebas Manuales - Motor de Texas Hold'em

Esta guía te permite verificar manualmente que el motor de poker funciona correctamente según las reglas oficiales de Texas Hold'em.

## 📋 **Requisitos Previos**

### Instalación y Setup
```bash
# 1. Clonar y navegar al proyecto
cd blind-ledger-core-backend

# 2. Instalar dependencias
go mod download

# 3. Ejecutar tests para verificar que todo funciona
go test ./internal/poker

# 4. Iniciar servidor (en terminal separado)
go run cmd/server/main.go

# 5. Iniciar Redis (requerido para WebSockets)
redis-server  # o docker run -p 6379:6379 redis
```

## 🎯 **Casos de Prueba Críticos**

### **CASO 1: Partida Heads-Up Básica (2 Jugadores)**

#### Setup Inicial
```bash
# Terminal 1: Conectar Jugador A
wscat -c ws://localhost:8080/ws/mesa_test

# Terminal 2: Conectar Jugador B  
wscat -c ws://localhost:8080/ws/mesa_test
```

#### Flujo de Prueba Completo
```json
# 1. Jugador A se une a la mesa
{"action": "join", "player_name": "Alice"}

# 2. Jugador B se une a la mesa  
{"action": "join", "player_name": "Bob"}

# 3. Ambos jugadores se marcan como "ready"
{"action": "set_ready", "ready": true}

# 4. Host (Alice) inicia el juego
{"action": "start_game"}
```

#### ✅ **Verificaciones Esperadas:**

**Después del start_game, verifica que:**
- Alice tiene exactamente 2 cartas privadas
- Bob tiene exactamente 2 cartas privadas  
- Pot inicial = 30 (SB: 10 + BB: 20)
- Alice (dealer/small blind) tiene bet=10, stack=990
- Bob (big blind) tiene bet=20, stack=980
- current_player debería ser Alice (small blind actúa primero)
- phase = "preflop"

#### **Simulación de Ronda Preflop:**
```json
# Alice (small blind) debe completar la apuesta al big blind
{"action": "poker_action", "poker_action": "call", "amount": 0}
# ✅ Verifica: Alice ahora tiene bet=20, stack=980

# Bob (big blind) puede hacer check o raise
{"action": "poker_action", "poker_action": "check", "amount": 0}  
# ✅ Verifica: phase cambia a "flop", aparecen 3 cartas comunitarias
```

#### **Simulación de Ronda Flop:**
```json
# Ahora Bob (primer jugador después del dealer) actúa primero
{"action": "poker_action", "poker_action": "check", "amount": 0}

# Alice puede hacer check o bet
{"action": "poker_action", "poker_action": "bet", "amount": 50}
# ✅ Verifica: current_bet = 50, Alice bet = 50, stack = 930

# Bob debe responder al bet
{"action": "poker_action", "poker_action": "call", "amount": 0}
# ✅ Verifica: phase cambia a "turn", aparece 1 carta comunitaria más
```

---

### **CASO 2: Evaluación de Manos (Showdown)**

#### Setup para Showdown Forzado
```json
# Después de completar todas las rondas de apuestas, 
# el juego debería llegar automáticamente a "showdown"

# ✅ Verifica en el showdown:
# - phase = "showdown"  
# - Se muestran las cartas de ambos jugadores
# - El ganador recibe todo el pot
# - Los stacks se actualizan correctamente
```

#### **Casos de Manos a Verificar:**

**🏆 Jerarquía de Manos (de mayor a menor):**
1. **Royal Flush**: A-K-Q-J-10 del mismo palo
2. **Straight Flush**: 5 cartas consecutivas del mismo palo  
3. **Four of a Kind**: 4 cartas del mismo valor
4. **Full House**: 3 de un valor + 2 de otro valor
5. **Flush**: 5 cartas del mismo palo
6. **Straight**: 5 cartas consecutivas
7. **Three of a Kind**: 3 cartas del mismo valor
8. **Two Pair**: 2 pares diferentes
9. **One Pair**: 2 cartas del mismo valor
10. **High Card**: Carta más alta

---

### **CASO 3: Side Pots (All-Ins Múltiples)**

#### Setup con 3 Jugadores
```bash
# Terminal 1: Alice (stack: 1000)
# Terminal 2: Bob (stack: 500)  
# Terminal 3: Carol (stack: 2000)
```

#### Escenario de All-In
```json
# Simulación de all-ins con diferentes stacks
# Alice: all-in 1000
{"action": "poker_action", "poker_action": "all_in", "amount": 0}

# Bob: all-in 500  
{"action": "poker_action", "poker_action": "all_in", "amount": 0}

# Carol: call 1000
{"action": "poker_action", "poker_action": "call", "amount": 1000}
```

#### ✅ **Verificaciones de Side Pots:**
- **Main Pot**: 500 × 3 = 1500 (todos los jugadores elegibles)
- **Side Pot**: 500 × 2 = 1000 (solo Alice y Carol elegibles)
- **Verificar distribución**: Ganador del main pot + side pot correspondiente

---

### **CASO 4: Auto-Restart de Manos**

#### Configuración
```json
# Verificar que auto-restart está habilitado por defecto
{"action": "get_table_state"}
# ✅ Verifica: auto_restart = true, restart_delay = 5s
```

#### Flujo de Auto-Restart
```json
# 1. Completar una mano hasta showdown
# 2. Esperar 5 segundos después del showdown
# ✅ Verifica: 
#   - phase vuelve a "preflop" automáticamente
#   - Se reparten nuevas cartas
#   - Blinds se colocan automáticamente
#   - Dealer button avanza al siguiente jugador
```

---

### **CASO 5: Buy-In y Configuración de Mesa**

#### Pruebas de Buy-In
```json
# Verificar buy-in personalizado
{"action": "join_with_buyin", "player_name": "Dave", "buy_in_amount": 1500}
# ✅ Verifica: Dave tiene stack = 1500

# Probar buy-in inválido (muy bajo)
{"action": "join_with_buyin", "player_name": "Eve", "buy_in_amount": 100}  
# ✅ Verifica: Error - "below minimum 500"
```

#### Configuración de Mesa
```json
# Cambiar configuración (solo host)
{
  "action": "update_table_config",
  "config": {
    "small_blind": 25,
    "big_blind": 50,
    "buy_in_amount": 2000,
    "min_buy_in": 1000,
    "max_buy_in": 5000
  }
}
```

---

## 🚨 **Casos de Error Comunes**

### **Errores Esperados (Comportamiento Correcto)**
```json
# Actuar fuera de turno
{"action": "poker_action", "poker_action": "call", "amount": 0}
# ✅ Espera: "not your turn"

# Check cuando hay apuesta
{"action": "poker_action", "poker_action": "check", "amount": 0}  
# ✅ Espera: "no puedes hacer check, hay una apuesta que igualar"

# Raise insuficiente
{"action": "poker_action", "poker_action": "raise", "amount": 5}
# ✅ Espera: "el raise mínimo es 20" (big blind)

# Jugador no ready intenta iniciar
{"action": "start_game"}
# ✅ Espera: "all players must be ready"
```

---

## 🧪 **Matriz de Pruebas de Regresión**

| Funcionalidad | Test Manual | Estado | Notas |
|---------------|-------------|---------|-------|
| **Heads-Up Básico** | CASO 1 | ✅ | Blinds, turnos, progresión |
| **Multi-Way (3+ jugadores)** | Extender CASO 1 | ⚠️ | Verificar orden de turnos |
| **Evaluación de Manos** | CASO 2 | ✅ | Todas las combinaciones |
| **Side Pots** | CASO 3 | ✅ | All-ins múltiples |
| **Auto-Restart** | CASO 4 | ✅ | Reinicio automático |
| **Buy-In Personalizado** | CASO 5 | ✅ | Rangos válidos |
| **Manejo de Errores** | Casos de Error | ✅ | Validaciones |
| **Desconexiones** | Manual | ⚠️ | Fold automático |

---

## 📊 **Checklist de Validación Final**

### **Reglas de Texas Hold'em** ✅
- [ ] Blinds se colocan correctamente
- [ ] Dealer button rota correctamente  
- [ ] Orden de acción correcto (preflop vs postflop)
- [ ] Evaluación de manos precisa
- [ ] Side pots funcionan con all-ins múltiples
- [ ] Showdown determina ganador correcto

### **Funcionalidades del Sistema** ✅  
- [ ] WebSocket funciona sin desconexiones
- [ ] Estados se sincronizan entre jugadores
- [ ] Auto-restart funciona después del delay
- [ ] Buy-in personalizado respeta rangos
- [ ] Errores se manejan correctamente

### **Casos Límite** ⚠️
- [ ] Todos los jugadores hacen fold excepto uno
- [ ] Jugador se desconecta durante su turno
- [ ] Múltiples all-ins con stacks diferentes
- [ ] Empates (split pot)
- [ ] Mesa llena (10 jugadores)

---

## 🔧 **Comandos de Debugging**

```bash
# Ver logs del servidor
go run cmd/server/main.go -v

# Ejecutar tests específicos
go test ./internal/poker -run TestSidePots -v

# Verificar race conditions
go test ./internal/poker -race

# Ejecutar fuzzing
go test ./internal/poker -fuzz=FuzzEvaluateHand -fuzztime=30s
```

---

## 📝 **Reportes de Issues**

**Formato para reportar problemas encontrados:**

```markdown
### 🐛 Issue: [Título Descriptivo]

**Pasos para Reproducir:**
1. [Paso 1]
2. [Paso 2]
3. [Paso 3]

**Resultado Esperado:**
[Qué debería pasar]

**Resultado Actual:**  
[Qué está pasando]

**Logs/Screenshots:**
[Incluir información adicional]

**Prioridad:** Alta/Media/Baja
```

---

Esta guía asegura que tu motor de Texas Hold'em cumple con los estándares profesionales y las reglas oficiales del poker. Cada caso de prueba está diseñado para validar aspectos críticos del juego y detectar bugs potenciales.