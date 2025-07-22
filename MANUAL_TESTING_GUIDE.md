# 🧪 **GUÍA EXHAUSTIVA DE PRUEBAS MANUALES - TEXAS HOLD'EM**

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
11. [**Checklist Final**](#checklist-final)

---

## 🔧 **PREPARACIÓN DEL ENTORNO**

### **1. Iniciar el Servidor**
```bash
cd /Users/zkcaleb/Documents/Blind\ Ledger/Code/blind-ledger-core-backend
go run cmd/server/main.go
```

**✅ Verificar:**
- Console muestra: `🚀 Servidor escuchando en :8080`
- `curl http://localhost:8080/health` responde `OK`

### **2. Abrir Interfaces de Prueba**
Abrir **múltiples pestañas** del navegador con:
- **Pestaña 1:** `http://localhost:8080/web/side-pots-test.html`
- **Pestaña 2:** `http://localhost:8080/web/poker-multi-test.html`
- **Pestaña 3:** `http://localhost:8080/web/poker-test.html` (si existe)

**✅ Verificar:**
- Todas las interfaces cargan correctamente
- Los logs muestran "Ready" en cada interfaz

---

## 🔗 **PRUEBAS BÁSICAS DE CONEXIÓN**

### **Test 1.1: Conexión Simple**
**Mesa:** `basic_connection_test`

**Pasos:**
1. Conectar Alice
2. Verificar estado: `Connected`
3. Desconectar Alice (cerrar pestaña)
4. Reconectar Alice

**✅ Resultados Esperados:**
- Alice aparece como "Connected" al conectarse
- Alice aparece como "Disconnected" al desconectarse
- Alice puede reconectarse sin problemas

### **Test 1.2: Múltiples Conexiones**
**Mesa:** `multi_connection_test`

**Pasos:**
1. Conectar Alice, Bob, Carol simultáneamente
2. Verificar que cada uno tenga estado independiente
3. Desconectar Bob solamente
4. Verificar que Alice y Carol siguen conectados

**✅ Resultados Esperados:**
- Cada jugador mantiene estado independiente
- Desconexión de uno no afecta a otros
- Los logs del servidor muestran conexiones/desconexiones correctas

---

## 🏠 **PRUEBAS DE LOBBY SYSTEM**

### **Test 2.1: Creación de Lobby**
**Mesa:** `lobby_test`

**Pasos:**
1. Alice se conecta (automáticamente es HOST)
2. Verificar que Alice tiene indicador de HOST
3. Bob se conecta
4. Verificar que Bob NO es HOST

**✅ Resultados Esperados:**
- Alice tiene borde naranja (host)
- Bob tiene borde normal
- Mesa en fase `lobby`

### **Test 2.2: Sistema Ready**
**Mesa:** `ready_test`

**Pasos:**
1. Conectar Alice (HOST) y Bob
2. Alice marca Ready → Botón "Start Game" se habilita
3. Bob marca Ready
4. Alice desmarca Ready → Botón "Start Game" se deshabilita
5. Bob marca Ready de nuevo
6. Alice marca Ready de nuevo

**✅ Resultados Esperados:**
- Solo HOST puede ver botón "Start Game"
- Botón solo se habilita cuando TODOS están ready
- Estado ready se sincroniza entre interfaces

### **Test 2.3: Inicio de Juego**
**Mesa:** `game_start_test`

**Pasos:**
1. Alice y Bob se conectan y marcan ready
2. Alice presiona "Start Game"
3. Verificar transición a fase `preflop`
4. Verificar que ambos jugadores tienen 2 cartas
5. Verificar que se colocaron blinds

**✅ Resultados Esperados:**
- Fase cambia de `lobby` a `preflop`
- Cada jugador recibe exactamente 2 cartas
- Small blind (10) y Big blind (20) están colocados
- El jugador correcto tiene el turno

---

## 👥 **PRUEBAS DE HEADS-UP (2 JUGADORES)**

### **Test 3.1: Distribución Inicial**
**Mesa:** `headsup_basic`

**Pasos:**
1. Alice y Bob → Ready → Start Game
2. Verificar posiciones del dealer button
3. Verificar blinds: Dealer = Small Blind, otro = Big Blind
4. Verificar turno inicial (Dealer actúa primero preflop)

**✅ Resultados Esperados:**
- En heads-up: Dealer = Small blind
- Dealer actúa primero en preflop
- Big blind actúa primero post-flop

### **Test 3.2: Ronda Completa Heads-up**
**Mesa:** `headsup_complete`

**Pasos:**
1. Alice (dealer/SB) y Bob (BB) inician
2. Alice call → Bob check → Flop aparece
3. Bob check → Alice bet 40 → Bob call → Turn aparece  
4. Bob check → Alice check → River aparece
5. Bob bet 60 → Alice call → Showdown

**✅ Resultados Esperados:**
- Flop: exactamente 3 cartas comunitarias
- Turn: total 4 cartas comunitarias  
- River: total 5 cartas comunitarias
- Showdown muestra ambas manos
- Ganador correcto recibe el pot

### **Test 3.3: All-in Heads-up**
**Mesa:** `headsup_allin`

**Pasos:**
1. Alice y Bob inician
2. Alice all-in (1000 fichas)
3. Bob call o fold
4. Si call: cartas se revelan inmediatamente
5. Verificar distribución correcta

**✅ Resultados Esperados:**
- All-in salta directamente a showdown si hay call
- Ganador recibe todo el pot
- Perdedor queda con 0 fichas

---

## 🎯 **PRUEBAS DE 3 JUGADORES**

### **Test 4.1: Rotación de Blinds**
**Mesa:** `three_players_blinds`

**Pasos:**
1. Alice, Bob, Carol → Start Game
2. Anotar quién es SB, BB, y UTG (Under The Gun)
3. Completar mano hasta showdown
4. **MANO 2:** Verificar que blinds rotaron correctamente
5. **MANO 3:** Verificar rotación completa

**✅ Resultados Esperados:**
- **Mano 1:** Alice=Dealer, Bob=SB, Carol=BB
- **Mano 2:** Bob=Dealer, Carol=SB, Alice=BB  
- **Mano 3:** Carol=Dealer, Alice=SB, Bob=BB

### **Test 4.2: Orden de Acción**
**Mesa:** `three_players_action`

**Pasos:**
1. Alice (Dealer), Bob (SB), Carol (BB)
2. **Preflop:** UTG (Alice) actúa primero
3. **Post-flop:** SB (Bob) actúa primero
4. Verificar orden en cada fase

**✅ Resultados Esperados:**
- **Preflop:** Alice → Bob → Carol
- **Post-flop:** Bob → Carol → Alice

### **Test 4.3: Jugador Fold**
**Mesa:** `three_players_fold`

**Pasos:**
1. Alice, Bob, Carol inician
2. Alice fold → Solo Bob y Carol continúan
3. Verificar que Alice no recibe más cartas comunitarias
4. Verificar que solo Bob y Carol pueden ganar

**✅ Resultados Esperados:**
- Alice sale de la mano inmediatamente
- Solo Bob y Carol compiten por el pot
- Alice no participa en showdown

---

## 🎪 **PRUEBAS DE 4+ JUGADORES**

### **Test 5.1: Mesa Completa (4 Jugadores)**
**Mesa:** `four_players_full`

**Pasos:**
1. Alice (Dealer), Bob (SB), Carol (BB), Dave (UTG)
2. Verificar orden de acción preflop: Dave → Alice → Bob → Carol
3. Verificar orden post-flop: Bob → Carol → Dave → Alice
4. Completar mano completa

**✅ Resultados Esperados:**
- Orden preflop: UTG → Dealer → SB → BB
- Orden post-flop: SB → BB → UTG → Dealer
- Todos los jugadores activos reciben cartas

### **Test 5.2: Múltiples Folds**
**Mesa:** `four_players_folds`

**Pasos:**
1. 4 jugadores inician
2. Dave fold, Alice fold → Solo Bob y Carol
3. Verificar que el juego continúa normalmente
4. Verificar que solo los jugadores activos compiten

**✅ Resultados Esperados:**
- Jugadores foldeados no participan más
- El juego continúa con jugadores restantes
- Solo jugadores activos pueden ganar

### **Test 5.3: Un Solo Ganador por Fold**
**Mesa:** `four_players_one_winner`

**Pasos:**
1. 4 jugadores inician
2. 3 jugadores hacen fold → Solo 1 queda
3. Verificar que el jugador restante gana automáticamente
4. Verificar que recibe todo el pot sin showdown

**✅ Resultados Esperados:**
- Mano termina inmediatamente con 1 jugador
- No hay showdown
- Ganador recibe todo el pot

---

## 🃏 **PRUEBAS DE EVALUACIÓN DE MANOS**

### **Test 6.1: Royal Flush vs Straight Flush**
**Mesa:** `hand_evaluation_royal`

**Configurar escenario:**
- **Cartas comunitarias:** 10♠ J♠ Q♠ K♠ A♠
- **Alice:** 9♠ 8♠ (Straight Flush)
- **Bob:** A♥ K♥ (Royal Flush usando comunitarias)

**✅ Resultado Esperado:**
- Bob gana con Royal Flush
- Alice tiene Straight Flush (segunda mejor mano)

### **Test 6.2: Four of a Kind vs Full House**
**Mesa:** `hand_evaluation_quads`

**Configurar escenario:**
- **Cartas comunitarias:** A♠ A♥ A♦ K♠ K♥
- **Alice:** A♣ Q♠ (Four of a Kind - Aces)
- **Bob:** K♦ Q♥ (Full House - Kings over Aces)

**✅ Resultado Esperado:**
- Alice gana con Four of a Kind
- Bob tiene Full House

### **Test 6.3: Empates - Split Pot**
**Mesa:** `hand_evaluation_tie`

**Configurar escenario:**
- **Cartas comunitarias:** A♠ K♠ Q♠ J♠ 10♥
- **Alice:** 2♦ 3♦ (Straight A-K-Q-J-10)
- **Bob:** 4♣ 5♣ (Straight A-K-Q-J-10)

**✅ Resultado Esperado:**
- Empate perfecto
- Pot se divide equitativamente entre Alice y Bob

### **Test 6.4: High Card vs Pair**
**Mesa:** `hand_evaluation_basic`

**Configurar escenario:**
- **Cartas comunitarias:** 2♠ 7♥ 9♦ J♠ K♥
- **Alice:** A♦ Q♣ (High Card - Ace)
- **Bob:** 2♣ 3♠ (Pair of 2s)

**✅ Resultado Esperado:**
- Bob gana con Pair of 2s
- Alice tiene solo High Card

### **Test 6.5: Straight con As bajo (Wheel)**
**Mesa:** `hand_evaluation_wheel`

**Configurar escenario:**
- **Cartas comunitarias:** A♠ 2♥ 3♦ 4♠ 5♥
- **Alice:** 6♦ 7♣ (Straight 7-high)
- **Bob:** 8♣ 9♠ (Straight 5-high / Wheel)

**✅ Resultado Esperado:**
- Alice gana con Straight 7-high
- Bob tiene Wheel (A-2-3-4-5)

### **Test 6.6: Flush con diferentes kickers**
**Mesa:** `hand_evaluation_flush`

**Configurar escenario:**
- **Cartas comunitarias:** A♠ K♠ Q♠ 7♠ 2♠
- **Alice:** J♠ 10♠ (Flush con J-high kicker)
- **Bob:** 9♠ 8♠ (Flush con 9-high kicker)

**✅ Resultado Esperado:**
- Alice gana con Flush J-high
- Bob tiene Flush 9-high

### **Test 6.7: Two Pair con diferentes kickers**
**Mesa:** `hand_evaluation_twopair`

**Configurar escenario:**
- **Cartas comunitarias:** A♠ A♥ K♦ K♠ 7♥
- **Alice:** Q♦ J♣ (Two Pair A's and K's, Q kicker)
- **Bob:** 10♣ 9♠ (Two Pair A's and K's, 10 kicker)

**✅ Resultado Esperado:**
- Alice gana con mejor kicker (Queen)
- Bob tiene mismo two pair pero kicker inferior

---

## 💰 **PRUEBAS DE SIDE POTS (ALL-INS)**

### **Test 7.1: Side Pot Básico (2 All-ins)**
**Mesa:** `sidepot_basic_two`

**Escenario:**
- **Alice:** 100 fichas → All-in 100
- **Bob:** 500 fichas → Call 100  

**Pasos:**
1. Alice all-in 100
2. Bob call 100
3. Ver showdown y distribución

**✅ Resultado Esperado:**
- **Pot único:** 200 fichas (100 + 100)
- Ganador recibe las 200 fichas completas

### **Test 7.2: Side Pot Complejo (3 All-ins diferentes)**
**Mesa:** `sidepot_complex_three`

**Escenario:**
- **Alice:** 100 fichas → All-in 100
- **Bob:** 500 fichas → All-in 500
- **Carol:** 1000 fichas → Call 500

**Pasos:**
1. Alice all-in 100
2. Bob all-in 500  
3. Carol call 500
4. Verificar side pots en showdown

**✅ Resultado Esperado:**
- **Side Pot 1:** 300 fichas (100×3) → Alice puede ganar
- **Side Pot 2:** 800 fichas (400×2) → Solo Bob y Carol pueden ganar
- **Total:** 1100 fichas

### **Test 7.3: Ganador Diferente por Side Pot**
**Mesa:** `sidepot_different_winners`

**Escenario:**
- **Alice:** 100 fichas, mano fuerte → All-in 100
- **Bob:** 500 fichas, mano débil → All-in 500
- **Carol:** 1000 fichas, mano media → Call 500

**Configurar para que:**
- Alice gane Side Pot 1 (mano más fuerte)
- Carol gane Side Pot 2 (mejor que Bob)

**✅ Resultado Esperado:**
- Alice gana 300 fichas
- Carol gana 800 fichas  
- Bob gana 0 fichas

### **Test 7.4: Side Pot con Empate**
**Mesa:** `sidepot_tie`

**Escenario:**
- **Alice:** 100 fichas → All-in 100 (mano fuerte)
- **Bob:** 500 fichas → All-in 500 (mano fuerte, igual que Alice)
- **Carol:** 1000 fichas → Call 500 (mano débil)

**✅ Resultado Esperado:**
- **Side Pot 1:** 300 fichas → Alice y Bob empatan, se divide (150 cada uno)
- **Side Pot 2:** 800 fichas → Bob gana completo

### **Test 7.5: Multiple Side Pots (4+ jugadores)**
**Mesa:** `sidepot_multiple_four`

**Escenario:**
- **Alice:** 50 fichas → All-in 50
- **Bob:** 200 fichas → All-in 200
- **Carol:** 500 fichas → All-in 500
- **Dave:** 1000 fichas → Call 500

**✅ Resultado Esperado:**
- **Side Pot 1:** 200 fichas (50×4) → Todos pueden ganar
- **Side Pot 2:** 600 fichas (150×4) → Bob, Carol, Dave pueden ganar
- **Side Pot 3:** 900 fichas (300×3) → Carol, Dave pueden ganar
- **Total:** 1700 fichas

---

## 🔄 **PRUEBAS DE AUTO-RESTART**

### **Test 8.1: Auto-restart Habilitado**
**Mesa:** `auto_restart_enabled`

**Pasos:**
1. 2-3 jugadores juegan hasta showdown
2. **NO tocar nada** después del showdown
3. Esperar 5 segundos
4. Verificar que nueva mano inicia automáticamente

**✅ Resultado Esperado:**
- A los 5 segundos: nueva mano inicia
- Nuevas cartas repartidas
- Blinds rotados correctamente
- Dealer button se mueve

### **Test 8.2: Auto-restart con Jugador Eliminado**
**Mesa:** `auto_restart_elimination`

**Pasos:**
1. Alice (100 fichas) vs Bob (1000 fichas)
2. Alice pierde todo en all-in → 0 fichas
3. Esperar 5 segundos
4. Verificar qué sucede

**✅ Resultado Esperado:**
- Auto-restart NO ocurre (menos de 2 jugadores con fichas)
- Mesa permanece en showdown
- Alice eliminada no participa

### **Test 8.3: Auto-restart con Side Pots**
**Mesa:** `auto_restart_sidepots`

**Pasos:**
1. Juego con side pots complejos
2. Completar showdown
3. Esperar auto-restart
4. Verificar nueva mano

**✅ Resultado Esperado:**
- Side pots se limpian correctamente
- Nueva mano inicia con stacks actualizados
- Blinds se calculan con fichas actuales

### **Test 8.4: Auto-restart después de Eliminación Múltiple**
**Mesa:** `auto_restart_multiple_elimination`

**Pasos:**
1. 4 jugadores: Alice (100), Bob (100), Carol (100), Dave (1000)
2. Alice, Bob, Carol pierden todo contra Dave
3. Esperar 5 segundos

**✅ Resultado Esperado:**
- Auto-restart NO ocurre (solo Dave tiene fichas)
- Mesa permanece en showdown
- Dave declarado ganador único

### **Test 8.5: Auto-restart con Nuevos Stacks**
**Mesa:** `auto_restart_new_stacks`

**Pasos:**
1. Alice (500) vs Bob (1500) → Alice gana y queda con 1000
2. Bob queda con 1000
3. Esperar auto-restart
4. Verificar blinds con nuevos stacks

**✅ Resultado Esperado:**
- Nueva mano inicia con stacks: Alice=1000, Bob=1000
- Blinds se calculan apropiadamente
- Dealer button rota correctamente

---

## ⚠️ **PRUEBAS DE EDGE CASES**

### **Test 9.1: Jugador Sin Fichas para Blinds**
**Mesa:** `edge_case_no_chips_blinds`

**Escenario:**
- Alice tiene 5 fichas, debe pagar Big Blind (20)

**✅ Resultado Esperado:**
- Alice hace all-in automático por 5 fichas
- Se crea side pot apropiado

### **Test 9.2: Todos All-in Preflop**
**Mesa:** `edge_case_all_allin_preflop`

**Pasos:**
1. Todos los jugadores hacen all-in en preflop
2. Verificar que todas las cartas comunitarias salen inmediatamente
3. Verificar showdown directo

**✅ Resultado Esperado:**
- Flop, Turn, River salen inmediatamente
- Showdown inmediato
- Side pots calculados correctamente

### **Test 9.3: Un Solo Jugador con Fichas**
**Mesa:** `edge_case_one_player_left`

**Pasos:**
1. Jugar hasta que solo 1 jugador tenga fichas
2. Verificar que auto-restart no ocurre
3. Verificar estado final

**✅ Resultado Esperado:**
- Mesa permanece en showdown
- No hay auto-restart
- Ganador único con todas las fichas

### **Test 9.4: Desconexión Durante Turno**
**Mesa:** `edge_case_disconnect_turn`

**Pasos:**
1. Alice tiene el turno
2. Cerrar pestaña de Alice (desconexión)
3. Verificar que pasa automáticamente al siguiente jugador

**✅ Resultado Esperado:**
- Alice hace fold automático
- Turno pasa al siguiente jugador
- Juego continúa normalmente

### **Test 9.5: Desconexión del Host**
**Mesa:** `edge_case_host_disconnect`

**Pasos:**
1. Alice (HOST) y Bob conectados en lobby
2. Alice se desconecta
3. Verificar qué pasa con el juego

**✅ Resultado Esperado:**
- Bob puede continuar o se asigna nuevo host
- Mesa no se rompe por desconexión de host

### **Test 9.6: Reconexión Durante Mano**
**Mesa:** `edge_case_reconnect_during_hand`

**Pasos:**
1. Alice, Bob jugando en preflop
2. Alice se desconecta (fold automático)
3. Alice se reconecta durante flop
4. Verificar estado

**✅ Resultado Esperado:**
- Alice permanece foldeada en esta mano
- Alice puede participar en próxima mano
- Estado consistente tras reconexión

### **Test 9.7: Small Blind Mayor que Stack**
**Mesa:** `edge_case_sb_greater_stack`

**Escenario:**
- Small blind = 10, Alice tiene 7 fichas, es SB

**✅ Resultado Esperado:**
- Alice all-in por 7 fichas como small blind
- Big blind normal de 20
- Side pot apropiado creado

### **Test 9.8: Big Blind Mayor que Stack**
**Mesa:** `edge_case_bb_greater_stack`

**Escenario:**
- Big blind = 20, Bob tiene 15 fichas, es BB

**✅ Resultado Esperado:**
- Bob all-in por 15 fichas como big blind
- Otros jugadores pueden call 15 o raise apropiadamente
- Side pot creado si hay raise

---

## ✅ **CHECKLIST FINAL DE VALIDACIÓN**

### **🎯 Funcionalidades Core**
- [ ] **Conexión/Desconexión:** Jugadores se conectan y desconectan sin problemas
- [ ] **Lobby System:** Ready/Start funciona correctamente
- [ ] **Repartición de Cartas:** 2 cartas por jugador, 5 comunitarias progresivas
- [ ] **Blinds:** Small y Big blind se colocan y rotan correctamente
- [ ] **Betting Rounds:** Call, raise, check, fold, all-in funcionan
- [ ] **Phase Progression:** Preflop → Flop → Turn → River → Showdown

### **🃏 Evaluación de Manos**
- [ ] **Royal Flush:** Detectado y gana vs todas las demás
- [ ] **Straight Flush:** Detectado y ordenado correctamente
- [ ] **Four of a Kind:** Detectado, kicker correcto
- [ ] **Full House:** Detectado, trips vs pair correcto
- [ ] **Flush:** Detectado, high card correcto
- [ ] **Straight:** Detectado, incluyendo A-2-3-4-5
- [ ] **Three of a Kind:** Detectado, kickers correctos
- [ ] **Two Pair:** Detectado, kicker correcto
- [ ] **One Pair:** Detectado, kickers correctos
- [ ] **High Card:** Funciona como último recurso
- [ ] **Empates:** Pot se divide correctamente

### **💰 Side Pots**
- [ ] **2 All-ins:** Pot único calculado correctamente
- [ ] **3+ All-ins:** Side pots múltiples calculados
- [ ] **Distribución:** Ganadores correctos por side pot
- [ ] **Edge Cases:** All-in menor que blinds manejado
- [ ] **Empates en Side Pots:** División correcta por pot

### **🔄 Auto-restart**
- [ ] **Restart Normal:** Ocurre después de 5 segundos
- [ ] **Sin Jugadores:** No restart con menos de 2 jugadores con fichas
- [ ] **Limpieza:** Side pots y estados se reinician correctamente
- [ ] **Rotación:** Dealer button y blinds rotan

### **🎮 Scenarios Multi-jugador**
- [ ] **Heads-up (2):** Funciona perfectamente
- [ ] **3 Jugadores:** Blinds y orden correcto
- [ ] **4+ Jugadores:** Escalable sin problemas
- [ ] **Eliminaciones:** Jugadores sin fichas manejados

### **⚠️ Edge Cases**
- [ ] **Desconexiones:** Fold automático durante turno
- [ ] **Sin Fichas Blinds:** All-in automático
- [ ] **Todos All-in:** Cartas salen inmediatamente
- [ ] **Un Ganador:** Game termina apropiadamente
- [ ] **Host Disconnect:** Transferencia o manejo apropiado
- [ ] **Reconexión:** Estado consistente

### **🔧 Performance & Stability**
- [ ] **No Memory Leaks:** Múltiples manos no degradan performance
- [ ] **Concurrent Tables:** Múltiples mesas funcionan independientemente
- [ ] **Error Handling:** Errores no rompen el sistema
- [ ] **Logs Útiles:** Server logs permiten debugging

---

## 🏆 **CRITERIOS DE ÉXITO**

### **✅ SISTEMA APROBADO SI:**
1. **Todas las secciones** del checklist pasan al 100%
2. **No hay bugs críticos** que rompan el flujo del juego  
3. **Side pots** se calculan y distribuyen matemáticamente correctos
4. **Evaluación de manos** es 100% precisa según reglas de Texas Hold'em
5. **Auto-restart** mantiene flujo continuo apropiadamente
6. **Multi-jugador** funciona desde 2 hasta 10+ jugadores sin problemas
7. **Edge cases** se manejan graciosamente sin crashes

### **❌ SISTEMA NECESITA REVISIÓN SI:**
- Cualquier evaluación de mano es incorrecta
- Side pots calculan mal las distribuciones (diferencia > 0)
- Auto-restart no funciona o causa problemas de estado
- Desconexiones rompen el juego permanentemente
- Blinds no rotan correctamente en múltiples manos
- Performance degrada significativamente con múltiples manos
- Crashes o errores no manejados

### **⚠️ SISTEMA NECESITA MEJORAS SI:**
- Funciona correctamente pero UX no es óptima
- Logs insuficientes para debugging
- Performance lenta pero funcional
- Edge cases mínimos no cubiertos

---

## 📝 **FORMATO DE REPORTE DE BUGS**

Para cada bug encontrado, usar el siguiente formato:

```markdown
### **Bug ID:** BUG-001
**Test:** Test 7.2 - Side Pot Complejo
**Severity:** CRÍTICO
**Status:** ABIERTO

**Pasos para Reproducir:**
1. Alice all-in 100 fichas
2. Bob all-in 500 fichas  
3. Carol call 500 fichas
4. Verificar showdown

**Resultado Esperado:**
- Side Pot 1: 300 fichas
- Side Pot 2: 800 fichas
- Total: 1100 fichas

**Resultado Actual:**
- Side Pot 1: 250 fichas ❌
- Side Pot 2: 850 fichas ❌
- Total: 1100 fichas ✅

**Logs del Servidor:**
```
2025-01-XX XX:XX:XX Side pots created: 2
2025-01-XX XX:XX:XX Side pot 1: Amount=250, EligiblePlayers=[0,1,2]
2025-01-XX XX:XX:XX Side pot 2: Amount=850, EligiblePlayers=[1,2]
```

**Screenshot/Evidence:**
[Adjuntar screenshot del navegador mostrando el problema]

**Impacto:**
- Cálculo incorrecto afecta distribución de fichas
- Jugadores pueden perder/ganar fichas incorrectamente

**Prioridad de Fix:** ALTA
```

---

## 🎲 **CASOS DE PRUEBA ESPECÍFICOS DE POKER**

### **Scenario A: Bad Beat con Side Pot**
- Alice: 100 fichas, AA (pocket aces)
- Bob: 500 fichas, 23 offsuit  
- Board: A 2 2 2 K
- Resultado: Bob gana con Four 2s vs Alice Full House

### **Scenario B: Royal Flush vs Straight Flush**
- Board: 10♠ J♠ Q♠ K♠ A♠
- Alice: 9♠ 8♠ (straight flush)
- Bob: 5♥ 6♥ (royal flush usando board)

### **Scenario C: Wheel vs Straight**
- Board: A 2 3 4 5
- Alice: 6 7 (straight 5-high)
- Bob: K Q (wheel A-2-3-4-5)

### **Scenario D: Flush vs Flush**
- Board: A♠ K♠ Q♠ 7♠ 2♠
- Alice: J♠ 10♠ (nut flush)
- Bob: 9♠ 8♠ (flush, 9-high)

---

## 📊 **MÉTRICAS DE CALIDAD**

### **Cobertura Mínima Requerida:**
- **Conexiones:** 100% casos base + 90% edge cases
- **Lobby System:** 100% flujos principales
- **Evaluación Manos:** 100% todas las combinaciones
- **Side Pots:** 100% matemática correcta
- **Multi-jugador:** 95% scenarios 2-10 jugadores
- **Auto-restart:** 100% flujo principal + 85% edge cases

### **Performance Benchmarks:**
- **Latencia:** < 100ms para acciones de poker
- **Throughput:** > 100 acciones/segundo por mesa
- **Memory:** < 50MB para 10 mesas activas
- **CPU:** < 10% utilización en carga normal

---

**🎯 Con esta guía exhaustiva puedes validar que tu sistema Texas Hold'em funciona perfectamente según las reglas oficiales del poker y está listo para producción. ¡Éxito en las pruebas!**