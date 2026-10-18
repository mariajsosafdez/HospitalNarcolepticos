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
	unattendedPenalty = 10 // base, por cada dormido sin atender al final

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

// PenaltyPercent es el multiplicador del día en porcentaje (100 = x1,0). Un
// día fuera de rango usa el día 1. La vista lo muestra en el reporte.
func PenaltyPercent(day int) int {
	if day >= 1 && day <= len(penaltyPercent) {
		return penaltyPercent[day-1]
	}
	return 100
}

// ScalePenalty multiplica una penalización base por el multiplicador del día,
// redondeando al entero más cercano. Un día fuera de rango usa el día 1.
func ScalePenalty(base, day int) int {
	return (base*PenaltyPercent(day) + 50) / 100
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
	case percent < 80:
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

	// Cuántas veces se cobró cada penalización. Los puntos de arriba salen de
	// aquí: se multiplica una sola vez y se redondea una sola vez por tipo.
	BedCharges      int `json:"bedCharges"`      // cobros de "cama sin alta"
	OverworkEvents  int `json:"overworkEvents"`  // doctores que llegaron a 4/4
	UnattendedCount int `json:"unattendedCount"` // dormidos sin cama a las 18:00
}

// recomputePenalties calcula los puntos de cada penalización a partir de
// cuántas veces se cobró: redondeo(cobros x base x multiplicador del día),
// UNA sola vez por tipo (no por cobro), así ×1,25 y ×1,75 son exactos.
func (s *ScoreBreakdown) recomputePenalties(day int) {
	s.BedPenalty = -ScalePenalty(s.BedCharges*bedPenalty, day)
	s.OverworkLost = -ScalePenalty(s.OverworkEvents*overworkPenalty, day)
	s.Unattended = -ScalePenalty(s.UnattendedCount*unattendedPenalty, day)
}

// MainPenalty dice qué tipo de penalización pesó más (en valor absoluto):
// "bed", "overwork" o "unattended". Si no hubo ninguna, "". En un empate gana
// el primero de esa lista. Sirve para escoger la frase del Director.
func (s ScoreBreakdown) MainPenalty() string {
	best, kind := 0, ""
	for _, c := range []struct {
		kind  string
		value int
	}{{"bed", -s.BedPenalty}, {"overwork", -s.OverworkLost}, {"unattended", -s.Unattended}} {
		if c.value > best {
			best, kind = c.value, c.kind
		}
	}
	return kind
}

// Total suma todas las filas.
func (s ScoreBreakdown) Total() int {
	return s.WaitPoints + s.SevereBonus + s.BedPenalty + s.OverworkLost + s.Unattended
}

// Attended cuenta los pacientes que sí llegaron a una cama (todas las notas
// de espera menos las de los que quedaron sin atender).
func (s ScoreBreakdown) Attended() int { return s.Excellent + s.Good + s.Regular + s.Bad }

// PossiblePoints son los puntos máximos del turno: 120 por episodio.
func PossiblePoints(episodes int) int { return episodes * maxPointsPerEpisode }

// PointsPercent dice qué porcentaje (0 a 100, entero) de los puntos posibles
// se logró. Sin episodios no hay nada que medir y devuelve 0. Es pura para
// probarla con una tabla.
func PointsPercent(points, episodes int) int {
	possible := PossiblePoints(episodes)
	if possible <= 0 {
		return 0
	}
	percent := (points*100 + possible/2) / possible // redondeo al más cercano
	return min(max(percent, 0), 100)
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
