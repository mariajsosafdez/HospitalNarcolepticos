package web

import (
	"fmt"
	"strings"

	"costenos-narcolepsia/hospital"
)

// DTO = "Data Transfer Object": structs pensados SOLO para viajar como JSON
// hacia el navegador. Las etiquetas `json:"..."` dicen cómo se llama cada
// campo en el JSON.
//
// ¿Por qué no poner las etiquetas JSON directamente en el paquete hospital?
// Porque el enunciado exige que el modelo no sepa nada de la vista. Si
// mañana cambiamos la web por otra interfaz, hospital no se toca: solo se
// reemplaza este paquete.

type patientDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Age       int    `json:"age"`
	Level     string `json:"level"`
	State     string `json:"state"`
	StateCode string `json:"stateCode"` // "awake", "hallway" o "bed" (para los colores en CSS)
	Location  string `json:"location"`
	Room      int    `json:"room"`
	Doctor    string `json:"doctor"`
}

type roomDTO struct {
	Number    int          `json:"number"`
	Capacity  int          `json:"capacity"`
	State     string       `json:"state"`
	Available bool         `json:"available"`
	Occupants []patientDTO `json:"occupants"`
}

type staffDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Available bool   `json:"available"`
	Episodes  int    `json:"episodes"`
}

type episodeDTO struct {
	ID          string `json:"id"`
	Time        string `json:"time"`
	PatientID   string `json:"patientId"`
	PatientName string `json:"patientName"`
	Location    string `json:"location"`
	Room        int    `json:"room"`
	AttendedBy  string `json:"attendedBy"`
	Role        string `json:"role"`
	Summary     string `json:"summary"`
}

type doctorEpisodesDTO struct {
	DoctorID   string       `json:"doctorId"`
	DoctorName string       `json:"doctorName"`
	Specialty  string       `json:"specialty"`
	Episodes   []episodeDTO `json:"episodes"`
}

type severeDTO struct {
	Patient  patientDTO `json:"patient"`
	Episodes int        `json:"episodes"`
}

type simulationDTO struct {
	Running bool       `json:"running"`
	Last    *resultDTO `json:"last"` // nil (null en JSON) si aún no se ha corrido ninguna
}

type resultDTO struct {
	Goroutines  int `json:"goroutines"`
	Episodes    int `json:"episodes"`
	NoRoom      int `json:"noRoom"`
	WakeUps     int `json:"wakeUps"`
	Relocations int `json:"relocations"`
}

type stateDTO struct {
	Hospital       string              `json:"hospital"`
	TakenAt        string              `json:"takenAt"`
	Patients       []patientDTO        `json:"patients"`
	Rooms          []roomDTO           `json:"rooms"`
	Staff          []staffDTO          `json:"staff"`
	History        []episodeDTO        `json:"history"`
	Hallway        []patientDTO        `json:"hallway"`
	DoctorEpisodes []doctorEpisodesDTO `json:"doctorEpisodes"`
	SevereReport   []severeDTO         `json:"severeReport"`
	Simulation     simulationDTO       `json:"simulation"`
}

type scenarioDTO struct {
	Log   string   `json:"log"`
	State stateDTO `json:"state"`
}

// --- Cuerpos de las peticiones (lo que el navegador envía) ---

type admitRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Level string `json:"level"`
}

type staffRequest struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Age       int    `json:"age"`
	Specialty string `json:"specialty"` // solo para doctores
}

type roomRequest struct {
	Number   int `json:"number"`
	Capacity int `json:"capacity"`
}

type sleepRequest struct {
	Location string `json:"location"`
}

type simulationRequest struct {
	Rounds     int `json:"rounds"`
	MaxDelayMs int `json:"maxDelayMs"`
}

// ---------------------------------------------------------------------------
// Conversión Snapshot (modelo) -> DTO (JSON)
// ---------------------------------------------------------------------------

// toStateDTO convierte la foto del hospital en el JSON que entiende la página.
// Se usa make(..., 0, n) para que las listas vacías salgan como [] y no como
// null en el JSON (así el JavaScript no tiene que revisar null).
func toStateDTO(s hospital.Snapshot) stateDTO {
	dto := stateDTO{
		Hospital:       s.HospitalName,
		TakenAt:        s.TakenAt.Format("15:04:05"),
		Patients:       make([]patientDTO, 0, len(s.Patients)),
		Rooms:          make([]roomDTO, 0, len(s.Rooms)),
		Staff:          make([]staffDTO, 0, len(s.Staff)),
		History:        make([]episodeDTO, 0, len(s.History)),
		Hallway:        make([]patientDTO, 0, len(s.Hallway)),
		DoctorEpisodes: make([]doctorEpisodesDTO, 0, len(s.DoctorEpisodes)),
		SevereReport:   make([]severeDTO, 0, len(s.SevereReport)),
	}
	for _, p := range s.Patients {
		dto.Patients = append(dto.Patients, toPatientDTO(p))
	}
	for _, r := range s.Rooms {
		room := roomDTO{
			Number:    r.Number,
			Capacity:  r.Capacity,
			State:     r.State.String(),
			Available: r.State == hospital.Available,
			Occupants: make([]patientDTO, 0, len(r.Occupants)),
		}
		for _, p := range r.Occupants {
			room.Occupants = append(room.Occupants, toPatientDTO(p))
		}
		dto.Rooms = append(dto.Rooms, room)
	}
	for _, a := range s.Staff {
		dto.Staff = append(dto.Staff, staffDTO(a)) // mismos campos y tipos: conversión directa
	}
	for _, e := range s.History {
		dto.History = append(dto.History, toEpisodeDTO(e))
	}
	for _, p := range s.Hallway {
		dto.Hallway = append(dto.Hallway, toPatientDTO(p))
	}
	for _, d := range s.DoctorEpisodes {
		group := doctorEpisodesDTO{
			DoctorID:   d.DoctorID,
			DoctorName: d.DoctorName,
			Specialty:  d.Specialty,
			Episodes:   make([]episodeDTO, 0, len(d.Episodes)),
		}
		for _, e := range d.Episodes {
			group.Episodes = append(group.Episodes, toEpisodeDTO(e))
		}
		dto.DoctorEpisodes = append(dto.DoctorEpisodes, group)
	}
	for _, entry := range s.SevereReport {
		dto.SevereReport = append(dto.SevereReport, severeDTO{Patient: toPatientDTO(entry.Patient), Episodes: entry.Episodes})
	}
	return dto
}

func toPatientDTO(p hospital.PatientInfo) patientDTO {
	return patientDTO{
		ID:        p.ID,
		Name:      p.Name,
		Age:       p.Age,
		Level:     p.Level.String(),
		State:     p.State.String(),
		StateCode: stateCode(p.State),
		Location:  p.Location,
		Room:      p.RoomNumber,
		Doctor:    p.DoctorName,
	}
}

// toEpisodeDTO solo usa datos que nunca cambian (ID y nombre del paciente y
// del personal), así que es seguro leerlos sin el candado del hospital.
func toEpisodeDTO(e hospital.EpisodeRecord) episodeDTO {
	dto := episodeDTO{
		ID:          e.ID(),
		Time:        e.Time().Format("2006-01-02 15:04:05"),
		PatientID:   e.Patient().ID(),
		PatientName: e.Patient().Name(),
		Location:    e.Location(),
		Room:        e.RoomNumber(),
		AttendedBy:  "nobody",
		Summary:     e.Summary(),
	}
	if a := e.Attender(); a != nil {
		dto.AttendedBy = a.Name()
		dto.Role = a.Role()
	}
	return dto
}

func toResultDTO(r hospital.SimulationResult) *resultDTO {
	return &resultDTO{
		Goroutines:  r.Goroutines,
		Episodes:    r.Episodes,
		NoRoom:      r.NoRoom,
		WakeUps:     r.WakeUps,
		Relocations: r.Relocations,
	}
}

// stateCode traduce el estado a una palabra corta para usarla como clase CSS.
func stateCode(s hospital.PatientState) string {
	switch s {
	case hospital.AsleepInHallway:
		return "hallway"
	case hospital.AsleepInBed:
		return "bed"
	default:
		return "awake"
	}
}

// parseLevel convierte el texto del formulario en la constante tipada.
func parseLevel(text string) (hospital.NarcolepsyLevel, error) {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "mild":
		return hospital.Mild, nil
	case "moderate":
		return hospital.Moderate, nil
	case "severe":
		return hospital.Severe, nil
	default:
		return 0, fmt.Errorf("unknown narcolepsy level %q: %w", text, hospital.ErrInvalidData)
	}
}
