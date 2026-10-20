# Sonido

> 2026-10-19: el juego se entrega sin sonido (decisión de la autora: "no creo que nos dé el tiempo"). Esta nota queda como registro de la búsqueda.
>
> 2026-10-19 (tarde): la autora pidió **probar en una rama aparte (`sonido`)** si se le puede meter sonido a última hora ("normaliza algunos de los sonidos [...] con ritmos relajados"). Es un experimento: si no le gusta, la rama no se une. Lo que hay en la rama está en la sección [Experimento en la rama sonido](#experimento-en-la-rama-sonido).

Dirección aprobada ([[Game Design#Registro de decisiones]]): *una pista tranquila de cumbia o champeta lo-fi en el menú y una más suave durante el turno, solo CC0 o CC-BY. Efectos: ronquido, pasos, ruedas de camilla, "ding" de alta.* Cálido, relajado, costeño, nunca de terror.

> **La IA no puede oír: todo es por defecto hasta que la autora lo confirme de oído.** Las descripciones salen de las páginas de las pistas, sus etiquetas y medidas (duración, volumen). Nada está conectado al juego todavía.

## Lista de sonidos (cues)

| Cue | Cuándo suena | Tipo |
|---|---|---|
| `music_menu` | Menú principal (2-3 min, en loop) | Música |
| `music_turno` | Durante el turno, más suave y larga, en loop | Música |
| `sfx_ronquido` | Un paciente/personal duerme (variantes) | Efecto repetido |
| `sfx_camilla` | Ruedas de camilla / carrito | Efecto repetido |
| `sfx_alta` | "Ding" suave cuando se da de alta | Efecto |
| `sfx_despierta` | Pop suave de "¡!" cuando alguien despierta | Efecto |
| `sfx_click` | Elegir/confirmar en menús | Efecto |
| `sfx_sin_cama` | Blip negativo suave si no hay cama | Efecto |
| `sfx_fin_turno` | Jingle corto y feliz al terminar el turno | Efecto |
| `sfx_pasos` | (opcional) pasos sobre baldosa | Efecto repetido |

## Candidatos

Los archivos están en `sound/audition/` (no se commitean). Detalle completo, créditos y duraciones: `sound/audition/README.md`.

| Cue | Archivo | Fuente | Licencia | Por qué |
|---|---|---|---|---|
| music_menu | `01_tropical_loop` | OpenGameArt, wipics | CC0 | Isla/conga/tambores; corto (18 s) y animado |
| music_menu | `02_beach_town_guitar` | OpenGameArt, Hitctrl | CC-BY 3.0 | Guitarra de pueblo de playa, repetible, volumen suave |
| music_menu | `03_feel_good_island` | OpenGameArt, Brandon Morris / AntumDeluge | OGA-BY 3.0 | Tema de isla alegre ya en loop |
| music_menu | `04_shop_theme_bossa` | OpenGameArt, CleytonKauffman | CC0 | Bossa nova repetible, ritmo latino relajado |
| music_turno | `01_calm_track` | OpenGameArt, pmiller | CC0 | Chillout de 4 min que evoluciona despacio, ya en loop |
| music_turno | `02_etirwer_guitarra` | OpenGameArt, Kistol | CC0 | Guitarra de nailon sencilla, loop sin cortes |
| music_turno | `03_serenade_guitar` | OpenGameArt, Pro Sensory | CC0 | Guitarra de serenata |
| music_turno | `04_green_salon` | OpenGameArt, Yubatake | CC-BY 4.0 | Bossa nova con vibráfono, 65 BPM, muy calmada |
| sfx_ronquido | `01_snore`, `02_snore_b`, `03_sleep_grunts`, `04_breath`, `05_breath_b` | OpenGameArt (rubberduck, crueltyinc) | CC0 | Únicos ronquidos/respiraciones CC0 que encontré |
| sfx_camilla | `01_rolling_loop` ... `05_creak` | OpenGameArt (rubberduck, themightyglider, AntumDeluge), Kenney | CC0 | Aproximaciones: no hay ruedas de camilla reales en CC0 |
| sfx_alta | `01`-`05` | Kenney Interface Sounds | CC0 | Confirmaciones y campanitas de vidrio |
| sfx_despierta | `01`-`04` | Kenney Interface Sounds | CC0 | "Pregunta" y "pluck" cortos |
| sfx_click | `01`-`04` | Kenney Interface Sounds / UI Audio | CC0 | Clics de interfaz |
| sfx_sin_cama | `01`-`04` | Kenney Interface Sounds | CC0 | `03` y `04` son los más suaves; `01`-`02` pueden sonar duros |
| sfx_fin_turno | `01`-`05` | Kenney Music Jingles | CC0 | Pizzicato, sax y steel drum (el más tropical) |
| sfx_pasos | `01`-`04` | Kenney RPG Audio | CC0 | Pasos sueltos; no son de baldosa |

## Qué escuchar primero

1. `music_menu/04_shop_theme_bossa` y `music_menu/02_beach_town_guitar`: son lo más cercano a un ambiente latino tranquilo con pista larga. `01_tropical_loop` solo dura 18 s y es la más animada.
2. `music_turno/01_calm_track` y `02_etirwer_guitarra`: ambas CC0 y ya en loop, así que no hay que cortar nada ni dar crédito.
3. `sfx_fin_turno/04_steel_a`: el steel drum es lo que más suena a Caribe.
4. `sfx_ronquido/02_snore_b` y `03_sleep_grunts`: los ronquidos más naturales; el `01` es muy bajito.
5. `sfx_camilla`: ninguno es una camilla real; dime si alguno sirve o buscamos otro.

Para escuchar (desde la raíz del repo): `powershell -ExecutionPolicy Bypass -File sound\audition\escuchar.ps1` (opcional `-Cue music_menu`).

## Qué no se consiguió

- Cumbia o champeta de verdad con licencia CC0/CC-BY: no apareció ninguna. Lo más cercano es bossa nova, guitarra e isla/steel drum. Si la autora quiere cumbia real, habría que buscar otra fuente o componerla.
- Ruedas de camilla reales y más ronquidos.
- Sitios: freepd.com (página de música devolvió 404), incompetech y Free Music Archive sin pistas confirmadas; freesound.org no se intentó (pide login). Pixabay se evitó a propósito.

## Pasos cuando la autora elija

1. Decir qué archivo por cue (el nombre de la carpeta y el número).
2. Música: si no es loop, cortar y comprobar el empalme (hoy no hay ffmpeg; hay un decodificador en Go para hacerlo). Pasar los WAV grandes a OGG.
3. Copiar solo las elegidas a `assets/sound/` (o la carpeta que decida la construcción) con nombres por cue, p. ej. `music_menu.ogg`, `sfx_alta_1.ogg`.
4. Créditos: añadir en la pantalla de Créditos y el README las líneas de las CC-BY/OGA-BY elegidas (están en `sound/audition/README.md`). Las CC0 no lo exigen, pero conviene nombrarlas.
5. Conectar al juego después: reproductor en `gameview/` (único paquete que importa Ebitengine; `audio` + `vorbis`), ajustes de volumen música/efectos ya previstos en Ajustes, y el modo `-agent` silencioso.
6. Guardar los archivos elegidos en git; `sound/audition/` sigue ignorado.

## Experimento en la rama sonido

**La IA no puede oír: todas estas elecciones son "por defecto: la autora confirma de oído".** Todo es CC0, así que no hay obligación de dar crédito, pero se nombran en Créditos y en el README.

### Qué se escogió

| Cue | Archivo (carpeta de `sound/audition/`) | Autor / fuente | Licencia | En el juego |
|---|---|---|---|---|
| `music_menu` | `04_shop_theme_bossa` | CleytonKauffman, OpenGameArt ("Shop Theme") | CC0 | Menú, créditos y ajustes, en bucle |
| `music_turno` | `01_calm_track` | pmiller, OpenGameArt ("Calm Track") | CC0 | Turno, Demo y reporte, en bucle. Pesa 3,4 MB (no es enorme), así que no hizo falta usar la de guitarra (`02_etirwer_guitarra`, 2 MB, queda de reserva) |
| `sfx_alta` | `03_glass_ding` | Kenney, Interface Sounds | CC0 | Alta (de quien sea, también en la Demo) |
| `sfx_despierta` | `03_pluck_a` | Kenney, Interface Sounds | CC0 | El "!" de quien despierta, solo si está en la cámara que se mira |
| `sfx_click` | `03_ui_click` | Kenney, UI Audio | CC0 | Todo clic, Enter, Espacio o Esc |
| `sfx_sin_cama` | `04_close` (no `01`/`02`: pueden sonar duros) | Kenney, Interface Sounds | CC0 | El aviso "No hay cama libre" |
| `sfx_fin_turno` | `04_steel_a` (steel drum) | Kenney, Music Jingles | CC0 | Al abrirse el reporte de fin de turno |
| `sfx_ronquido` | `02_snore_b` | rubberduck, OpenGameArt | CC0 | Muy bajito y de vez en cuando (cada 6 a 14 s, sorteado con la semilla del turno) si hay alguien durmiendo en la cámara que se mira |
| `sfx_camilla` | `01_rolling_loop` | rubberduck, OpenGameArt | CC0 | Bajito, en bucle, mientras la cámara muestre una camilla |

Los OGG originales (sin tocar) están en `gameview/assets/sound/` con el nombre del cue (`music_menu.ogg`, `sfx_alta.ogg`...): 6,9 MB en total. No hay `sfx_pasos` (era opcional).

### Cómo se igualó el volumen

Se **mide** cada archivo (`measureLoudness`, en `gameview/sound_mix.go`): el pico y el RMS "de la parte fuerte" (ventanas de 10 ms, solo las que están a menos de 20 dB de la más fuerte, para que un silencio al inicio no engañe). La ganancia es `objetivo - RMS medido`, sin dejar que ningún pico pase de -1 dBFS. Objetivos: música -20 dBFS, efectos -18 dBFS (la música queda 2 dB abajo), ronquido -30 y camilla -32 (suenan seguido, de fondo). La ganancia se aplica al cargar el sonido (`gainReader`), sin tocar el archivo. La tabla está en `gameview/sound_gain.go`, y dos tests (`TestEverySoundHasGain`, `TestSoundsAreNormalized`) fallan si falta un archivo o si un archivo no queda en su objetivo. Para volver a medir: `SOUND_MEASURE=1 go test ./gameview -run TestMeasureSounds -v`.

| Archivo | RMS antes | Pico antes | Objetivo | Ganancia | RMS después | Pico después |
|---|---|---|---|---|---|---|
| `music_menu` | -16,4 | -0,3 | -20 | -3,6 dB | -20,0 | -3,9 |
| `music_turno` | -14,6 | -0,0 | -20 | -5,4 dB | -20,0 | -5,4 |
| `sfx_alta` | -15,7 | -1,0 | -18 | -2,3 dB | -18,0 | -3,3 |
| `sfx_despierta` | -16,5 | -0,0 | -18 | -1,5 dB | -18,0 | -1,5 |
| `sfx_click` | -17,5 | -1,5 | -18 | -0,5 dB | -18,0 | -2,1 |
| `sfx_sin_cama` | -14,8 | -0,9 | -18 | -3,2 dB | -18,0 | -4,1 |
| `sfx_fin_turno` | -16,0 | -2,6 | -18 | -2,0 dB | -18,0 | -4,6 |
| `sfx_ronquido` | -23,5 | -5,7 | -30 | -6,5 dB | -30,0 | -12,2 |
| `sfx_camilla` | -26,6 | -13,4 | -32 | -5,4 dB | -32,0 | -18,8 |

(Todo en dBFS. Los archivos originales venían fuertes: la música a -15 dB RMS y con picos pegados a 0, por eso todas las ganancias son negativas.)

### Cómo suena el conjunto (reglas)

- **Fundido**: al cambiar de pista (menú a turno y al revés) la música baja en 0,25 s y la nueva sube en otros 0,25 s (`musicStep`, `fadeVolume`).
- **Pausa**: la música baja al 40 % (en 0,3 s) y vuelve al salir.
- **Ajustes, fila "Sonido: Prendido / Apagado"** (por defecto Prendido, se guarda): apagado no suena nada y la música se pausa.
- **Modo pruebas (`-agent`)**: siempre en silencio; el volcado trae `sounds` y `music` (ver [[Arquitectura#Puerta de pruebas (-agent)]]).
- Si el computador no tiene tarjeta de sonido o falla, el juego sigue sin sonido (no se cae).

### Qué escuchar primero (la autora)

1. **El volumen de la música contra los efectos** en un turno (la música a -20 y los efectos a -18 dBFS es solo un número: de oído puede quedar alta o baja; se cambia en `targetMusicDB` y `targetSFXDB` y se vuelve a medir).
2. **`music_turno` (Calm Track)**: es la que más tiempo se oye; ¿relaja o cansa? Si cansa, la alternativa es `02_etirwer_guitarra`.
3. **El ronquido y la camilla**: son aproximaciones (ya se dijo que no hay ruedas de camilla reales); ¿molestan a ese volumen o ni se oyen?
4. **El clic con cada Enter/clic**: suena en todo clic, también si se toca el vacío; si es demasiado, se puede quitar de las teclas.
5. **El steel drum al terminar el turno** (`sfx_fin_turno`) sobre la música del turno que sigue sonando.

### Pendiente / no verificado

- Nadie ha oído nada (ni la IA ni, todavía, la autora). Lo único comprobado es que decodifica, que los volúmenes medidos quedan en su objetivo y que el juego arranca con el audio real sin caerse.
- No se probó el empalme del bucle de las dos músicas (las páginas dicen que ya vienen en bucle). Si se oye un salto al reiniciar, hay que cortarlas.

### Volúmenes de música y efectos (2026-10-19)

Ajustes vuelve a tener las dos barras aprobadas (Volumen música 70 % y Volumen efectos 80 % por defecto, de 10 en 10), además del interruptor **Sonido** (silencio total). Cada porcentaje pasa por `volumeGain` (`sound_mix.go`): **ganancia = (porcentaje / 100)²**. El oído no es lineal, así que una curva cuadrática hace que cada paso de la barra se sienta parejo: 50 % es -12 dB, 10 % es -40 dB, 0 % es silencio exacto y 100 % deja el nivel ya normalizado de cada archivo. La música multiplica `fundido × pausa × volumen`; los efectos (incluidos el ronquido y la camilla) multiplican su volumen, y `Sound.SetVolumes` se llama en cada tick, así lo que se arrastra se oye al instante. Un guardado sin estos campos toma 70 / 80. Como siempre, los valores por defecto son de la IA: la autora confirma de oído.
