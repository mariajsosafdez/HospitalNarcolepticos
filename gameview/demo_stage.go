package gameview

import (
	"math"

	"costenos-narcolepsia/game"
)

// Puesta en escena de la Demo. game.Demo dice QUÉ pasa y cuánto va (beat y
// progreso de 0 a 1); aquí se decide DÓNDE se ve: por qué lado de la cámara
// entra el personal, dónde se para junto al paciente y por dónde se va con la
// camilla. Todo son funciones puras para poder probarlas con tablas.

// entersFromLeft dice por qué lado de la cámara entra y sale el personal: por
// el lado más cercano al paciente (x es el centro del paciente, en el fondo).
func entersFromLeft(x float64) bool { return x < feedW/2 }

// lerp mezcla dos valores: t = 0 da a, t = 1 da b.
func lerp(a, b, t float64) float64 { return a + (b-a)*t }

// offscreenX es el centro de un sprite de ancho w justo FUERA de la cámara,
// por el lado izquierdo o el derecho.
func offscreenX(left bool, w int) float64 {
	if left {
		return -float64(w) / 2
	}
	return feedW + float64(w)/2
}

// standX es el centro del personal cuando está junto al paciente: pegado al
// lado por donde entró, sin salirse de la cámara.
func standX(patientLeft, patientWidth int, left bool, staffW int) float64 {
	x := float64(patientLeft - 4)
	if !left {
		x = float64(patientLeft + patientWidth + 4)
	}
	half := float64(staffW) / 2
	return math.Min(math.Max(x, half), feedW-half)
}

// stagedX da el centro del personal en un beat: entra caminando desde el
// borde, espera de pie junto al paciente o se va hacia el borde. progress va
// de 0 a 1 dentro del beat. walking dice si debe usarse el sprite de caminar.
func stagedX(kind game.BeatKind, progress, stand, edge float64) (x float64, walking bool) {
	switch kind {
	case game.BeatArrive:
		return lerp(edge, stand, progress), true
	case game.BeatLeave:
		return lerp(stand, edge, progress), true
	}
	return stand, false // BeatNoBed: de pie junto al paciente
}

// stretcherX da el centro de la camilla mientras se lleva al paciente: sale
// de donde estaba hasta quedar fuera de la cámara. progress es la fracción del
// traslado (StaffMember.CooldownFraction).
func stretcherX(from float64, left bool, w int, progress float64) float64 {
	return lerp(from, offscreenX(left, w), progress)
}
