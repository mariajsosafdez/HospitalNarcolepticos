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

## Art log

| Fecha | Qué | Dónde | Estado |
|---|---|---|---|
