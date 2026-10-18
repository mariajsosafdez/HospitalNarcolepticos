# Arquitectura

Cómo está armado el código del juego y por qué. La lógica detallada está en [[Lógica del juego]]; esta nota es el mapa para no perderse.

## Mapa de paquetes

```
hospital/    el modelo del enunciado          SIN CAMBIOS (git diff main -- hospital/ vacío)
game/        reglas del juego                 solo librería estándar + hospital
gameview/    ventana, dibujo y entrada        ÚNICO paquete que importa Ebitengine
web/         dashboard viejo                  SIN CAMBIOS, solo con -web
main.go      arma el escenario y arranca cosas (sin lógica de negocio)
game_on.go   etiqueta !nogame: abre la ventana (importa gameview)
game_off.go  etiqueta nogame: avisa que el juego se compiló fuera
```

Dependencias permitidas (la flecha es "importa a"):

```
main ──► hospital, web
main ──► game, gameview      (solo en game_on.go)
gameview ──► game, Ebitengine
game ──► hospital
web ──► hospital
hospital ──► (nada)
```

`hospital` y `game` nunca importan Ebitengine. Por eso las reglas se pueden probar con `go test` sin abrir una ventana, y el programa se puede compilar sin la librería gráfica.

## Cómo `game` usa `hospital` sin cambiarlo

- Cada día es un `hospital.NewHospital` nuevo (`game.NewShift`), así el cupo de los doctores es diario y el reporte solo muestra ese día.
- `game` solo llama métodos públicos del modelo: `AdmitPatient`, `HireStaff`, `AddRoom`, `RegisterEpisode`, `AssignRoom`, `WakeUpPatient`, `PatientsInHallway`, `Rooms`, `SevereReport`, `History`, `StateOf`.
- El estado de **juego** de cada paciente (`Wandering`, `AsleepWaiting`, `AwakeInBed`...) vive en `game.GamePatient`. El modelo solo se entera de lo que el jugador reporta, así que a veces los dos estados difieren a propósito (tabla de la sección 3.1 de [[Lógica del juego]]). Hay un test que recorre turnos completos y comprueba que la combinación juego/modelo siempre es una de las permitidas.
- El tiempo es por **ticks** (`game.Clock`): 10 800 ticks = 6 h de juego. Ninguna regla lee el reloj del computador. La semilla de `math/rand` la pone quien crea el turno (`main` usa el reloj; los tests, un número fijo). Con la misma semilla la partida es igual.
- `EpisodeRecord` guarda `time.Now()` real. `game` guarda aparte un `map` de ID de episodio a tick de juego (`episodeTimes`) para mostrar la hora de juego en el reporte.

## El truco de `StaffMember`

`RegisterEpisode` despacha al personal por turnos (round-robin) y toma al primero cuyo `IsAvailable()` dé `true`. En el juego **elige el jugador**. Solución sin tocar `hospital`:

1. `game.StaffMember` envuelve a un `*hospital.Doctor` o `*hospital.Orderly` y cumple `hospital.Attender` (los 5 métodos). El hospital lo contrata con `HireStaff` y no sabe que existe.
2. Su `IsAvailable()` es `selected && cooldown terminado && el tipo real está disponible`. Solo el elegido tiene `selected = true`, y solo durante la llamada a `RegisterEpisode`.
3. `Attend` delega en el doctor o camillero real. Por eso los episodios del historial guardan al `*Doctor` real, no al envoltorio.
4. `Shift.Assign` comprueba después que el atendedor del último episodio es el elegido (`record.Attender() != m.inner` sería un bug).

Consecuencia que hay que saber explicar: `HireStaff` llena `Doctors()` con un *type assertion* `a.(*Doctor)`. Como el juego contrata `*StaffMember`, `h.Doctors()` sale **vacío** en el hospital del juego. Por eso `StaffMember` guarda su propio `*Doctor` y el reporte 5.2 llama `doctor.MyEpisodes()` directamente.

Todo corre en una sola goroutine (el `Update` del juego), así que `selected` nunca lo ve otro hilo. El hospital igual tiene su `sync.Mutex`.

## Banderas y etiquetas de `main`

| Comando | Qué hace |
|---|---|
| `go run .` | Escenario + simulación por consola y luego abre la ventana del juego. Sin servidor web. |
| `go run . -web` | Escenario + simulación + dashboard web (comportamiento anterior). Sin ventana de juego. |
| `go run -tags nogame .` | Igual que `go run .` pero **sin compilar Ebitengine**: al final solo imprime que el juego está fuera. |

Por qué existe `nogame`: en Linux Ebitengine necesita cgo y las librerías de desarrollo de X11/OpenGL. Quien revisa el proyecto debe poder correr siempre el escenario obligatorio, aunque no pueda compilar la parte gráfica. Con la etiqueta, `game_on.go` (que importa `gameview`) no entra en la compilación y `game_off.go` ocupa su lugar. `go vet -tags nogame ./...` y `go build -tags nogame .` deben pasar siempre.

La versión de Ebitengine está fijada en **v2.8.8** porque pide `go 1.22.0`; las versiones nuevas piden Go 1.24 o 1.25 y obligarían a subir la versión mínima de Go del proyecto.

## Guardado

`game.SaveStore` recibe la ruta del archivo al crearse (`NewSaveStore(path)`). El valor por defecto de `DefaultSavePath()` es `%AppData%\HospitalCostenos\save.json`. Los tests y la puerta de pruebas usan una carpeta temporal y **nunca** tocan el guardado real.

## Puerta de pruebas (-agent)

> Estado: **planeada, aún no existe.** Se implementa junto con la primera pantalla (el menú), porque una puerta sin pantallas no tiene nada que mostrar.

Objetivo: que un agente tester pueda revisar el juego sin una persona mirando. Se activa con banderas en `main` y corre sin ventana visible y en silencio (solo escribe los archivos que se le pidan).

| Bandera | Qué hace |
|---|---|
| `-agent` | Activa el modo de pruebas. Sin esta bandera las demás se ignoran |
| `-scene <nombre>` | Salta directo a una escena (`menu`, `demo`, `day1`...) |
| `-seed <n>` | Semilla del azar: la misma semilla da la misma partida |
| `-day <1-5>` | Día que se juega |
| `-input "<guion>"` | Lista de pasos separados por `;` o saltos de línea. Pasos: `wait N` (N ticks), `click x,y`, `key <tecla>`, `shot <archivo>` (captura a 1×), `shot4 <archivo>` (captura a 4×), `dump <archivo>` (JSON de `game.Snapshot`) |
| `-frames N` | Cuántos ticks correr antes de terminar |
| `-quit` | Cierra al terminar el guion |

Reglas de la puerta:

- **Determinista:** misma semilla y mismo guion producen las mismas capturas y el mismo JSON. Nada lee el reloj del computador.
- **Silenciosa:** sin sonido ni mensajes en consola, salvo errores.
- **Nunca escribe el guardado real:** usa una ruta temporal inyectada en `SaveStore`.
- El volcado JSON es `game.Snapshot` (reloj, estado de juego y del modelo de cada paciente, cámara, espera, estado del personal y su cooldown, cupos, habitaciones, desglose de puntos y mensajes). Ya existe y está probado.

Ejemplo previsto:

```
go run . -agent -scene demo -seed 1 -input "wait 120; shot out/demo-1.png; dump out/demo.json" -quit
```
