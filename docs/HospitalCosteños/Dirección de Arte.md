# Dirección de Arte

> Dirección general aprobada el 2026-10-17: "me gusta la dirección de arte, la manilla de cada nivel como la recomendaste". La paleta exacta y la fuente las propone el agente de arte, y quedan registradas en el [[#Art log]].

## Estilo

- **Pixel art**, ambiente **cálido y relajado**. El sistema de cámaras se inspira en FNAF, pero **no** su paleta ni su terror.
- La cámara se ve con grano suave, tinte cálido, viñeta leve, un punto `REC` y el sello `CAM 2 · 14:15`. Nada de estática agresiva ni sustos.
- La luz de las ventanas cambia con la hora del turno: **mediodía (12–14) → tarde dorada (14–16) → atardecer naranja (16–18)**.
- La paleta es limitada (unos 24–32 colores): terracota, mango, coral, crema en las paredes, menta o verde agua en lo de hospital, azul mar de acento. Se guarda en `art/palette.hex`.

## Medidas

| Cosa | Tamaño a 1× |
|---|---|
| Pantalla del juego | 480×270 (se escala ×2, ×3 o ×4) |
| Frames de Figma | 1920×1080 (= 480×270 × 4) |
| Personajes adultos | ~16×24 px |
| Retrato del Director | ~48×48 px |
| Iconos del panel | 8×8 o 12×12 px |

## Distribución de la pantalla de cámaras

- **Cámara** a la izquierda (~70 % del ancho).
- **Panel de personal** a la derecha: una tarjeta por persona con nombre, rol, estado, barra de cooldown y cupo `x/4` en los doctores.
- **Mapa** abajo a la derecha, con los botones `CAM 1A`, `CAM 2`… (ver [[Mapa de Cámaras]]).
- **Barra superior:** hora del turno, día y estrellas actuales.
- No hay "oficina": siempre se juega desde las cámaras.

### Pantalla de cámaras: medidas ✅

- La imagen de la cámara ocupa (4,20) 336×246 a 1×. Sus fondos miden **168×123** y se dibujan **a ×2**: así los personajes de 16×24 se ven de 32×48 y los cuartos quedan a escala de persona. El panel, el mapa y la barra superior van a 1×.
- Panel de personal en x = 344..476 (tarjetas de 132×36); mapa en (344,156) 132×110; barra superior 480×16.
- Encima de la cámara, a 1×: tinte por hora, viñeta tramada (35 %) y grano (12 %, 2 cuadros).
- Cada fondo trae 3 cuadros: mediodía, tarde dorada y atardecer.
- Medidas exactas por capa: notas de Figma en la sección "Cámaras — Demo (2026-10-17)".
- Retoques del 2026-10-17: los rótulos de la cámara (`REC`, sello y lugar) llevan **borde 0 de 1 px**; los pisos ya no tienen mancha de sol (solo cambian las ventanas); **una sola cama por habitación**, en (34,72) del fondo; los botones bloqueados del mapa muestran `ui_candado` al pasar el mouse.
- Con 4 o 5 personas (días 3–5) el panel usa `ui_tarjeta_compacta` (132×24) para que quepan todas sin pasar páginas.

## Manillas por nivel ✅

| Nivel | Color |
|---|---|
| Leve | verde |
| Moderado | amarillo |
| Severo | rojo |

## Texto

- En **español con sabor costeño** (ver [[Textos del Director]]).
- Fuente pixel con licencia libre (OFL o similar) que tenga **á é í ó ú ü ñ ¿ ¡**. Si es de Google Fonts, Figma ya la tiene disponible.

## Flujo de trabajo del arte (sin Aseprite)

1. Cada sprite es un archivo de texto en `art/sprites/<nombre>.txt`: una cuadrícula donde cada carácter es un color de `art/palette.hex` (`.` = transparente). Las animaciones son varios cuadros en el mismo archivo.
2. `go run ./tools/sprites` convierte esos archivos en PNG en `art/png/`, a 1× y a 4×.
3. Las pantallas se arman en **Figma con Figwright**, usando esos PNG y la fuente real.
4. Las copias PNG para revisar van en `docs/HospitalCosteños/preview/<tema>/`, así se ven aquí mismo en Obsidian.

## Paleta ✅

31 colores en `art/palette.hex` (clave de un carácter, `#rrggbb`, nombre). Muestrario: `preview/paleta/paleta.png`.

- **Contornos y sombras:** `0` cacao, `1` café oscuro, `2` ladrillo oscuro (nada de negro puro); grises cálidos `x`, `X`.
- **Cálidos:** terracota `t` `T`, coral `c` `C`, naranja `o`, mango `m` `M`, amarillo sol `y`.
- **Paredes:** crema `w` `W`, arena `k` `s`. **Hospital:** menta `a` `A`, verde agua `q`. **Acento:** azul mar `b`, cielo `B`, noche `n`.
- **Atardecer:** rosa `p`, morado `v`. **Vegetación:** `l` `g` `G`.
- **Manillas:** Leve `L` verde, Moderado `Y` amarillo, Severo `R` rojo.
- **Luz por hora:** mediodía = `B` `w` `y`; tarde dorada = `M` `m`; atardecer = `o` `p` `v` `n`.

## Fuente ✅

- **Pixelify Sans** (Stefie Justprince), licencia **SIL OFL 1.1**, Google Fonts: https://fonts.google.com/specimen/Pixelify+Sans. Subconjuntos latin y latin-ext; probada en Figma con á é í ó ú ü ñ Ñ ¿ ¡. Variable 400–700 (Regular y Bold).
  Tamaños a 1×: **18** Bold (título), **10** (opciones del menú, línea de ayuda y **toda línea que lleve ¡**: a 8 px el ¡ se confunde con una i), **8** (subtítulo, descripciones).
  **Regla de pantalla (2026-10-17):** en el juego se quita la ¡ inicial y se deja el ! final ("No hay cama libre!"), porque la ¡ de Pixelify Sans parece una i.
- **VT323** (Peter Hull), **SIL OFL 1.1**, Google Fonts: https://fonts.google.com/specimen/VT323. Solo para el sello de cámara (`REC  CAM 1A · 17:45`), 10 px a 1×.
- Aún no se empaqueta el archivo `.ttf` en el repo.

## Art log

| Fecha | Qué | Dónde | Estado |
|---|---|---|---|
| 2026-10-17 | Paleta cálida de 31 colores | `art/palette.hex`, `preview/paleta/paleta.png`; Figma: página "Hospital Costeños — Pantallas", sección "Paleta (2026-10-17)" | ✅ Aprobada junto con el menú |
| 2026-10-17 | Fuentes: Pixelify Sans + VT323 (OFL) | Sección "Fuente" de esta nota | ✅ Aprobadas junto con el menú |
| 2026-10-17 | Sprites del menú: `menu_fondo`, `palmera` (2 cuadros), `banca` (antes `banca_sombrero`: se quitó el sombrero), `ui_panel_menu`, `ui_seleccion`, `ui_flecha` (2), `ui_rec` (2; ya no sale en el menú, se guarda para las cámaras) | `art/sprites/*.txt` → `art/png/` (`go run ./tools/sprites`) | ✅ Aprobado junto con el menú |
| 2026-10-17 | Vista del menú principal: `Menu / default`, `Menu / hover Demo`, `Menu / sin partida guardada` | Figma: página "Hospital Costeños — Pantallas", sección "Menú principal (2026-10-17)"; `preview/menu/menu-default.png`, `menu-hover-demo.png`, `menu-sin-partida.png` | ✅ Aprobado con cambios: quitar el sombrero ("yo creo que es mejor quitarlo"), quitar esquinas de cámara y REC ("nop"), líneas que empiezan con ¡ a tamaño 10 ("ok"), atardecer fijo, descripciones aprobadas ("me gustan como estan") |
| 2026-10-17 | Sprites del elenco de la Demo: `pac_{yesid,yeimy,rafa,wilfrido,marelvis}_{camina,quieto,dormido,cama,sentado}`, `doc_{karen,alvaro}_{camina,quieto,camilla}`, `cam_chepe_{camina,quieto,camilla}`, `cama`, `director_retrato` (48×48), `fx_zzz` (3), `fx_alerta` (2). Caminar: 4 cuadros mirando a la derecha (a la izquierda se refleja en el código) | `art/sprites/*.txt` → `art/png/`; hojas en `preview/elenco/*.png` | ✅ Aprobado con retoques (2026-10-17): "si". Ver la fila de retoques abajo |
| 2026-10-17 | Fondos de cámara (3 cuadros por hora): `fondo_cam2_cafeteria`, `fondo_cam8_champeta`, `fondo_habitacion` (7A–7C); capas `camara_vineta`, `camara_grano`; UI `ui_barra_superior`, `ui_estrella`, `ui_tarjeta`, `ui_chip_estado`, `ui_boton_cam`, `ui_mapa` | `art/sprites/*.txt` → `art/png/` | ✅ Aprobado con retoques (2026-10-17): "si". Ver la fila de retoques abajo |
| 2026-10-17 | Vistas de cámaras: `Cam / zona común`, `Cam / paciente seleccionado`, `Cam / sin cama`, `Cam / habitación`, `Cam / Demo controles`, `Panel / estados del personal`, `Ref / luz por hora (CAM 2)` | Figma: página "Hospital Costeños — Pantallas", sección "Cámaras — Demo (2026-10-17)"; `preview/camaras/*.png` | ✅ Aprobado con retoques (2026-10-17): "si". Ver la fila de retoques abajo |
| 2026-10-17 | **Retoques pedidos por la autora** a menú, elenco y cámaras | `art/sprites/*.txt` → `art/png/`; Figma: secciones "Menú principal" y "Cámaras — Demo"; `preview/menu/`, `preview/camaras/`, `preview/elenco/` | Hechos (2026-10-17), esperando visto bueno: ✔ ayuda del menú "Flechas: elegir · Enter: dale" ("quita los ¡!"); ✔ borde 0 de 1 px en REC, sello y lugar de la cámara ("hazle borde negro al texto"); ✔ batas sin cambios ("deja las batas así"); ✔ `pac_wilfrido_dormido`: la nevera queda detrás de él ("un poco más adelante de la nevera"); ✔ los 10 `pac_*_cama` / `pac_*_sentado` redibujados: acostados bajo la cobija con el brazo abajo y la manilla, y sentados con los brazos abajo ("ponles los brazos hacia abajo normales"); el sentado parpadea en vez de estirarse; ✔ `pac_yeimy_dormido`: brazos doblados en el codo que salen del hombro ("corrigele los brazos"); ✔ cobija sin huecos y con borde: el fallo de la última columna estaba en las cuadrículas (la cobija de las camas terminaba en azul en la columna 27 y la olla de Marelvis en la columna 16), no en `tools/sprites`; camillas de Karen, Álvaro y Chepe con la misma cobija que Luz Dary y el Mono; ✔ sin reflejos de ventana en los pisos (cafetería, champeta, habitación); ✔ una sola cama por habitación y más decoración (matas, cuadro del mar, florero, radio, silla, sombrero vueltiao, esterilla), vistas `Cam / habitación` y `Cam / habitación dormido`; ✔ candado rojo `ui_candado` al pasar el mouse por una cámara bloqueada, vista `Mapa / hover cámara bloqueada`; ✔ fondos nuevos `fondo_cam4_pasillo_oeste` y `fondo_cam3_pasillo_este` (vista `Ref / pasillos`) y `Cam / sin cama` rehecha con Wilfrido dormido en CAM 4 |
| 2026-10-17 | Pantallas del turno: `Director / mensaje`, `Reporte / fin de turno`, `Pausa`, `Ajustes`, `Panel / 5 personas`; sprite nuevo `ui_tarjeta_compacta` (132×24) | Figma: página "Hospital Costeños — Pantallas", sección "Pantallas del turno (2026-10-17)"; `preview/turno/*.png`; textos en [[Textos del Director#Pantallas del turno]] | ✅ Aprobado (2026-10-18): "sip". En la Demo y el Día 5 solo "Menú"; "Salir al menú" abandona el turno sin guardar |
| 2026-10-18 | Retoque de Yeimy (`pac_yeimy_dormido`, cuadro con los brazos arriba) | `art/sprites/pac_yeimy_dormido.txt` | Pendiente: "en el frame donde sube los brazos, la parte de abajo no sea un solo pixel, sino 3, ya que parece que se le separan los brazos, sumado a eso, subele un poco más los brazos cuando ella los alce, tipo unos 2 pixeles más, asi se ve más cuando baila" |
