package game

import (
	"math/rand"
	"slices"
)

// CameraID identifica una cámara: "1A", "2", "7B"... En pantalla se muestra
// como "CAM 1A".
type CameraID string

// Reception es la cámara por donde entran y salen los pacientes.
const Reception CameraID = "1A"

// CameraKind distingue las zonas comunes (donde caminan y se duermen los
// pacientes) de las habitaciones (donde duermen ya atendidos).
type CameraKind int

const (
	Common CameraKind = iota
	RoomCam
)

// Camera describe una cámara del plano (ver "Mapa de Cámaras" en la bóveda).
type Camera struct {
	ID         CameraID
	Name       string
	Kind       CameraKind
	Neighbours []CameraID // cámaras vecinas del plano (solo zonas comunes)
	Spots      int        // puntos distintos donde puede estar un paciente (silla, mostrador...)
	RoomNumber int        // solo habitaciones: número de habitación del modelo
}

// cameras es el plano completo. Los vecinos salen del boceto del plano:
//
//	[6 Patio]   [8 Champeta]   [2 Cafetería]
//	    |             |              |
//	[4 Pasillo O]--[1B Espera]--[3 Pasillo E]--[5 Rx]
//	                  |
//	              [1A Recepción]
//
// La cafetería no tenía línea en el boceto; se une al Pasillo Este porque
// queda justo encima. Las habitaciones 7A-7C cuelgan del Pasillo Oeste y
// 7D-7F del Pasillo Este, pero los pacientes nunca caminan hasta ellas.
var cameras = map[CameraID]Camera{
	"1A": {ID: "1A", Name: "Recepción", Neighbours: []CameraID{"1B"}, Spots: 4},
	"1B": {ID: "1B", Name: "Sala de espera", Neighbours: []CameraID{"1A", "4", "3", "8"}, Spots: 6},
	"2":  {ID: "2", Name: "Cafetería", Neighbours: []CameraID{"3"}, Spots: 5},
	"3":  {ID: "3", Name: "Pasillo Este", Neighbours: []CameraID{"1B", "2", "5"}, Spots: 4},
	"4":  {ID: "4", Name: "Pasillo Oeste", Neighbours: []CameraID{"1B", "6"}, Spots: 4},
	"5":  {ID: "5", Name: "Radiología", Neighbours: []CameraID{"3"}, Spots: 4},
	"6":  {ID: "6", Name: "Patio de las palmeras", Neighbours: []CameraID{"4"}, Spots: 5},
	"8":  {ID: "8", Name: "Salón de champeta", Neighbours: []CameraID{"1B"}, Spots: 5},
	"7A": {ID: "7A", Name: "Habitación 101", Kind: RoomCam, RoomNumber: 101},
	"7B": {ID: "7B", Name: "Habitación 102", Kind: RoomCam, RoomNumber: 102},
	"7C": {ID: "7C", Name: "Habitación 103", Kind: RoomCam, RoomNumber: 103},
	"7D": {ID: "7D", Name: "Habitación 104", Kind: RoomCam, RoomNumber: 104},
	"7E": {ID: "7E", Name: "Habitación 105", Kind: RoomCam, RoomNumber: 105},
	"7F": {ID: "7F", Name: "Habitación 106", Kind: RoomCam, RoomNumber: 106},
}

// CameraByID busca una cámara del plano.
func CameraByID(id CameraID) (Camera, bool) {
	c, ok := cameras[id]
	return c, ok
}

// RoomCamera devuelve la cámara de una habitación del modelo (101 -> "7A").
func RoomCamera(roomNumber int) (CameraID, bool) {
	for id, c := range cameras {
		if c.Kind == RoomCam && c.RoomNumber == roomNumber {
			return id, true
		}
	}
	return "", false
}

// sleepZoneFactor es el modificador de probabilidad de dormirse según la
// zona, en porcentaje: la cafetería x2 (el sancocho) y el patio x1,5.
func sleepZoneFactor(id CameraID) float64 {
	switch id {
	case "2":
		return 2
	case "6":
		return 1.5
	default:
		return 1
	}
}

// World es el conjunto de cámaras DESBLOQUEADAS en el día actual.
type World struct {
	common []CameraID // zonas comunes, en orden del plano
	rooms  []CameraID
}

// NewWorld arma el mundo con las zonas comunes y las habitaciones del día.
func NewWorld(common []CameraID, roomNumbers []int) *World {
	w := &World{common: slices.Clone(common)}
	for _, n := range roomNumbers {
		if id, ok := RoomCamera(n); ok {
			w.rooms = append(w.rooms, id)
		}
	}
	return w
}

// Has dice si la cámara existe en el mundo del día.
func (w *World) Has(id CameraID) bool {
	return slices.Contains(w.common, id) || slices.Contains(w.rooms, id)
}

// Unlocked devuelve todas las cámaras disponibles (comunes y habitaciones).
func (w *World) Unlocked() []CameraID {
	return append(slices.Clone(w.common), w.rooms...)
}

// Neighbours devuelve las vecinas de una cámara que están desbloqueadas.
func (w *World) Neighbours(id CameraID) []CameraID {
	var result []CameraID
	for _, n := range cameras[id].Neighbours {
		if slices.Contains(w.common, n) {
			result = append(result, n)
		}
	}
	return result
}

// RandomMove sortea a dónde va un paciente cuando le toca moverse: a una
// cámara vecina o se queda donde está (todas con la misma probabilidad).
func (w *World) RandomMove(from CameraID, rng *rand.Rand) CameraID {
	options := append(w.Neighbours(from), from)
	return options[rng.Intn(len(options))]
}

// StepToward da UN paso desde from hacia la cámara to por el camino más
// corto (búsqueda en anchura). Si ya está en to, o no hay camino, devuelve
// from. Se usa para que los pacientes que se van caminen hasta Recepción.
func (w *World) StepToward(from, to CameraID) CameraID {
	if from == to {
		return from
	}
	// previous guarda "por qué cámara llegué a esta", para reconstruir el camino.
	previous := map[CameraID]CameraID{from: from}
	queue := []CameraID{from}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, n := range w.Neighbours(current) {
			if _, seen := previous[n]; seen {
				continue
			}
			previous[n] = current
			if n == to {
				// Retrocedemos hasta el primer paso después de from.
				step := n
				for previous[step] != from {
					step = previous[step]
				}
				return step
			}
			queue = append(queue, n)
		}
	}
	return from
}

// randomBetween devuelve un entero al azar entre min y max (ambos incluidos).
func randomBetween(rng *rand.Rand, min, max int) int {
	return min + rng.Intn(max-min+1)
}
