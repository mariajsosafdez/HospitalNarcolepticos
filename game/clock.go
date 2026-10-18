// Package game contiene las REGLAS del juego de cámaras: reloj, pacientes,
// personal, puntos, días, reporte y modo Demo. Solo usa la librería estándar
// y el paquete hospital (que NO se modifica). No dibuja nada: la ventana vive
// en gameview y solo lee el estado de aquí y le manda acciones.
package game

import "fmt"

// Constantes de tiempo. Todo el juego se mide en TICKS (1 tick = 1/60 s real)
// y nunca se lee el reloj del computador, así que con la misma semilla la
// partida siempre sale igual.
const (
	TicksPerSecond = 60
	TicksPerMinute = 30                            // 1 minuto de juego = 0,5 s real = 30 ticks
	ShiftMinutes   = 6 * 60                        // turno de 12:00 a 18:00
	ShiftTicks     = ShiftMinutes * TicksPerMinute // 10 800 ticks = 180 s reales
	shiftStartHour = 12
)

// MinutesToTicks convierte minutos de juego a ticks.
func MinutesToTicks(minutes int) int { return minutes * TicksPerMinute }

// TicksToMinutes convierte ticks a minutos de juego (se descartan los
// ticks sobrantes).
func TicksToMinutes(ticks int) int { return ticks / TicksPerMinute }

// Clock es el reloj del turno: cuenta ticks y sabe si está en pausa.
type Clock struct {
	tick   int
	paused bool
}

// Advance suma un tick, salvo si el juego está en pausa o el turno terminó.
func (c *Clock) Advance() {
	if c.paused || c.Over() {
		return
	}
	c.tick++
}

// Pause y Resume congelan y reanudan el reloj.
func (c *Clock) Pause()  { c.paused = true }
func (c *Clock) Resume() { c.paused = false }

// Paused dice si el reloj está en pausa.
func (c *Clock) Paused() bool { return c.paused }

// Tick devuelve cuántos ticks han pasado desde las 12:00.
func (c *Clock) Tick() int { return c.tick }

// Over es true al llegar a las 18:00.
func (c *Clock) Over() bool { return c.tick >= ShiftTicks }

// Now devuelve la hora de juego en formato "14:15".
func (c *Clock) Now() string { return FormatGameTime(c.tick) }

// FormatGameTime convierte ticks desde las 12:00 en "HH:MM". Es una función
// pura para poder probarla con una tabla.
func FormatGameTime(tick int) string {
	minutes := TicksToMinutes(tick)
	return fmt.Sprintf("%02d:%02d", shiftStartHour+minutes/60, minutes%60)
}
