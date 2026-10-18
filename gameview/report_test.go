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
	if attendedText(1) != "1 paciente atendido" || attendedText(4) != "4 pacientes atendidos" || attendedText(0) != "0 pacientes atendidos" {
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

	noBed := noBedLines(r)
	want := []string{"Al cerrar el turno: nadie.", "Durante el turno: 1 vez.", "Wilfrido Berrio, Pasillo Oeste", "(después pasó a la 101)"}
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

func TestNoBedStillInHallway(t *testing.T) {
	r := game.Report{
		Hallway: []game.PatientLine{{ID: "P-004", Name: "Wilfrido Berrio", Location: "Pasillo Oeste"}},
		NoBed:   []game.NoBedLine{{ID: "P-004", Name: "Wilfrido Berrio", Location: "Pasillo Oeste"}},
	}
	lines := noBedLines(r)
	if lines[0].left != "Al cerrar el turno: 1 paciente." || len(lines) != 3 {
		t.Errorf("noBedLines = %+v", lines)
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
