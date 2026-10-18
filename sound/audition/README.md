# Candidatos de sonido (para escuchar)

Los archivos de audio de esta carpeta **no se commitean** (`.gitignore`); solo este README y `escuchar.ps1`.
La IA no puede oír: las descripciones salen de la página de cada pista (texto, etiquetas) y de medidas
(duración, volumen pico). **Todo es "por defecto, la autora confirma de oído".**

Escuchar todo, uno tras otro, con el nombre impreso (desde la raíz del repo):

```
powershell -ExecutionPolicy Bypass -File sound\audition\escuchar.ps1
powershell -ExecutionPolicy Bypass -File sound\audition\escuchar.ps1 -Cue music_menu
```

En cada carpeta hay el original (`.ogg` / `.wav`) y una copia `.wav` para escuchar: los efectos completos y,
de las músicas largas, un recorte de 30 s (`_30s.wav`, tomado desde ~20 % de la pista) hecho con un decodificador
de Go (no había ffmpeg). Duración = la del archivo original completo. Todos decodifican bien (comprobado).

Licencias: las leí en la página de cada pista; los packs de Kenney traen además `License.txt` (CC0) dentro del
zip, el de bicicletas trae `LICENSE.txt` (CC0). Los demás zips/archivos de OpenGameArt no traen licencia interna.

## Música del menú (`music_menu/`)

| Archivo | Fuente | Autor | Licencia | Dur. | Cómo debería sonar |
|---|---|---|---|---|---|
| `01_tropical_loop` | https://opengameart.org/content/tropical-loop | wipics | CC0 | 0:18 | "Tropical loop con tambores africanos"; etiquetas tropical, isla, conga, sintetizadores. Corto y animado, quizá más alegre que tranquilo. Pico 0.99 (fuerte) |
| `02_beach_town_guitar` | https://opengameart.org/content/beach-town-guitar-theme | Hitctrl | CC-BY 3.0 | 2:11 | "Tema simple de pueblo de playa con guitarra", se puede repetir. Pico 0.58 (suave) |
| `03_feel_good_island` | https://opengameart.org/content/feel-good-island-loop | Brandon Morris (recorte en loop: AntumDeluge) | OGA-BY 3.0 (la descripción dice que también hay CC0) | 0:52 | "Feel Good Island", tema de isla alegre; ya viene recortado para repetir |
| `04_shop_theme_bossa` | https://opengameart.org/content/shop-theme | CleytonKauffman | CC0 | 1:28 | Bossa nova / música de ascensor, repetible. Es la versión normal (el zip trae también una "radio", con sonido de parlante) |

Créditos CC-BY / OGA-BY si se escogen:
- `02`: "Beach Town Guitar Theme" por Hitctrl, CC BY 3.0 (https://creativecommons.org/licenses/by/3.0/), vía OpenGameArt.org.
- `03`: "Feel Good Island" de Brandon Morris, recortado en loop por AntumDeluge, OGA-BY 3.0, vía OpenGameArt.org.

## Música del turno (`music_turno/`), más suave y larga

| Archivo | Fuente | Autor | Licencia | Dur. | Cómo debería sonar |
|---|---|---|---|---|---|
| `01_calm_track` | https://opengameart.org/content/calm-track | pmiller | CC0 | 4:09 | Etiquetas calm, chillout; "va evolucionando despacio". Viene la versión ya en loop |
| `02_etirwer_guitarra` | https://opengameart.org/content/etirwer | Kistol | CC0 | 2:10 | Guitarra acústica de nailon, sencilla y en loop sin cortes (según la autora). Pico 0.63 |
| `03_serenade_guitar` | https://opengameart.org/content/serenade-guitar | Pro Sensory | CC0 | 2:09 | Una guitarra (serenata); la página no dice más. El WAV pesa 22 MB; convendría pasarlo a OGG |
| `04_green_salon` | https://opengameart.org/content/green-salon | Yubatake | CC-BY 4.0 | 5:01 | Bossa nova con vibráfono, 65 BPM, muy tranquila. No está hecha para loop |

Crédito: "Green Salon" por Yubatake, CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/), vía OpenGameArt.org.

## Efectos

Todos CC0. Kenney = https://kenney.nl/assets/<pack> (autor Kenney, `License.txt` CC0 dentro del zip).
Rubberduck = OpenGameArt.

| Cue | Archivo | Fuente / autor | Dur. | Cómo debería sonar |
|---|---|---|---|---|
| sfx_ronquido | `01_snore` | creature SFX, rubberduck (opengameart.org/content/80-cc0-creature-sfx) | 0.44 s | Un ronquido corto, suave (pico 0.22) |
| | `02_snore_b` | creature SFX #2, rubberduck (opengameart.org/content/80-cc0-creture-sfx-2) | 0.93 s | Otro ronquido, más largo y más fuerte |
| | `03_sleep_grunts` | Little Grunts, crueltyinc (opengameart.org/content/little-grunts) | 6.3 s | Ruidos reales de alguien durmiendo, grabados con un rastreador de sueño; bajito (0.28). Se puede cortar en trozos |
| | `04_breath`, `05_breath_b` | las mismas dos packs, rubberduck | 0.67 s / 0.36 s | Respiración, para mezclar con el ronquido (`05` suena fuerte) |
| sfx_camilla | `01_rolling_loop` | 30 CC0 SFX loops, rubberduck (opengameart.org/content/30-cc0-sfx-loops) | 1.06 s | "Rolling sound" en loop; bajito (0.21). Es lo más cercano a ruedas |
| | `02_spinning_wheel` | Spinning Wheel, themightyglider (opengameart.org/content/spinning-wheel-0) | 1.16 s | Una rueda girando. Derivado de un sonido de freesound (stransky); su licencia original no la verifiqué |
| | `03_bicycle_spokes` | Bicycle Sounds, AntumDeluge (opengameart.org/content/bicycle-sounds) | 1.16 s | Rayos de rueda de bicicleta |
| | `04_wood_squeak`, `05_creak` | 100 CC0 metal and wood SFX, rubberduck; Kenney rpg-audio | 0.48 s / 0.66 s | Chirridos de madera, para mezclar encima de las ruedas |
| sfx_alta | `01_confirmation_a`, `02_confirmation_b` | Kenney interface-sounds | 0.29 s / 0.54 s | Confirmación suave, tipo "ding" de interfaz |
| | `03_glass_ding`, `04_glass_small` | Kenney interface-sounds (glass) | 0.28 s / 0.12 s | Toquecito de vidrio, campanita corta |
| | `05_bong` | Kenney interface-sounds | 0.12 s | "Bong" corto |
| sfx_despierta | `01_question_a`, `02_question_b` | Kenney interface-sounds | 0.49 s / 0.33 s | Sonido tipo pregunta/atención, suave |
| | `03_pluck_a`, `04_pluck_b` | Kenney interface-sounds | 0.10 s / 0.16 s | "Pop" de cuerda pulsada, muy corto |
| sfx_click | `01_select` | Kenney interface-sounds (select_002) | 0.04 s | Clic seco |
| | `02_click_a` | Kenney interface-sounds (click_001) | 0.10 s | Clic de interfaz |
| | `03_ui_click`, `04_ui_rollover` | Kenney ui-audio | 0.09 s / 0.23 s | Clic y "rollover" |
| sfx_sin_cama | `01_error_a`, `02_error_b` | Kenney interface-sounds (error_004, error_006) | 0.10 s / 0.50 s | Error de interfaz; ojo, puede sonar más duro que "suave" |
| | `03_back`, `04_close` | Kenney interface-sounds | 0.06 s / 0.31 s | Más suaves: "atrás" / "cerrar" |
| sfx_fin_turno | `01_pizzicato_a`, `02_pizzicato_b` | Kenney music-jingles (Pizzicato) | 0.49 s / 0.56 s | Jingle corto de cuerdas pulsadas, alegre |
| | `03_sax` | Kenney music-jingles (Sax) | 0.39 s | Saxofón corto (pico 0.31, bajito) |
| | `04_steel_a`, `05_steel_b` | Kenney music-jingles (Steel) | 0.93 s / 0.91 s | Jingle de "steel drum", tropical; los más cercanos al ambiente costeño |
| sfx_pasos | `01_footstep_a` ... `04_footstep_d` | Kenney rpg-audio (footstep00..03) | 0.24-0.28 s | Pasos sueltos de un juego de rol (probablemente suelo duro/tierra, no baldosa) |

## No conseguido

- **Cumbia / champeta de verdad**: ninguna pista CC0/CC-BY de OpenGameArt lleva esas etiquetas (búsqueda "cumbia": 0; "champeta": 0). Lo más cercano: bossa nova, guitarra, isla, steel drum.
- **Ruedas de camilla**: no hay una de verdad en CC0; las de arriba son aproximaciones.
- **Ronquidos**: solo hay 2 ronquidos reales en CC0 más grabaciones de sueño y respiración.
- "Latin Loop" (Macro, CC-BY 3.0, electrónica): descartada, es WAV de 24 bits y Ebitengine solo lee 8/16.
- freepd.com (la página de música del mundo devolvió 404), incompetech y Free Music Archive: no pude confirmar pistas latinas/tranquilas con licencia clara, no los usé. freesound.org no se intentó (pide login).
