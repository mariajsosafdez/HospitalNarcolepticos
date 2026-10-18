package gameview

import (
	"math"
	"math/rand"

	"costenos-narcolepsia/game"
)

// Los pacientes de game solo saben en qué CÁMARA están, no en qué rincón de
// la sala. Dónde están parados y cómo caminan es cosa de la vista: cada
// paciente tiene un walker que pasea hacia puntos al azar del piso. El azar
// sale de un rand con semilla (la del turno), así que con -seed la vista
// también es igual cada vez.

// floorRect es la zona del piso (coordenadas del fondo de 168x123) por
// donde se puede caminar: los PIES del paciente se quedan dentro.
var floorRect = rect{16, 68, 136, 50}

const (
	walkSpeed    = 0.4 // píxeles del fondo por tick (24 por segundo)
	minPauseTick = 60
	maxPauseTick = 240
)

// walker es la posición de un paciente dentro de su cámara.
type walker struct {
	cam     game.CameraID
	x, y    float64 // pies del paciente (centro abajo), en coordenadas del fondo
	tx, ty  float64 // hacia dónde camina
	pause   int     // ticks que le faltan quieto
	left    bool    // mira a la izquierda
	walking bool
	settled bool // ya eligió dónde tumbarse al dormirse (ver pickSleepSpot)
}

// randomFloorPoint sortea un punto del piso.
func randomFloorPoint(floor rect, rng *rand.Rand) (float64, float64) {
	return float64(floor.x + rng.Intn(floor.w)), float64(floor.y + rng.Intn(floor.h))
}

// newWalker pone un paciente en un punto al azar de la cámara.
func newWalker(cam game.CameraID, floor rect, rng *rand.Rand) *walker {
	w := &walker{cam: cam}
	w.x, w.y = randomFloorPoint(floor, rng)
	w.tx, w.ty = w.x, w.y
	w.pause = minPauseTick + rng.Intn(maxPauseTick-minPauseTick)
	return w
}

// step avanza el walker un tick. Solo camina si el paciente está en pie
// (moving); si se durmió o está en cama se queda donde está.
func (w *walker) step(moving bool, floor rect, rng *rand.Rand) {
	if !moving {
		w.walking = false
		return
	}
	if w.pause > 0 {
		w.pause--
		w.walking = false
		if w.pause == 0 {
			w.tx, w.ty = randomFloorPoint(floor, rng)
		}
		return
	}
	dx, dy := w.tx-w.x, w.ty-w.y
	dist := math.Hypot(dx, dy)
	if dist <= walkSpeed {
		w.x, w.y = w.tx, w.ty
		w.walking = false
		w.pause = minPauseTick + rng.Intn(maxPauseTick-minPauseTick)
		return
	}
	w.x += dx / dist * walkSpeed
	w.y += dy / dist * walkSpeed
	w.walking = true
	if math.Abs(dx) > 0.5 {
		w.left = dx < 0
	}
}

// updateWalkers avanza a todos los pacientes un tick de la vista.
func (c *camsScene) updateWalkers() {
	for _, g := range c.shift.Patients() {
		id := g.Patient().ID()
		w, ok := c.walkers[id]
		if !ok || w.cam != g.Camera() {
			w = newWalker(g.Camera(), c.floor(), c.rng)
			c.walkers[id] = w
		}
		state := g.State()
		w.step(state == game.Wandering || state == game.Arriving, c.floor(), c.rng)

		// Al dormirse en el piso se busca un lugar libre; al despertar, se
		// olvida para la próxima vez.
		if state == game.AsleepWaiting || state == game.NoBedWaiting {
			if !w.settled {
				c.settleSleeper(g, w)
			}
		} else {
			w.settled = false
		}
	}
}

// sleepSize da el tamaño del sprite dormido de un paciente (el del
// placeholder, 16x24, si todavía no tiene arte).
func (c *camsScene) sleepSize(g *game.GamePatient) (int, int) {
	name := "pac_" + spriteKey(g.Patient().ID(), g.Patient().Name()) + "_dormido"
	if c.env.Assets.Has(name) {
		return c.env.Assets.FrameSize(name, 2)
	}
	return 16, 24
}

// settleSleeper mueve a un paciente recién dormido a un lugar libre del piso:
// que su sprite no tape a otro dormido de la cámara ni a un obstáculo.
func (c *camsScene) settleSleeper(g *game.GamePatient, w *walker) {
	w.settled = true
	sw, sh := c.sleepSize(g)
	var taken []rect
	for _, o := range c.shift.Patients() {
		ow := c.walkers[o.Patient().ID()]
		if o == g || ow == nil || !ow.settled || o.Camera() != g.Camera() {
			continue
		}
		ow2, oh := c.sleepSize(o)
		taken = append(taken, footBox(ow.x, ow.y, ow2, oh))
	}
	w.x, w.y = pickSleepSpot(w.x, w.y, sw, sh, c.floor(), taken, sleepObstacles[g.Camera()], c.rng)
	w.tx, w.ty = w.x, w.y
}

// floor devuelve la zona por donde caminan los pacientes. En la Demo la caja
// de controles tapa la parte de abajo de la cámara, así que se camina solo
// por arriba de ella.
func (c *camsScene) floor() rect {
	if c.demo != nil {
		return rect{floorRect.x, floorRect.y, floorRect.w, 22}
	}
	return floorRect
}
