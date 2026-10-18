package game

import (
	"errors"
	"testing"
)

// playPart anima la parte siguiente de la Demo hasta el final y devuelve el
// evento que haya pasado en el camino.
func playPart(t *testing.T, d *Demo) DemoEvent {
	t.Helper()
	ev, err := d.Start()
	if err != nil {
		t.Fatalf("Start (part %d): %v", d.Part(), err)
	}
	last, err := d.Finish()
	if err != nil {
		t.Fatalf("Finish (part %d): %v", d.Part(), err)
	}
	if last.Kind != EventNone {
		ev = last
	}
	return ev
}

// patientAt busca un paciente en el Snapshot.
func patientAt(snap Snapshot, id string) PatientDump {
	for _, p := range snap.Patients {
		if p.ID == id {
			return p
		}
	}
	return PatientDump{}
}

// El reloj de la Demo NO corre solo: sin ninguna parte animándose, diez mil
// ticks de la Demo dejan la hora del turno exactamente igual.
func TestDemoClockDoesNotAdvanceWhileWaiting(t *testing.T) {
	d := mustNewDemo(t)
	check := func(label string) {
		t.Helper()
		before := d.Shift().Clock().Tick()
		for i := 0; i < 10000; i++ {
			if _, err := d.Tick(); err != nil {
				t.Fatal(err)
			}
		}
		if got := d.Shift().Clock().Tick(); got != before {
			t.Errorf("%s: the clock moved while waiting: %d -> %d", label, before, got)
		}
	}
	check("before the first part")
	for d.Part() < DemoParts {
		playPart(t, d)
		check("between parts")
	}
}

// El reloj SÍ corre mientras una parte se anima, y se detiene cuando acaba.
func TestDemoClockRunsOnlyWhilePlaying(t *testing.T) {
	d := mustNewDemo(t)
	if _, err := d.Start(); err != nil {
		t.Fatal(err)
	}
	start := d.Shift().Clock().Tick()
	for i := 0; i < 10; i++ {
		if _, err := d.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	if d.Shift().Clock().Tick() <= start {
		t.Error("the clock should run while a part plays")
	}
	if _, err := d.Finish(); err != nil {
		t.Fatal(err)
	}
	if d.Playing() {
		t.Error("Finish should leave the part finished")
	}
	end := d.Shift().Clock().Tick()
	d.Tick()
	if d.Shift().Clock().Tick() != end {
		t.Error("the clock moved after the part finished")
	}
}

// El turno nunca termina dentro de la Demo, ni jugando las 8 partes ni
// dejando pasar mucho tiempo.
func TestDemoShiftNeverFinishes(t *testing.T) {
	d := mustNewDemo(t)
	for d.Part() < DemoParts {
		playPart(t, d)
		if d.Shift().Finished() {
			t.Fatalf("the shift finished after part %d", d.Part())
		}
	}
	for i := 0; i < ShiftTicks*2; i++ {
		d.Tick()
	}
	snap := d.Shift().Snapshot()
	if snap.Over || snap.Tick >= ShiftTicks/2 {
		t.Errorf("end of the demo: over=%v at tick %d (the whole shift is %d)", snap.Over, snap.Tick, ShiftTicks)
	}
}

// Después de cada una de las 8 partes el turno queda en el estado esperado.
func TestDemoEachPartEndsInTheExpectedState(t *testing.T) {
	type want struct {
		state PatientState
		cam   CameraID
		room  int
	}
	tests := []struct {
		part     int
		patients map[string]want
		noBed    int
	}{
		{1, map[string]want{"P-001": {Wandering, "1A", 0}, "P-004": {Wandering, "1A", 0}}, 0},
		{2, map[string]want{"P-001": {InBed, "7A", 101}, "P-002": {Wandering, "1A", 0}}, 0},
		{3, map[string]want{"P-001": {InBed, "7A", 101}, "P-002": {InBed, "7B", 102}, "P-003": {Wandering, "1A", 0}}, 0},
		{4, map[string]want{"P-002": {InBed, "7B", 102}, "P-003": {InBed, "7C", 103}, "P-004": {Wandering, "1A", 0}}, 0},
		{5, map[string]want{"P-003": {InBed, "7C", 103}, "P-004": {NoBedWaiting, "4", 0}}, 1},
		{6, map[string]want{"P-001": {Wandering, "4", 0}, "P-004": {NoBedWaiting, "4", 0}}, 1},
		{7, map[string]want{"P-001": {Wandering, "4", 0}, "P-004": {InBed, "7A", 101}}, 1},
		{8, map[string]want{"P-004": {InBed, "7A", 101}, "P-005": {Wandering, "1A", 0}}, 1},
	}
	d := mustNewDemo(t)
	for _, tt := range tests {
		if tt.part > 1 {
			playPart(t, d)
		}
		if d.Part() != tt.part {
			t.Fatalf("Part() = %d, want %d", d.Part(), tt.part)
		}
		snap := d.Shift().Snapshot()
		for id, w := range tt.patients {
			p := patientAt(snap, id)
			if p.State != w.state || p.Camera != w.cam || p.Room != w.room {
				t.Errorf("part %d, %s: state=%v cam=%s room=%d, want %v %s %d", tt.part, id, p.State, p.Camera, p.Room, w.state, w.cam, w.room)
			}
		}
		if snap.NoBedCount != tt.noBed {
			t.Errorf("part %d: NoBedCount = %d, want %d", tt.part, snap.NoBedCount, tt.noBed)
		}
	}
	if !d.Done() {
		t.Error("the demo should be done after part 8")
	}
	if _, err := d.Start(); err == nil {
		t.Error("Start after the last part should fail")
	}
}

// Quién atiende a quién es lo mismo que en main.go, y en 2.4 la doctora
// descansa sin traslado mientras Wilfrido sigue en el piso.
func TestDemoStaffMatchesTheConsoleScenario(t *testing.T) {
	d := mustNewDemo(t)
	for d.Part() < 5 {
		playPart(t, d)
	}
	s := d.Shift()
	if got := s.FindStaff("D-001").State(); got != StaffResting {
		t.Errorf("Karen after 2.4 = %v, want Resting (no stretcher)", got)
	}
	attended := map[string]string{}
	for _, e := range s.Hospital().History() {
		attended[e.Patient().ID()] = e.Attender().ID()
	}
	want := map[string]string{"P-001": "D-001", "P-002": "D-002", "P-003": "O-001", "P-004": "D-001"}
	for p, staff := range want {
		if attended[p] != staff {
			t.Errorf("%s was attended by %q, want %q", p, attended[p], staff)
		}
	}
}

// Cada parte da el evento "sin cama" solo en la 2.4 (el aviso de la vista).
func TestDemoNoBedEventOnlyInPart24(t *testing.T) {
	d := mustNewDemo(t)
	for d.Part() < DemoParts {
		ev := playPart(t, d)
		wantEvent := d.Part() == 5
		if (ev.Kind == EventNoBed) != wantEvent {
			t.Errorf("part %d: event %v", d.Part(), ev)
		}
		if wantEvent && ev.Patient != "P-004" {
			t.Errorf("no-bed event patient = %q", ev.Patient)
		}
	}
}

// Una parte no se puede empezar si otra se está animando.
func TestDemoStartWhilePlayingFails(t *testing.T) {
	d := mustNewDemo(t)
	if _, err := d.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Start(); err == nil {
		t.Error("Start while a part is playing should fail")
	}
}

// Los beats de cámara: cada parte empieza en la cámara de la tabla.
func TestDemoCamerasFollowTheEvents(t *testing.T) {
	want := map[int][]CameraID{
		2: {"2", "7A"}, 3: {"8", "7B"}, 4: {"5", "7C"}, 5: {"4"}, 6: {"7A"}, 7: {"4", "7A"},
	}
	d := mustNewDemo(t)
	for d.Part() < 7 {
		if _, err := d.Start(); err != nil {
			t.Fatal(err)
		}
		var seen []CameraID
		for d.Playing() {
			if cam := d.Shift().CurrentCamera(); len(seen) == 0 || seen[len(seen)-1] != cam {
				seen = append(seen, cam)
			}
			if _, err := d.Tick(); err != nil {
				t.Fatal(err)
			}
		}
		if cam := d.Shift().CurrentCamera(); seen[len(seen)-1] != cam {
			seen = append(seen, cam)
		}
		w := want[d.Part()]
		if len(seen) != len(w) {
			t.Fatalf("part %d: cameras %v, want %v", d.Part(), seen, w)
		}
		for i := range w {
			if seen[i] != w[i] {
				t.Errorf("part %d: cameras %v, want %v", d.Part(), seen, w)
			}
		}
	}
}

func TestForceMove(t *testing.T) {
	s := newDemoShift(t)
	if err := s.ForceMove("P-001", "2"); err != nil {
		t.Fatal(err)
	}
	if got := s.FindPatient("P-001").Camera(); got != "2" {
		t.Errorf("camera = %s", got)
	}
	if err := s.ForceMove("P-001", "99"); !errors.Is(err, ErrUnknownCamera) {
		t.Errorf("unknown camera: %v", err)
	}
	if err := s.ForceMove("zz", "2"); !errors.Is(err, ErrUnknownPatient) {
		t.Errorf("unknown patient: %v", err)
	}
	s.ForceSleep("P-001", "2", "x")
	if err := s.ForceMove("P-001", "4"); !errors.Is(err, ErrWrongState) {
		t.Errorf("asleep patient: %v", err)
	}
}
