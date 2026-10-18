# Plan de construcción

Orden de trabajo decidido con la autora: **la Demo primero**, luego el Día 1 con reporte ([[Game Design#Registro de decisiones]], "podríamos cambiar de orden el día 1 + reporte por la demo"). Entrega del repo: 2026-10-19 23:59. Sustentación: 2026-10-20.

Una pantalla solo se programa cuando su vista de Figma está ✅ aprobada en el Art log de [[Dirección de Arte]]. La estructura del código está en [[Arquitectura]] y las reglas en [[Lógica del juego]].

## Hitos

| # | Hito | Qué incluye | Estado |
|---|---|---|---|
| 0 | Base de reglas | Paquete `game/` completo y probado, banderas `-web` y `nogame`, ventana vacía de Ebitengine | ✅ hecho |
| 1 | **Demo** | Menú, vista de cámara, panel de personal, modo Demo con "Siguiente paso" y "Auto", reporte con las 4 consultas. Usa `game.NewDemo` | ✅ jugable: menú, cámaras, panel, mapa, Demo en 4 pasos y puerta de pruebas. Reporte provisional (sin vista aprobada) |
| 2 | **Día 1 + reporte** | Turno jugable de 3 min con 4 cámaras comunes + 3 habitaciones, asignar, mover a cama, dar de alta, pausa, reporte de fin de turno con estrellas y mensaje del Director | en curso: el Día 1 ya se juega (asignar, mover a cama, alta, pausa mínima, reporte provisional). Falta la vista de Pausa y del Reporte |
| 3 | **Días 2 a 5** | Más cámaras, doctores, camilleros y habitaciones por día, penalizaciones que suben, autoguardado y desbloqueo de días, ajustes | pendiente |
| 4 | **Sonido** | Música y efectos CC0/CC-BY, volúmenes de los ajustes. Cada elección "por defecto, la autora confirma de oído" | pendiente |
| 5 | **Tutorial** | Mensajes de Don Ernesto Cassiani en el Día 1 | pendiente |
| 6 | **Easter eggs** | Personajes extra, solo si sobra tiempo | pendiente |

Transversal: la **puerta de pruebas** (`-agent`, ver [[Arquitectura#Puerta de pruebas (-agent)]]) se construye con la primera pantalla y cada pantalla nueva debe poder alcanzarse con ella.

## Cosas abiertas que afectan al plan

- El Elenco tiene 10 pacientes, pero los días 3 a 5 piden 14 (doctores x 4 + 2). Por ahora esos días usan los 10. Hay que decidir si se aprueban 4 personajes más.
- La cámara de la cafetería no tenía línea en el boceto del plano; en `game/world.go` quedó unida al Pasillo Este. Revisar con la autora.
- La Demo muestra los lugares con los textos en inglés de `main.go` (para que las 4 consultas coincidan con la consola).

## Registro de avance

| Fecha | Qué | Commits | Cómo verlo |
|---|---|---|---|
| 2026-10-17 | Paquete `game/`: reloj, pacientes, `StaffMember`, turno con acciones, puntos y estrellas, días 1-5, reporte, Demo, guardado, `Snapshot` y 26 tests | `feat(game): add game rules layer...` | `go test ./game` |
| 2026-10-17 | `go run .` abre la ventana (480x270 x3, fondo crema), `-web` para el dashboard, etiqueta `nogame` | `feat: open the game window by default...` | `go run .`, `go run . -web`, `go run -tags nogame .` |
| 2026-10-17 | README, [[Arquitectura]] y este plan | `docs: add architecture and build plan...` | leer las notas |
| 2026-10-17 | Menú, pantalla de cámaras (turno y Demo), panel de personal, mapa, avisos, pausa y reporte provisionales, fuentes y arte embebidos, puerta de pruebas `-agent` | `feat(gameview): ...` | `go run .`; `go run . -agent -scene demo -step 4 -input guion.txt -quit` (ver [[Arquitectura#Puerta de pruebas (-agent)]]) |

Provisionales hasta que haya vista aprobada: "Cómo jugar", "Ajustes" y "Créditos" (aviso "Próximamente"), Pausa, Reporte (en la Demo y al final del turno), paginación del panel con más de 3 personas (días 3 a 5), avisos de error distintos al de "sin cama".
