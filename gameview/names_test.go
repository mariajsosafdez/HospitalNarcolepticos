package gameview

import (
	"testing"

	"costenos-narcolepsia/game"
)

func TestDisplayStaffText(t *testing.T) {
	tests := []struct {
		id, name, specialty string
		wantName, wantSpec  string
	}{
		{"D-001", "Dr. Karen Ospina", "Neurology", "Dra. Karen Ospina", "Neurología"},
		{"D-002", "Dr. Alvaro Mendoza", "Sleep Medicine", "Dr. Álvaro Mendoza", "Medicina del sueño"},
		{"D-003", "Dr. Luz Dary Fontalvo", "General", "Dra. Luz Dary Fontalvo", "Medicina general"},
		{"O-001", "Chepe Barrios", "", "Chepe Barrios", "Camillero"},
		{"X-001", "Alguien Nuevo", "", "Alguien Nuevo", ""},
	}
	for _, tt := range tests {
		if got := displayName(tt.id, tt.name); got != tt.wantName {
			t.Errorf("displayName(%s) = %q, want %q", tt.id, got, tt.wantName)
		}
		if got := displaySpecialty(tt.id, tt.specialty); got != tt.wantSpec {
			t.Errorf("displaySpecialty(%s) = %q, want %q", tt.id, got, tt.wantSpec)
		}
	}
}

func TestPatientLabelName(t *testing.T) {
	tests := []struct{ id, full, want string }{
		{"P-001", "Yesid Pacheco", "Yesid Pacheco"},
		{"P-007", `Abelardo "Mañe" Julio`, "Mañe Julio"},
		{"P-008", "Kelly Johana Ruiz", "Kelly Ruiz"},
		{"P-099", "Juan Carlos Pérez Gómez", "Juan Gómez"},
		{"P-100", "Solo", "Solo"},
	}
	for _, tt := range tests {
		if got := patientLabelName(tt.id, tt.full); got != tt.want {
			t.Errorf("patientLabelName(%s) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

// Los diez pacientes del elenco caben enteros en la etiqueta (60 px).
func TestCastLabelNamesFit(t *testing.T) {
	fonts, err := LoadFonts()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := game.ConfigForDay(5)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range cfg.Patients {
		name := patientLabelName(p.ID, p.Name)
		if w := fonts.Width(name, fontPixel, 8); w > 60 {
			t.Errorf("%s: %q mide %d px", p.ID, name, w)
		}
	}
}
