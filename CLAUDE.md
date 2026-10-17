# HospitalCosteños: reglas para Claude y sus agentes

Proyecto de POO en Go (Universidad EIA) para Maria José Sosa. Sobre el modelo del enunciado (`hospital/`) se construye un juego pixel art de vigilancia de cámaras, en la rama **`game`**.

Entrega del enlace del repo: 2026-10-19 23:59. Sustentación individual: 2026-10-20. La autora tiene que poder explicar **cada línea**.

## Reglas que no se rompen

1. **`hospital/` no cambia.** Ni una línea (lo exige el bonus de interfaz). `git diff main -- hospital/` debe salir vacío.
2. La lógica del juego va en `game/` y solo usa la librería estándar. El dibujo y la entrada van en `gameview/`, que es el **único** paquete que importa Ebitengine (`github.com/hajimehoshi/ebiten/v2`).
3. `main.go` solo arma el escenario, imprime y arranca cosas: `go run .` abre el juego y `go run . -web` el dashboard viejo (sin juego).
4. Identificadores en **inglés**. Comentarios en **español**, explicativos y pensados para estudiar. Los textos del juego van en **español con sabor costeño**.
5. Código simple y defendible: nada de patrones rebuscados. Funciones puras con tests de tabla para tiempos, puntos y probabilidades.
6. Una pantalla solo se programa si su vista de Figma está **✅ aprobada** en el Art log (`docs/HospitalCosteños/Dirección de Arte.md`).

## Comandos de verificación (antes de cada commit)

```
gofmt -l .          # no imprime nada
go vet ./...        # no imprime nada
go test ./...
go test -race ./... # necesita gcc en el PATH (ya instalado)
```

## Git

- Rama `game`. Mensajes de commit en inglés, estilo convencional (`feat(game): ...`).
- **Nunca** agregar `Co-Authored-By` ni ninguna mención a Claude en los commits.
- Los agentes hacen commit solo de sus propios archivos y **nunca hacen push**: eso lo hace la sesión principal.
- Los finales de línea se fuerzan a LF (`.gitattributes`). No editar archivos con `Get-Content`/`Set-Content` de PowerShell 5.1, porque rompen el UTF-8.

## Documentación (bóveda de Obsidian en `docs/HospitalCosteños/`)

| Nota | Qué tiene |
|---|---|
| `Game Design.md` | Reglas del juego y **registro de decisiones** con las palabras de la autora |
| `Lógica del juego.md` | Especificación de la capa `game` (estados, `StaffMember`, acciones, puntos, tests) |
| `Mapa de Cámaras.md` | Cámaras y vecinas |
| `Dirección de Arte.md` | Estilo, medidas, flujo de sprites y **Art log** |
| `Elenco.md`, `Textos del Director.md` | Personajes y diálogos |

Lee las notas **por secciones**, no completas.

## Arte (sin Aseprite)

- La paleta está en `art/palette.hex`. Los sprites son cuadrículas de texto en `art/sprites/*.txt` y `go run ./tools/sprites` los convierte en PNG en `art/png/`.
- Figma se usa con Figwright (`mcp__figwright__*`; primero `ping`).
- Los PNG para revisar van en `docs/HospitalCosteños/preview/<tema>/`.
