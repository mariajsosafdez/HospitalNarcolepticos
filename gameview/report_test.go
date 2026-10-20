package gameview

import (
	"strings"
	"testing"

	"costenos-narcolepsia/game"
)

// demoReport juega la Demo completa y devuelve su reporte.
func demoReport(t *testing.T) game.Report {
	t.Helper()
	d, err := game.NewDemo()
	if err != nil {
		t.Fatal(err)
	}
	for !d.Done() {
		if _, err := d.Start(); err != nil {
			t.Fatal(err)
		}
		if _, err := d.Finish(); err != nil {
			t.Fatal(err)
		}
	}
	return d.Shift().Report()
}

func TestReportTexts(t *testing.T) {
	if got := signedPoints(200); got != "+200" {
		t.Errorf("signedPoints(200) = %q", got)
	}
	if signedPoints(-4) != "-4" || signedPoints(0) != "0" {
		t.Error("signedPoints with negatives or zero")
	}
	if attendedText(1) != "1 paciente atendido" || attendedText(4) != "4 pacientes atendidos" || attendedText(0) != "Nadie fue atendido" {
		t.Error("attendedText")
	}
	if got := reportSubtitle(1, false, 4); got != "Día 1  ·  de 12 a 6  ·  4 pacientes atendidos" {
		t.Errorf("subtitle = %q", got)
	}
	if got := reportSubtitle(1, true, 4); !strings.HasPrefix(got, "Demo  ·") {
		t.Errorf("demo subtitle = %q", got)
	}
	if countText(1, "vez", "veces") != "1 vez" || countText(0, "vez", "veces") != "0 veces" || countText(2, "vez", "veces") != "2 veces" {
		t.Error("countText")
	}
}

func TestCapWithMore(t *testing.T) {
	var lines []rline
	for i := 0; i < 7; i++ {
		lines = append(lines, rline{left: "x"})
	}
	got := capWithMore(lines, 5)
	if len(got) != 5 || got[4].left != "y 3 más" {
		t.Errorf("capWithMore = %d lines, last %q", len(got), got[len(got)-1].left)
	}
	if len(capWithMore(lines, 7)) != 7 || len(capWithMore(lines[:2], 5)) != 2 {
		t.Error("a short list must not change")
	}
}

func TestDemoReportCards(t *testing.T) {
	r := demoReport(t)

	noBed := noBedLines(r, testMeasure)
	want := []string{"Al cerrar el turno: nadie.", "Durante el turno: 1 vez.", "Atendidos, sin cama (5.1): 0", "Sin atender: 0", "Wilfrido Berrio, Pasillo Oeste", "(después pasó a la 101)"}
	if len(noBed) != len(want) {
		t.Fatalf("noBedLines = %+v", noBed)
	}
	for i, w := range want {
		if noBed[i].left != w {
			t.Errorf("noBed line %d = %q, want %q", i, noBed[i].left, w)
		}
	}

	docs := doctorLines(r)
	if len(docs) != 4 || docs[0].left != "Dra. Karen Ospina" || docs[0].right != "2" || docs[1].left != "Yesid, Wilfrido" || !docs[1].indent {
		t.Errorf("doctorLines = %+v", docs)
	}
	if docs[2].left != "Dr. Álvaro Mendoza" || docs[2].right != "1" || docs[3].left != "Yeimy" {
		t.Errorf("doctorLines second doctor = %+v", docs[2:])
	}

	cells := bedCells(r)
	if len(cells) != 3 || cells[0].number != 101 || cells[0].who != "Wilfrido Berrio" || cells[1].who != "Yeimy Padilla" || cells[2].who != "Rafa Cantillo" {
		t.Errorf("bedCells = %+v", cells)
	}

	rows, footer := severeLines(r)
	if len(rows) != 3 || footer != "3 severos, 2 atendidos por doctor" {
		t.Errorf("severeLines = %+v, %q", rows, footer)
	}
	got := map[string]string{}
	for _, row := range rows {
		got[row.left] = row.right
	}
	if got["Yeimy Padilla"] != "cama 102" || got["Wilfrido Berrio"] != "cama 101" || got["Marelvis Ortega"] != "despierta" {
		t.Errorf("severe statuses = %v", got)
	}
}

func TestSevereStatus(t *testing.T) {
	tests := []struct {
		s    game.SevereLine
		want string
	}{
		{game.SevereLine{Name: "Yeimy Padilla", Room: 102}, "cama 102"},
		{game.SevereLine{Name: "Yeimy Padilla", Asleep: true}, "dormida"},
		{game.SevereLine{Name: "Wilfrido Berrio", Asleep: true}, "dormido"},
		{game.SevereLine{Name: "Marelvis Ortega"}, "despierta"},
		{game.SevereLine{Name: "Wilfrido Berrio"}, "despierto"},
	}
	for _, tt := range tests {
		if got := severeStatus(tt.s); got != tt.want {
			t.Errorf("severeStatus(%+v) = %q, want %q", tt.s, got, tt.want)
		}
	}
}

// testMeasure mide 5 px por letra, basta para probar.
func testMeasure(s string) int { return len(s) * 5 }

func TestNoBedLinesGroups(t *testing.T) {
	p := func(name string) game.PatientLine { return game.PatientLine{Name: name} }
	tests := []struct {
		name string
		r    game.Report
		want []string
	}{
		{"nobody", game.Report{}, []string{"Al cerrar el turno: nadie.", "Durante el turno: 0 veces.", "Atendidos, sin cama (5.1): 0", "Sin atender: 0"}},
		{"only unattended counts in the header", game.Report{Waiting: []game.PatientLine{p("Luz Dary Pérez")}},
			[]string{"Al cerrar el turno: 1 paciente.", "Durante el turno: 0 veces.", "Atendidos, sin cama (5.1): 0", "Sin atender: 1", "Luz"}},
		{"both groups add up", game.Report{
			Hallway: []game.PatientLine{p("Wilfrido Berrio")},
			Waiting: []game.PatientLine{p("Yeimy Padilla"), p("Rafa Cantillo")},
			NoBed:   []game.NoBedLine{{Name: "Wilfrido Berrio", Location: "Cafetería"}}},
			[]string{"Al cerrar el turno: 3 pacientes.", "Durante el turno: 1 vez.", "Atendidos, sin cama (5.1): 1", "Wilfrido", "Sin atender: 2", "Yeimy, Rafa"}},
	}
	for _, tt := range tests {
		got := noBedLines(tt.r, testMeasure)
		if len(got) < len(tt.want) {
			t.Errorf("%s: %d lines, want at least %d: %+v", tt.name, len(got), len(tt.want), got)
			continue
		}
		for i, w := range tt.want {
			if got[i].left != w {
				t.Errorf("%s: line %d = %q, want %q", tt.name, i, got[i].left, w)
			}
		}
	}
}

func TestNoBedLinesOverflow(t *testing.T) {
	var waiting []game.PatientLine
	for _, n := range []string{"Aaaaaaaaaaaa Z", "Bbbbbbbbbbbb Z", "Cccccccccccc Z", "Dddddddddddd Z"} {
		waiting = append(waiting, game.PatientLine{Name: n})
	}
	got := noBedLines(game.Report{Waiting: waiting}, testMeasure)
	if len(got) > reportNoBedLines {
		t.Fatalf("%d lines, max %d", len(got), reportNoBedLines)
	}
	if names := got[len(got)-1].left; !strings.Contains(names, "más") {
		t.Errorf("names line %q should end with 'y N más'", names)
	}
}

func TestLocationText(t *testing.T) {
	tests := []struct{ in, want string }{
		{"cafeteria (sancocho line)", "Cafetería"},
		{"champeta dance hall", "Salón de champeta"},
		{"radiology queue", "Fila de radiología"},
		{"hallway 2, second floor", "Pasillo Oeste"},
		{"Patio de las palmeras", "Patio de las palmeras"}, // turno normal: ya está en español
		{"", ""},
	}
	for _, tt := range tests {
		if got := locationText(tt.in); got != tt.want {
			t.Errorf("locationText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMultiplierText(t *testing.T) {
	tests := map[int]string{100: "×1", 125: "×1,25", 150: "×1,5", 175: "×1,75", 200: "×2"}
	for percent, want := range tests {
		if got := multiplierText(percent); got != want {
			t.Errorf("multiplierText(%d) = %q, want %q", percent, got, want)
		}
	}
}

func TestUnattendedText(t *testing.T) {
	measure := func(s string) int { return len(s) * 5 } // 5 px por letra, basta para probar
	tests := []struct {
		names []string
		maxW  int
		want  string
	}{
		{nil, 100, ""},
		{[]string{"Yesid Mendoza"}, 100, "Yesid"},
		{[]string{"Yesid Mendoza", "Wilfrido Berrio"}, 100, "Yesid, Wilfrido"},
		{[]string{"Yesid Mendoza", "Wilfrido Berrio", "Karina Pérez"}, 100, "Yesid y 2 más"},
		{[]string{"Yesid Mendoza", "Wilfrido Berrio", "Karina Pérez"}, 200, "Yesid, Wilfrido, Karina"},
	}
	for _, tt := range tests {
		if got := unattendedText(tt.names, tt.maxW, measure); got != tt.want {
			t.Errorf("unattendedText(%v, %d) = %q, want %q", tt.names, tt.maxW, got, tt.want)
		}
	}
}

// Los eventos "sin cama durante el turno" van al final, con el espacio que
// sobre, y el "y N más" cuenta personas.
func TestNoBedLinesEventsCap(t *testing.T) {
	ev := func(room int) game.NoBedLine {
		return game.NoBedLine{Name: "Ana", Location: "Cafetería", LaterRoom: room}
	}
	tests := []struct {
		name   string
		noBed  []game.NoBedLine
		length int    // líneas en total (4 de base: sin nadie)
		last   string // última línea
	}{
		{"fits", []game.NoBedLine{ev(101)}, 6, "(después pasó a la 101)"},
		{"two do not fit", []game.NoBedLine{ev(0), ev(0), ev(0)}, 6, "y 2 más"},
		{"four with rooms", []game.NoBedLine{ev(101), ev(102), ev(103), ev(104)}, 5, "y 4 más"},
	}
	for _, tt := range tests {
		got := noBedLines(game.Report{NoBed: tt.noBed}, testMeasure)
		if len(got) != tt.length || got[len(got)-1].left != tt.last {
			t.Errorf("%s: %d lines, last %q; want %d, %q", tt.name, len(got), got[len(got)-1].left, tt.length, tt.last)
		}
	}
}
