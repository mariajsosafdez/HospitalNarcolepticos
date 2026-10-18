# Errores de la IA (para AI_USAGE.md)

Registro honesto de errores reales que cometieron Claude y sus agentes durante el juego, y cómo se corrigieron. El enunciado pide "al menos dos casos donde el código generado estaba mal o no era idiomático". Aquí van todos, para elegir los mejores al escribir `AI_USAGE.md`.

| Fecha | Quién | Qué pasó | Cómo se detectó | Arreglo |
|---|---|---|---|---|
| 2026-10-17 | Builder | `go get` de Ebitengine subió sin avisar `go 1.22` → `go 1.25` en `go.mod` | El mismo agente, al revisar el diff | Fijar Ebitengine v2.8.8, que acepta Go 1.22 |
| 2026-10-17 | Builder | Un comando de shell con comillas mal escapadas corrompió un `Printf` de `main.go` | El compilador | Usar las herramientas de edición en vez de heredocs grandes |
| 2026-10-17 | Builder | Un test esperaba la cámara equivocada después del alta (1B en vez de 4) | El test falló; el código estaba bien | Corregir el test |
| 2026-10-17 | Builder | Inventó un nombre de función de Ebitengine (`SetRunnableOnUnfocusedWindow`) que no existe | El compilador | Usar el real, `SetRunnableOnUnfocused` |
| 2026-10-17 | Builder | La vista adivinaba quién empujaba la camilla con "el personal en traslado", y en la Demo fallaba | Capturas del modo de prueba | Agregar `GamePatient.Carrier()` en `game/` |
| 2026-10-17 | Builder | Todo el texto se veía borroso: se dibujaba a 480×270 y luego se agrandaba | La sesión principal, al revisar las capturas ×4 | Dibujar a la resolución real de la ventana |
| 2026-10-17 | Builder | Al arreglar lo anterior, la imagen de la cámara quedó con doble zoom | Capturas del modo de prueba | `Scale = 1` dentro de `drawFeed` |
| 2026-10-17 | Builder | Después del cambio de resolución, el mouse quedó mal: el hover y los clics caían en otro lado (menú, mapa, pausa, Demo) | **La autora, jugando** | Commit `94ba621`: una sola función pura `windowToGame` convierte el cursor, con test de tabla para escalas 1–4. El modo de prueba ahora hace clic por ese mismo camino y prueba a ×2 y ×3. Lección: el modo de prueba hacía clic sin pasar por la conversión real del mouse, por eso los tests no lo vieron |
| 2026-10-17 | Arte | Sprites de cama y camilla con huecos en el contorno (faltaba la última columna en algunos) | La autora, mirando las imágenes, y un script de revisión de bordes | Redibujar las cuadrículas; el convertidor no tenía error |
| 2026-10-18 | Builder | La pausa no tenía selección con teclado y no había forma de elegir pacientes con teclas | **La autora, jugando** | Flechas + Enter en la pausa y para elegir paciente y personal (commit `b6a9c2c`) |
| 2026-10-18 | Builder | Mañe salía como un rectángulo: el juego buscaba el sprite por el primer nombre ("abelardo") y los archivos se llaman `pac_mane_*` | **La autora, jugando** | Buscar el sprite por ID, con un test que exige que exista el sprite de cada personaje |
| 2026-10-18 | Builders en paralelo | Dos builders compartían el área de preparación (index) de git: el de pantallas hizo commit y se llevó los archivos del de sonido (commit `12b521d`) | El builder de sonido, al ver que su commit quedaba vacío | No reescribir la historia; para la próxima, cada agente hace `git commit <archivos>` con rutas explícitas o se trabaja en worktrees separados |
| 2026-10-18 | Builder | Un `sed` pensado para la tarjeta compacta también cambió la tarjeta normal | Comparación automática de la captura contra la vista aprobada | Restaurar la tarjeta normal |
| 2026-10-18 | Builder | Supuso que los puntos del reporte iban alineados a la derecha, pero en el diseño van a la izquierda | Comparación de la captura con la vista aprobada | Alinearlos como en el diseño |
| 2026-10-18 | Builder | Python en Windows escribió finales de línea CRLF en archivos `.go` | `gofmt -l` | `gofmt -w` antes de cada commit |
| 2026-10-17 | Sesión principal | Un `python -` en un comando quedó esperando entrada y colgó la terminal | El comando no terminó en 2 minutos | Detenerlo y hacer las ediciones a mano |

## Otros datos para AI_USAGE.md

- Los agentes de arte generaron las cuadrículas grandes (fondos y personajes) con **scripts temporales de Python que no están en el repo**. Las cuadrículas `.txt` son la fuente real y cada una lo dice en un comentario.
- Se trabajó con un equipo de agentes: arte (Opus), builder y tester (Sonnet), coordinados por la sesión principal. La autora aprobó cada pantalla antes de programarla (ver el registro de decisiones en [[Game Design]]).
