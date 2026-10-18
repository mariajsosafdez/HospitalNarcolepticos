package game

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"costenos-narcolepsia/hospital"
)

// newDemoShift arma el turno de la Demo (sin azar, con los datos del escenario).
func newDemoShift(t *testing.T) *Shift {
	t.Helper()
	s, err := NewShift(DemoConfig(), 1)
	if err != nil {
		t.Fatalf("NewShift: %v", err)
	}
	return s
}

// newCustomShift arma un turno sin azar con la cantidad de doctores,
// camilleros, habitaciones y pacientes que pida el test.
func newCustomShift(t *testing.T, day, doctors, orderlies, rooms, patients int) *Shift {
	t.Helper()
	cfg := DemoConfig()
	cfg.Day = day
	cfg.Doctors = slices.Clone(castDoctors[:doctors])
	cfg.Orderlies = slices.Clone(castOrderlies[:orderlies])
	cfg.Rooms = nil
	for i := 0; i < rooms; i++ {
		cfg.Rooms = append(cfg.Rooms, 101+i)
	}
	cfg.Patients = slices.Clone(castPatients[:patients])
	s, err := NewShift(cfg, 1)
	if err != nil {
		t.Fatalf("NewShift: %v", err)
	}
	return s
}

// advance avanza n ticks.
func advance(s *Shift, n int) {
	for i := 0; i < n; i++ {
		s.Tick()
	}
}

// mustSleepAndAssign duerme a un paciente en la cámara 1B y lo asigna.
func mustSleepAndAssign(t *testing.T, s *Shift, patientID, staffID string) {
	t.Helper()
	if err := s.ForceSleep(patientID, "1B", "waiting room"); err != nil {
		t.Fatalf("ForceSleep %s: %v", patientID, err)
	}
	if err := s.Assign(patientID, staffID); err != nil {
		t.Fatalf("Assign %s/%s: %v", patientID, staffID, err)
	}
}

func TestChosenStaffIsDispatchedEvenWhenRoundRobinPointsElsewhere(t *testing.T) {
	s := newDemoShift(t)
	// Cuatro asignaciones; el round-robin del hospital empieza en D-001, pero
	// el jugador elige otro cada vez. El episodio debe quedar con el elegido.
	steps := []struct{ patient, staff string }{
		{"P-001", "O-001"}, // el round-robin apuntaba a D-001
		{"P-002", "D-002"}, // ahora apunta a D-001 (después de O-001) y se elige D-002
	}
	for _, step := range steps {
		mustSleepAndAssign(t, s, step.patient, step.staff)
	}
	history := s.Hospital().History()
	if len(history) != len(steps) {
		t.Fatalf("history has %d episodes, want %d", len(history), len(steps))
	}
	for i, step := range steps {
		if got := history[i].Attender().ID(); got != step.staff {
			t.Errorf("episode %d attended by %s, want %s", i, got, step.staff)
		}
		if got := history[i].Patient().ID(); got != step.patient {
			t.Errorf("episode %d is for %s, want %s", i, got, step.patient)
		}
	}
}

func TestStaffMemberOnlyAvailableWhenSelected(t *testing.T) {
	s := newDemoShift(t)
	for _, m := range s.Staff() {
		if m.IsAvailable() {
			t.Errorf("%s looks available without being selected", m.ID())
		}
		if !m.Free() {
			t.Errorf("%s should be free at tick 0", m.ID())
		}
	}
	// El truco de la sección 4.3: HireStaff solo llena Doctors() con *Doctor,
	// y el juego contrata StaffMember, así que la lista sale vacía.
	if got := len(s.Hospital().Doctors()); got != 0 {
		t.Errorf("Doctors() has %d entries, want 0", got)
	}
	if got := len(s.Hospital().Staff()); got != 3 {
		t.Errorf("Staff() has %d entries, want 3", got)
	}
}

func TestNoBedKeepsPatientWhereTheyFell(t *testing.T) {
	s := newCustomShift(t, 1, 2, 1, 1, 3) // una sola cama
	mustSleepAndAssign(t, s, "P-001", "O-001")

	if err := s.ForceSleep("P-002", "8", "champeta dance hall"); err != nil {
		t.Fatal(err)
	}
	err := s.Assign("P-002", "D-001")
	if !errors.Is(err, hospital.ErrNoRoomAvailable) {
		t.Fatalf("Assign error = %v, want ErrNoRoomAvailable", err)
	}
	g := s.FindPatient("P-002")
	if g.State() != NoBedWaiting {
		t.Errorf("game state = %v, want NoBedWaiting", g.State())
	}
	if g.Camera() != "8" {
		t.Errorf("camera = %s, want 8 (where they fell asleep)", g.Camera())
	}
	if got := s.Hospital().StateOf(g.Patient()); got != hospital.AsleepInHallway {
		t.Errorf("model state = %v, want AsleepInHallway", got)
	}
	if got := g.Patient().Location(); got != "champeta dance hall" {
		t.Errorf("model location = %q", got)
	}
	// El doctor sí atendió: el episodio existe y entró en cooldown.
	if got := len(s.Hospital().History()); got != 2 {
		t.Errorf("history has %d episodes, want 2", got)
	}
	// Sin cama no hay nada que cargar: va directo a descansar, sin traslado.
	d := s.FindStaff("D-001")
	if d.State() != StaffResting {
		t.Errorf("D-001 state = %v, want Resting (no stretcher to push)", d.State())
	}
	if g.Carrier() != nil {
		t.Error("a patient without a bed must have no carrier")
	}
	if d.TransferUntil() != s.Clock().Tick() {
		t.Errorf("TransferUntil = %d, want now (%d)", d.TransferUntil(), s.Clock().Tick())
	}
	advance(s, doctorRestTicks-1)
	if d.Free() {
		t.Error("still resting one tick before the normal rest time ends")
	}
	advance(s, 1)
	if !d.Free() || d.State() != StaffAvailable {
		t.Errorf("after the normal rest D-001 state = %v, want Available", d.State())
	}
	if s.NoBedCount() != 1 {
		t.Errorf("NoBedCount = %d, want 1", s.NoBedCount())
	}
	// Y sigue contando la espera.
	advance(s, MinutesToTicks(30))
	if got := g.WaitMinutes(s.Clock().Tick()); got < 30 {
		t.Errorf("wait = %d min, want >= 30", got)
	}
}

func TestDischargeFreesTheRoomAndMoveToBedUsesIt(t *testing.T) {
	s := newCustomShift(t, 1, 2, 1, 1, 3)
	mustSleepAndAssign(t, s, "P-001", "O-001")
	advance(s, orderlyTransferTicks) // llega a la cama
	if st := s.FindPatient("P-001").State(); st != InBed {
		t.Fatalf("P-001 state = %v, want InBed", st)
	}
	if err := s.ForceSleep("P-002", "2", "cafeteria"); err != nil {
		t.Fatal(err)
	}
	if err := s.Assign("P-002", "D-001"); !errors.Is(err, hospital.ErrNoRoomAvailable) {
		t.Fatalf("want no room, got %v", err)
	}
	// Sin cama libre, MoveToBed falla y no cambia nada.
	advance(s, doctorTransferTicks+doctorRestTicks)
	if err := s.MoveToBed("P-002", "D-002"); !errors.Is(err, hospital.ErrNoRoomAvailable) {
		t.Fatalf("MoveToBed without room = %v", err)
	}
	if st := s.FindPatient("P-002").State(); st != NoBedWaiting {
		t.Errorf("state = %v, want NoBedWaiting", st)
	}

	if err := s.ForceWake("P-001"); err != nil {
		t.Fatal(err)
	}
	s.Tick()
	if err := s.Discharge("P-001"); err != nil {
		t.Fatalf("Discharge: %v", err)
	}
	room := s.Hospital().Rooms()[0]
	if room.State() != hospital.Available {
		t.Errorf("room state = %v, want Available", room.State())
	}
	if err := s.MoveToBed("P-002", "D-002"); err != nil {
		t.Fatalf("MoveToBed: %v", err)
	}
	if len(s.Hospital().History()) != 2 {
		t.Errorf("MoveToBed must not create an episode, history = %d", len(s.Hospital().History()))
	}
	advance(s, doctorTransferTicks)
	if st := s.FindPatient("P-002").State(); st != InBed {
		t.Errorf("P-002 state = %v, want InBed", st)
	}
}

func TestCannotDischargeWhileAsleep(t *testing.T) {
	s := newDemoShift(t)
	mustSleepAndAssign(t, s, "P-001", "O-001")
	tests := []struct {
		name string
		do   func()
	}{
		{"asleep waiting", func() {}},
		{"carried", func() {}},
		{"in bed", func() { advance(s, orderlyTransferTicks) }},
	}
	// Se prueba en AsleepWaiting con otro paciente y luego en Carried/InBed.
	if err := s.ForceSleep("P-002", "2", "cafeteria"); err != nil {
		t.Fatal(err)
	}
	if err := s.Discharge("P-002"); !errors.Is(err, ErrWrongState) {
		t.Errorf("discharge asleep-waiting = %v, want ErrWrongState", err)
	}
	for _, tt := range tests[1:] {
		tt.do()
		if err := s.Discharge("P-001"); !errors.Is(err, ErrWrongState) {
			t.Errorf("%s: discharge = %v, want ErrWrongState", tt.name, err)
		}
	}
	if err := s.Discharge("P-999"); !errors.Is(err, ErrUnknownPatient) {
		t.Errorf("unknown patient = %v", err)
	}
}

func TestCooldownTimings(t *testing.T) {
	tests := []struct {
		name                string
		staffID             string
		transfer, rest      int
		wantTransferSeconds int
		wantRestSeconds     int
	}{
		{"doctor 5s + 15s", "D-001", doctorTransferTicks, doctorRestTicks, 5, 15},
		{"orderly 3s + 5s", "O-001", orderlyTransferTicks, orderlyRestTicks, 3, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.transfer != tt.wantTransferSeconds*60 || tt.rest != tt.wantRestSeconds*60 {
				t.Fatalf("constants = %d/%d ticks", tt.transfer, tt.rest)
			}
			s := newDemoShift(t)
			mustSleepAndAssign(t, s, "P-001", tt.staffID)
			m := s.FindStaff(tt.staffID)
			if m.Free() || m.IsAvailable() {
				t.Error("just used staff must not be free")
			}
			if err := s.ForceSleep("P-002", "2", "cafeteria"); err != nil {
				t.Fatal(err)
			}
			if err := s.Assign("P-002", tt.staffID); !errors.Is(err, ErrStaffBusy) {
				t.Errorf("Assign while busy = %v, want ErrStaffBusy", err)
			}
			advance(s, tt.transfer-1)
			if m.State() != StaffTransferring {
				t.Errorf("state = %v, want Transferring", m.State())
			}
			advance(s, 1)
			if m.State() != StaffResting {
				t.Errorf("state = %v, want Resting", m.State())
			}
			advance(s, tt.rest-1)
			if m.Free() {
				t.Error("still resting one tick before the end")
			}
			advance(s, 1)
			if !m.Free() || m.State() != StaffAvailable {
				t.Errorf("after cooldown state = %v, want Available", m.State())
			}
		})
	}
}

func TestOverworkIsChargedOnlyOnce(t *testing.T) {
	// Un solo doctor y 5 pacientes: al cuarto paciente llega a 4/4.
	s := newCustomShift(t, 3, 1, 0, 6, 5)
	for i := 1; i <= 4; i++ {
		id := fmt.Sprintf("P-%03d", i)
		mustSleepAndAssign(t, s, id, "D-001")
		advance(s, doctorTransferTicks+doctorRestTicks)
	}
	m := s.FindStaff("D-001")
	if m.PatientCount() != 4 {
		t.Fatalf("patients = %d, want 4", m.PatientCount())
	}
	want := -ScalePenalty(overworkPenalty, 3) // día 3: x1,5 = -75
	if got := s.Score().OverworkLost; got != want {
		t.Errorf("OverworkLost = %d, want %d", got, want)
	}
	if want != -75 {
		t.Errorf("day 3 overwork = %d, want -75", want)
	}
	// Volver a revisar no cobra otra vez.
	s.checkOverwork(m)
	s.checkOverwork(m)
	if got := s.Score().OverworkLost; got != want {
		t.Errorf("charged again: %d", got)
	}
	if m.State() != StaffNoQuota {
		t.Errorf("state = %v, want NoQuota", m.State())
	}
	// Sin cupo no se puede elegir.
	if err := s.ForceSleep("P-005", "1B", "waiting room"); err != nil {
		t.Fatal(err)
	}
	if err := s.Assign("P-005", "D-001"); !errors.Is(err, ErrStaffNoQuota) {
		t.Errorf("Assign to full doctor = %v, want ErrStaffNoQuota", err)
	}
	if len(s.Overworked()) != 1 {
		t.Errorf("overworked = %v", s.Overworked())
	}
}

func TestClock(t *testing.T) {
	var c Clock
	if c.Now() != "12:00" {
		t.Errorf("start = %s", c.Now())
	}
	c.Pause()
	for i := 0; i < 100; i++ {
		c.Advance()
	}
	if c.Tick() != 0 {
		t.Errorf("paused clock advanced to %d", c.Tick())
	}
	c.Resume()
	for i := 0; i < ShiftTicks-1; i++ {
		c.Advance()
	}
	if c.Over() {
		t.Error("over one tick early")
	}
	c.Advance()
	if !c.Over() || c.Now() != "18:00" || c.Tick() != 10800 {
		t.Errorf("end: over=%v now=%s tick=%d", c.Over(), c.Now(), c.Tick())
	}
	c.Advance()
	if c.Tick() != 10800 {
		t.Errorf("clock moved past 18:00: %d", c.Tick())
	}
}

func TestFormatGameTime(t *testing.T) {
	tests := []struct {
		tick int
		want string
	}{
		{0, "12:00"},
		{30, "12:01"},
		{MinutesToTicks(135), "14:15"},
		{MinutesToTicks(59), "12:59"},
		{ShiftTicks, "18:00"},
	}
	for _, tt := range tests {
		if got := FormatGameTime(tt.tick); got != tt.want {
			t.Errorf("FormatGameTime(%d) = %s, want %s", tt.tick, got, tt.want)
		}
	}
}

func TestShiftPauseStopsTheGame(t *testing.T) {
	s := newDemoShift(t)
	s.Pause()
	advance(s, 500)
	if s.Clock().Tick() != 0 {
		t.Errorf("tick = %d while paused", s.Clock().Tick())
	}
	s.Resume()
	advance(s, ShiftTicks+50)
	if !s.Finished() || s.Clock().Now() != "18:00" {
		t.Errorf("finished=%v now=%s", s.Finished(), s.Clock().Now())
	}
	if err := s.Discharge("P-001"); !errors.Is(err, ErrShiftOver) {
		t.Errorf("action after the shift = %v, want ErrShiftOver", err)
	}
}

func TestRateWait(t *testing.T) {
	tests := []struct {
		minutes int
		rating  WaitRating
		points  int
	}{
		{0, Excellent, 100}, {20, Excellent, 100},
		{21, Good, 70}, {45, Good, 70},
		{46, Regular, 40}, {90, Regular, 40},
		{91, Bad, 10}, {400, Bad, 10},
	}
	for _, tt := range tests {
		rating, points := RateWait(tt.minutes)
		if rating != tt.rating || points != tt.points {
			t.Errorf("RateWait(%d) = %v/%d, want %v/%d", tt.minutes, rating, points, tt.rating, tt.points)
		}
	}
}

func TestScalePenalty(t *testing.T) {
	tests := []struct {
		base, day, want int
	}{
		{50, 1, 50}, {50, 2, 63}, {50, 3, 75}, {50, 4, 88}, {50, 5, 100},
		{2, 1, 2}, {2, 5, 4}, {50, 0, 50}, {50, 9, 50},
	}
	for _, tt := range tests {
		if got := ScalePenalty(tt.base, tt.day); got != tt.want {
			t.Errorf("ScalePenalty(%d, day %d) = %d, want %d", tt.base, tt.day, got, tt.want)
		}
	}
}

// TestPenaltyRoundedOncePerType: los puntos de cada tipo salen de
// redondeo(cobros x base x multiplicador), una sola vez, para los 5 días.
func TestPenaltyRoundedOncePerType(t *testing.T) {
	tests := []struct {
		day, bed, overwork, unattended int // cobros
		wantBed, wantOverwork, wantUn  int
	}{
		{1, 4, 1, 1, -8, -50, -10},
		{2, 4, 1, 1, -10, -63, -13}, // cama: 4 x 2 x 1,25 = 10; sin atender: 12,5 -> 13
		{2, 1, 0, 2, -3, 0, -25},    // 1 cobro: 2,5 -> 3; 2 sin atender: 25 exacto
		{3, 3, 2, 3, -9, -150, -45},
		{4, 2, 1, 1, -7, -88, -18}, // 2 x 2 x 1,75 = 7; 17,5 -> 18; 87,5 -> 88
		{5, 5, 3, 2, -20, -300, -40},
		{5, 0, 0, 0, 0, 0, 0},
	}
	for _, tt := range tests {
		sc := ScoreBreakdown{BedCharges: tt.bed, OverworkEvents: tt.overwork, UnattendedCount: tt.unattended}
		sc.recomputePenalties(tt.day)
		if sc.BedPenalty != tt.wantBed || sc.OverworkLost != tt.wantOverwork || sc.Unattended != tt.wantUn {
			t.Errorf("day %d %+v: got %d/%d/%d, want %d/%d/%d", tt.day, tt, sc.BedPenalty, sc.OverworkLost, sc.Unattended, tt.wantBed, tt.wantOverwork, tt.wantUn)
		}
	}
}

func TestMainPenaltyAndOneStarLine(t *testing.T) {
	tests := []struct {
		sc    ScoreBreakdown
		cause string
		line  string
	}{
		{ScoreBreakdown{BedPenalty: -20, OverworkLost: -50, Unattended: -10}, "overwork", endLinesOneStar[2]},
		{ScoreBreakdown{BedPenalty: -30, OverworkLost: -25}, "bed", endLinesOneStar[1]},
		{ScoreBreakdown{Unattended: -30, BedPenalty: -4}, "unattended", endLinesOneStar[0]},
		{ScoreBreakdown{BedPenalty: -10, Unattended: -10}, "bed", endLinesOneStar[1]}, // empate: gana cama
		{ScoreBreakdown{}, "", endLinesOneStar[3]},
	}
	for _, tt := range tests {
		if got := tt.sc.MainPenalty(); got != tt.cause {
			t.Errorf("MainPenalty(%+v) = %q, want %q", tt.sc, got, tt.cause)
		}
		if got := ReportLine(1, tt.cause, 7); got != tt.line {
			t.Errorf("ReportLine(1, %q) = %q", tt.cause, got)
		}
	}
	// Con más de 1★ la causa no cambia la frase.
	if got := ReportLine(3, "bed", 7); got != EndLine(3, 7) {
		t.Errorf("3 stars: %q", got)
	}
}

func TestStars(t *testing.T) {
	tests := []struct {
		points, episodes, want int
	}{
		{0, 10, 1}, {-80, 10, 1},
		{359, 10, 1}, {360, 10, 2}, // 30 % de 1200
		{600, 10, 3}, {840, 10, 4}, {959, 10, 4}, {960, 10, 5}, {1200, 10, 5}, // 5★ desde el 80 %
		{50, 0, 5},
	}
	for _, tt := range tests {
		if got := Stars(tt.points, tt.episodes); got != tt.want {
			t.Errorf("Stars(%d, %d) = %d, want %d", tt.points, tt.episodes, got, tt.want)
		}
	}
}

func TestSleepChance(t *testing.T) {
	tests := []struct {
		level hospital.NarcolepsyLevel
		cam   CameraID
		want  float64
	}{
		{hospital.Mild, "1B", 0.03},
		{hospital.Moderate, "1B", 0.06},
		{hospital.Severe, "1B", 0.10},
		{hospital.Mild, "2", 0.06},     // cafetería x2
		{hospital.Severe, "2", 0.20},   // cafetería x2
		{hospital.Moderate, "6", 0.09}, // patio x1,5
	}
	for _, tt := range tests {
		got := SleepChance(tt.level, tt.cam)
		if diff := got - tt.want; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("SleepChance(%v, %s) = %v, want %v", tt.level, tt.cam, got, tt.want)
		}
	}
}

// En los últimos 20 min de juego (17:40 a 18:00) ya no hay ataques de sueño
// nuevos: sería una penalización garantizada.
func TestSleepWindowOpen(t *testing.T) {
	tests := []struct {
		name string
		tick int
		want bool
	}{
		{"start of the shift", 0, true},
		{"17:39", MinutesToTicks(5*60 + 39), true},
		{"one tick before 17:40", MinutesToTicks(5*60+40) - 1, true},
		{"17:40 exactly", MinutesToTicks(5*60 + 40), false},
		{"17:59", MinutesToTicks(5*60 + 59), false},
		{"18:00", ShiftTicks, false},
	}
	for _, tt := range tests {
		if got := sleepWindowOpen(tt.tick); got != tt.want {
			t.Errorf("%s: sleepWindowOpen(%d) = %v, want %v", tt.name, tt.tick, got, tt.want)
		}
	}
	if noSleepFinalMinutes*TicksPerMinute != 10*TicksPerSecond {
		t.Error("the final window must be 10 real seconds")
	}
}

// Ningún paciente se duerme en la ventana final, aunque la probabilidad fuera
// muy alta: se recorre el turno completo y se revisa la hora de cada sueño.
func TestNoNewSleepInTheFinalWindow(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		cfg, err := ConfigForDay(5)
		if err != nil {
			t.Fatal(err)
		}
		s, err := NewShift(cfg, seed)
		if err != nil {
			t.Fatal(err)
		}
		slept := 0
		for !s.Clock().Over() {
			before := 0
			for _, g := range s.Patients() {
				before += g.sleeps
			}
			s.Tick()
			after := 0
			for _, g := range s.Patients() {
				after += g.sleeps
			}
			if after > before {
				slept++
				if s.Clock().Tick()-1 >= ShiftTicks-600 { // 600 ticks = 10 s reales
					t.Fatalf("seed %d: someone fell asleep at tick %d, inside the final window", seed, s.Clock().Tick())
				}
			}
		}
		if slept == 0 {
			t.Errorf("seed %d: nobody ever slept, the test checks nothing", seed)
		}
	}
}

func TestWorldNeighboursAndSteps(t *testing.T) {
	day1, _ := ConfigForDay(1)
	w := NewWorld(day1.Cameras, day1.Rooms)
	if got := w.Neighbours("1B"); !slices.Equal(got, []CameraID{"1A", "3"}) {
		t.Errorf("day 1 neighbours of 1B = %v (4 and 8 are still locked)", got)
	}
	full := NewWorld(DemoConfig().Cameras, nil)
	tests := []struct{ from, to, want CameraID }{
		{"6", "1A", "4"},
		{"4", "1A", "1B"},
		{"5", "1A", "3"},
		{"8", "1A", "1B"},
		{"1B", "1A", "1A"},
		{"1A", "1A", "1A"},
	}
	for _, tt := range tests {
		if got := full.StepToward(tt.from, tt.to); got != tt.want {
			t.Errorf("StepToward(%s -> %s) = %s, want %s", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestDayConfigs(t *testing.T) {
	wantDoctors := []int{2, 2, 3, 3, 3}
	wantRooms := []int{3, 4, 4, 5, 6}
	wantCams := []int{4, 5, 6, 7, 8}
	for day := 1; day <= MaxDay; day++ {
		cfg, err := ConfigForDay(day)
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Doctors) != wantDoctors[day-1] || len(cfg.Rooms) != wantRooms[day-1] || len(cfg.Cameras) != wantCams[day-1] {
			t.Errorf("day %d: %d doctors, %d rooms, %d cameras", day, len(cfg.Doctors), len(cfg.Rooms), len(cfg.Cameras))
		}
		severe := 0
		for _, p := range cfg.Patients {
			if p.Level == hospital.Severe {
				severe++
			}
		}
		if severe < 2 {
			t.Errorf("day %d has %d severe patients, want >= 2", day, severe)
		}
		if _, err := NewShift(cfg, 7); err != nil {
			t.Errorf("day %d: NewShift: %v", day, err)
		}
	}
	if _, err := ConfigForDay(0); !errors.Is(err, ErrInvalidDay) {
		t.Errorf("day 0 = %v", err)
	}
	if _, err := ConfigForDay(6); !errors.Is(err, ErrInvalidDay) {
		t.Errorf("day 6 = %v", err)
	}
}

func TestAwakeInBedPenalty(t *testing.T) {
	s := newCustomShift(t, 2, 2, 1, 3, 3) // día 2: n cobros de -2 x 1,25, redondeado UNA vez
	mustSleepAndAssign(t, s, "P-001", "O-001")
	advance(s, orderlyTransferTicks)
	if err := s.ForceWake("P-001"); err != nil {
		t.Fatal(err)
	}
	s.Tick() // pasa a AwakeInBed
	tests := []struct {
		afterMinutes int
		want         int
	}{
		{15, 0},  // dentro de la gracia
		{24, 0},  // todavía no se cumplen los 10 min extra
		{25, -3}, // primer cobro: 2,5 -> 3
		{34, -3},
		{35, -5}, // 2 cobros: 4 x 1,25 = 5 exacto (antes, por cobro, daba 6)
		{45, -8}, // 3 cobros: 6 x 1,25 = 7,5 -> 8
	}
	elapsed := 0
	for _, tt := range tests {
		advance(s, MinutesToTicks(tt.afterMinutes-elapsed))
		elapsed = tt.afterMinutes
		if got := s.Score().BedPenalty; got != tt.want {
			t.Errorf("after %d min: BedPenalty = %d, want %d", tt.afterMinutes, got, tt.want)
		}
	}
}

func TestSevereBonusOnlyForDoctors(t *testing.T) {
	s := newDemoShift(t)
	mustSleepAndAssign(t, s, "P-002", "O-001") // severa, pero la atiende un camillero
	if got := s.Score().SevereBonus; got != 0 {
		t.Errorf("orderly gave a bonus: %d", got)
	}
	mustSleepAndAssign(t, s, "P-004", "D-001") // severo con doctor
	if got := s.Score().SevereBonus; got != 20 {
		t.Errorf("SevereBonus = %d, want 20", got)
	}
}

func TestWaitIsScoredWhenReachingBed(t *testing.T) {
	s := newDemoShift(t)
	mustSleepAndAssign(t, s, "P-001", "D-001") // espera = 5 s de traslado = 10 min de juego
	advance(s, doctorTransferTicks)
	sc := s.Score()
	if sc.Excellent != 1 || sc.WaitPoints != 100 || sc.Episodes != 1 {
		t.Errorf("score = %+v", sc)
	}
	if got := s.FindPatient("P-001").WaitMinutes(s.Clock().Tick()); got != 10 {
		t.Errorf("wait = %d min, want 10", got)
	}
}

// consoleScenario repite los pasos 1 a 3 de runScenario (main.go) directamente
// sobre el modelo, con el reparto por turnos del hospital, y devuelve las
// 4 consultas ya normalizadas para poder compararlas.
func consoleScenario(t *testing.T) queryResult {
	t.Helper()
	h := hospital.NewHospital("scenario")
	for _, d := range []struct {
		id, name string
		age      int
		spec     string
	}{{"D-001", "Dr. Karen Ospina", 41, "Neurology"}, {"D-002", "Dr. Alvaro Mendoza", 55, "Sleep Medicine"}} {
		doc, err := hospital.NewDoctor(d.id, d.name, d.age, d.spec)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.HireDoctor(doc); err != nil {
			t.Fatal(err)
		}
	}
	orderly, _ := hospital.NewOrderly("O-001", "Chepe Barrios", 33)
	if err := h.HireStaff(orderly); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{101, 102, 103} {
		room, _ := hospital.NewRoom(n, 1)
		if err := h.AddRoom(room); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range DemoConfig().Patients {
		p, _ := hospital.NewPatient(d.ID, d.Name, d.Age, d.Level)
		if err := h.AdmitPatient(p); err != nil {
			t.Fatal(err)
		}
	}
	for i, id := range []string{"P-002", "P-004", "P-005"} {
		p, _ := h.FindPatient(id)
		doctor := h.Doctors()[[]int{0, 0, 1}[i]]
		if err := doctor.DiagnosePatient(p); err != nil {
			t.Fatal(err)
		}
	}
	for _, a := range []struct{ id, location string }{
		{"P-001", "cafeteria (sancocho line)"},
		{"P-002", "champeta dance hall"},
		{"P-003", "radiology queue"},
		{"P-004", "hallway 2, second floor"},
	} {
		p, _ := h.FindPatient(a.id)
		_ = h.RegisterEpisode(p, a.location) // el 4.º devuelve ErrNoRoomAvailable
	}
	p1, _ := h.FindPatient("P-001")
	if err := h.WakeUpPatient(p1); err != nil {
		t.Fatal(err)
	}
	for _, p := range h.PatientsInHallway() {
		if _, err := h.AssignRoom(p); err != nil {
			t.Fatal(err)
		}
	}
	return queriesOf(h, h.Doctors())
}

// queryResult guarda las 4 consultas en una forma comparable (sin horas reales).
type queryResult struct {
	Hallway  []string
	Episodes map[string][]string // doctor -> "paciente|lugar|cuarto"
	Rooms    []string
	Severe   []string
}

func queriesOf(h *hospital.Hospital, doctors []*hospital.Doctor) queryResult {
	q := queryResult{Episodes: map[string][]string{}}
	for _, p := range h.PatientsInHallway() {
		q.Hallway = append(q.Hallway, p.ID()+"|"+p.Location())
	}
	for _, d := range doctors {
		q.Episodes[d.ID()] = nil
		for _, e := range d.MyEpisodes() {
			q.Episodes[d.ID()] = append(q.Episodes[d.ID()], fmt.Sprintf("%s|%s|%d", e.Patient().ID(), e.Location(), e.RoomNumber()))
		}
	}
	for _, r := range h.Rooms() {
		line := fmt.Sprintf("%d|%v", r.Number(), r.State())
		for _, p := range r.Occupants() {
			line += "|" + p.ID()
		}
		q.Rooms = append(q.Rooms, line)
	}
	for p, n := range h.SevereReport() {
		q.Severe = append(q.Severe, fmt.Sprintf("%s|%d", p.ID(), n))
	}
	slices.SortFunc(q.Severe, cmp.Compare[string])
	return q
}

func TestDemoEndStateMatchesTheConsoleScenario(t *testing.T) {
	want := consoleScenario(t)

	demo, err := NewDemo()
	if err != nil {
		t.Fatal(err)
	}
	for !demo.Done() {
		playPart(t, demo)
	}
	s := demo.Shift()
	var doctors []*hospital.Doctor
	for _, m := range s.Staff() {
		if m.IsDoctor() {
			doctors = append(doctors, m.Doctor())
		}
	}
	got := queriesOf(s.Hospital(), doctors)

	// El juego guarda la ubicación de cada cuarto como "room N" sin cambiar
	// nada: la consulta 5.1 debe estar vacía en los dos.
	if !slices.Equal(got.Hallway, want.Hallway) {
		t.Errorf("5.1 hallway = %v, want %v", got.Hallway, want.Hallway)
	}
	for id, wantEpisodes := range want.Episodes {
		if !slices.Equal(got.Episodes[id], wantEpisodes) {
			t.Errorf("5.2 %s = %v, want %v", id, got.Episodes[id], wantEpisodes)
		}
	}
	if !slices.Equal(got.Rooms, want.Rooms) {
		t.Errorf("5.3 rooms = %v, want %v", got.Rooms, want.Rooms)
	}
	if !slices.Equal(got.Severe, want.Severe) {
		t.Errorf("5.4 severe = %v, want %v", got.Severe, want.Severe)
	}

	// El reporte del juego trae lo mismo que las consultas.
	r := s.Report()
	if len(r.Hallway) != 0 || r.FreeRooms != 0 || len(r.Rooms) != 3 {
		t.Errorf("report: hallway=%d freeRooms=%d rooms=%d", len(r.Hallway), r.FreeRooms, len(r.Rooms))
	}
	if r.NoBedCount != 1 {
		t.Errorf("NoBedCount = %d, want 1", r.NoBedCount)
	}
	var episodes int
	for _, d := range r.Doctors {
		episodes += len(d.Episodes)
		for _, e := range d.Episodes {
			if e.GameTime == "" {
				t.Errorf("episode %s has no game time", e.ID)
			}
		}
	}
	if episodes != 3 {
		t.Errorf("doctors' episodes in the report = %d, want 3 (the orderly's one is not a doctor's)", episodes)
	}
}

func TestDemoSleepLocationsMapToCameras(t *testing.T) {
	demo, err := NewDemo()
	if err != nil {
		t.Fatal(err)
	}
	wantCam := map[string]CameraID{"P-001": "2", "P-002": "8", "P-003": "5", "P-004": "4"}
	// Se juegan las partes 2.1 a 2.4 y se mira dónde se durmió cada uno.
	for demo.Part() < 5 {
		playPart(t, demo)
	}
	s := demo.Shift()
	for id, cam := range wantCam {
		for _, e := range s.Hospital().History() {
			if e.Patient().ID() == id {
				if got := s.FindPatient(id).sleepLocation; got != e.Location() {
					t.Errorf("%s sleep location %q != episode location %q", id, got, e.Location())
				}
			}
		}
		// P-004 sigue en su cámara (sin cama); los demás ya están en habitaciones.
		if id == "P-004" && s.FindPatient(id).Camera() != cam {
			t.Errorf("P-004 camera = %s, want %s", s.FindPatient(id).Camera(), cam)
		}
	}
}

func TestSameSeedSameGame(t *testing.T) {
	run := func(seed int64) string {
		cfg, _ := ConfigForDay(1)
		s, err := NewShift(cfg, seed)
		if err != nil {
			t.Fatal(err)
		}
		advance(s, 4000)
		raw, err := json.Marshal(s.Snapshot())
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	// Se juega dos veces con la misma semilla: si la lógica es determinista,
	// las dos fotos (snapshots) en JSON tienen que ser idénticas.
	first, second := run(42), run(42)
	if first != second {
		t.Error("same seed gave different games")
	}
	if first == run(43) {
		t.Error("different seeds gave the same game")
	}
}

// autoPlay es un jugador automático sencillo para probar turnos completos.
func autoPlay(s *Shift) {
	for _, g := range s.Patients() {
		id := g.Patient().ID()
		switch g.State() {
		case AwakeInBed:
			_ = s.Discharge(id)
		case AsleepWaiting:
			for _, m := range s.Staff() {
				if m.Free() {
					_ = s.Assign(id, m.ID())
					break
				}
			}
		case NoBedWaiting:
			for _, m := range s.Staff() {
				if m.Free() {
					_ = s.MoveToBed(id, m.ID())
					break
				}
			}
		}
	}
}

// checkModelMatchesGame comprueba la tabla de la sección 3.1: el estado de
// juego de cada paciente y su estado en el modelo deben ser compatibles.
func checkModelMatchesGame(t *testing.T, s *Shift) {
	t.Helper()
	want := map[PatientState][]hospital.PatientState{
		Arriving:      {hospital.Awake},
		Wandering:     {hospital.Awake},
		AsleepWaiting: {hospital.Awake},
		NoBedWaiting:  {hospital.AsleepInHallway},
		Carried:       {hospital.AsleepInBed},
		InBed:         {hospital.AsleepInBed},
		AwakeInBed:    {hospital.AsleepInBed},
		Gone:          {hospital.Awake},
	}
	for _, g := range s.Patients() {
		got := s.Hospital().StateOf(g.Patient())
		if !slices.Contains(want[g.State()], got) {
			t.Fatalf("tick %d: %s game state %v but model state %v", s.Clock().Tick(), g.Patient().ID(), g.State(), got)
		}
	}
}

func TestFullShiftsAreConsistent(t *testing.T) {
	for day := 1; day <= MaxDay; day++ {
		t.Run(fmt.Sprintf("day %d", day), func(t *testing.T) {
			cfg, _ := ConfigForDay(day)
			s, err := NewShift(cfg, int64(day)*101)
			if err != nil {
				t.Fatal(err)
			}
			for tick := 0; tick < ShiftTicks; tick++ {
				s.Tick()
				if tick%10 == 0 {
					autoPlay(s)
				}
				if tick%97 == 0 {
					checkModelMatchesGame(t, s)
				}
			}
			if !s.Finished() {
				t.Fatal("shift did not finish at 18:00")
			}
			r := s.Report()
			if r.Stars < 1 || r.Stars > 5 {
				t.Errorf("stars = %d", r.Stars)
			}
			if len(r.Severe) < 2 {
				t.Errorf("severe report has %d patients", len(r.Severe))
			}
			if len(s.Patients()) != len(cfg.Patients) {
				t.Errorf("%d patients arrived, want %d", len(s.Patients()), len(cfg.Patients))
			}
			if _, err := json.Marshal(s.Snapshot()); err != nil {
				t.Errorf("snapshot JSON: %v", err)
			}
		})
	}
}

func TestSnapshotJSON(t *testing.T) {
	s := newDemoShift(t)
	mustSleepAndAssign(t, s, "P-001", "D-001")
	raw, err := json.Marshal(s.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, want := range []string{`"clock":"12:00"`, `"state":"Carried"`, `"hospitalState":"Asleep in bed"`, `"state":"Transferring"`, `"quota":4`} {
		if !strings.Contains(text, want) {
			t.Errorf("snapshot JSON lacks %s", want)
		}
	}
	snap := s.Snapshot()
	if len(snap.Patients) != 5 || len(snap.Staff) != 3 || len(snap.Rooms) != 3 {
		t.Errorf("snapshot sizes: %d patients, %d staff, %d rooms", len(snap.Patients), len(snap.Staff), len(snap.Rooms))
	}
}

func TestSaveStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "save.json")
	store := NewSaveStore(path)

	data, err := store.Load()
	if err != nil {
		t.Fatalf("Load of a missing file: %v", err)
	}
	if data.LastUnlockedDay != 1 || data.Settings != DefaultSettings() {
		t.Errorf("new data = %+v", data)
	}
	if got := DefaultSettings(); got.BlinkOnSleep || got.BlinkOnWake || got.MusicVolume != 70 || got.SfxVolume != 80 || got.Fullscreen {
		t.Errorf("default settings = %+v", got)
	}

	data.RecordShift(1, 3)
	data.RecordShift(1, 2) // no baja las estrellas
	data.Settings.Fullscreen = true
	if err := store.Save(data); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != data {
		t.Errorf("loaded %+v, want %+v", loaded, data)
	}
	if loaded.BestStars[0] != 3 || loaded.LastUnlockedDay != 2 {
		t.Errorf("progress = %+v", loaded)
	}
	data.RecordShift(MaxDay, 5)
	if data.LastUnlockedDay > MaxDay {
		t.Errorf("unlocked day %d beyond the last", data.LastUnlockedDay)
	}
}

func TestMoveRulesAfterDischarge(t *testing.T) {
	// Tras el alta no puede dormirse en al menos 10 min de juego.
	s := newCustomShift(t, 1, 2, 1, 3, 3)
	mustSleepAndAssign(t, s, "P-001", "O-001")
	advance(s, orderlyTransferTicks)
	_ = s.ForceWake("P-001")
	s.Tick()
	if err := s.Discharge("P-001"); err != nil {
		t.Fatal(err)
	}
	g := s.FindPatient("P-001")
	if g.State() != Wandering {
		t.Fatalf("state = %v", g.State())
	}
	if want := s.Clock().Tick() + MinutesToTicks(10); g.noSleepUntil != want {
		t.Errorf("noSleepUntil = %d, want %d", g.noSleepUntil, want)
	}
	// La habitación 101 (CAM 7A) sale al Pasillo Oeste (CAM 4).
	if cam := g.Camera(); cam != "4" {
		t.Errorf("left the room through %s, want 4", cam)
	}
}

// TestUnattendedAtEndSubtractPoints: al terminar el turno, cada paciente que
// sigue dormido sin atender RESTA puntos, y la resta crece con el día.
func TestUnattendedAtEndSubtractPoints(t *testing.T) {
	for _, day := range []int{1, 5} {
		s := newCustomShift(t, day, 2, 1, 3, 3)
		if err := s.ForceSleep("P-001", "2", "cafeteria (sancocho line)"); err != nil {
			t.Fatal(err)
		}
		advance(s, ShiftTicks)
		if !s.Finished() {
			t.Fatalf("day %d: shift did not finish", day)
		}
		want := -ScalePenalty(unattendedPenalty, day) // día 1: -10, día 5: -20
		if got := s.Score().Unattended; got != want {
			t.Errorf("day %d: Unattended = %d, want %d", day, got, want)
		}
	}
}

// La barra del panel crece durante el traslado, baja durante el descanso y
// queda en 0 cuando el personal está libre.
func TestCooldownFraction(t *testing.T) {
	s := newDemoShift(t)
	m := s.FindStaff("O-001")
	if got := m.CooldownFraction(); got != 0 {
		t.Fatalf("free staff fraction = %v, want 0", got)
	}
	mustSleepAndAssign(t, s, "P-001", "O-001")
	if got := m.CooldownFraction(); got != 0 {
		t.Errorf("fraction at the start of the transfer = %v, want 0", got)
	}
	advance(s, orderlyTransferTicks/2)
	if got := m.CooldownFraction(); got != 0.5 {
		t.Errorf("fraction halfway through the transfer = %v, want 0.5", got)
	}
	advance(s, orderlyTransferTicks/2)
	if got := m.CooldownFraction(); got != 1 {
		t.Errorf("fraction at the start of the rest = %v, want 1", got)
	}
	advance(s, orderlyRestTicks/2)
	if got := m.CooldownFraction(); got != 0.5 {
		t.Errorf("fraction halfway through the rest = %v, want 0.5", got)
	}
	advance(s, orderlyRestTicks/2)
	if got := m.CooldownFraction(); got != 0 {
		t.Errorf("fraction after the rest = %v, want 0", got)
	}
}

func TestSaveStoreExists(t *testing.T) {
	store := NewSaveStore(filepath.Join(t.TempDir(), "save.json"))
	if store.Exists() {
		t.Error("Exists before saving")
	}
	if err := store.Save(NewSaveData()); err != nil {
		t.Fatal(err)
	}
	if !store.Exists() {
		t.Error("Exists is false after saving")
	}
}

// El paciente en camilla sabe quién lo empuja; fuera de la camilla, nadie.
func TestCarrier(t *testing.T) {
	s := newDemoShift(t)
	if err := s.ForceSleep("P-001", "2", "cafeteria"); err != nil {
		t.Fatal(err)
	}
	if s.FindPatient("P-001").Carrier() != nil {
		t.Error("a sleeping patient has no carrier")
	}
	if err := s.Assign("P-001", "O-001"); err != nil {
		t.Fatal(err)
	}
	if got := s.FindPatient("P-001").Carrier(); got != s.FindStaff("O-001") {
		t.Errorf("carrier = %v, want O-001", got)
	}
	advance(s, orderlyTransferTicks)
	if s.FindPatient("P-001").Carrier() != nil {
		t.Error("after arriving in bed there is no carrier")
	}
}

// "Pacientes atendidos" cuenta personas distintas: si un paciente se duerme y
// llega a cama dos veces, es uno solo (las notas sí cuentan los dos episodios).
func TestAttendedCountsDistinctPatients(t *testing.T) {
	s := newDemoShift(t)
	g := s.FindPatient("P-001")
	s.markAttended(g)
	s.markAttended(g)
	s.markAttended(s.FindPatient("P-002"))
	if got := s.Report().Attended; got != 2 {
		t.Errorf("Attended = %d, want 2", got)
	}
}
