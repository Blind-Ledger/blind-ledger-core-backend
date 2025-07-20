# 🧪 Guía Completa de Pruebas - Blind Ledger Poker Tournaments

Esta guía te mostrará cómo hacer pruebas manuales completas de tu sistema de torneos de poker.

## 🚀 Paso 1: Iniciar el Servidor

```bash
# En la raíz del proyecto
go run cmd/server/main.go
```

Deberías ver:
```
🔍 Configuración Redis → Addr="localhost:6379", DB=0
🚀 Servidor escuchando en :8080
📡 WebSocket: ws://localhost:8080/ws/{tableId}
💚 Health check: http://localhost:8080/health
```

## 🎮 Opciones de Testing

### **Opción A: Cliente Go Interactivo (Recomendado)**

#### 1. Iniciar múltiples clientes
```bash
# Terminal 1 - Alice
cd test-client && go run main.go Alice table1

# Terminal 2 - Bob  
cd test-client && go run main.go Bob table1

# Terminal 3 - Charlie
cd test-client && go run main.go Charlie table1

# Terminal 4 - Diana
cd test-client && go run main.go Diana table1
```

#### 2. Escenario de Prueba Completo

**En Terminal 1 (Alice):**
```
> create_tournament weekly "Weekly Tournament" 100 standard
> register weekly
> list_tournaments
```

**En Terminal 2 (Bob):**
```
> register weekly
> tournament_info weekly
```

**En Terminal 3 (Charlie):**
```
> register weekly
> state
```

**En Terminal 4 (Diana):**
```
> register weekly
> tournament_info weekly
```

**De vuelta en Terminal 1 (Alice):**
```
> start_tournament weekly
> state
```

**Ahora todos pueden jugar poker:**
```
> call
> raise 50
> fold
> all_in
> state
```

### **Opción B: Cliente Web**

1. Abre `web/poker-test.html` en tu navegador
2. Abre múltiples pestañas con diferentes nombres de jugador
3. Sigue el mismo escenario usando la interfaz web

### **Opción C: wscat (Para usuarios avanzados)**

```bash
# Instalar wscat
npm install -g wscat

# Conectar
wscat -c ws://localhost:8080/ws/table1

# Crear torneo
{"type":"tournament_create","version":1,"payload":{"tournament_id":"test1","tournament_name":"Test Tournament","buy_in":100,"tournament_type":"standard"}}

# Registrarse
{"type":"tournament_register","version":1,"payload":{"tournament_id":"test1","player":"Alice"}}
```

## 🎯 Escenarios de Prueba Específicos

### **A. Torneos Básicos**

1. **Creación y Registro:**
   - Crear torneo con diferentes tipos (standard/turbo)
   - Registrar 4+ jugadores
   - Verificar prize pool y player count
   - Intentar registrar jugador duplicado (debería fallar)

2. **Inicio de Torneo:**
   - Iniciar con mínimo de jugadores
   - Verificar que se crean mesas balanceadas
   - Comprobar que se colocan blinds automáticamente

### **B. Poker Real**

1. **Acciones Básicas:**
   - Call, raise, fold, all-in
   - Verificar turnos correctos
   - Comprobar distribución de pot

2. **Fases del Juego:**
   - Preflop → Flop → Turn → River → Showdown
   - Verificar cartas comunitarias
   - Comprobar evaluación de manos

### **C. Funcionalidades Avanzadas**

1. **Blinds Progresivos:**
   - Esperar 10 minutos (standard) o 5 minutos (turbo)
   - Verificar que blinds suben automáticamente
   - Comprobar que todas las mesas se actualizan

2. **Eliminación de Jugadores:**
   - Jugar hasta que alguien pierda todas las fichas
   - Verificar eliminación automática
   - Comprobar reorganización de mesas

## 📊 Comandos Útiles Durante las Pruebas

### **Cliente Go:**
```
help              # Mostrar todos los comandos
state             # Ver estado actual de la mesa
list_tournaments  # Ver todos los torneos
tournament_info <id> # Info específica de torneo
call              # Hacer call
raise 50          # Raise 50 fichas
fold              # Retirarse
all_in            # Apostar todo
quit              # Salir
```

### **Cliente Web:**
- Usar botones de la interfaz
- Ver log de mensajes en tiempo real
- Estado del juego se actualiza automáticamente

## 🔍 Verificaciones Importantes

### **Durante el Registro:**
- ✅ Players count incrementa correctamente
- ✅ Prize pool = players × buy-in
- ✅ Estado del torneo = "registering"
- ✅ Error al registrar jugador duplicado

### **Al Iniciar Torneo:**
- ✅ Estado cambia a "active"
- ✅ Se crean mesas balanceadas
- ✅ Blinds se colocan automáticamente
- ✅ Turnos rotan correctamente

### **Durante el Poker:**
- ✅ Solo el jugador actual puede actuar
- ✅ Pot se actualiza correctamente
- ✅ Cartas se reparten aleatoriamente
- ✅ Hand evaluator determina ganadores reales
- ✅ Stacks se actualizan tras cada mano

### **Blinds Progresivos:**
- ✅ Timer funciona correctamente
- ✅ Blinds suben en todas las mesas
- ✅ Antes se agregan en niveles altos

## 🚨 Problemas Comunes y Soluciones

### **"Connection refused"**
- Verificar que el servidor esté corriendo
- Comprobar puerto 8080
- Revisar que Redis esté disponible

### **"Player already at table"**
- Usar nombres de jugador únicos
- Cerrar conexiones anteriores

### **"Not your turn"**
- Verificar el campo `current_player` en el estado
- Respetar turnos rotativos

### **Comandos no funcionan**
- Verificar sintaxis exacta
- Usar nombres de torneo correctos
- Comprobar que estás registrado en el torneo

## 📈 Testing de Performance

### **Carga de Jugadores:**
```bash
# Crear múltiples clientes simultáneos
for i in {1..10}; do
  cd test-client && go run main.go "Player$i" "table$i" &
done
```

### **Múltiples Torneos:**
```bash
# En diferentes terminales, crear varios torneos
create_tournament t1 "Tournament 1" 50 standard
create_tournament t2 "Tournament 2" 100 turbo  
create_tournament t3 "Tournament 3" 200 standard
```

## 🎭 Escenarios de Edge Cases

1. **Desconexión de Jugadores:**
   - Cerrar terminal durante partida
   - Verificar que otros jugadores continúan

2. **All-in Scenarios:**
   - Múltiples all-ins
   - Side pots
   - Evaluación correcta de manos

3. **Mesa Final:**
   - Reducir jugadores hasta ≤6
   - Verificar transición a mesa final

## ✅ Checklist de Testing Completo

- [ ] Crear torneos standard y turbo
- [ ] Registrar múltiples jugadores
- [ ] Iniciar torneo correctamente
- [ ] Jugar varias manos de poker
- [ ] Verificar blinds progresivos
- [ ] Comprobar eliminación de jugadores
- [ ] Probar todas las acciones de poker
- [ ] Verificar evaluación de manos
- [ ] Testing con cliente web
- [ ] Prueba de múltiples torneos simultáneos

¡Con esta guía tienes todo lo necesario para probar completamente tu sistema de torneos de poker! 🎉