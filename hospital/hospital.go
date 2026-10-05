package hospital

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// Hospital es el objeto central: conoce a todos los pacientes, al personal,
// los cuartos y el historial completo de episodios.
//
// Guarda punteros (*Patient, *Room, *Doctor) para trabajar siempre con los
// objetos originales: si un cuarto cambia de estado, todos los que lo
// referencian ven el cambio.
//
// Usamos slices (y no maps) porque necesitamos conservar el ORDEN de llegada:
// el primer cuarto disponible, los pacientes en el orden en que fueron
// admitidos, y el personal en orden para despacharlo por turnos. Con pocos
// elementos, buscar recorriendo el slice es simple y suficientemente rápido.
type Hospital struct {
	name      string
	doctors   []*Doctor       // solo los doctores (para la consulta por doctor)
	staff     []Attender      // TODO el personal: doctores y camilleros juntos (polimorfismo)
	patients  []*Patient      // pacientes admitidos, en orden de llegada
	rooms     []*Room         // cuartos, en orden de registro
	history   []EpisodeRecord // todos los episodios del hospital
	nextStaff int             // posición del siguiente miembro del personal en el turno
}

// NewHospital crea un hospital vacío.
func NewHospital(name string) *Hospital {
	return &Hospital{name: name}
}

// Name devuelve el nombre del hospital.
func (h *Hospital) Name() string { return h.name }

// ---------------------------------------------------------------------------
// Registro de pacientes, personal y cuartos
// ---------------------------------------------------------------------------

// AdmitPatient admite a un paciente. Falla si es nil o si su ID ya existe.
func (h *Hospital) AdmitPatient(p *Patient) error {
	if p == nil {
		return fmt.Errorf("admit patient: %w", ErrInvalidData)
	}
	if h.findPatient(p.ID()) != nil {
		return fmt.Errorf("admit patient %s: %w", p.ID(), ErrDuplicateID)
	}
	h.patients = append(h.patients, p)
	return nil
}

// HireDoctor contrata a un doctor. Internamente usa HireStaff, porque un
// doctor también es un Attender.
func (h *Hospital) HireDoctor(d *Doctor) error {
	if d == nil {
		return fmt.Errorf("hire doctor: %w", ErrInvalidData)
	}
	return h.HireStaff(d)
}

// HireStaff contrata a CUALQUIER tipo de personal que cumpla Attender
// (Doctor, Orderly o un tipo nuevo como Nurse) y lo agrega al turno.
func (h *Hospital) HireStaff(a Attender) error {
	if a == nil {
		return fmt.Errorf("hire staff: %w", ErrInvalidData)
	}
	if h.findStaff(a.ID()) != nil {
		return fmt.Errorf("hire staff %s: %w", a.ID(), ErrDuplicateID)
	}
	h.staff = append(h.staff, a)

	// "Type assertion": preguntamos si el valor guardado en la interfaz es,
	// en concreto, un *Doctor. Si lo es (ok == true), también lo guardamos en
	// la lista de doctores para poder consultar sus episodios. El despacho
	// de emergencias NO usa esto: trabaja solo con la interfaz.
	if d, ok := a.(*Doctor); ok {
		h.doctors = append(h.doctors, d)
	}
	return nil
}

// AddRoom registra un cuarto nuevo. Falla si es nil o si el número ya existe.
func (h *Hospital) AddRoom(r *Room) error {
	if r == nil {
		return fmt.Errorf("add room: %w", ErrInvalidData)
	}
	for _, existing := range h.rooms {
		if existing.number == r.number {
			return fmt.Errorf("add room %d: %w", r.number, ErrDuplicateID)
		}
	}
	h.rooms = append(h.rooms, r)
	return nil
}

// ---------------------------------------------------------------------------
// Getters: devuelven COPIAS de los slices para que nadie de afuera pueda
// agregar o quitar elementos de las listas internas del hospital.
// ---------------------------------------------------------------------------

// Patients devuelve los pacientes admitidos.
func (h *Hospital) Patients() []*Patient { return slices.Clone(h.patients) }

// Doctors devuelve los doctores contratados.
func (h *Hospital) Doctors() []*Doctor { return slices.Clone(h.doctors) }

// Staff devuelve todo el personal (doctores, camilleros...) como Attender.
func (h *Hospital) Staff() []Attender { return slices.Clone(h.staff) }

// Rooms devuelve los cuartos registrados.
func (h *Hospital) Rooms() []*Room { return slices.Clone(h.rooms) }

// History devuelve todos los episodios registrados.
func (h *Hospital) History() []EpisodeRecord { return slices.Clone(h.history) }

// FindPatient busca un paciente por su ID.
func (h *Hospital) FindPatient(id string) (*Patient, error) {
	p := h.findPatient(id)
	if p == nil {
		return nil, fmt.Errorf("find patient %q: %w", id, ErrPatientNotFound)
	}
	return p, nil
}

// ---------------------------------------------------------------------------
// Episodios de sueño
// ---------------------------------------------------------------------------

// RegisterEpisode registra que un paciente se quedó dormido en location.
// Pasos:
//  1. El paciente sufre el ataque y queda dormido en el pasillo.
//  2. Se busca un cuarto libre (si no hay, el paciente sigue en el pasillo).
//  3. Se despacha al siguiente miembro del personal disponible (polimorfismo).
//  4. Se guarda el episodio en el historial.
//
// El episodio SIEMPRE queda registrado (el paciente sí se durmió). Si no hubo
// cuarto o no hubo personal, se devuelve un error explicándolo, que se puede
// revisar con errors.Is(err, ErrNoRoomAvailable).
func (h *Hospital) RegisterEpisode(p *Patient, location string) error {
	if p == nil || location == "" {
		return fmt.Errorf("register episode: %w", ErrInvalidData)
	}
	if !h.hasPatient(p) {
		return fmt.Errorf("register episode of %s: %w", p.ID(), ErrPatientNotFound)
	}
	if p.state != Awake {
		return fmt.Errorf("register episode of %s: %w", p.ID(), ErrAlreadyAsleep)
	}

	// 1. Ataque de sueño.
	p.SufferSleepAttack(location)

	// 2. Buscar cama. Si falla, guardamos el error pero seguimos: el paciente
	//    necesita ser atendido aunque se quede en el pasillo.
	_, roomErr := h.AssignRoom(p)

	// 3. Despachar personal sin saber si es Doctor u Orderly.
	record, staffErr := h.attend(p, location)

	// 4. Guardar en el historial del hospital.
	h.history = append(h.history, record)

	// errors.Join junta los dos errores en uno solo (y devuelve nil si
	// ambos son nil). errors.Is sigue funcionando con cada uno por separado.
	return errors.Join(roomErr, staffErr)
}

// attend despacha a un miembro del personal para atender al paciente y
// devuelve el registro del episodio. Si nadie puede atender, el episodio
// se registra igual, pero sin attender.
func (h *Hospital) attend(p *Patient, location string) (EpisodeRecord, error) {
	attender, err := h.dispatchAttender()
	if err != nil {
		return newEpisodeRecord(p, nil, location), fmt.Errorf("attend %s: %w", p.ID(), err)
	}
	// Llamada POLIMÓRFICA: Go ejecuta Doctor.Attend u Orderly.Attend según
	// el tipo concreto que haya dentro de la variable attender.
	record, err := attender.Attend(p, location)
	if err != nil {
		return newEpisodeRecord(p, nil, location), err
	}
	return record, nil
}

// dispatchAttender elige al siguiente miembro del personal disponible,
// por turnos (round-robin): empieza en nextStaff y da la vuelta al slice.
// Solo usa métodos de la interfaz: no sabe ni le importa el tipo concreto.
func (h *Hospital) dispatchAttender() (Attender, error) {
	n := len(h.staff)
	for i := 0; i < n; i++ {
		idx := (h.nextStaff + i) % n // % n hace que después del último vuelva al primero
		candidate := h.staff[idx]
		if candidate.IsAvailable() {
			h.nextStaff = (idx + 1) % n // el próximo turno empieza en el siguiente
			return candidate, nil
		}
	}
	return nil, ErrNoStaffAvailable
}

// WakeUpPatient despierta a un paciente del hospital (y libera su cuarto).
func (h *Hospital) WakeUpPatient(p *Patient) error {
	if p == nil {
		return fmt.Errorf("wake up: %w", ErrInvalidData)
	}
	if !h.hasPatient(p) {
		return fmt.Errorf("wake up %s: %w", p.ID(), ErrPatientNotFound)
	}
	return p.WakeUp()
}

// ---------------------------------------------------------------------------
// Consultas requeridas (sección 5). Todas DEVUELVEN datos; ninguna imprime.
// ---------------------------------------------------------------------------

// PatientsInHallway es la consulta 5.1: devuelve los pacientes dormidos en
// un pasillo, para poder enviar a un camillero.
func (h *Hospital) PatientsInHallway() []*Patient {
	var result []*Patient
	for _, p := range h.patients {
		if p.state == AsleepInHallway {
			result = append(result, p)
		}
	}
	return result
}

// AssignRoom es la consulta 5.3: busca el primer cuarto disponible y acuesta
// ahí al paciente (el cuarto puede quedar Occupied y el paciente AsleepInBed).
// Si no hay cuarto libre devuelve ErrNoRoomAvailable: el paciente sigue en el
// pasillo, el programa no se cae y no se inventa ningún cuarto.
func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	if p == nil {
		return nil, fmt.Errorf("assign room: %w", ErrInvalidData)
	}
	if !h.hasPatient(p) {
		return nil, fmt.Errorf("assign room to %s: %w", p.ID(), ErrPatientNotFound)
	}
	if p.state != AsleepInHallway {
		return nil, fmt.Errorf("assign room to %s: %w", p.ID(), ErrNotInHallway)
	}
	for _, r := range h.rooms {
		if r.IsAvailable() {
			if err := r.Occupy(p); err != nil {
				return nil, fmt.Errorf("assign room to %s: %w", p.ID(), err)
			}
			return r, nil
		}
	}
	return nil, fmt.Errorf("assign room to %s %s: %w", p.ID(), p.Name(), ErrNoRoomAvailable)
}

// SevereReport es la consulta 5.4: cruza dos fuentes, los pacientes Severe
// y el historial de episodios, y devuelve cuántos episodios tuvo HOY cada uno.
//
// Se usa un map porque es justo una relación "paciente -> cantidad": buscar
// y sumar por clave es directo. Los pacientes Severe sin episodios aparecen
// con 0. Ojo: un map de Go NO tiene orden; quien imprima debe ordenar.
func (h *Hospital) SevereReport() map[*Patient]int {
	report := make(map[*Patient]int)
	for _, p := range h.patients {
		if p.level == Severe {
			report[p] = 0
		}
	}
	now := time.Now()
	for _, record := range h.history {
		// "comma ok": ok es true solo si el paciente está en el reporte (es Severe).
		if _, ok := report[record.patient]; ok && sameDay(record.time, now) {
			report[record.patient]++
		}
	}
	return report
}

// ---------------------------------------------------------------------------
// Ayudantes privados
// ---------------------------------------------------------------------------

// findPatient devuelve el paciente con ese ID, o nil si no existe.
func (h *Hospital) findPatient(id string) *Patient {
	for _, p := range h.patients {
		if p.ID() == id {
			return p
		}
	}
	return nil
}

// findStaff devuelve el miembro del personal con ese ID, o nil si no existe.
func (h *Hospital) findStaff(id string) Attender {
	for _, a := range h.staff {
		if a.ID() == id {
			return a
		}
	}
	return nil
}

// hasPatient dice si ese paciente (el mismo puntero) está admitido aquí.
func (h *Hospital) hasPatient(p *Patient) bool {
	return slices.Contains(h.patients, p)
}

// sameDay dice si dos instantes caen en el mismo día del calendario.
func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
