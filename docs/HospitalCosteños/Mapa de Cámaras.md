# Mapa de Cámaras

> ✅ Propuesta de cámaras aprobada el 2026-10-17: "Me gusta la propuesta de cámaras que me dices." Los pacientes se mueven entre cámaras.

Inspirado en FNAF 1: un plano del edificio en la esquina inferior derecha con botones `CAM 1A`, `CAM 1B`… El jugador está en la **Sala de monitoreo** (el equivalente a la oficina de FNAF) y ve una cámara a la vez.

Hay dos tipos de cámara:

- **Zonas comunes**: por aquí caminan los pacientes y aquí se duermen. Esos lugares son el `location` del episodio en el modelo.
- **Habitaciones**: aquí duermen los pacientes ya atendidos y aquí hay que ver cuándo despiertan.

## Cámaras

| Cam   | Lugar                 | Tipo            | Qué pasa ahí                                                 |
| ----- | --------------------- | --------------- | ------------------------------------------------------------ |
| 1A    | Entrada / Recepción   | común           | Llegan los pacientes nuevos                                  |
| 1B    | Sala de espera        | común           | Sillas, ventilador de techo; sitio clásico para dormirse     |
| 2     | Cafetería             | común           | El sancocho. Mucha probabilidad de dormirse después de comer |
| 3     | Pasillo Este          | común / pasillo | Donde se quedan los que no consiguen cama                    |
| 4     | Pasillo Oeste         | común / pasillo | Igual que el 3                                               |
| 5     | Radiología            | común           | La "fila de radiología" del ejemplo del enunciado            |
| 6     | Patio de las palmeras | común           | Guiño al enunciado: se duermen de pie contra una palmera     |
| 8     | Salón de champeta     | común           | Sala de terapia de baile. Aquí se duerme Yeimy en la Demo (el "champeta dance hall" de `main.go`) ✅ |
| 7A    | Habitación 101        | habitación      |                                                              |
| 7B    | Habitación 102        | habitación      |                                                              |
| 7C    | Habitación 103        | habitación      |                                                              |
| 7D–7F | Habitaciones 104–106  | habitación      | Se desbloquean en días siguientes                            |

## Boceto del plano

```
 ┌──────────── HOSPITAL DE LOS COSTEÑOS ────────────┐
 │  [6 Patio]   [8 Champeta]   [2 Cafetería]          │
 │      │                   │                        │
 │ [4 Pasillo O]──[1B Espera]──[3 Pasillo E]──[5 Rx] │
 │      │              │              │              │
 │ [7A][7B][7C]   [1A Recepción]   [7D][7E][7F]      │
 │                     ▲                             │
 │               (Tú: Monitoreo)                     │
 └───────────────────────────────────────────────────┘
```

## Ideas de ambiente por cámara

- Imagen de cámara con grano suave y un leve tinte cálido, **no** estática de terror.
- Sello de hora en la esquina (`CAM 2 · 14:15`), como cámara real.
- La luz de las ventanas cambia con la hora del turno (mediodía → atardecer).
- Sonido ambiente distinto por zona: murmullo en espera, cubiertos en cafetería, brisa en el patio.

## Por decidir

- ¿Cuántas cámaras en el Día 1? Propuesta: **1A, 1B, 2, 3 y 3 habitaciones** (7 cámaras) para no abrumar.
- ✅ Los pacientes **se mueven** entre cámaras: "Los pacientes se van a mover por las cámaras, la idea es que no sean estáticos."
