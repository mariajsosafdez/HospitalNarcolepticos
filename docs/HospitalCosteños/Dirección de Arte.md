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

## Paleta (propuesta, esperando aprobación)

31 colores en `art/palette.hex` (clave de un carácter, `#rrggbb`, nombre). Muestrario: `preview/paleta/paleta.png`.

- **Contornos y sombras:** `0` cacao, `1` café oscuro, `2` ladrillo oscuro (nada de negro puro); grises cálidos `x`, `X`.
- **Cálidos:** terracota `t` `T`, coral `c` `C`, naranja `o`, mango `m` `M`, amarillo sol `y`.
- **Paredes:** crema `w` `W`, arena `k` `s`. **Hospital:** menta `a` `A`, verde agua `q`. **Acento:** azul mar `b`, cielo `B`, noche `n`.
- **Atardecer:** rosa `p`, morado `v`. **Vegetación:** `l` `g` `G`.
- **Manillas:** Leve `L` verde, Moderado `Y` amarillo, Severo `R` rojo.
- **Luz por hora:** mediodía = `B` `w` `y`; tarde dorada = `M` `m`; atardecer = `o` `p` `v` `n`.

## Fuente (propuesta, esperando aprobación)

- **Pixelify Sans** (Stefie Justprince), licencia **SIL OFL 1.1**, Google Fonts: https://fonts.google.com/specimen/Pixelify+Sans. Subconjuntos latin y latin-ext; probada en Figma con á é í ó ú ü ñ Ñ ¿ ¡. Variable 400–700 (Regular y Bold).
  Tamaños a 1×: **18** Bold (título), **10** (opciones del menú), **8** (subtítulo, descripciones, ayudas).
- **VT323** (Peter Hull), **SIL OFL 1.1**, Google Fonts: https://fonts.google.com/specimen/VT323. Solo para el sello de cámara (`REC  CAM 1A · 17:45`), 10 px a 1×.
- Aún no se empaqueta el archivo `.ttf` en el repo.

## Art log

| Fecha | Qué | Dónde | Estado |
|---|---|---|---|
| 2026-10-17 | Paleta cálida de 31 colores | `art/palette.hex`, `preview/paleta/paleta.png`; Figma: página "Hospital Costeños — Pantallas", sección "Paleta (2026-10-17)" | Esperando aprobación |
| 2026-10-17 | Fuentes: Pixelify Sans + VT323 (OFL) | Sección "Fuente" de esta nota | Esperando aprobación |
| 2026-10-17 | Sprites del menú: `menu_fondo`, `palmera` (2 cuadros), `banca_sombrero`, `ui_panel_menu`, `ui_seleccion`, `ui_flecha` (2), `ui_rec` (2) | `art/sprites/*.txt` → `art/png/` (`go run ./tools/sprites`) | Esperando aprobación |
| 2026-10-17 | Vista del menú principal: `Menu / default`, `Menu / hover Demo`, `Menu / sin partida guardada` | Figma: página "Hospital Costeños — Pantallas", sección "Menú principal (2026-10-17)"; `preview/menu/menu-default.png`, `menu-hover-demo.png`, `menu-sin-partida.png` | Esperando aprobación |
