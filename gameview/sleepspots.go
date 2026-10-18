package gameview

import (
	"math/rand"

	"costenos-narcolepsia/game"
)

// Dónde se tumba un paciente cuando se duerme. Mientras camina puede estar en
// cualquier punto del piso, pero al dormirse su sprite (hasta 47 px de ancho)
// no puede tapar a otro dormido, salirse de la cámara ni caer sobre el tronco
// de una palmera. Todo esto es cosa de la vista: game no sabe de rincones.

const (
	sleepMargin   = 8  // distancia mínima del sprite dormido a los bordes de la cámara
	sleepGap      = 2  // aire entre dos dormidos
	sleepTries    = 40 // puntos al azar que se prueban antes de rendirse
	lyingMinRatio = 2  // un sprite acostado es al menos 2 veces más ancho que alto
)

// sleepObstacles son los muebles contra los que no se tumba nadie (coordenadas
// del fondo de 168x123, medidas sobre el primer cuadro de cada PNG de
// art/png). Cada rectángulo cubre un mueble que se para en el piso: mesas,
// sillas, bancas, mostradores, matas, el dispensador, la máquina de rayos X.
// Un paciente que camina puede pasar junto a ellos, pero al dormirse su caja
// no los toca (ver sleepScore). Las claves son todas las cámaras comunes. En
// el Patio: los troncos de las palmeras (base cerca de (44,104) y (126,92)),
// la jardinera y la banca. Dormir "contra una palmera"
// es al lado del tronco, no encima.
var sleepObstacles = map[game.CameraID][]rect{
	"1A": {
		{2, 86, 36, 18},   // banca de la entrada
		{80, 46, 62, 17},  // mostrador de recepción
		{148, 56, 12, 16}, // mata de la derecha
		{61, 46, 12, 16},  // mata junto a la puerta
	},
	"1B": {
		{22, 44, 68, 18},  // fila de sillas del fondo
		{21, 50, 10, 16},  // mata de la izquierda
		{93, 75, 55, 21},  // fila de sillas del frente
		{132, 24, 16, 38}, // dispensador de agua
	},
	"2": {
		{8, 70, 52, 28},   // mesa izquierda con sus sillas
		{102, 82, 52, 27}, // mesa derecha con sus sillas
		{106, 28, 40, 28}, // mostrador de la olla
		{150, 34, 11, 25}, // mata de la derecha
	},
	"3": {
		{2, 96, 16, 22},   // mata grande de abajo
		{107, 50, 29, 10}, // banca
	},
	"4": {
		{35, 49, 29, 11},  // banca
		{75, 45, 10, 17},  // dispensador
		{149, 95, 17, 23}, // mata grande de abajo
	},
	"5": {
		{28, 20, 10, 59}, // columna del aparato de rayos X
		{40, 56, 60, 30}, // camilla de la máquina
	},
	"6": {
		{40, 60, 10, 46},  // palmera izquierda
		{124, 60, 8, 34},  // palmera derecha
		{144, 86, 18, 20}, // jardinera
		{66, 54, 38, 16},  // banca
	},
	"8": {
		{2, 64, 10, 13},   // silla de arriba
		{6, 82, 10, 13},   // silla de abajo
		{146, 46, 12, 32}, // ventilador de pie
	},
}

// footBox es el rectángulo de un sprite de w x h cuyos pies (centro abajo)
// están en (x, y).
func footBox(x, y float64, w, h int) rect {
	return rect{int(x) - w/2, int(y) - h, w, h}
}

// overlapArea es el área en que se cruzan dos rectángulos (0 si no se tocan).
func overlapArea(a, b rect) int {
	w := min(a.x+a.w, b.x+b.w) - max(a.x, b.x)
	h := min(a.y+a.h, b.y+b.h) - max(a.y, b.y)
	if w <= 0 || h <= 0 {
		return 0
	}
	return w * h
}

// grow agranda un rectángulo n píxeles por cada lado.
func (r rect) grow(n int) rect { return rect{r.x - n, r.y - n, r.w + 2*n, r.h + 2*n} }

// sleepScore mide qué tan mal cae un sprite de w x h con los pies en (x, y):
// área que tapa de otros dormidos u obstáculos, más un castigo enorme si se
// sale de la zona permitida. 0 es un lugar libre.
func sleepScore(x, y float64, w, h int, floor rect, taken, obstacles []rect) int {
	box := footBox(x, y, w, h)
	score := 0
	if box.x < sleepMargin || box.x+box.w > feedW-sleepMargin {
		score += 100000
	}
	if int(y) < floor.y || int(y) > floor.y+floor.h {
		score += 100000
	}
	for _, t := range taken {
		score += overlapArea(box, t.grow(sleepGap))
	}
	for _, o := range obstacles {
		score += overlapArea(box, o)
	}
	return score
}

// pickSleepSpot elige los pies del paciente dormido. Prueba primero donde se
// quedó dormido (corrido hacia adentro si el sprite se saldría), y después
// puntos al azar del piso; se queda con el primero libre. Si no hay ninguno
// libre devuelve el que menos estorba. Con la misma semilla da lo mismo.
func pickSleepSpot(x, y float64, w, h int, floor rect, taken, obstacles []rect, rng *rand.Rand) (float64, float64) {
	minX, maxX := float64(sleepMargin+w/2), float64(feedW-sleepMargin-w/2)
	cx := min(max(x, minX), maxX)
	cy := min(max(y, float64(floor.y)), float64(floor.y+floor.h))

	bestX, bestY := cx, cy
	best := sleepScore(cx, cy, w, h, floor, taken, obstacles)
	for i := 0; i < sleepTries && best > 0; i++ {
		px, py := randomFloorPoint(floor, rng)
		s := sleepScore(px, py, w, h, floor, taken, obstacles)
		if s < best {
			bestX, bestY, best = px, py, s
		}
	}
	return bestX, bestY
}

// isLying dice si el sprite (w x h) es de alguien acostado en el piso (los
// de Yesid, Nayeli, Kelly, Eder, Yuranis y Wilfrido son 2 o más veces más
// anchos que altos; Rafa, Mañe, Marelvis y Yeimy no).
func isLying(w, h int) bool { return w >= lyingMinRatio*h }

// zzzSpot dice dónde va la esquina del efecto zzz de un dormido en el piso.
// Acostado: encima de la cabeza, que en los sprites queda a la IZQUIERDA.
// De pie o en silla: arriba del hombro derecho, como siempre. fxW es el
// ancho del efecto.
func zzzSpot(x, y, w, h, fxW int) (int, int) {
	if isLying(w, h) {
		return min(max(x+1, 0), feedW-fxW), y - 16
	}
	return min(x+w-5, feedW-fxW), y - 16
}
