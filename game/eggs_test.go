package game

import (
	"testing"

	"costenos-narcolepsia/hospital"
)

// RollEgg con semillas fijas: unas traen easter egg y otras no, y siempre es
// el mismo resultado para la misma semilla.
func TestRollEggFixedSeeds(t *testing.T) {
	var with, without []int64
	for seed := int64(1); seed <= 40; seed++ {
		if _, ok := RollEgg(seed); ok {
			with = append(with, seed)
		} else {
			without = append(without, seed)
		}
	}
	if len(with) == 0 || len(without) == 0 {
		t.Fatalf("seeds 1-40 should mix: with=%v without=%v", with, without)
	}
	for _, seed := range []int64{with[0], without[0]} {
		a, okA := RollEgg(seed)
		b, okB := RollEgg(seed)
		if a != b || okA != okB {
			t.Errorf("seed %d is not deterministic", seed)
		}
	}
}

// Con muchas semillas la frecuencia se acerca al 15 %.
func TestRollEggFrequency(t *testing.T) {
	const runs = 20000
	count := 0
	seen := map[string]bool{}
	for seed := int64(0); seed < runs; seed++ {
		if egg, ok := RollEgg(seed); ok {
			count++
			seen[egg.ID] = true
		}
	}
	rate := float64(count) / runs
	if rate < 0.13 || rate > 0.17 {
		t.Errorf("egg rate = %.3f, want about %.2f", rate, eggChance)
	}
	if len(seen) != len(castEggs) {
		t.Errorf("only %d of %d eggs ever came up", len(seen), len(castEggs))
	}
}

// Un turno recibe como mucho UN easter egg (WithEgg suma uno y el sorteo
// devuelve uno), y la configuración original no se modifica.
func TestWithEggAddsExactlyOne(t *testing.T) {
	cfg, err := ConfigForDay(3)
	if err != nil {
		t.Fatal(err)
	}
	before := len(cfg.Patients)
	egg, _ := EggByNumber(2)
	got := cfg.WithEgg(egg)
	if len(got.Patients) != before+1 || len(cfg.Patients) != before {
		t.Errorf("patients: original %d, with egg %d, want %d and %d", len(cfg.Patients), len(got.Patients), before, before+1)
	}
	eggs := 0
	for _, p := range got.Patients {
		if IsEgg(p.ID) {
			eggs++
		}
	}
	if eggs != 1 {
		t.Errorf("%d eggs in the shift, want 1", eggs)
	}
}

func TestEggCastIsValid(t *testing.T) {
	want := []struct {
		id, name string
		level    hospital.NarcolepsyLevel
	}{
		{"E-001", "Misifu", hospital.Severe},
		{"E-002", "Alejo Urrego", hospital.Mild},
		{"E-003", "Samu Giraldo", hospital.Mild},
		{"E-004", "Cote", hospital.Severe},
		{"E-005", "Sebs", hospital.Moderate},
	}
	for i, w := range want {
		egg, ok := EggByNumber(i + 1)
		if !ok || egg.ID != w.id || egg.Name != w.name || egg.Level != w.level {
			t.Errorf("egg %d = %+v, want %+v", i+1, egg, w)
		}
		// El modelo (que no se toca) tiene que aceptar al paciente.
		if _, err := hospital.NewPatient(egg.ID, egg.Name, egg.Age, egg.Level); err != nil {
			t.Errorf("%s: NewPatient: %v", egg.Name, err)
		}
	}
	if _, ok := EggByNumber(0); ok {
		t.Error("egg 0 should not exist")
	}
	if _, ok := EggByNumber(6); ok {
		t.Error("egg 6 should not exist")
	}
}

// Misifu y Cote se duermen 3 veces más seguido que cualquier severo.
func TestFrequentSleepers(t *testing.T) {
	tests := []struct {
		id    string
		level hospital.NarcolepsyLevel
		cam   CameraID
		want  float64
	}{
		{CatID, hospital.Severe, "1B", 0.30},
		{CoteID, hospital.Severe, "1B", 0.30},
		{CoteID, hospital.Severe, "2", 0.60}, // cafetería x2
		{CatID, hospital.Severe, "2", 0.60},
		{"P-002", hospital.Severe, "1B", 0.10}, // un severo normal no cambia
		{"E-002", hospital.Mild, "1B", 0.03},
		{"E-005", hospital.Moderate, "1B", 0.06},
	}
	for _, tt := range tests {
		got := sleepChanceFor(tt.id, tt.level, tt.cam)
		if diff := got - tt.want; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("sleepChanceFor(%s, %v, %s) = %v, want %v", tt.id, tt.level, tt.cam, got, tt.want)
		}
	}
	if got := noSleepMinutesAfterDischarge(CatID); got != 0 {
		t.Errorf("cat cooldown = %d, want 0", got)
	}
	if got := noSleepMinutesAfterDischarge(CoteID); got != 0 {
		t.Errorf("Cote cooldown = %d, want 0", got)
	}
	if got := noSleepMinutesAfterDischarge("P-001"); got != 10 {
		t.Errorf("normal cooldown = %d, want 10", got)
	}
}

// Después del alta, Cote puede volver a dormirse de una vez; Yesid, no.
// Y en los últimos 20 min nadie se duerme, tampoco los easter eggs.
func TestDischargeCooldownForEggs(t *testing.T) {
	cfg := DemoConfig()
	cfg.Patients = append(cfg.Patients[:1:1], castEggs[3]) // Yesid y Cote
	cfg.ArriveAtStart = CoteID
	s, err := NewShift(cfg, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"P-001", CoteID} {
		if err := s.ForceSleep(id, "1B", "waiting room"); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Assign("P-001", "D-001"); err != nil {
		t.Fatal(err)
	}
	if err := s.Assign(CoteID, "O-001"); err != nil {
		t.Fatal(err)
	}
	advance(s, doctorTransferTicks)
	for _, id := range []string{"P-001", CoteID} {
		_ = s.ForceWake(id)
	}
	s.Tick()
	for _, id := range []string{"P-001", CoteID} {
		if err := s.Discharge(id); err != nil {
			t.Fatal(err)
		}
	}
	now := s.Clock().Tick()
	if g := s.FindPatient("P-001"); g.noSleepUntil != now+MinutesToTicks(10) {
		t.Errorf("Yesid noSleepUntil = %d, want %d", g.noSleepUntil, now+MinutesToTicks(10))
	}
	if g := s.FindPatient(CoteID); g.noSleepUntil != now {
		t.Errorf("Cote noSleepUntil = %d, want %d (no wait)", g.noSleepUntil, now)
	}
	if sleepWindowOpen(ShiftTicks - 1) {
		t.Error("final 20 minutes must stay closed for everyone")
	}
}

// ArriveAtStart y ArrivedIDs: el gato llega en el tick 0 y la vista se entera.
func TestArrivedIDsAndArriveAtStart(t *testing.T) {
	cfg, _ := ConfigForDay(1)
	cfg = cfg.WithEgg(castEggs[0])
	cfg.ArriveAtStart = CatID
	s, err := NewShift(cfg, 7)
	if err != nil {
		t.Fatal(err)
	}
	ids := s.ArrivedIDs()
	if len(ids) == 0 || ids[0] != CatID {
		t.Fatalf("arrived at tick 0 = %v, want the cat first", ids)
	}
	if s.FindPatient(CatID).Patient().Age() != 6 {
		t.Error("the cat should be 6")
	}
}
