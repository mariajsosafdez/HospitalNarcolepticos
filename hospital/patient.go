package hospital

// NarcolepsyLevel es el nivel de narcolepsia de un paciente.
//
// En vez de usar textos sueltos como "Severe" (donde un error de tipeo como
// "Sevr" compilaría sin problema), declaramos un TIPO NUEVO basado en int
// y unas constantes de ese tipo. Si alguien escribe Sevr, el compilador
// avisa de inmediato porque esa constante no existe.
type NarcolepsyLevel int

// iota vale 0 en la primera constante y aumenta de 1 en 1:
// Mild = 0, Moderate = 1, Severe = 2.
const (
	Mild NarcolepsyLevel = iota
	Moderate
	Severe
)

// String convierte el nivel en texto legible. Al tener este método,
// NarcolepsyLevel cumple la interfaz fmt.Stringer, así que fmt.Println(nivel)
// imprime "Severe" en lugar de "2".
func (l NarcolepsyLevel) String() string {
	switch l {
	case Mild:
		return "Mild"
	case Moderate:
		return "Moderate"
	case Severe:
		return "Severe"
	default:
		return "Unknown"
	}
}

// PatientState indica dónde y cómo está el paciente en este momento.
type PatientState int

// Awake = 0 (despierto), AsleepInHallway = 1 (dormido en un pasillo, bloqueando
// el paso), AsleepInBed = 2 (dormido en una cama de un cuarto).
const (
	Awake PatientState = iota
	AsleepInHallway
	AsleepInBed
)

// String devuelve el estado en texto legible para imprimirlo.
func (s PatientState) String() string {
	switch s {
	case Awake:
		return "Awake"
	case AsleepInHallway:
		return "Asleep in hallway"
	case AsleepInBed:
		return "Asleep in bed"
	default:
		return "Unknown"
	}
}
