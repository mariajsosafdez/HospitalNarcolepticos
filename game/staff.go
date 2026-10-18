package game

import "costenos-narcolepsia/hospital"

// StaffState es el estado de un miembro del personal tal como lo ve el panel.
type StaffState int

const (
	StaffAvailable    StaffState = iota // puede recibir una asignación
	StaffTransferring                   // empujando la camilla
	StaffResting                        // descansando tras el traslado
	StaffNoQuota                        // doctor con 4/4 pacientes a cargo
)

func (s StaffState) String() string {
	switch s {
	case StaffAvailable:
		return "Available"
	case StaffTransferring:
		return "Transferring"
	case StaffResting:
		return "Resting"
	case StaffNoQuota:
		return "NoQuota"
	default:
		return "Unknown"
	}
}

// MarshalText hace que el estado salga como texto en el JSON.
func (s StaffState) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// Tiempos de traslado y descanso en ticks (aprobados por la autora):
// doctor 5 s + 15 s, camillero 3 s + 5 s de tiempo real (60 ticks por segundo).
const (
	doctorTransferTicks  = 5 * TicksPerSecond
	doctorRestTicks      = 15 * TicksPerSecond
	orderlyTransferTicks = 3 * TicksPerSecond
	orderlyRestTicks     = 5 * TicksPerSecond

	// DoctorQuota es el cupo del doctor. Es el mismo número que
	// maxPatientsPerDoctor del modelo, que no es público.
	DoctorQuota = 4
)

// StaffMember envuelve a un miembro del personal del hospital y le agrega lo
// que solo existe en el juego: el cooldown y la selección del jugador.
//
// Cumple la interfaz hospital.Attender, así que el hospital lo contrata con
// HireStaff y lo usa sin saber que existe (polimorfismo dentro de
// polimorfismo): por dentro delega en el *Doctor o *Orderly de verdad.
type StaffMember struct {
	inner           hospital.Attender // el *Doctor o *Orderly de verdad
	doctor          *hospital.Doctor  // nil si es camillero (cupo y reporte 5.2)
	transferUntil   int               // tick en que termina el traslado
	busyUntil       int               // tick en que termina traslado + descanso
	selected        bool              // true SOLO durante el despacho elegido
	overworkCharged bool              // ya se cobró el exceso de trabajo
	now             func() int        // reloj del turno
}

// NewDoctorMember envuelve a un doctor. now entrega el tick actual.
func NewDoctorMember(d *hospital.Doctor, now func() int) *StaffMember {
	return &StaffMember{inner: d, doctor: d, now: now}
}

// NewOrderlyMember envuelve a un camillero.
func NewOrderlyMember(o *hospital.Orderly, now func() int) *StaffMember {
	return &StaffMember{inner: o, now: now}
}

// Verificación en tiempo de compilación: StaffMember cumple Attender.
var _ hospital.Attender = (*StaffMember)(nil)

func (s *StaffMember) ID() string   { return s.inner.ID() }
func (s *StaffMember) Name() string { return s.inner.Name() }
func (s *StaffMember) Role() string { return s.inner.Role() }

// IsAvailable: el hospital solo "ve" disponible al que eligió el jugador,
// y solo si no está en cooldown y su tipo real también está disponible
// (por ejemplo, un doctor con cupo lleno no lo está).
func (s *StaffMember) IsAvailable() bool {
	return s.selected && s.Free()
}

// Attend delega en el Doctor u Orderly real.
func (s *StaffMember) Attend(p *hospital.Patient, location string) (hospital.EpisodeRecord, error) {
	return s.inner.Attend(p, location)
}

// Free dice si el jugador puede elegir a esta persona ahora mismo: no está en
// cooldown y (si es doctor) le queda cupo. A diferencia de IsAvailable, no
// depende de la selección.
func (s *StaffMember) Free() bool {
	return s.now() >= s.busyUntil && s.inner.IsAvailable()
}

// IsDoctor dice si el miembro envuelve a un doctor.
func (s *StaffMember) IsDoctor() bool { return s.doctor != nil }

// Doctor devuelve el *hospital.Doctor real (nil si es camillero).
func (s *StaffMember) Doctor() *hospital.Doctor { return s.doctor }

// State calcula el estado que muestra el panel.
func (s *StaffMember) State() StaffState {
	now := s.now()
	switch {
	case now < s.transferUntil:
		return StaffTransferring
	case now < s.busyUntil:
		return StaffResting
	case !s.inner.IsAvailable():
		return StaffNoQuota
	default:
		return StaffAvailable
	}
}

// PatientCount devuelve cuántos pacientes tiene a cargo (siempre 0 para un camillero).
func (s *StaffMember) PatientCount() int {
	if s.doctor == nil {
		return 0
	}
	return len(s.doctor.Patients())
}

// TransferUntil y BusyUntil exponen los cronómetros para la barra del panel.
func (s *StaffMember) TransferUntil() int { return s.transferUntil }
func (s *StaffMember) BusyUntil() int     { return s.busyUntil }

// cooldownTicks devuelve la duración del traslado y del descanso de esta persona.
func (s *StaffMember) cooldownTicks() (transfer, rest int) {
	if s.IsDoctor() {
		return doctorTransferTicks, doctorRestTicks
	}
	return orderlyTransferTicks, orderlyRestTicks
}

// startCooldown inicia traslado + descanso. Con shortRest el descanso se
// reduce a la mitad (lo usa MoveToBed, el "cooldown corto").
func (s *StaffMember) startCooldown(shortRest bool) {
	transfer, rest := s.cooldownTicks()
	if shortRest {
		rest /= 2
	}
	s.transferUntil = s.now() + transfer
	s.busyUntil = s.transferUntil + rest
}
