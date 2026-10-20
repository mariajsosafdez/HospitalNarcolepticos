# Lógica del juego (capa `game`)

Especificación de **todo lo que el juego agrega a nivel de lógica** y de cómo usa el paquete `hospital`. Es la nota que hay que dominar para la defensa.

> Estado: **borrador**. Los números marcados *(propuesto)* se ajustan al probar. Las decisiones confirmadas están en [[Game Design#Registro de decisiones]].
	
---

## 0. Principios

1. **El paquete `hospital` no cambia ni una línea.** Lo exige el enunciado (sección 7) para el bonus de interfaz. Se puede comprobar con `git diff main -- hospital/`, que debe salir vacío.
2. **`game` tiene las reglas, la vista solo dibuja.** La ventana del juego lee el estado de `game` y le manda acciones ("asignar", "dar de alta"). No toma decisiones.
3. **El jugador es quien registra en el sistema.** El paquete `hospital` es "el sistema del hospital": solo se entera de las cosas cuando el jugador actúa. Por eso a veces el juego y el modelo muestran estados distintos (ver §3.1). Esa diferencia es intencional y se explica en la defensa.
4. **Determinista.** El tiempo avanza por *ticks* y el azar usa una semilla. Con la misma semilla, la partida es igual: eso permite tests y un modo Demo reproducible.
5. **Una sola goroutine para el juego.** Toda la lógica corre dentro del bucle del juego (`Update`), así que la capa `game` no necesita candados propios. El hospital igual tiene su `sync.Mutex`, que sigue protegiéndolo.

---

## 1. Paquetes

```
hospital/          modelo del enunciado            (SIN CAMBIOS)
game/              reglas del juego                (NUEVO)
  clock.go           reloj del turno
  patient.go         GamePatient: paciente + estado de juego
  staff.go           StaffMember: envoltorio que cumple hospital.Attender
  world.go           cámaras, movimiento, aparición de pacientes
  score.go           puntuación y estrellas
  shift.go           Shift: une todo, recibe acciones, avanza el tiempo
  days.go            configuración de cada día
  demo.go            guion del modo Demo
  game_test.go
gameview/          dibujo pixel art y entrada con Ebitengine (NUEVO)
main.go            + abre la ventana del juego al final (cambio mínimo)
web/               como está hoy                     (SIN CAMBIOS)
```

✅ **Motor: Ebitengine** (`github.com/hajimehoshi/ebiten/v2`). Es la única dependencia de terceros, y el enunciado la permite porque es solo para la interfaz gráfica. Solo la importa `gameview`; `game` y `hospital` siguen usando únicamente la librería estándar.

✅ **Se abre con `go run .`** Nuevo orden de `main`:

1. Escenario obligatorio por consola (igual que hoy).
2. Simulación concurrente (igual que hoy).
3. *(Solo con `go run . -web`)* ✅ El dashboard web, como antes, y **sin** juego. Con `go run .` no hay dashboard.
4. Se abre la ventana del juego con `ebiten.RunGame`. Ebitengine exige correr en la goroutine principal, por eso va de último.

`main` sigue sin lógica de negocio: solo arranca cosas. Al cerrar la ventana del juego, el programa termina.

**Sprites sin Aseprite** ✅: cada sprite se describe como una cuadrícula de texto (un carácter = un color de la paleta) y un script pequeño en Go (`tools/sprites`) la convierte en PNG. Así el arte queda en el repo como texto, se revisa en git y se puede regenerar. Si quieres retocar a mano, LibreSprite y Piskel son editores pixel art gratuitos que abren esos PNG.

---

## 2. Reloj del turno (`clock.go`)

| Concepto | Valor |
|---|---|
| Tick | 1/60 s real |
| Turno | 12:00 → 18:00 de juego = 180 s reales = 10 800 ticks |
| 1 tick | 2 s de juego |
| 1 min de juego | 0,5 s real (30 ticks) |
| 1 h de juego | 30 s reales |

- `Clock.Advance()`: suma un tick, salvo si el juego está en pausa.
- `Clock.Now()`: hora de juego (por ejemplo `14:15`).
- `Clock.Over()`: `true` al llegar a las 18:00.

**Ojo con las horas del reporte.** `EpisodeRecord` guarda `time.Now()` real. El juego guarda por su lado la hora de juego de cada episodio (un `map` de ID de episodio → hora de juego) y la usa al mostrar el reporte. No se toca `EpisodeRecord`.

---

## 3. Pacientes (`patient.go`)

`GamePatient` contiene un `*hospital.Patient` y le agrega lo que solo existe en el juego: cámara actual, posición en pantalla, estado de juego, cronómetro de espera, hora en que debe despertar y sprite.

### 3.1 Máquina de estados

| Estado de juego | Estado en `hospital` | Qué ve el jugador | Pasa a… |
|---|---|---|---|
| `Arriving` | `Awake` (recién admitido con `AdmitPatient`) | Entra por Recepción (CAM 1A) | `Wandering` |
| `Wandering` | `Awake` | Camina entre cámaras | `AsleepWaiting` cuando le da el ataque de sueño |
| `AsleepWaiting` | **`Awake`** ⚠️ | Dormido en el suelo/silla, "zzz". Corre su espera | `Carried` cuando el jugador asigna personal; `NoBedWaiting` si no había cama |
| `NoBedWaiting` | `AsleepInHallway` | Ya fue atendido pero no había cama: sigue dormido **en la misma cámara donde se durmió**. La espera sigue | `Carried` cuando hay cama y el jugador lo manda (manual) |
| `Carried` | `AsleepInBed` | El personal se lo lleva en camilla (unos segundos) | `InBed` |
| `InBed` | `AsleepInBed` | Durmiendo en su habitación | `AwakeInBed` al llegar su hora de despertar |
| `AwakeInBed` | **`AsleepInBed`** ⚠️ | Sentado en la cama, despierto. Aparece el botón "Dar de alta" | `Wandering` cuando el jugador le da de alta |
| `Gone` | `Awake` | Se fue del hospital | (fin) |

⚠️ **Las dos filas donde juego y modelo no coinciden:**

- **Dormido sin atender:** en el juego ya se durmió, pero el sistema no lo sabe hasta que el jugador lo reporta. `RegisterEpisode` exige que el paciente esté `Awake` en el modelo, y por eso se llama justo al asignar personal.
- **Despierto sin alta:** en el juego ya se despertó, pero para el sistema la cama sigue ocupada hasta que alguien le da de alta. Ahí se llama a `WakeUpPatient`, que libera la cama.

**Sobre el nombre `AsleepInHallway`:** en el modelo significa "dormido fuera de una cama", sea cual sea el lugar. Un paciente puede dormirse en **cualquier** zona común (cafetería, patio, sala de espera…), y si no hay cama se queda **ahí mismo**: el juego no lo lleva al pasillo. El lugar real está en `Patient.Location()`.

### 3.2 Movimiento

- Las cámaras forman un grafo: cada zona común conoce sus vecinas (ver [[Mapa de Cámaras]]).
- ✅ **Movimiento por ticks aleatorios** (decisión de la autora, para que sea más difícil saber dónde está cada paciente). Cada paciente tiene su propio `nextMoveTick`. Al llegar ese tick, se mueve a una cámara vecina (o se queda), y se sortea el siguiente: entre 5 y 25 min de juego *(propuesto)*, o sea entre 2,5 y 12,5 s reales. Así no se mueven todos a la vez ni a ritmo fijo.
- Dentro de la cámara también se mueven a puntos al azar (silla, mostrador, ventana…), para que no queden quietos.
- Los pacientes no entran a las habitaciones por su cuenta: solo llegan ahí en camilla.

### 3.3 Ataque de sueño

- Cada *N* minutos de juego, un paciente en `Wandering` tira el dado: `probabilidad = base(nivel) × modificador(zona)`.
- Base *(propuesto)*: Leve 3 %, Moderado 6 %, Severo 10 %.
- Modificador de zona *(propuesto)*: Cafetería ×2 (el sancocho), Patio ×1,5, el resto ×1.
- ✅ Un paciente recién dado de alta no puede dormirse otra vez durante **al menos 10 min de juego** (5 s), si no se ha ido del hospital.
- ✅ **Sin ataques nuevos en los últimos 10 s reales** (los últimos 20 min de juego, de 17:40 a 18:00; constante `noSleepFinalMinutes`, función `sleepWindowOpen`): "los pacientes no deberían dormirse faltando menos de 10 segundos para que se acabe el juego, porque sino seria garantizar una penalización" (la autora, 2026-10-18). La Demo no se ve afectada: no tiene sueño al azar.

### 3.4 Despertar

Al llegar a la cama se sortea cuánto va a dormir *(propuesto)*: Leve 30–60 min, Moderado 1–2 h, Severo 2–3 h de juego. Al cumplirse pasa a `AwakeInBed`. **No** se despierta solo en el pasillo.

### 3.5 Después del alta

✅ Al recibir el alta, el paciente **se queda o se va**:

- **Se queda:** vuelve a `Wandering` y puede dormirse otra vez (después de 10 min). Eso le da números interesantes al `SevereReport`.
- **Se va:** camina a Recepción (CAM 1A) y sale (`Gone`).

Probabilidad de irse *(propuesto)*: Leve 60 %, Moderado 40 %, Severo 20 %. Los severos se quedan más porque necesitan más cuidado. Eso también hace que tengan más episodios en el reporte.

---

## 4. Personal (`staff.go`)

### 4.1 El problema

`RegisterEpisode` despacha personal **solo**: recorre `[]Attender` por turnos (round-robin) y toma al primero cuyo `IsAvailable()` dé `true`. En el juego, **el jugador elige** a quién mandar.

### 4.2 La solución: `StaffMember`, un envoltorio que cumple `Attender`

El juego no contrata directamente al `*Doctor` ni al `*Orderly`: contrata un `StaffMember` que los **contiene** y cumple la misma interfaz.

```go
// StaffMember envuelve a un miembro del personal del hospital y le agrega
// lo que solo existe en el juego: el cooldown y la selección del jugador.
type StaffMember struct {
	inner         hospital.Attender // el *Doctor o *Orderly de verdad
	doctor        *hospital.Doctor  // nil si es camillero (para cupo y reporte 5.2)
	busyUntil     int               // tick en que termina traslado + descanso
	selected      bool              // true SOLO durante el despacho elegido
	now           func() int        // reloj del turno
}

func (s *StaffMember) ID() string   { return s.inner.ID() }
func (s *StaffMember) Name() string { return s.inner.Name() }
func (s *StaffMember) Role() string { return s.inner.Role() }

// IsAvailable: el hospital solo "ve" disponible al que eligió el jugador,
// y solo si no está en cooldown y su tipo real también está disponible
// (por ejemplo, un doctor con cupo lleno no lo está).
func (s *StaffMember) IsAvailable() bool {
	return s.selected && s.now() >= s.busyUntil && s.inner.IsAvailable()
}

// Attend delega en el Doctor u Orderly real (polimorfismo dentro de polimorfismo).
func (s *StaffMember) Attend(p *hospital.Patient, location string) (hospital.EpisodeRecord, error) {
	return s.inner.Attend(p, location)
}
```

Flujo de una asignación:

1. El jugador elige paciente y personal.
2. `game` valida (§5) y pone `selected = true` **solo** en ese `StaffMember`.
3. Llama `h.RegisterEpisode(p, location)`. Como solo uno está disponible, el round-robin lo elige a él.
4. Pone `selected = false` y le asigna su cooldown.

Por qué es seguro: todo pasa en la goroutine del juego, en orden, así que nadie más puede ver `selected = true`.

Por qué sirve en la defensa: es **otro tipo que cumple `Attender`** y que el hospital usa sin saber que existe. Es justo la pregunta del enunciado ("agregue un tipo nuevo que cumpla `Attender`").

### 4.3 Consecuencia a tener en cuenta

`HireStaff` usa *type assertion* `a.(*Doctor)` para llenar la lista de doctores. Como el juego contrata `*StaffMember`, **`h.Doctors()` saldrá vacío en el hospital del juego**. Por eso:

- El juego guarda su propio puntero al `*hospital.Doctor` (campo `doctor`) y para el reporte 5.2 llama directamente `doctor.MyEpisodes()`. La consulta del enunciado se sigue usando igual.
- Los episodios del historial guardan como `attender` al `*Doctor` real (porque `Doctor.Attend` crea el registro), no al envoltorio. El conteo por ID de `Snapshot` sigue funcionando.

Alternativa descartada: contratar al `*Doctor` directo. El hospital elegiría por turnos y el jugador no podría decidir.

### 4.4 Estados del personal (los que se ven en el panel)

| Estado | Condición |
|---|---|
| Disponible | `now >= busyUntil` y (si es doctor) cupo < 4 |
| En traslado | durante el viaje de la camilla |
| Descansando | después del traslado, hasta `busyUntil` |
| Sin cupo | doctor con 4/4 pacientes a cargo (no se puede elegir) |

| Tiempos *(propuesto)* | Doctor | Camillero |
|---|---|---|
| Traslado | 10 min de juego (5 s) | **6 min de juego (3 s)** ✅ |
| Descanso después | 30 min de juego (15 s) | 10 min de juego (5 s) |

✅ Tiempos aprobados, con el cambio "que el camillero sea de 3s, no de 5s". Lo apliqué al **traslado** del camillero; falta confirmar si también aplica a su descanso.

### 4.5 Cupo del doctor (regla del modelo)

- `maxPatientsPerDoctor = 4`. Cuenta **pacientes distintos a cargo**, no episodios. Nunca se liberan durante el turno.
- Si el doctor atiende a **su propio** paciente otra vez, **no** gasta cupo (`DiagnosePatient` devuelve `nil` sin agregar).
- Si atiende a un paciente que ya es de **otro** doctor, tampoco gasta cupo: el paciente sigue siendo del primero (`AttendSleepEmergency` solo diagnostica si no tiene doctor).
- El panel muestra `x/4` con `len(doctor.Patients())`.
- **Exceso de trabajo:** cuando un doctor llega a 4/4 se descuentan puntos **una sola vez** por doctor (§7). Así conviene repartir pacientes entre los doctores y usar a los camilleros.

---

## 5. Acciones del jugador (`shift.go`)

Cada acción es un método de `Shift` que devuelve `error` si no se puede hacer. La vista muestra ese error como un mensaje amable.

### `Assign(patientID, staffID)`

- Valida: el paciente está en `AsleepWaiting`, y el personal está Disponible.
- Hace el flujo de §4.2.
- Resultado:
  - `nil` → paciente en `Carried`, con cama.
  - `errors.Is(err, hospital.ErrNoRoomAvailable)` → el paciente queda en `NoBedWaiting`, en la misma cámara (y para el modelo, en `AsleepInHallway`). Se muestra el error legible, igual que en el escenario del enunciado. El personal **sí** atendió (el episodio quedó registrado con él), así que entra en cooldown.
  - `errors.Is(err, hospital.ErrNoStaffAvailable)` → no debería pasar porque se valida antes. Si pasa es un bug y lo atrapa un test.

### `MoveToBed(patientID, staffID)` ✅

- Para un paciente en `NoBedWaiting` cuando ya hay cama libre. **Siempre es manual**: el paciente no pasa solo a la cama que se libera.
- Necesita personal Disponible (cualquiera), que entra en cooldown corto.
- Llama `h.AssignRoom(p)`. **No** crea un episodio nuevo: es el mismo episodio, solo se le consiguió cama. Coincide con el paso 3 del escenario obligatorio.

### `Discharge(patientID)`

- Solo si el paciente está en `AwakeInBed`.
- Llama `h.WakeUpPatient(p)`, que libera la cama. El paciente vuelve a `Wandering`.

### Otras

`SwitchCamera(cam)`, `Pause()` / `Resume()`.

---

## 6. Días y aparición de pacientes (`days.go`, `world.go`)

**Regla de la autora:** el número de pacientes de cada partida no debe pasar mucho del cupo total de los doctores, para que los doctores se usen y los camilleros sean el apoyo.

Fórmula original: `pacientes = doctores × 4 + 2`. ✅ **Cambio:** el Elenco tiene 10 pacientes, así que **todos los días traen 10** ("okeyyy"). La dificultad sube por cámaras, camas y penalizaciones. Si sobra tiempo se agregan personajes y se sube a 12.

| Día | Doctores | Camilleros | Habitaciones | Cámaras comunes | Pacientes |
|---|---|---|---|---|---|
| 1 | 2 | 1 | 3 | 1A, 1B, 2, 3 | 10 |
| 2 | 2 | 1 | 4 | + 4 | 10 |
| 3 | 3 | 1 | 4 | + 5 | 10 |
| 4 | 3 | 2 | 5 | + 6 | 10 |
| 5 | 3 | 2 | 6 | todas | 10 |

- Los pacientes llegan repartidos entre 12:00 y ~16:00. Al llegar se llama `h.AdmitPatient(p)`.
- Reparto de niveles *(propuesto)*: 40 % Leve, 35 % Moderado, 25 % Severo (al menos 2 severos, como el enunciado).
- Nombres costeños de una lista fija.

**Easter eggs** (`eggs.go`, ✅ aprobados 2026-10-19): `RollEgg(seed)` es pura y da 15 % de que un turno normal (días 1 a 5; nunca la Demo ni "Cómo jugar") traiga **un** paciente extra de `castEggs` (E-001 Misifu, el gato, 6 años; E-002 Alejo Urrego y E-003 Samu Giraldo, Leves; E-004 Cote, Severo; E-005 Sebs, Moderado). Usa un azar aparte (`seed ^ eggSeedSalt`) para no cambiar el del turno; `DayConfig.WithEgg` lo agrega y llega a una hora al azar como los demás. **Misifu y Cote** (`hasFrequentSleep`) tienen 3 veces la probabilidad de dormirse (`frequentSleepFactor`) y no tienen la espera de 10 min tras el alta (`noSleepMinutesAfterDischarge`), pero respetan los últimos 20 min sin sueño. Al llegar el gato, la vista muestra `CatArrivalLine` de Don Ernesto (`Shift.ArrivedIDs`).

---

## 7. Puntuación (`score.go`)

**Espera** = desde que el paciente se duerme hasta que llega a una cama *(propuesto: si se queda en el pasillo, sigue contando)*.

| Evento | Puntos *(propuesto)* |
|---|---|
| Espera ≤ 20 min de juego | +100 (Excelente) |
| Espera ≤ 45 min | +70 (Bien) |
| Espera ≤ 90 min | +40 (Regular) |
| Espera > 90 min | +10 (Mal) |
| Severo atendido por doctor | +20 |
| Despierto en cama sin alta | −2 por cada 10 min de juego, después de 15 min de gracia |
| Doctor llega a 4/4 (exceso de trabajo) | −50, una vez por doctor |
| Paciente dormido sin atender al terminar el turno (`AsleepWaiting` o `NoBedWaiting`) | **−10** ✅ ("cambiemoslo mejor por -10, porque realmente no se atendieron"), con multiplicador del día |
| Paciente en camilla al terminar el turno | su espera se califica normal: sí fue atendido |

✅ Tabla aprobada.

**Las penalizaciones suben con el día** (decisión de la autora). Multiplicador de las filas que restan *(propuesto)*:

| Día | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| × penalización | 1,0 | 1,25 | 1,5 | 1,75 | 2,0 |

Ejemplo: en el Día 5, el exceso de trabajo de un doctor resta 100 en vez de 50.

**Estrellas:** `porcentaje = puntos / (episodios × 120)`. 1★ < 30 %, 2★ < 50 %, 3★ < 70 %, 4★ < **80 %**, 5★ el resto. Si el resultado es negativo, se queda en 1★ (nunca 0).

✅ **Cambios del 2026-10-18** (después de la prueba del tester):
- El límite de 5★ bajó de 85 % a **80 %**: un turno perfecto sin severos daba 83 % y se quedaba en 4★ ("okeyy").
- **Redondeo:** las restas de cada tipo se suman primero y el multiplicador del día se aplica y redondea **una sola vez** por tipo, no en cada cobro ("okey"). Así ×1,25 y ×1,75 son exactos.
- **Cama sin alta:** el primer −2 llega a los **25 min** despierto (15 de gracia + 10), y luego cada 10 min. Queda escrito así para no tener dos lecturas.
- **Implementación:** `ScoreBreakdown` guarda cuántas veces se cobró cada tipo (`BedCharges`, `OverworkEvents`, `UnattendedCount`) y `recomputePenalties` saca los puntos de cada tipo con `redondeo(cobros × base × multiplicador)`. Ejemplo Día 2: 4 cobros de cama = 10 (no 12); 1 dormido sin atender = 12,5 → 13; 2 = 25.
- **El reporte explica cada resta** (✅ aprobado): "Penalizaciones" muestra "Día N · ×M", cada fila trae el número de cobros y los nombres de quienes costaron "Sin cama o sin atender" van bajo esa fila (`Report.Unattended`). La caja 5.1 sigue mostrando solo `PatientsInHallway()`.
- **Frase de 1★:** `Report.Cause` (= `ScoreBreakdown.MainPenalty`, la penalización de mayor valor absoluto) escoge la frase del Director en `game.ReportLine`; empate: cama, exceso de trabajo, sin atender.

✅ **Con 1★ aparece un mensaje divertido del Director** regañando por la mala administración (varios textos al azar). Con 5★, uno de felicitación.

---

## 8. Fin del turno

Al llegar las 18:00 el turno se congela y se muestra el reporte, armado con **las 4 consultas del enunciado**:

| Sección del reporte | Llamada |
|---|---|
| Dormidos en el pasillo | `h.PatientsInHallway()` |
| Episodios por doctor | `doctor.MyEpisodes()` de cada doctor (§4.3) |
| Camas | estado de `h.Rooms()` + cuántas veces no hubo cama |
| Pacientes severos | `h.SevereReport()` |

Más: estrellas, puntos, desglose y los doctores que hicieron exceso de trabajo.

**Por qué los "sin atender" no salen en la consulta 5.1.** El modelo solo sabe lo que el jugador le reporta: un paciente que se durmió y nadie atendió (`AsleepWaiting`) sigue `Awake` para el hospital, así que `h.PatientsInHallway()` no lo ve. El reporte (`Report.Hallway` y `Report.Waiting`) muestra los dos grupos por separado ("Atendidos, sin cama (5.1)" y "Sin atender") y suma ambos en "Al cerrar el turno"; `hospital/` no cambia y el primer grupo es exactamente la consulta 5.1.

---

## 9. Modo Demo (`demo.go`)

Un turno sin azar que sigue un guion fijo con **los mismos datos que `runScenario` en `main.go`** (Karen atiende a Yesid y a Wilfrido, Álvaro a Yeimy, Chepe a Rafa y luego lleva a Wilfrido). Desde 2026-10-18 es una **demo guiada de 8 partes** (los textos aprobados están en [[Textos del Director#Demo guiada ✅]]):

| Parte | Qué pasa (cámara) |
|---|---|
| 1 | Se abre el turno (CAM 1A). No se anima nada |
| 2.1 | Yesid se duerme en CAM 2, Karen entra, se lo lleva en la camilla, corte a CAM 7A: en la cama 101 |
| 2.2 | Igual con Yeimy (CAM 8), Álvaro y la 102 (CAM 7B) |
| 2.3 | Igual con Rafa (CAM 5), Chepe y la 103 (CAM 7C) |
| 2.4 | Wilfrido se duerme en CAM 4, Karen llega, aviso "No hay cama libre!", lo atiende y se va a descansar (sin camilla); él se queda en el piso |
| 3.1 | CAM 7A: Yesid despierta ("!"), "Dar de alta" se resalta y se pulsa solo: la 101 queda libre |
| 3.2 | Chepe lleva a Wilfrido de CAM 4 a la 101 (CAM 7A) |
| 4 | El reporte con las 4 consultas |

**El reloj nunca corre solo.** La Demo es una máquina de estados (`Demo`): cada parte es una lista de *beats* (`Beat`: `BeatMingle`, `BeatSleep`, `BeatArrive`, `BeatCarry`, `BeatNoBed`, `BeatLeave`, `BeatBed`, `BeatWake`, `BeatHighlight`, `BeatDischarge`, `BeatReport`) con una duración en ticks de la Demo. `Demo.Tick` es el **único** lugar donde se llama a `Shift.Tick`, y solo mientras una parte se anima; entre partes todo espera a `Demo.Start` (botón **Siguiente paso**, o **Auto**, que arranca la siguiente tras 0,75 s). Mientras la camilla va camino a la cama el reloj del turno avanza 4 ticks por cada tick de la Demo (`carrySpeed`), para que el traslado de 5 s (doctor) o 3 s (camillero) se vea en ~1,2 s. El turno no llega ni a la mitad (la Demo entera gasta ~2 000 de los 10 800 ticks) y nunca termina.

Cada beat, al empezar, ejecuta **las mismas acciones de §5** (`Assign`, `MoveToBed`, `Discharge`) más las ayudas del guion (`ForceMove`, `ForceSleep`, `ForceWake`): la animación (personal que entra, camilla que sale) la pone la vista leyendo `Demo.Beat()` y `Demo.BeatProgress()`; `game` no sabe de píxeles. `Demo.Finish` corre lo que falte de la parte de golpe (el botón **Siguiente paso** durante una animación la salta). El aviso de "sin cama" sale como `DemoEvent`.

Tests: `game/demo_test.go` (el reloj no se mueve con 10 000 ticks de espera, cada una de las 8 partes termina en el estado esperado, el turno nunca termina, quién atiende a quién, cámaras por parte) y `TestDemoEndStateMatchesTheConsoleScenario` (las 4 consultas dan lo mismo que `main.go`).

---

## 10. Ajustes y guardado

| Ajuste | Por defecto |
|---|---|
| Parpadeo de cámara cuando alguien se duerme | **apagado** |
| Parpadeo de cámara cuando alguien se despierta | **apagado** |

El parpadeo no se acaba solo (la autora: "las cámaras no están parpadeando para siempre, solo lo hacen un momento"): el botón del mapa parpadea mientras haya un dormido sin atender (mango) o un despierto en cama (menta) en esa cámara y el jugador no la haya mirado; mirarla lo apaga y un paciente nuevo lo prende otra vez.
| Volumen música / efectos | 70 % / 80 % *(propuesto)* |
| Pantalla completa | apagado |

✅ **Autoguardado**: un archivo JSON pequeño con el último día desbloqueado, las mejores estrellas por día y los ajustes. Se guarda solo al terminar cada turno y al cambiar un ajuste.

✅ **Cada día es un hospital nuevo** (`hospital.NewHospital`). Así el cupo de los doctores es **diario**, como dijo la autora, y el reporte de cada turno solo muestra ese día.

---

## 11. Tests que debería tener `game`

| Test | Qué comprueba |
|---|---|
| El elegido es el despachado | Con 3 `StaffMember`, `Assign` hace que el episodio quede con el elegido, aunque el round-robin apunte a otro |
| Sin cama → se queda donde está | `Assign` con todas las camas ocupadas deja al paciente en `NoBedWaiting` en su misma cámara y en `AsleepInHallway` en el modelo |
| Alta libera cama | `Discharge` deja la habitación `Available` |
| No se da de alta dormido | `Discharge` en `InBed` devuelve error |
| Cooldown | Un `StaffMember` recién usado no es elegible hasta `busyUntil` |
| Exceso de trabajo una vez | Al llegar a 4/4 se descuenta una sola vez |
| Reloj | 10 800 ticks = 18:00, y la pausa no avanza |
| Demo = escenario | Al final de la demo, las 4 consultas dan lo mismo que el escenario de `main.go` |
| Modelo intacto | (manual) `git diff main -- hospital/` vacío |

---

## 12. Preguntas probables en la defensa

- **¿Por qué el paciente sale `Awake` en el modelo si en pantalla está dormido?** Porque el modelo es el sistema del hospital y solo registra lo que el jugador reporta (§3.1).
- **¿Cómo eliges al personal sin cambiar `hospital`?** Con `StaffMember`, que cumple `Attender` y solo se muestra disponible si fue elegido (§4.2).
- **¿Por qué `h.Doctors()` sale vacío en el juego?** Por la *type assertion* de `HireStaff` (§4.3).
- **¿Hay data races?** La lógica del juego corre en una sola goroutine y el hospital tiene su propio mutex.
- **¿Qué pasa si un doctor llena su cupo?** `IsAvailable()` del `*Doctor` da `false`, y por eso el del envoltorio también. El panel lo muestra "Sin cupo" y hay penalización (§4.5).
