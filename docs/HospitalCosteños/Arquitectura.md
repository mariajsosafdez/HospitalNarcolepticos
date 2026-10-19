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
| `demo_stage.go` | **Funciones puras** de la puesta en escena de la Demo: por qué lado entra el personal, dónde se para junto al paciente y por dónde sale la camilla (tests de tabla) |
| `layout.go`, `textfit.go` | **Funciones puras** con tests de tabla: medidas de las vistas, luz por hora, cuadro de cada sprite, a qué botón le da clic el mouse, reporte en líneas, ajuste de texto |
| `assets.go`, `text.go` | Carga de PNG, paleta y fuentes; dibujo de texto (sombra, borde) y rectángulos |
| `input.go` | `Input`: lo único que las pantallas saben del teclado y el mouse |
| `options.go`, `script.go`, `agent.go` | La puerta de pruebas |

**Arte embebido.** `go:embed` solo puede leer archivos que estén en la carpeta del paquete o debajo, y `gameview/` no puede alcanzar `../art`. Por eso hay un paquete diminuto `art` (`art/art.go`) que embebe `png/*.png` y `palette.hex` y expone un `embed.FS`. Los PNG se leen solo a 1x. Para que el juego vea un PNG nuevo hay que volver a compilar (`go run .` lo hace solo). Los fondos de cámara se buscan por nombre (`fondo_cam<id>_*`), así que cuando el arte agrega uno, la cámara lo recoge sin tocar código; mientras no exista, se pinta un relleno liso con el rótulo. Lo mismo con los sprites de personajes (`pac_<nombre>_*`, `doc_*`, `cam_*`): sin PNG sale una figura simple.

**Fuentes.** Pixelify Sans (variable, eje `wght` con `SetVariation`) y VT323, embebidas con `go:embed` y dibujadas con `text/v2` de Ebitengine. Las dos son SIL OFL 1.1 y sus `OFL-*.txt` están junto a los `.ttf`.

**Texto nítido: se dibuja a tamaño de pantalla.** Las dos fuentes son vectoriales (no una cuadrícula de píxeles exacta): dibujadas a 8 o 10 px en el lienzo de 480x270 y agrandadas después, salían suavizadas y borrosas (y el borde de 1 px del rótulo, un manchón). Ahora `App.Layout` elige el mayor factor entero `Env.Scale` que cabe en la ventana y devuelve 480x270 por ese factor; todo lo que se dibuja (`FillRect`, `DrawSprite`, `Text`) multiplica sus coordenadas por `Scale` (arte sin suavizar) y el texto usa letras de `tamaño * Scale`, igual que Figma al exportar a x4. Las coordenadas de las pantallas siguen siendo las de 480x270. La imagen de la cámara (168x123) se pinta con `Scale = 1` y se agranda aparte. Las capturas de `-agent` ya no se repintan: salen de la imagen que acaba de pintar `App.Draw` (ver el cursor, abajo). El sello de la cámara vuelve a 10 px (como la nota de Figma) con borde de 1 píxel del juego.

**Pantalla completa nítida (2026-10-18).** Antes `Layout` devolvía 480x270 por el factor entero y Ebitengine reescalaba esa imagen a la ventana con un factor NO entero (con suavizado) cuando la pantalla no era múltiplo exacto (por ejemplo 2560x1440 o 1920x1080 con zoom de Windows al 150 %): de ahí lo borroso. Ahora `Layout`/`LayoutF` devuelven el tamaño REAL de la ventana en píxeles (`LayoutF` multiplica por `Monitor().DeviceScaleFactor()`), así que Ebitengine no reescala nada. `fitScreen` elige el mayor factor entero `k` y el corrimiento que centra el lienzo de 480k x 270k; `Draw` pinta el lienzo aparte y lo pega con vecino más cercano sobre bordes del color `0` de la paleta. El cursor se convierte con `viewport{scale, offX, offY}` (`windowToGame` resta el corrimiento; el borde oscuro da coordenadas fuera de 0..479). Prueba: `go run . -agent -scene menu -screen 1501x899 -input g.txt -quit` con `shotwin` (queda 1501x899, lienzo de 1440x810 en el (30,44)). Lo único que no se pudo comprobar sin una pantalla real es el factor de zoom del sistema (`DeviceScaleFactor`).

**Sin cama no hay traslado (2026-10-18).** Si `Assign` termina en `ErrNoRoomAvailable`, el personal atendió pero no carga a nadie: `StaffMember.startRestOnly` lo deja directo en "Descansando" (descanso normal, `transferUntil` = ahora). La camilla solo se dibuja para `GamePatient.Carrier()`, que en ese caso es `nil`.

**Escala fraccionaria: la pantalla se llena (2026-10-18, segunda vuelta).** El factor entero dejaba bordes grandes y el juego chico ("está un poco popó"). Ahora `fitScreen(w, h)` devuelve `s = min(w/480, h/270)` como decimal (2,846 en 1366x768), el tamaño del lienzo redondeado y el corrimiento que lo centra; solo sobra una franja si la pantalla no es 16:9 (ninguna en 1920x1080 ni 2560x1440). Todo se dibuja directo a esa escala y nada se reescala después: los rectángulos (`FillRect`/`StrokeRect`) y las imágenes (`imageGeo`) redondean sus BORDES a píxeles enteros de pantalla (así dos cosas vecinas no dejan rendija, y los marcos miden `max(1, round(grosor*s))`); las imágenes usan vecino más cercano, de modo que algunos píxeles del arte miden un píxel de pantalla más que otros pero no hay suavizado; el texto usa la letra de tamaño `size*s` (redondeado a cuartos de píxel para la caché de caras) puesta en píxel entero, y los dígitos de sprite van por `imageGeo`. `viewport` guarda la escala decimal y el corrimiento; `windowToGame` divide y redondea hacia abajo, `gameToWindow` apunta al centro del píxel (tests de tabla con 1, 1,5, 16/9, 2,25, 4 y bordes, y ida y vuelta de los 480x270 píxeles). La imagen de la cámara sigue pintándose a 168x123 y se pega a x2 con `imageGeo`. La ventana ahora se puede redimensionar (`WindowResizingModeEnabled`). `shot`/`shot4` siguen sacando el lienzo (480x270 reescalado sin suavizar desde el lienzo real). Se revisó con `-screen` 1920x1080, 1366x768, 2560x1440, 1280x800 y 1500x900. Lo que solo la autora puede ver: su pantalla real con el zoom de Windows (`DeviceScaleFactor`).

**El cursor llega en el espacio de `Layout`, no en 480x270.** `ebiten.CursorPosition()` devuelve coordenadas del tamaño que devolvió `Layout` (480x270 por `Scale`). Las pantallas hacen toda su detección (hover, clics) en píxeles del juego, así que el cursor se divide por `Scale` en **un solo sitio**: `windowToGame` (en `input.go`, con test de tabla para escalas 1 a 4, tamaños impares y cursor fuera de la ventana). `Input.moveToWindow` es el único camino del cursor a las pantallas; lo usan el juego de verdad (`readReal`) y la puerta de pruebas. Con la versión anterior (cursor sin dividir) en una ventana de 960x540 el hover caía el doble de abajo y a la derecha: "Nuevo turno" señalaba "Créditos" y los botones del mapa, de la Demo y de la pausa no respondían. El tamaño de la ventana se elige con `layoutScale` (pura, con test).

**Teclado para atender.** Con un turno normal: Izquierda/Derecha eligen al paciente dormido de la cámara (de izquierda a derecha), Arriba/Abajo marcan una tarjeta del personal (borde blanco), Enter la manda. Sin paciente elegido, Enter da de alta al despierto de la habitación. En la pausa: Arriba/Abajo (o Izquierda/Derecha) marcan Continuar, Ajustes o Salir al menú, Enter ejecuta y Esc sigue; el mouse marca lo que toca.

**Nombres en español en la vista.** La Demo usa los datos en inglés del enunciado ("Dr. Karen Ospina / Neurology"), que no se tocan. `gameview/names.go` busca por ID (D-001 a D-003) el nombre y la especialidad del Elenco y se usa en las tarjetas del panel y en el reporte.

**Texto en pantalla sin ¡.** Decisión de la autora (2026-10-17): lo que se ve en el juego pierde el `¡` de apertura y conserva el `!` del cierre, porque en Pixelify Sans el `¡` parece una `i`. Se hace en un solo lugar: `screenText` (en `layout.go`, con test de tabla) lo usan `drawText` y la medición de ancho. Los textos del código y de las notas conservan el español correcto.

**Dos relojes.** `game.Clock` cuenta los ticks del turno y se detiene en la pausa. `Env.Tick` cuenta los ticks de la interfaz (parpadeos, animaciones) y sigue corriendo. Nada lee el reloj del computador, por eso las capturas de la puerta de pruebas son iguales cada vez.

**Adiciones pequeñas a `game/`** (con tests): `StaffMember.CooldownFraction()` (la barra del panel crece en el traslado y baja en el descanso), `GamePatient.Carrier()` (quién empuja la camilla, para dibujar a la persona correcta) y `SaveStore.Exists()` (para dejar "Continuar" gris sin partida).

**Guardado.** En el juego de verdad la vista usa `game.DefaultSavePath()`. En `-agent` no hay `SaveStore` (es `nil`) salvo que se pase `-savepath` (un archivo de prueba, nunca el real): el progreso vive solo en memoria y el guardado real no se toca. Se autoguarda al terminar cada turno y con cada cambio de ajustes.

### Pantallas del turno (2026-10-18)

| Archivo | Qué tiene |
|---|---|
| `message.go` | Caja del Director (tutorial, inicio de día, exceso de trabajo). Mientras está abierta el reloj del turno se detiene (`shift.Pause`). |
| `report.go` | Reporte de fin de turno. Las funciones puras (`noBedLines`, `doctorLines`, `bedCells`, `severeLines`, `capWithMore`, `signedPoints`...) arman el texto de cada caja y tienen tests de tabla. |
| `pause.go` | Pausa (Continuar / Ajustes / Salir al menú) y el parpadeo de los botones del mapa (`blinkAlpha`, `stateEvent`). |
| `settings.go` | Ajustes: volúmenes (de a 10), parpadeo al dormirse y al despertarse, pantalla completa. Cada cambio llama `Env.settingsChanged` (autoguarda). |
| `play.go` | `newShiftScene` y `newPracticeScene`: la única manera de armar un turno, usada por el menú, el reporte y la puerta de pruebas. |

**Flujo.** Menú -> turno (`camsScene`) -> `finishShift` guarda (`Env.recordShift`: mejores estrellas + día siguiente desbloqueado) -> `reportScene` -> "Siguiente día" arma el día N+1 con el mensaje del Director de ese día, o "Menú". En la Demo (paso 4) y el Día 5 solo hay "Menú". "Salir al menú" desde la pausa abandona el turno sin guardarlo. "Cómo jugar" abre un Día 1 con los 9 mensajes del tutorial y al terminar vuelve al menú. El tutorial del Día 1 de "Nuevo turno" sale solo la primera vez (`SaveData.TutorialSeen`).

**Medidas.** Salen de las vistas aprobadas en `preview/turno/`, comparadas pixel a pixel con capturas `-scale 4` (la nota de Figma no estaba al alcance del builder cuando se programó: ver Plan de construcción).

**Habitaciones.** Cada habitación puede tener su fondo propio `fondo_habitacion_<número>`. La cama (32x30) y la placa del número (9x4) de cada una están en la tabla `roomArts` (`layout.go`), que repite los comentarios `// cama: x,y` y `// placa: x,y` de `art/sprites/fondo_habitacion_<n>.txt`; `TestRoomArtsMatchSprites` lee esos archivos y falla si dejan de coincidir. Se usa una tabla y no se lee el comentario al ejecutar porque los `.txt` no van en el ejecutable. Si falta el fondo propio se usa `fondo_habitacion` con la cama en (34, 72).

**Dónde se tumban los dormidos.** Al dormirse en el piso, la vista (`settleSleeper`, `sleepspots.go`) busca un lugar libre con `pickSleepSpot`: el sprite dormido no se solapa con otro dormido de la cámara, queda a 8 px de los bordes y lejos de los muebles de la cámara: mesas, sillas, bancas, mostradores, matas, el dispensador, la máquina de rayos X, los troncos y la jardinera del Patio (`sleepObstacles`, un rectángulo por mueble en las ocho cámaras comunes; quien camina sí puede pasar junto a ellos). Es determinista con la semilla. Los acostados (sprite al menos 2 veces más ancho que alto) llevan el zzz sobre la cabeza, que está a la izquierda (`zzzSpot`). Con la puerta de pruebas: `sleep P-004 2; sleep P-005 2; sleep P-001 2; wait 20; shotwin ...` y `-cam 6` para el patio.

**Avisos.** `noticeY` pone el aviso abajo, o arriba (y=36) si abajo taparía al paciente del que habla o el botón "Dar de alta". Abrir un mensaje del Director borra el aviso. Un aviso por asignación fallida no suelta la selección. El reporte cuenta "pacientes atendidos" por persona distinta (`Shift.markAttended`); los puntos posibles siguen siendo por episodio.

**Sprites por ID.** `spriteKeys` (`layout.go`) da la clave de los PNG por ID de persona (P-007 es `mane`, no `abelardo`). `TestEveryCastMemberHasSprites` comprueba que todo el elenco tenga sus PNG embebidos.

**Easter eggs.** `withEgg` (`play.go`) agrega el easter egg al `DayConfig` (el de `-egg` o el de `game.RollEgg(seed)`); "Cómo jugar" no lo llama. Sus sprites son `pac_egg<N>_*` (`spriteKeys` por ID E-00N). Sobre cualquier camilla con Misifu, `drawCatHead` pinta `pac_egg1_camilla` (7x6) en (14, 8) del cuadro (reflejada si la camilla va reflejada). El mensaje del gato es `msgCat`: `checkCatArrival` mira `Shift.ArrivedIDs`. Para verlos: `go run . -agent -scale 2 -scene day -egg 1 -cam 1B -input guion.txt -quit` con `wait 2; click 240 100` (cierra el mensaje), `sleep E-001 1B`, `wait 40`, `shotwin ruta`; seleccionar con un clic en el recuadro de `dump` (`boxes`), clic en la tarjeta de Chepe (350, 124), `wait 60` y `shotwin` muestran la camilla con la cabeza del gato; `wait 140; click 356 221` abre la habitación 101 con el paciente en la cama.

**Dígitos en pixel art.** Pixelify Sans dibuja el 5 como una S, así que `drawText` (`text.go`) parte la línea en tandas (`splitRuns`, `digits.go`): los dígitos van con los sprites `ui_digitos_8` (Pixelify 8 Regular y Bold), `ui_digitos_10b` y `ui_digitos_18b` (`digitSpecFor` da avance, alto y línea base) y lo demás con la fuente. Cada cuadro se vuelve una máscara blanca (`Assets.digitMask`) y se tiñe con `ColorScale`, así funcionan `TextShadow` y `TextOutline`. `Fonts.lineWidth` (y `Width`) suman el avance de cada dígito (`digitSpec.advanceOf`), para que el centrado y la alineación a la derecha sigan bien. **El "1" avanza 1 px menos** (`narrowOne`): su tinta es un píxel más angosta y con el avance completo "12" salía "1 2". El test `TestDigitAdvanceMatchesInk` mide la tinta de cada PNG y comprueba la tabla; los demás dígitos no cambian. Otros tamaños y VT323 no cambian. Prueba: `go run . -agent -scale 3 -scene day -day 5 -input guion.txt -quit` (`wait 5; shotwin ruta`).

## Puerta de pruebas (-agent)

Un tester (persona o agente) puede saltar a una pantalla, mandar clics y teclas, sacar capturas y volcar el estado sin mirar la ventana. Con `-agent` **no se imprime el escenario ni la simulación** (silencio total, solo errores) y no se escribe el guardado real.

**Dónde viven las banderas.** `gameview.RegisterFlags` las define y `game_on.go` las registra en una variable global (`agentOpts`) antes de `flag.Parse()`. `main.go` solo hace `if agentMode() { runAgent(); return }`: sin lógica. Con la etiqueta `nogame` no existen (`game_off.go`).

| Bandera | Qué hace |
|---|---|
| `-agent` | Activa el modo de pruebas. Sin esta bandera las demás se ignoran |
| `-scene` | `menu`, `cams`, `demo`, `day`... o `credits` (tabla abajo) |
| `-day N` | Día que se juega (1 a 5) en `-scene day` |
| `-seed N` | Semilla del azar del turno y de la vista (por defecto 1) |
| `-cam ID` | Cámara con la que empieza (`2`, `7A`...) |
| `-step N` | Solo `demo`: **parte** de la Demo (1 a 8) con la que empieza: juega de golpe, sin mostrarlas, las partes 2 a N, así que `-step 1` es el arranque (parte 1) y `-step 8` abre el reporte. Numeración: 1, 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 4 = 1 a 8. Para ver una animación a medias: `-step 1` y un guion con `click 245 250` (Siguiente paso), `wait N`, `shot`. En `-scene tutorial` es la línea (1 a 9) |
| `-input archivo` | Guion (ver abajo) |
| `-frames N` | Ticks extra que corren después del guion |
| `-quit` | Cierra al terminar (con `-quit` la ventana se saca de la pantalla). Sin `-quit` la ventana sigue abierta |
| `-saved N` | Finge una partida guardada con último día desbloqueado N (activa "Continuar") |
| `-intro` | Con `-scene day`: abre con el mensaje del Director del día (tutorial en el Día 1 la primera vez, inicio de día en 2 a 5). Sin la bandera el turno corre sin mensajes |
| `-savepath archivo` | Guardado de PRUEBA que se lee y se escribe (ajustes, días desbloqueados, estrellas). Nunca puede ser el guardado real (falla si lo es). Sin la bandera no hay guardado |
| `-egg N` | Fuerza al easter egg N (1 Misifu, 2 Alejo Urrego, 3 Samu Giraldo, 4 Cote, 5 Sebs) dentro de un turno normal (`-scene day`, `report`, `pause`...). Llega en el tick 0 (`DayConfig.ArriveAtStart`) en vez de a una hora al azar. Con 1 sale de entrada el mensaje de Don Ernesto (un clic lo cierra). Sin la bandera manda el sorteo de la semilla (15 %; con `-seed 39` sale Misifu, con 3 o 10 Cote, con 5 Alejo; con la semilla 1, ninguno). No afecta la Demo ni `tutorial` |
| `-scale N` | Factor de dibujo fijo (1 a 4). Sin la bandera, 2 (ventana de 960x540). Sirve para probar a 2x y 3x lo que ve la ventana real |
| `-screen WxH` | Finge una imagen de ventana de ese tamaño (por ejemplo `1501x899` o `2560x1440`, mínimo 480x270) para probar la pantalla completa con bordes. Con `shotwin` sale la ventana entera, bordes incluidos; `shot`/`shot4` siguen sacando solo el lienzo |

| `-scene` | Qué abre |
|---|---|
| `menu` | El menú. Sin `-saved`, "Continuar" sale gris |
| `day` | Turno normal del `-day`, con azar de `-seed` (los pacientes llegan, caminan y se duermen solos) |
| `cams` | La pantalla de cámaras con el escenario de la Demo (5 pacientes, 3 habitaciones, sin azar y sin caja de la Demo). Ideal para probar con `sleep`/`wake` |
| `dayintro` | El mensaje de inicio del día `-day` (2 a 5) sobre el turno detenido |
| `tutorial` | "Cómo jugar": los 9 mensajes sobre un Día 1, empezando en el mensaje `-step` (1 a 9) |
| `director` | El mensaje de exceso de trabajo (Dra. Karen) sobre un Día 1 |
| `report` | Juega el día `-day` entero con `game.AutoStep` (jugador automático), guarda como un turno real y abre el reporte |
| `pause` | Un Día 1 en la pausa (`key Down`, `key Enter` abre los Ajustes) |
| `settings` | El menú con los Ajustes abiertos (usar con `-savepath`) |
| `credits` | Los Créditos (`credits.go`), con el menú detrás: `key Escape`, `key Enter` o `click` vuelven al menú con "Créditos" elegido. El desfile depende solo de los ticks: `wait N` antes del `shot` lo mueve (16 px/s, 60 ticks = 16 px) |
| `demo` | La Demo guiada con su caja de controles, después de la parte `-step` (ver abajo "Demo guiada"). `dump` trae `demoPart` (1 a 8), `demoStep` (el "Paso N de 4"), `demoPlaying` (una parte se anima: el reloj corre) y `demoBeat` (`mingle`, `sleep`, `arrive`, `carry`, `nobed`, `leave`, `bed`, `wake`, `highlight`, `discharge`) |

**Guion.** Un paso por línea (o varios con `;`); `#` empieza un comentario. Un paso desconocido o con argumentos mal escritos **hace fallar el arranque** con el número de línea.

| Paso | Qué hace |
|---|---|
| `wait N` | Corre N ticks |
| `move X Y` | Pone el mouse en el cuadro (X, Y) del lienzo de 480x270 (gasta 1 tick). Entra como **posición de ventana** (`gameToWindow`: el medio del cuadro por `-scale`) y pasa por la misma conversión que el cursor real |
| `click X Y` | Igual que `move`, y hace clic (1 tick) |
| `key Nombre` | Pulsa una tecla: `Up Down Left Right Enter Escape Space M 1`..`8` |
| `shot ruta` | Captura PNG a 1x (480x270), reescalada sin suavizar desde lo que pintó `Draw` |
| `shot4 ruta` | Captura PNG a 4x (1920x1080), idem. Con `-scale 4` es exacta |
| `shotwin ruta` | Captura tal cual la ventana: 480x270 por `-scale`. Es lo que ve la autora |
| `dump ruta` | JSON: `ui` (escena, cámara, paciente elegido, `boxes` con el rectángulo donde se le da clic a cada paciente, `staffCursor`, botón bajo el mouse, aviso, pausa o reporte, paso de la Demo...) y `game` (`game.Snapshot` completo) |
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
click 172 214          # botón "Dar de alta" (habitación 101: centrado bajo la cama, y = 208 a 221)
dump out/final.json
```

```
go run . -agent -scene cams -cam 2 -input out/ejemplo.txt -quit
go run . -agent -scene demo -step 8 -input script.txt -quit     # reporte de la Demo (parte 4 de 4)
go run . -agent -scene demo -step 5 -scale 2 -input script.txt -quit     # tras la parte 2.4: Wilfrido sin cama
go run . -agent -scene menu -saved 3 -input script.txt -quit    # "Continuar" activo
```

**Estados alcanzables.** Menú (normal, Demo elegida con `key Down`, sin y con partida), cámaras (zona común, paciente elegido, aviso sin cama, habitación con dormido y con despierto, candado en cámara bloqueada con `move`), panel de personal (los 4 estados; 4 y 5 personas con `-scene day -day 3` o `-day 4`), Demo (las 8 partes, sus animaciones a medias con `click 245 250; wait N`, Auto con `click 305 250`), pausa (`key Escape` o `-scene pause`), ajustes, créditos (`-scene credits`, o desde el menú con `key Up; key Enter`), mensajes del Director, reporte (`-scene report`, parte 8 de la Demo, o al terminar un turno) y parpadeo de los botones del mapa (con los ajustes prendidos en un `-savepath`, `wait 2; sleep P-001 8; wait 20; dump`: el dump lista `blinking`). El volcado trae también `save` (ajustes, día desbloqueado, estrellas), `message` y `blinking`.

**Demo guiada.** `game.Demo` (ver [[Lógica del juego#9. Modo Demo (`demo.go`)]]) decide qué pasa y cuánto va; `gameview` solo dibuja. `camsScene.updateDemo` llama a `Demo.Tick` (único lugar donde corre el reloj) en vez de `Shift.Tick`. Lo que se ve viene de `Demo.Beat()` y `Demo.BeatProgress()`: `drawStagedStaff` pinta al personal entrando, de pie junto al paciente o yéndose (sprites `*_camina` y `*_quieto`), `carryPlacement` hace salir la camilla de la cámara al ritmo del traslado (`StaffMember.CooldownFraction`) y `drawDischargeButton` resalta y hunde "Dar de alta". La cámara la cambia sola cada beat (`Beat.Camera`); con una parte animándose se ignoran las teclas 1 a 8 y el mapa, y en espera el jugador puede cambiarla a mano (la siguiente parte la devuelve). "Siguiente paso" durante una animación la salta (`Demo.Finish`); Auto (que pasa a "Pausa") espera 0,75 s entre partes. Los 8 recuadros ("Paso 2 de 4 · Yesid (1 de 4)") y sus textos están en `demoParts` (`layout.go`). **VISUAL, pendiente de la autora:** en la Demo el botón "Dar de alta" va en y = 186 (entre la cama y la caja de la Demo, dibujado encima de ella) porque en y = 208 la caja lo tapa.

**Panel de 4 o 5 personas.** Con más de 3 personas (días 3 a 5) las tarjetas son `ui_tarjeta_compacta` (132x24), sin páginas.
