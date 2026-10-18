package game

import (
	"cmp"
	"slices"
)

// El reporte de fin de turno se arma con LAS 4 CONSULTAS DEL ENUNCIADO. Estos
// structs solo copian los datos para que la vista los muestre sin tocar el modelo.

// PatientLine es un paciente dormido fuera de una cama (consulta 5.1).
type PatientLine struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// EpisodeLine es un episodio atendido por un doctor (consulta 5.2).
type EpisodeLine struct {
	ID          string `json:"id"`
	PatientID   string `json:"patientId"`
	PatientName string `json:"patientName"`
	Location    string `json:"location"`
	RoomNumber  int    `json:"roomNumber"` // 0 = se quedó sin cama
	GameTime    string `json:"gameTime"`   // hora de JUEGO (EpisodeRecord guarda la hora real)
}

// DoctorLine agrupa los episodios de un doctor.
type DoctorLine struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Specialty string        `json:"specialty"`
	Patients  int           `json:"patients"` // cupo usado, x/4
	Episodes  []EpisodeLine `json:"episodes"`
}

// RoomLine es el estado de una cama (consulta 5.3).
type RoomLine struct {
	Number    int      `json:"number"`
	State     string   `json:"state"`
	Occupants []string `json:"occupants"` // "P-004 Wilfrido Berrio"
}

// SevereLine es un paciente severo con sus episodios del día (consulta 5.4).
type SevereLine struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Episodes int    `json:"episodes"`
}

// Report es el reporte completo del turno.
type Report struct {
	Day        int            `json:"day"`
	Hallway    []PatientLine  `json:"hallway"`
	Doctors    []DoctorLine   `json:"doctors"`
	Rooms      []RoomLine     `json:"rooms"`
	FreeRooms  int            `json:"freeRooms"`
	NoBedCount int            `json:"noBedCount"`
	Severe     []SevereLine   `json:"severe"`
	Score      ScoreBreakdown `json:"score"`
	Total      int            `json:"total"`
	Stars      int            `json:"stars"`
	Overworked []string       `json:"overworked"`
}

// Report arma el reporte con el estado de este momento. Al terminar el turno
// (18:00) es el reporte final.
func (s *Shift) Report() Report {
	r := Report{
		Day:        s.cfg.Day,
		NoBedCount: s.noBedCount,
		Score:      s.score,
		Total:      s.score.Total(),
		Stars:      s.Stars(),
		Overworked: s.Overworked(),
	}

	// Consulta 5.1: dormidos fuera de una cama.
	for _, p := range s.h.PatientsInHallway() {
		r.Hallway = append(r.Hallway, PatientLine{ID: p.ID(), Name: p.Name(), Location: p.Location()})
	}

	// Consulta 5.2: episodios de cada doctor. h.Doctors() sale vacío en el
	// hospital del juego (HireStaff recibe StaffMember, no *Doctor), por eso
	// se usa el puntero al doctor real que guarda cada StaffMember.
	for _, m := range s.staff {
		if !m.IsDoctor() {
			continue
		}
		line := DoctorLine{ID: m.ID(), Name: m.Name(), Specialty: m.doctor.Specialty(), Patients: m.PatientCount()}
		for _, e := range m.doctor.MyEpisodes() {
			line.Episodes = append(line.Episodes, EpisodeLine{
				ID:          e.ID(),
				PatientID:   e.Patient().ID(),
				PatientName: e.Patient().Name(),
				Location:    e.Location(),
				RoomNumber:  e.RoomNumber(),
				GameTime:    FormatGameTime(s.episodeTimes[e.ID()]),
			})
		}
		r.Doctors = append(r.Doctors, line)
	}

	// Consulta 5.3: estado de las camas.
	for _, room := range s.h.Rooms() {
		line := RoomLine{Number: room.Number(), State: room.State().String()}
		for _, p := range room.Occupants() {
			line.Occupants = append(line.Occupants, p.ID()+" "+p.Name())
		}
		if room.IsAvailable() {
			r.FreeRooms++
		}
		r.Rooms = append(r.Rooms, line)
	}

	// Consulta 5.4: el map no tiene orden, así que se ordena por ID.
	for p, episodes := range s.h.SevereReport() {
		r.Severe = append(r.Severe, SevereLine{ID: p.ID(), Name: p.Name(), Episodes: episodes})
	}
	slices.SortFunc(r.Severe, func(a, b SevereLine) int { return cmp.Compare(a.ID, b.ID) })
	return r
}
