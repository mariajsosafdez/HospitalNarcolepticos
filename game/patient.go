package game

import (
	"math/rand"

	"costenos-narcolepsia/hospital"
)

// PatientState es el estado de JUEGO de un paciente (§3.1 de la especificación).
// No es lo mismo que hospital.PatientState: el modelo solo se entera de lo
// que el jugador reporta, así que a veces los dos estados no coinciden.
type PatientState int

const (
	Arriving      PatientState = iota // entra por Recepción
	Wandering                         // camina entre cámaras
	AsleepWaiting                     // dormido, nadie lo ha atendido (el modelo lo ve Awake)
	NoBedWaiting                      // atendido pero sin cama: sigue dormido donde se durmió
	Carried                           // en camilla rumbo a su habitación
	InBed                             // durmiendo en su habitación
	AwakeInBed                        // despierto, esperando el alta (el modelo lo ve AsleepInBed)
	Gone                              // ya salió del hospital
)

func (s PatientState) String() string {
	switch s {
	case Arriving:
		return "Arriving"
	case Wandering:
		return "Wandering"
	case AsleepWaiting:
		return "AsleepWaiting"
	case NoBedWaiting:
		return "NoBedWaiting"
	case Carried:
		return "Carried"
	case InBed:
		return "InBed"
	case AwakeInBed:
		return "AwakeInBed"
	case Gone:
		return "Gone"
	default:
		return "Unknown"
	}
}

// MarshalText hace que los estados salgan como texto en el JSON del Snapshot.
func (s PatientState) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// GamePatient envuelve al paciente del modelo y le agrega lo que solo existe
// en el juego: cámara, estado de juego y cronómetros (todos en ticks).
type GamePatient struct {
	patient *hospital.Patient
	state   PatientState
	camera  CameraID
	spot    int // punto dentro de la cámara (silla, mostrador...)

	arrivalTick    int          // tick en que entra al hospital
	nextMoveTick   int          // cuándo se mueve de nuevo
	nextSleepCheck int          // cuándo vuelve a tirar el dado del sueño
	noSleepUntil   int          // tras el alta no puede dormirse hasta este tick
	leaving        bool         // dado de alta y rumbo a Recepción para irse
	sleepLocation  string       // lugar que se le reporta al hospital
	sleepTick      int          // tick en que se durmió (empieza a correr su espera)
	arrivesAtBed   int          // tick en que la camilla llega a la habitación
	carrier        *StaffMember // quién empuja la camilla (solo en Carried)
	wakeTick       int          // tick en que despierta en su cama
	awakeSince     int          // tick en que quedó AwakeInBed
	bedPenalties   int          // cuántos cobros por cama ocupada ya se le hicieron
	waitMinutes    int          // espera final (cuando llegó a cama)
	sleeps         int          // cuántas veces se ha dormido en el turno
}

// Patient devuelve el paciente del modelo.
func (g *GamePatient) Patient() *hospital.Patient { return g.patient }

// State devuelve el estado de juego.
func (g *GamePatient) State() PatientState { return g.state }

// Camera devuelve la cámara donde está.
func (g *GamePatient) Camera() CameraID { return g.camera }

// Spot devuelve el punto de la cámara donde está parado.
func (g *GamePatient) Spot() int { return g.spot }

// SleepTick devuelve el tick en que se durmió por última vez.
func (g *GamePatient) SleepTick() int { return g.sleepTick }

// WaitMinutes devuelve cuánto lleva esperando (o esperó) una cama, en minutos
// de juego, mirando el tick actual si todavía espera.
func (g *GamePatient) WaitMinutes(now int) int {
	switch g.state {
	case AsleepWaiting, NoBedWaiting, Carried:
		return TicksToMinutes(now - g.sleepTick)
	case InBed, AwakeInBed:
		return g.waitMinutes
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// Funciones puras (se prueban con tablas)
// ---------------------------------------------------------------------------

// baseSleepChance es la probabilidad de dormirse en cada chequeo, según el nivel.
func baseSleepChance(level hospital.NarcolepsyLevel) float64 {
	switch level {
	case hospital.Mild:
		return 0.03
	case hospital.Moderate:
		return 0.06
	default:
		return 0.10
	}
}

// SleepChance combina el nivel del paciente con la zona donde está.
func SleepChance(level hospital.NarcolepsyLevel, cam CameraID) float64 {
	return baseSleepChance(level) * sleepZoneFactor(cam)
}

// sleepMinutes sortea cuánto duerme en cama: Leve 30-60 min, Moderado 1-2 h,
// Severo 2-3 h de juego.
func sleepMinutes(level hospital.NarcolepsyLevel, rng *rand.Rand) int {
	switch level {
	case hospital.Mild:
		return randomBetween(rng, 30, 60)
	case hospital.Moderate:
		return randomBetween(rng, 60, 120)
	default:
		return randomBetween(rng, 120, 180)
	}
}

// leaveChance es la probabilidad de irse del hospital tras el alta. Los
// severos se quedan más porque necesitan más cuidado.
func leaveChance(level hospital.NarcolepsyLevel) float64 {
	switch level {
	case hospital.Mild:
		return 0.60
	case hospital.Moderate:
		return 0.40
	default:
		return 0.20
	}
}

// Tiempos de movimiento y sueño, en minutos de juego.
const (
	minMoveMinutes     = 5
	maxMoveMinutes     = 25
	sleepCheckMinutes  = 10 // cada cuánto tira el dado del sueño
	arrivingMinutes    = 5  // cuánto tarda en pasar de Arriving a Wandering
	noSleepAfterMinute = 10 // tras el alta, mínimo 10 min sin dormirse

	// noSleepFinalMinutes: en los últimos 20 min de juego (10 s reales, de
	// 17:40 a 18:00) nadie se duerme de nuevo: un paciente que se durmiera
	// ahí ya no se alcanza a atender y sería una penalización garantizada.
	noSleepFinalMinutes = 20
)

// nextMoveDelay sortea cuántos ticks faltan para el próximo movimiento.
func nextMoveDelay(rng *rand.Rand) int {
	return MinutesToTicks(randomBetween(rng, minMoveMinutes, maxMoveMinutes))
}

// Carrier devuelve a quien empuja la camilla del paciente (nil si no va en
// camilla). La vista lo usa para dibujar a la persona correcta.
func (g *GamePatient) Carrier() *StaffMember {
	if g.state != Carried {
		return nil
	}
	return g.carrier
}

// sleepWindowOpen dice si en el tick now todavía pueden darse ataques de
// sueño nuevos: no en los últimos noSleepFinalMinutes del turno. Es pura para
// poder probarla con una tabla.
func sleepWindowOpen(now int) bool {
	return now < ShiftTicks-MinutesToTicks(noSleepFinalMinutes)
}
