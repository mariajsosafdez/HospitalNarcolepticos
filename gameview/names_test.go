package gameview

import "testing"

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
