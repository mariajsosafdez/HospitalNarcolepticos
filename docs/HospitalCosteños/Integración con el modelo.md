# Integración con el modelo

Resumen corto. **El detalle completo está en [[Lógica del juego]].**

## La regla que manda

Del enunciado, sección 7: *el paquete `hospital` no debe cambiar al agregar la interfaz y no puede tener código de visualización.* Todo lo que es "de juego" va en un paquete aparte, `game`.

## Qué ya da el modelo

| Juego | Modelo |
|---|---|
| Llega un paciente | `AdmitPatient(p)` |
| El jugador asigna personal a un dormido | `RegisterEpisode(p, location)` |
| Se lleva a la primera cama libre | `AssignRoom` (lo llama `RegisterEpisode`) |
| Sin cama → se queda en el pasillo + error | `ErrNoRoomAvailable` |
| Mover del pasillo a una cama que se liberó | `AssignRoom(p)` |
| "Dar de alta" | `WakeUpPatient(p)` libera la cama |
| Cupo del doctor | `Doctor.IsAvailable()`, `Doctor.Patients()` |
| Reporte final | las 4 consultas |

## Huecos y cómo quedaron

| # | Hueco | Decisión |
|---|---|---|
| 1 | El jugador elige quién atiende, pero el modelo elige solo | Envoltorio `StaffMember` que cumple `Attender` (ver [[Lógica del juego#4. Personal (`staff.go`)]]) |
| 2 | Dormirse ≠ ser atendido | ✅ Confirmado: el estado "dormido sin atender" vive solo en `game`; `RegisterEpisode` se llama al asignar |
| 3 | Despertar ≠ liberar cama | ✅ Confirmado: la cama sigue ocupada hasta "Dar de alta" → `WakeUpPatient` |
| 4 | Cupo del doctor (4, no se libera) | ✅ Confirmado: pacientes por partida cerca del cupo total de doctores, y penalización por "exceso de trabajo" al llenar el cupo |
| 5 | `EpisodeRecord` guarda hora real | El juego guarda aparte la hora de juego de cada episodio |
| 6 | Concurrencia | Lógica del juego en una sola goroutine; el hospital mantiene su mutex |
| 7 | `h.Doctors()` sale vacío con el envoltorio | El juego guarda sus propios `*Doctor` para el reporte 5.2 |
