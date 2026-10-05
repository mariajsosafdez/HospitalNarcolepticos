package hospital

import "fmt"

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

// Patient es un paciente costeño con narcolepsia.
//
// Person va EMBEBIDO (sin nombre de campo): Patient "tiene" un id, un nombre
// y una edad sin volver a declararlos, y hereda (por promoción) los métodos
// ID(), Name() y Age(). Se embebe por VALOR: cada paciente tiene su propia
// copia de Person dentro de su struct.
//
// Todos los demás campos son privados; solo se cambian mediante métodos,
// así nadie puede, por ejemplo, poner state = AsleepInBed sin darle un cuarto.
type Patient struct {
	Person
	level           NarcolepsyLevel
	state           PatientState
	currentLocation string
	assignedDoctor  *Doctor // doctor que lo diagnosticó; nil si aún no tiene
	room            *Room   // cuarto donde duerme; nil si no tiene cuarto
}

// NewPatient crea un paciente despierto en la recepción.
// Devuelve un error si los datos no son válidos (constructor con validación).
func NewPatient(id, name string, age int, level NarcolepsyLevel) (*Patient, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		// %w "envuelve" el error original: el que llama puede seguir
		// preguntando errors.Is(err, ErrInvalidData).
		return nil, fmt.Errorf("new patient %q: %w", id, err)
	}
	if level < Mild || level > Severe {
		return nil, fmt.Errorf("new patient %q: unknown narcolepsy level: %w", id, ErrInvalidData)
	}
	return &Patient{
		Person:          person,
		level:           level,
		state:           Awake,
		currentLocation: "reception",
	}, nil
}

// Los métodos que MODIFICAN al paciente usan receptor por PUNTERO (p *Patient):
// así trabajan sobre el paciente original y no sobre una copia. Los getters
// también usan puntero por consistencia (regla de Go: si un tipo tiene algún
// método con receptor puntero, todos deberían tenerlo).

// Level devuelve el nivel de narcolepsia del paciente.
func (p *Patient) Level() NarcolepsyLevel { return p.level }

// State devuelve el estado actual del paciente.
func (p *Patient) State() PatientState { return p.state }

// Location devuelve el último lugar conocido del paciente.
func (p *Patient) Location() string { return p.currentLocation }

// AssignedDoctor devuelve el doctor a cargo del paciente, o nil si no tiene.
func (p *Patient) AssignedDoctor() *Doctor { return p.assignedDoctor }

// Room devuelve el cuarto donde duerme el paciente, o nil si no tiene.
func (p *Patient) Room() *Room { return p.room }

// SufferSleepAttack registra que el paciente se quedó dormido de repente en
// un lugar del hospital. Queda dormido en el pasillo hasta que se le asigne
// una cama. Si el paciente ya estaba dormido no cambia nada (no se puede
// dormir dos veces); el hospital valida ese caso antes y devuelve un error.
func (p *Patient) SufferSleepAttack(location string) {
	if p.state != Awake {
		return
	}
	p.state = AsleepInHallway
	p.currentLocation = location
}

// WakeUp despierta al paciente. Si estaba en una cama, el paciente le pide
// a su cuarto que lo libere (los objetos se comunican entre sí).
func (p *Patient) WakeUp() error {
	if p.state == Awake {
		return fmt.Errorf("wake up %s: %w", p.id, ErrAlreadyAwake)
	}
	if p.room != nil {
		// Release también pone p.room en nil (ver room.go).
		if err := p.room.Release(p); err != nil {
			return fmt.Errorf("wake up %s: %w", p.id, err)
		}
	}
	p.state = Awake
	return nil
}

// moveToBed es privado: solo Room.Occupy lo llama, para que el cuarto y el
// paciente siempre queden de acuerdo (el cuarto tiene al paciente y el
// paciente sabe en qué cuarto está).
func (p *Patient) moveToBed(r *Room) {
	p.room = r
	p.state = AsleepInBed
	p.currentLocation = fmt.Sprintf("room %d", r.number)
}

// leaveRoom es privado: solo Room.Release lo llama.
func (p *Patient) leaveRoom() {
	p.room = nil
}
