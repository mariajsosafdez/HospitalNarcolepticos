package game

// WaitRating es la nota que recibe una espera.
type WaitRating int

const (
	Excellent WaitRating = iota
	Good
	Regular
	Bad
)

func (r WaitRating) String() string {
	switch r {
	case Excellent:
		return "Excellent"
	case Good:
		return "Good"
	case Regular:
		return "Regular"
	default:
		return "Bad"
	}
}

// Puntos de la tabla aprobada (sección 7 de la especificación).
const (
	pointsExcellent   = 100
	pointsGood        = 70
	pointsRegular     = 40
	pointsBad         = 10
	severeDoctorBonus = 20
	overworkPenalty   = 50 // base, antes del multiplicador del día
	bedPenalty        = 2  // base, por cada 10 min tras la gracia
	bedGraceMinutes   = 15
	bedPenaltyEvery   = 10

	maxPointsPerEpisode = 120 // para calcular el porcentaje de estrellas
)

// RateWait traduce los minutos de espera a nota y puntos.
func RateWait(minutes int) (WaitRating, int) {
	switch {
	case minutes <= 20:
		return Excellent, pointsExcellent
	case minutes <= 45:
		return Good, pointsGood
	case minutes <= 90:
		return Regular, pointsRegular
	default:
		return Bad, pointsBad
	}
}

// penaltyPercent es el multiplicador de las penalizaciones por día, en
// porcentaje (x1,0 a x2,0). Se usan enteros para no redondear mal.
var penaltyPercent = [...]int{100, 125, 150, 175, 200}

// ScalePenalty multiplica una penalización base por el multiplicador del día,
// redondeando al entero más cercano. Un día fuera de rango usa el día 1.
func ScalePenalty(base, day int) int {
	percent := 100
	if day >= 1 && day <= len(penaltyPercent) {
		percent = penaltyPercent[day-1]
	}
	return (base*percent + 50) / 100
}

// Stars convierte puntos y episodios en estrellas: porcentaje =
// puntos / (episodios x 120). Nunca baja de 1 estrella. Sin episodios no hubo
// nada que hacer mal, así que son 5.
func Stars(points, episodes int) int {
	if episodes <= 0 {
		return 5
	}
	// Se compara en enteros: points*100 / (episodes*120) < umbral.
	percent := float64(points) * 100 / float64(episodes*maxPointsPerEpisode)
	switch {
	case percent < 30:
		return 1
	case percent < 50:
		return 2
	case percent < 70:
		return 3
	case percent < 85:
		return 4
	default:
		return 5
	}
}

// ScoreBreakdown es el desglose de puntos del turno. Las penalizaciones se
// guardan con signo negativo.
type ScoreBreakdown struct {
	Excellent    int `json:"excellent"` // cuántas esperas de cada nota
	Good         int `json:"good"`
	Regular      int `json:"regular"`
	Bad          int `json:"bad"`
	WaitPoints   int `json:"waitPoints"`
	SevereBonus  int `json:"severeBonus"`
	BedPenalty   int `json:"bedPenalty"`   // despierto en cama sin alta
	OverworkLost int `json:"overworkLost"` // exceso de trabajo de los doctores
	Unattended   int `json:"unattended"`   // dormidos sin cama al terminar el turno
	Episodes     int `json:"episodes"`     // episodios que cuentan para las estrellas
}

// Total suma todas las filas.
func (s ScoreBreakdown) Total() int {
	return s.WaitPoints + s.SevereBonus + s.BedPenalty + s.OverworkLost + s.Unattended
}

// addWait registra la espera de un paciente que llegó a una cama.
func (s *ScoreBreakdown) addWait(minutes int) {
	rating, points := RateWait(minutes)
	switch rating {
	case Excellent:
		s.Excellent++
	case Good:
		s.Good++
	case Regular:
		s.Regular++
	default:
		s.Bad++
	}
	s.WaitPoints += points
	s.Episodes++
}
