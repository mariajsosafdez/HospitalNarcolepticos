# Arquitectura

Cómo está armado el código del juego y por qué. La lógica detallada está en [[Lógica del juego]]; esta nota es el mapa para no perderse.

## Mapa de paquetes

```
hospital/    el modelo del enunciado          SIN CAMBIOS (git diff main -- hospital/ vacío)
game/        reglas del juego                 solo librería estándar + hospital
gameview/    ventana, pantallas, dibujo, entrada y puerta de pruebas   ÚNICO paquete que importa Ebitengine
art/         PNG y paleta embebidos (art/art.go)                          solo go:embed
web/         dashboard viejo                  SIN CAMBIOS, solo con -web
main.go      arma el escenario y arranca cosas (sin lógica de negocio)
game_on.go   etiqueta !nogame: abre la ventana (importa gameview)
game_off.go  etiqueta nogame: avisa que el juego se compiló fuera
```

Dependencias permitidas (la flecha es "importa a"):

```
main ──► hospital, web
main ──► game, gameview      (solo en game_on.go)
gameview ──► game, hospital, art, Ebitengine
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

## La vista (`gameview/`)

| Archivo | Qué tiene |
|---|---|
| `app.go` | `App` (lo que Ebitengine llama: `Update`, `Draw`, `Layout`), `Env` (arte, fuentes, guardado, reloj de interfaz), la interfaz `Scene` y `Run` |
| `menu.go` | Pantalla del menú |
| `cams.go`, `cams_draw.go` | Pantalla de cámaras (turno y Demo): lógica de entrada y dibujo por capas |
| `walkers.go` | Cómo caminan los pacientes dentro de una sala (la lógica solo sabe la cámara) |
| `layout.go`, `textfit.go` | **Funciones puras** con tests de tabla: medidas de las vistas, luz por hora, cuadro de cada sprite, a qué botón le da clic el mouse, reporte en líneas, ajuste de texto |
| `assets.go`, `text.go` | Carga de PNG, paleta y fuentes; dibujo de texto (sombra, borde) y rectángulos |
| `input.go` | `Input`: lo único que las pantallas saben del teclado y el mouse |
| `options.go`, `script.go`, `agent.go` | La puerta de pruebas |

**Arte embebido.** `go:embed` solo puede leer archivos que estén en la carpeta del paquete o debajo, y `gameview/` no puede alcanzar `../art`. Por eso hay un paquete diminuto `art` (`art/art.go`) que embebe `png/*.png` y `palette.hex` y expone un `embed.FS`. Los PNG se leen solo a 1x. Para que el juego vea un PNG nuevo hay que volver a compilar (`go run .` lo hace solo). Los fondos de cámara se buscan por nombre (`fondo_cam<id>_*`), así que cuando el arte agrega uno, la cámara lo recoge sin tocar código; mientras no exista, se pinta un relleno liso con el rótulo. Lo mismo con los sprites de personajes (`pac_<nombre>_*`, `doc_*`, `cam_*`): sin PNG sale una figura simple.

**Fuentes.** Pixelify Sans (variable, eje `wght` con `SetVariation`) y VT323, embebidas con `go:embed` y dibujadas con `text/v2` de Ebitengine. Las dos son SIL OFL 1.1 y sus `OFL-*.txt` están junto a los `.ttf`.

**Texto en pantalla sin ¡.** Decisión de la autora (2026-10-17): lo que se ve en el juego pierde el `¡` de apertura y conserva el `!` del cierre, porque en Pixelify Sans el `¡` parece una `i`. Se hace en un solo lugar: `screenText` (en `layout.go`, con test de tabla) lo usan `drawText` y la medición de ancho. Los textos del código y de las notas conservan el español correcto.

**Dos relojes.** `game.Clock` cuenta los ticks del turno y se detiene en la pausa. `Env.Tick` cuenta los ticks de la interfaz (parpadeos, animaciones) y sigue corriendo. Nada lee el reloj del computador, por eso las capturas de la puerta de pruebas son iguales cada vez.

**Adiciones pequeñas a `game/`** (con tests): `StaffMember.CooldownFraction()` (la barra del panel crece en el traslado y baja en el descanso), `GamePatient.Carrier()` (quién empuja la camilla, para dibujar a la persona correcta) y `SaveStore.Exists()` (para dejar "Continuar" gris sin partida).

**Guardado.** En el juego de verdad la vista usa `game.DefaultSavePath()`. En `-agent` no hay `SaveStore` (es `nil`): el progreso vive solo en memoria y el guardado real no se toca.

## Puerta de pruebas (-agent)

Un tester (persona o agente) puede saltar a una pantalla, mandar clics y teclas, sacar capturas y volcar el estado sin mirar la ventana. Con `-agent` **no se imprime el escenario ni la simulación** (silencio total, solo errores) y no se escribe el guardado real.

**Dónde viven las banderas.** `gameview.RegisterFlags` las define y `game_on.go` las registra en una variable global (`agentOpts`) antes de `flag.Parse()`. `main.go` solo hace `if agentMode() { runAgent(); return }`: sin lógica. Con la etiqueta `nogame` no existen (`game_off.go`).

| Bandera | Qué hace |
|---|---|
| `-agent` | Activa el modo de pruebas. Sin esta bandera las demás se ignoran |
| `-scene` | `menu`, `cams`, `demo` o `day` (tabla abajo) |
| `-day N` | Día que se juega (1 a 5) en `-scene day` |
| `-seed N` | Semilla del azar del turno y de la vista (por defecto 1) |
| `-cam ID` | Cámara con la que empieza (`2`, `7A`...) |
| `-step N` | Solo `demo`: paso de la Demo (1 a 4) con el que empieza (ejecuta los anteriores sin mostrarlos) |
| `-input archivo` | Guion (ver abajo) |
| `-frames N` | Ticks extra que corren después del guion |
| `-quit` | Cierra al terminar (con `-quit` la ventana se saca de la pantalla). Sin `-quit` la ventana sigue abierta |
| `-saved N` | Finge una partida guardada con último día desbloqueado N (activa "Continuar") |

| `-scene` | Qué abre |
|---|---|
| `menu` | El menú. Sin `-saved`, "Continuar" sale gris |
| `day` | Turno normal del `-day`, con azar de `-seed` (los pacientes llegan, caminan y se duermen solos) |
| `cams` | La pantalla de cámaras con el escenario de la Demo (5 pacientes, 3 habitaciones, sin azar y sin caja de la Demo). Ideal para probar con `sleep`/`wake` |
| `demo` | La Demo con su caja de controles, arrancando en el paso `-step` |

**Guion.** Un paso por línea (o varios con `;`); `#` empieza un comentario. Un paso desconocido o con argumentos mal escritos **hace fallar el arranque** con el número de línea.

| Paso | Qué hace |
|---|---|
| `wait N` | Corre N ticks |
| `move X Y` | Pone el mouse en (X, Y) del lienzo de 480x270 (gasta 1 tick) |
| `click X Y` | Mueve el mouse y hace clic (1 tick) |
| `key Nombre` | Pulsa una tecla: `Up Down Left Right Enter Escape Space M 1`..`8` |
| `shot ruta` | Captura PNG a 1x (480x270) |
| `shot4 ruta` | Captura PNG a 4x (1920x1080) |
| `dump ruta` | JSON: `ui` (escena, cámara, paciente elegido, botón bajo el mouse, aviso, pausa o reporte, paso de la Demo...) y `game` (`game.Snapshot` completo) |
| `sleep ID CAM` | Duerme al paciente en esa cámara (`ForceSleep`, la misma ayuda de la Demo) |
| `wake ID` | Despierta a un paciente que duerme en cama (`ForceWake`) |

Reglas: **determinista** (misma semilla y mismo guion dan las mismas capturas y el mismo JSON), **silenciosa**, y la captura se toma cuando el cuadro ya está pintado (el motor no avanza mientras espera). Todo corre lo más rápido que da la máquina: el tiempo del juego son ticks, no segundos.

**Ejemplo** (`out/ejemplo.txt`): duerme a Yesid en la cafetería, lo elige con un clic, escoge a la Dra. Karen, espera a que llegue a la cama, mira la habitación 101, lo despierta y lo da de alta.

```
sleep P-001 2
wait 30
shot4 out/1-dormido.png
click 178 205          # clic sobre el paciente (con la semilla 1 cae ahí)
shot4 out/2-elegido.png
click 400 50           # tarjeta de la Dra. Karen
wait 400               # traslado de 5 s = 300 ticks
key 7                  # cámara 7A
wake P-001
wait 3
shot4 out/3-despierto.png
click 96 248           # botón "Dar de alta"
dump out/final.json
```

```
go run . -agent -scene cams -cam 2 -input out/ejemplo.txt -quit
go run . -agent -scene demo -step 4 -input script.txt -quit     # reporte de la Demo
go run . -agent -scene menu -saved 3 -input script.txt -quit    # "Continuar" activo
```

**Estados alcanzables.** Menú (normal, Demo elegida con `key Down`, sin y con partida, "Próximamente"), cámaras (zona común, paciente elegido, aviso sin cama, habitación con dormido y con despierto, candado en cámara bloqueada con `move`), panel de personal (los 4 estados), Demo (pasos 1 a 4, Auto), pausa (`key Escape`) y reporte (paso 4 de la Demo, o al terminar un turno con `wait 10800`).

**Pendiente de diseño.** Un paso de la vista no está en Figma: con más de 3 personas (días 3 a 5) el panel muestra 3 tarjetas por página y un botón "Más >" (en la vista aprobada solo caben 3).
