# Sonido

> 2026-10-19: el juego se entrega sin sonido (decisión de la autora: "no creo que nos dé el tiempo"). Esta nota queda como registro de la búsqueda.

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
