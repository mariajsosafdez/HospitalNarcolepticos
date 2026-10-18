package game

import "slices"

// Snapshot es una copia plana de TODO el estado del turno, con etiquetas
// JSON. La puerta de pruebas (-agent) la vuelca para que un tester revise el
// juego sin mirar la pantalla.

// PatientDump es un paciente con su estado de juego Y su estado en el modelo.
type PatientDump struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Level         string       `json:"level"`
	State         PatientState `json:"state"`         // estado de juego
	HospitalState string       `json:"hospitalState"` // estado en el modelo (puede diferir)
	Camera        CameraID     `json:"camera"`
	Spot          int          `json:"spot"`
	Location      string       `json:"location"` // Patient.Location() del modelo
	Room          int          `json:"room"`     // 0 = sin cama
	Doctor        string       `json:"doctor"`   // ID del doctor a cargo ("" = ninguno)
	WaitMinutes   int          `json:"waitMinutes"`
	WaitTicks     int          `json:"waitTicks"`
	WakeTick      int          `json:"wakeTick"` // -1 = no aplica
	Leaving       bool         `json:"leaving"`
	Sleeps        int          `json:"sleeps"`
}

// StaffDump es un miembro del personal con su cooldown y su cupo.
type StaffDump struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Role          string     `json:"role"`
	State         StaffState `json:"state"`
	BusyUntil     int        `json:"busyUntil"`
	TransferUntil int        `json:"transferUntil"`
	CooldownLeft  int        `json:"cooldownLeft"` // ticks que faltan
	Patients      int        `json:"patients"`     // x de x/4 (0 en camilleros)
	Quota         int        `json:"quota"`        // 4 en doctores, 0 en camilleros
	Overworked    bool       `json:"overworked"`
}

// RoomDump es una habitación del modelo.
type RoomDump struct {
	Number    int      `json:"number"`
	Camera    CameraID `json:"camera"`
	State     string   `json:"state"`
	Occupants []string `json:"occupants"`
}

// Snapshot es el volcado completo.
type Snapshot struct {
	Day        int            `json:"day"`
	Seed       int64          `json:"seed"`
	Tick       int            `json:"tick"`
	Clock      string         `json:"clock"`
	Paused     bool           `json:"paused"`
	Over       bool           `json:"over"`
	Camera     CameraID       `json:"camera"`
	Unlocked   []CameraID     `json:"unlocked"`
	Pending    int            `json:"pending"` // pacientes que aún no llegan
	Patients   []PatientDump  `json:"patients"`
	Staff      []StaffDump    `json:"staff"`
	Rooms      []RoomDump     `json:"rooms"`
	Score      ScoreBreakdown `json:"score"`
	Total      int            `json:"total"`
	Stars      int            `json:"stars"`
	NoBedCount int            `json:"noBedCount"`
	Messages   []Message      `json:"messages"`
}

// Snapshot copia el estado actual del turno.
func (s *Shift) Snapshot() Snapshot {
	now := s.clock.Tick()
	snap := Snapshot{
		Day:        s.cfg.Day,
		Seed:       s.seed,
		Tick:       now,
		Clock:      s.clock.Now(),
		Paused:     s.clock.Paused(),
		Over:       s.finished,
		Camera:     s.camera,
		Unlocked:   s.world.Unlocked(),
		Pending:    len(s.pending),
		Score:      s.score,
		Total:      s.score.Total(),
		Stars:      s.Stars(),
		NoBedCount: s.noBedCount,
		Messages:   slices.Clone(s.messages),
	}
	for _, g := range s.patients {
		p := g.patient
		dump := PatientDump{
			ID:            p.ID(),
			Name:          p.Name(),
			Level:         p.Level().String(),
			State:         g.state,
			HospitalState: s.h.StateOf(p).String(),
			Camera:        g.camera,
			Spot:          g.spot,
			Location:      p.Location(),
			WaitMinutes:   g.WaitMinutes(now),
			WakeTick:      -1,
			Leaving:       g.leaving,
			Sleeps:        g.sleeps,
		}
		switch g.state {
		case AsleepWaiting, NoBedWaiting, Carried:
			dump.WaitTicks = now - g.sleepTick
		case InBed:
			dump.WakeTick = g.wakeTick
		}
		if p.Room() != nil {
			dump.Room = p.Room().Number()
		}
		if d := p.AssignedDoctor(); d != nil {
			dump.Doctor = d.ID()
		}
		snap.Patients = append(snap.Patients, dump)
	}
	for _, m := range s.staff {
		dump := StaffDump{
			ID:            m.ID(),
			Name:          m.Name(),
			Role:          m.Role(),
			State:         m.State(),
			BusyUntil:     m.busyUntil,
			TransferUntil: m.transferUntil,
			CooldownLeft:  max(m.busyUntil-now, 0),
			Patients:      m.PatientCount(),
			Overworked:    m.overworkCharged,
		}
		if m.IsDoctor() {
			dump.Quota = DoctorQuota
		}
		snap.Staff = append(snap.Staff, dump)
	}
	for _, room := range s.h.Rooms() {
		cam, _ := RoomCamera(room.Number())
		dump := RoomDump{Number: room.Number(), Camera: cam, State: room.State().String()}
		for _, p := range room.Occupants() {
			dump.Occupants = append(dump.Occupants, p.ID())
		}
		snap.Rooms = append(snap.Rooms, dump)
	}
	return snap
}
