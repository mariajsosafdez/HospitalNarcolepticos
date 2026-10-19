package game

import (
	"math/rand"
	"slices"

	"costenos-narcolepsia/hospital"
)

// Easter eggs: cinco pacientes extra que casi nunca aparecen (decisión de la
// autora, 2026-10-19). Ver "Easter eggs" en Elenco.md.

// IDs de los easter eggs. Empiezan por E- para no chocar con los P-xxx.
const (
	CatID  = "E-001" // Misifu, el gato: el Director comenta su llegada
	CoteID = "E-004" // Cote
)

// eggChance es la probabilidad de que un turno normal traiga un easter egg.
const eggChance = 0.15

// frequentSleepFactor multiplica la probabilidad de dormirse de los dos
// easter eggs severos (Misifu y Cote): "se duermen mucho más seguido".
const frequentSleepFactor = 3

// eggSeedSalt separa el azar del easter egg del azar del turno: así sacar (o
// no) un easter egg no cambia cuándo llegan ni cómo se mueven los demás.
const eggSeedSalt = 0x0e66

// castEggs son los cinco easter eggs, en orden: el número N de la bandera
// -egg N y de los sprites pac_egg<N>_* es su posición + 1. La edad del gato
// (6) la acepta hospital.NewPatient porque solo exige edad >= 0.
var castEggs = []PatientData{
	{CatID, "Misifu", 6, hospital.Severe},
	{"E-002", "Alejo Urrego", 22, hospital.Mild},
	{"E-003", "Samu Giraldo", 22, hospital.Mild},
	{CoteID, "Cote", 22, hospital.Severe},
	{"E-005", "Sebs", 22, hospital.Moderate},
}

// EggByNumber devuelve el easter egg número n (1 a 5).
func EggByNumber(n int) (PatientData, bool) {
	if n < 1 || n > len(castEggs) {
		return PatientData{}, false
	}
	return castEggs[n-1], true
}

// IsEgg dice si un ID es de un easter egg.
func IsEgg(id string) bool {
	return slices.ContainsFunc(castEggs, func(e PatientData) bool { return e.ID == id })
}

// RollEgg decide si el turno de esta semilla trae un easter egg y cuál. Es
// pura: misma semilla, mismo resultado. Nunca devuelve más de uno.
func RollEgg(seed int64) (PatientData, bool) {
	rng := rand.New(rand.NewSource(seed ^ eggSeedSalt))
	if rng.Float64() >= eggChance {
		return PatientData{}, false
	}
	return castEggs[rng.Intn(len(castEggs))], true
}

// WithEgg devuelve una copia de la configuración con el easter egg agregado
// como un paciente más (llega a una hora al azar como los otros).
func (c DayConfig) WithEgg(egg PatientData) DayConfig {
	c.Patients = append(slices.Clone(c.Patients), egg)
	return c
}

// hasFrequentSleep dice si el paciente se duerme mucho más seguido y sin
// la espera tras el alta: solo Misifu y Cote.
func hasFrequentSleep(id string) bool { return id == CatID || id == CoteID }

// sleepChanceFor es la probabilidad de dormirse de ESTE paciente: la normal
// (nivel y zona) y, para Misifu y Cote, multiplicada por frequentSleepFactor.
func sleepChanceFor(id string, level hospital.NarcolepsyLevel, cam CameraID) float64 {
	chance := SleepChance(level, cam)
	if hasFrequentSleep(id) {
		chance = min(chance*frequentSleepFactor, 1)
	}
	return chance
}

// noSleepMinutesAfterDischarge son los minutos sin dormirse tras el alta:
// 10 para todos, 0 para Misifu y Cote.
func noSleepMinutesAfterDischarge(id string) int {
	if hasFrequentSleep(id) {
		return 0
	}
	return noSleepAfterMinute
}
