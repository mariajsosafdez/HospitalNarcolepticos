package gameview

import (
	"io/fs"
	"testing"

	"costenos-narcolepsia/art"
	"costenos-narcolepsia/game"
)

// Todo el elenco tiene que tener sus sprites en los PNG embebidos: si falta
// uno, este test falla en vez de que la pantalla muestre un rectángulo de
// relleno (el bug de Abelardo "Mañe", que buscaba pac_abelardo_*).
func TestEveryCastMemberHasSprites(t *testing.T) {
	cfg, err := game.ConfigForDay(game.MaxDay) // el día 5 trae a todos
	if err != nil {
		t.Fatal(err)
	}
	has := func(name string) bool {
		_, err := fs.Stat(art.Files, "png/"+name+".png")
		return err == nil
	}
	for _, p := range cfg.Patients {
		key := spriteKey(p.ID, p.Name)
		for _, suffix := range []string{"camina", "quieto", "dormido", "cama", "sentado"} {
			if name := "pac_" + key + "_" + suffix; !has(name) {
				t.Errorf("%s (%s): missing %s.png", p.Name, p.ID, name)
			}
		}
	}
	for _, d := range cfg.Doctors {
		key := spriteKey(d.ID, d.Name)
		for _, suffix := range []string{"camina", "quieto", "camilla"} {
			if name := "doc_" + key + "_" + suffix; !has(name) {
				t.Errorf("%s (%s): missing %s.png", d.Name, d.ID, name)
			}
		}
	}
	for _, o := range cfg.Orderlies {
		key := spriteKey(o.ID, o.Name)
		for _, suffix := range []string{"camina", "quieto", "camilla"} {
			if name := "cam_" + key + "_" + suffix; !has(name) {
				t.Errorf("%s (%s): missing %s.png", o.Name, o.ID, name)
			}
		}
	}
}

func TestSpriteKey(t *testing.T) {
	tests := []struct{ id, name, want string }{
		{"P-007", `Abelardo "Mañe" Julio`, "mane"},
		{"D-003", "Dra. Luz Dary Fontalvo", "luzdary"},
		{"O-002", `Jhon Jairo "El Mono" Arrieta`, "mono"},
		{"D-001", "Dr. Karen Ospina", "karen"}, // el nombre en inglés de la Demo
		{"P-099", "Nuevo Paciente", "nuevo"},   // sin tabla: primera palabra
	}
	for _, tt := range tests {
		if got := spriteKey(tt.id, tt.name); got != tt.want {
			t.Errorf("spriteKey(%s, %q) = %q, want %q", tt.id, tt.name, got, tt.want)
		}
	}
	if !isFemale("Kelly Johana Ruiz") || isFemale(`Abelardo "Mañe" Julio`) {
		t.Error("isFemale broke")
	}
}
