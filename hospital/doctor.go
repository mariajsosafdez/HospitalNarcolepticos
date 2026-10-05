package hospital

import (
	"fmt"
	"slices"
)

// maxPatientsPerDoctor es cuántos pacientes puede tener un doctor a su cargo.
// Al llegar a ese número el doctor deja de estar disponible (IsAvailable).
const maxPatientsPerDoctor = 4

// Doctor es un médico del hospital. Puede diagnosticar pacientes y atender
// emergencias de sueño.
//
// Igual que Patient, embebe Person: no repite id, name ni age, y gracias a
// los métodos promovidos ID() y Name() ya cumple parte de la interfaz Attender.
type Doctor struct {
	Person
	specialty string
	patients  []*Patient      // pacientes a su cargo (diagnosticados por él)
	episodes  []EpisodeRecord // episodios que atendió: la relación vive DENTRO del doctor
}

// NewDoctor crea un doctor validando sus datos.
func NewDoctor(id, name string, age int, specialty string) (*Doctor, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new doctor %q: %w", id, err)
	}
	if specialty == "" {
		return nil, fmt.Errorf("new doctor %q: empty specialty: %w", id, ErrInvalidData)
	}
	return &Doctor{Person: person, specialty: specialty}, nil
}

// Specialty devuelve la especialidad del doctor.
func (d *Doctor) Specialty() string { return d.specialty }

// Patients devuelve una copia de la lista de pacientes a cargo del doctor.
func (d *Doctor) Patients() []*Patient { return slices.Clone(d.patients) }

// Role cumple la interfaz Attender.
func (d *Doctor) Role() string { return "Doctor" }

// IsAvailable cumple la interfaz Attender: el doctor está disponible
// mientras no tenga el máximo de pacientes a su cargo.
func (d *Doctor) IsAvailable() bool {
	return len(d.patients) < maxPatientsPerDoctor
}

// DiagnosePatient pone al paciente a cargo de este doctor.
// Falla si el paciente ya tiene otro doctor o si este doctor está lleno.
func (d *Doctor) DiagnosePatient(p *Patient) error {
	if p == nil {
		return fmt.Errorf("diagnose: %w", ErrInvalidData)
	}
	if p.assignedDoctor == d {
		return nil // ya es su paciente: no hay nada que hacer
	}
	if p.assignedDoctor != nil {
		return fmt.Errorf("diagnose %s: %w", p.id, ErrAlreadyAssigned)
	}
	if !d.IsAvailable() {
		return fmt.Errorf("diagnose %s by %s: %w", p.id, d.id, ErrDoctorFull)
	}
	d.patients = append(d.patients, p)
	p.assignedDoctor = d // se puede porque estamos dentro del mismo paquete
	return nil
}

// AttendSleepEmergency valida que de verdad haya una emergencia (el paciente
// está dormido). Si el paciente todavía no tiene doctor, este doctor lo
// diagnostica y lo toma a su cargo.
func (d *Doctor) AttendSleepEmergency(p *Patient) error {
	if p == nil {
		return fmt.Errorf("attend emergency: %w", ErrInvalidData)
	}
	if p.state == Awake {
		return fmt.Errorf("attend emergency of %s: %w", p.id, ErrPatientAwake)
	}
	if p.assignedDoctor == nil {
		if err := d.DiagnosePatient(p); err != nil {
			return fmt.Errorf("attend emergency of %s: %w", p.id, err)
		}
	}
	return nil
}

// Attend cumple la interfaz Attender: atiende la emergencia, crea el
// registro del episodio y lo guarda en la propia historia del doctor.
func (d *Doctor) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if err := d.AttendSleepEmergency(p); err != nil {
		// Con error devolvemos el "valor cero" de EpisodeRecord (vacío).
		return EpisodeRecord{}, err
	}
	record := newEpisodeRecord(p, d, location)
	d.episodes = append(d.episodes, record)
	return record, nil
}

// MyEpisodes es la consulta 5.2: devuelve todos los episodios en los que
// intervino este doctor. No recibe el hospital como parámetro porque el
// doctor guarda sus propios episodios. Devuelve una copia para proteger
// el slice interno.
func (d *Doctor) MyEpisodes() []EpisodeRecord {
	return slices.Clone(d.episodes)
}
