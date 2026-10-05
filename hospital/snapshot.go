package hospital

import (
	"cmp"
	"slices"
	"time"
)

// Snapshot es una "foto" del hospital en un instante: COPIAS de los datos
// en valores simples, armadas mientras se tiene el candado.
//
// ¿Para qué sirve? Mientras la simulación cambia pacientes y cuartos desde
// varias goroutines, leer p.State() directamente desde otra goroutine (por
// ejemplo, el servidor web) sería una carrera de datos. Con Snapshot, el
// lector recibe sus propias copias y puede usarlas con calma sin candado.
//
// Estos structs tienen campos exportados a propósito: NO son entidades del
// modelo (no tienen comportamiento ni reglas que proteger), son solo
// "paquetes de datos" de lectura para quien quiera mostrar el estado.
type Snapshot struct {
	HospitalName   string
	TakenAt        time.Time
	Patients       []PatientInfo
	Rooms          []RoomInfo
	Staff          []StaffInfo
	History        []EpisodeRecord // EpisodeRecord es inmutable: copiarlo es seguro
	Hallway        []PatientInfo   // consulta 5.1
	DoctorEpisodes []DoctorEpisodes
	SevereReport   []SevereEntry // consulta 5.4, ya ordenada por ID
}

// PatientInfo es la copia de los datos de un paciente.
type PatientInfo struct {
	ID         string
	Name       string
	Age        int
	Level      NarcolepsyLevel
	State      PatientState
	Location   string
	RoomNumber int    // 0 si no tiene cuarto
	DoctorName string // "" si no tiene doctor
}

// RoomInfo es la copia de los datos de un cuarto.
type RoomInfo struct {
	Number    int
	Capacity  int
	State     RoomState
	Occupants []PatientInfo
}

// StaffInfo es la copia de los datos de un miembro del personal.
type StaffInfo struct {
	ID        string
	Name      string
	Role      string
	Available bool
	Episodes  int // cuántos episodios ha atendido
}

// DoctorEpisodes agrupa los episodios de un doctor (consulta 5.2).
type DoctorEpisodes struct {
	DoctorID   string
	DoctorName string
	Specialty  string
	Episodes   []EpisodeRecord
}

// SevereEntry es una fila del reporte de severidad (consulta 5.4).
type SevereEntry struct {
	Patient  PatientInfo
	Episodes int
}

// Snapshot toma el candado UNA sola vez y copia todo el estado. Así la foto
// es consistente: nunca mezcla datos de antes y después de un cambio.
func (h *Hospital) Snapshot() Snapshot {
	h.mu.Lock()
	defer h.mu.Unlock()

	s := Snapshot{HospitalName: h.name, TakenAt: time.Now()}

	for _, p := range h.patients {
		s.Patients = append(s.Patients, patientInfo(p))
	}
	for _, r := range h.rooms {
		info := RoomInfo{Number: r.number, Capacity: r.capacity, State: r.state}
		for _, p := range r.occupants {
			info.Occupants = append(info.Occupants, patientInfo(p))
		}
		s.Rooms = append(s.Rooms, info)
	}
	for _, a := range h.staff {
		s.Staff = append(s.Staff, StaffInfo{
			ID:        a.ID(),
			Name:      a.Name(),
			Role:      a.Role(),
			Available: a.IsAvailable(),
			Episodes:  h.countEpisodesByLocked(a),
		})
	}
	s.History = slices.Clone(h.history)

	// Consultas, usando las versiones "Locked" (ya tenemos el candado).
	for _, p := range h.patientsInHallwayLocked() {
		s.Hallway = append(s.Hallway, patientInfo(p))
	}
	for _, d := range h.doctors {
		s.DoctorEpisodes = append(s.DoctorEpisodes, DoctorEpisodes{
			DoctorID:   d.ID(),
			DoctorName: d.Name(),
			Specialty:  d.specialty,
			Episodes:   d.MyEpisodes(),
		})
	}
	for p, count := range h.severeReportLocked() {
		s.SevereReport = append(s.SevereReport, SevereEntry{Patient: patientInfo(p), Episodes: count})
	}
	// El map no tiene orden: ordenamos por ID para que la foto sea estable.
	slices.SortFunc(s.SevereReport, func(a, b SevereEntry) int {
		return cmp.Compare(a.Patient.ID, b.Patient.ID)
	})
	return s
}

// patientInfo copia los datos de un paciente. Requiere el candado tomado.
func patientInfo(p *Patient) PatientInfo {
	info := PatientInfo{
		ID:       p.ID(),
		Name:     p.Name(),
		Age:      p.Age(),
		Level:    p.level,
		State:    p.state,
		Location: p.currentLocation,
	}
	if p.room != nil {
		info.RoomNumber = p.room.number
	}
	if p.assignedDoctor != nil {
		info.DoctorName = p.assignedDoctor.Name()
	}
	return info
}

// countEpisodesByLocked cuenta los episodios del historial atendidos por a.
// Compara por ID, así funciona con cualquier tipo de Attender.
func (h *Hospital) countEpisodesByLocked(a Attender) int {
	count := 0
	for _, record := range h.history {
		if record.attender != nil && record.attender.ID() == a.ID() {
			count++
		}
	}
	return count
}
