package game

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestDayIntro(t *testing.T) {
	if _, ok := DayIntro(1); ok {
		t.Error("day 1 opens with the tutorial, not with an intro")
	}
	if _, ok := DayIntro(MaxDay + 1); ok {
		t.Error("day 6 does not exist")
	}
	for day := 2; day <= MaxDay; day++ {
		text, ok := DayIntro(day)
		if !ok || text == "" {
			t.Errorf("day %d has no intro", day)
		}
	}
	if text, _ := DayIntro(3); !strings.Contains(text, "Luz Dary") {
		t.Errorf("day 3 intro = %q", text)
	}
	if len(TutorialLines) != 9 {
		t.Errorf("tutorial has %d lines, want 9", len(TutorialLines))
	}
}

func TestEndLinePicksFromTheRightList(t *testing.T) {
	tests := []struct {
		stars int
		list  []string
	}{
		{1, endLinesOneStar},
		{2, endLinesMiddle},
		{3, endLinesMiddle},
		{4, endLinesMiddle},
		{5, endLinesFiveStars},
	}
	for _, tt := range tests {
		for seed := int64(0); seed < 30; seed++ {
			line := EndLine(tt.stars, seed)
			if !slices.Contains(tt.list, line) {
				t.Fatalf("EndLine(%d, %d) = %q, not in its list", tt.stars, seed, line)
			}
			if EndLine(tt.stars, seed) != line {
				t.Fatalf("EndLine(%d, %d) is not deterministic", tt.stars, seed)
			}
		}
		// Con muchas semillas salen todas las frases de la lista.
		seen := map[string]bool{}
		for seed := int64(0); seed < 200; seed++ {
			seen[EndLine(tt.stars, seed)] = true
		}
		if len(seen) != len(tt.list) {
			t.Errorf("%d stars: %d different lines over 200 seeds, want %d", tt.stars, len(seen), len(tt.list))
		}
	}
}

func TestOverworkLine(t *testing.T) {
	got := OverworkLine("Dra. Luz Dary Fontalvo")
	want := "¡Oiga! Dra. Luz Dary Fontalvo ya tiene cuatro pacientes a cargo. Repártame el trabajo, que con este calor nadie aguanta tanto."
	if got != want {
		t.Errorf("OverworkLine = %q", got)
	}
}

func TestPointsPercent(t *testing.T) {
	tests := []struct {
		points, episodes, want int
	}{
		{0, 0, 0},     // sin episodios no hay nada que medir
		{350, 4, 73},  // 350 de 480
		{480, 4, 100}, // todo perfecto
		{600, 4, 100}, // nunca pasa de 100
		{-30, 4, 0},   // ni baja de 0
		{60, 1, 50},   // 60 de 120
		{1, 1, 1},     // 0,83 % redondea a 1
		{120, 10, 10}, // 120 de 1200
	}
	for _, tt := range tests {
		if got := PointsPercent(tt.points, tt.episodes); got != tt.want {
			t.Errorf("PointsPercent(%d, %d) = %d, want %d", tt.points, tt.episodes, got, tt.want)
		}
	}
	if PossiblePoints(4) != 480 {
		t.Errorf("PossiblePoints(4) = %d", PossiblePoints(4))
	}
}

func TestSaveProgressRules(t *testing.T) {
	data := NewSaveData()
	if data.HasProgress() {
		t.Error("a new save has no progress")
	}
	data.Settings.Fullscreen = true // cambiar ajustes no es progreso
	if data.HasProgress() {
		t.Error("settings alone are not progress")
	}
	data.RecordShift(1, 2)
	if !data.HasProgress() || data.LastUnlockedDay != 2 || data.BestStars[0] != 2 {
		t.Errorf("after day 1: %+v", data)
	}
	// Repetir un día con peores estrellas no baja nada ni retrocede el desbloqueo.
	data.RecordShift(1, 1)
	data.RecordShift(3, 4) // saltarse un día no desbloquea de más: solo el siguiente al jugado
	if data.BestStars[0] != 2 || data.BestStars[2] != 4 || data.LastUnlockedDay != 4 {
		t.Errorf("after replays: %+v", data)
	}
	// Un día fuera de rango no cambia nada.
	before := data
	data.RecordShift(0, 5)
	data.RecordShift(MaxDay+1, 5)
	if data != before {
		t.Errorf("out of range day changed the save: %+v", data)
	}
}

func TestTutorialSeenIsSaved(t *testing.T) {
	store := NewSaveStore(filepath.Join(t.TempDir(), "save.json"))
	data := NewSaveData()
	data.TutorialSeen = true
	if err := store.Save(data); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil || !loaded.TutorialSeen {
		t.Errorf("loaded %+v, err %v", loaded, err)
	}
}

// Juega los 5 días enteros con el jugador automático y guarda los resultados
// como lo hace la vista: se desbloquea el siguiente día y quedan las mejores
// estrellas.
func TestAutoPlayAllDaysAndSave(t *testing.T) {
	store := NewSaveStore(filepath.Join(t.TempDir(), "save.json"))
	data := NewSaveData()
	for day := 1; day <= MaxDay; day++ {
		cfg, err := ConfigForDay(day)
		if err != nil {
			t.Fatal(err)
		}
		s, err := NewShift(cfg, int64(day))
		if err != nil {
			t.Fatal(err)
		}
		actions := 0
		for tick := 0; tick < ShiftTicks+5; tick++ {
			s.Tick()
			actions += AutoStep(s)
		}
		if !s.Finished() {
			t.Fatalf("day %d did not finish", day)
		}
		if actions == 0 {
			t.Errorf("day %d: the auto player did nothing", day)
		}
		r := s.Report()
		if r.Day != day || r.Stars < 1 || r.Stars > 5 {
			t.Errorf("day %d report: day %d, stars %d", day, r.Day, r.Stars)
		}
		if r.Attended != r.Score.Attended() || r.Possible != PossiblePoints(r.Score.Episodes) {
			t.Errorf("day %d: attended %d possible %d", day, r.Attended, r.Possible)
		}
		data.RecordShift(day, r.Stars)
		if err := store.Save(data); err != nil {
			t.Fatal(err)
		}
		wantNext := min(day+1, MaxDay)
		if data.LastUnlockedDay != wantNext {
			t.Errorf("after day %d the last unlocked day is %d, want %d", day, data.LastUnlockedDay, wantNext)
		}
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	for day, stars := range loaded.BestStars {
		if stars < 1 || stars > 5 {
			t.Errorf("best stars of day %d = %d", day+1, stars)
		}
	}
	if loaded.LastUnlockedDay != MaxDay {
		t.Errorf("last unlocked day = %d, want %d", loaded.LastUnlockedDay, MaxDay)
	}
}

// El reporte de la Demo trae lo que muestra la vista del reporte: quién se
// quedó sin cama y a qué cama pasó después, y el estado de los severos.
func TestReportNoBedAndSevereDetails(t *testing.T) {
	d := mustNewDemo(t)
	for d.StepIndex() < 4 { // los cuatro sueños: el cuarto, Wilfrido, no tiene cama
		if _, err := d.Next(); err != nil {
			t.Fatal(err)
		}
	}
	r := d.Shift().Report()
	if len(r.NoBed) != 1 || r.NoBed[0].ID != "P-004" || r.NoBed[0].LaterRoom != 0 {
		t.Fatalf("NoBed before the discharge = %+v", r.NoBed)
	}
	for !d.Done() {
		if _, err := d.Next(); err != nil {
			t.Fatal(err)
		}
	}
	r = d.Shift().Report()
	if len(r.NoBed) != 1 || r.NoBed[0].LaterRoom != 101 {
		t.Errorf("NoBed after the move = %+v, want a later room 101", r.NoBed)
	}
	if r.Attended != 4 || r.Percent < 0 || r.Percent > 100 {
		t.Errorf("attended %d, percent %d", r.Attended, r.Percent)
	}
	rooms := map[string]int{}
	for _, sv := range r.Severe {
		rooms[sv.ID] = sv.Room
	}
	if rooms["P-002"] != 102 || rooms["P-004"] != 101 || rooms["P-005"] != 0 {
		t.Errorf("severe rooms = %v", rooms)
	}
	for _, sv := range r.Severe {
		if sv.Asleep {
			t.Errorf("%s should not be asleep at the end of the Demo", sv.ID)
		}
	}
}

func mustNewDemo(t *testing.T) *Demo {
	t.Helper()
	d, err := NewDemo()
	if err != nil {
		t.Fatal(err)
	}
	return d
}
