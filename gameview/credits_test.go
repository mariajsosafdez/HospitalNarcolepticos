package gameview

import (
	"strings"
	"testing"

	"costenos-narcolepsia/game"
)

func TestParadeX(t *testing.T) {
	tests := []struct {
		name     string
		x0, tick int
		want     int
	}{
		{"en el tick 0 está donde empieza", 40, 0, 40},
		{"un segundo: 16 px a la derecha", 40, 60, 56},
		{"medio segundo: 8 px", 40, 30, 48},
		{"sale por la derecha y vuelve por la izquierda", 540, 60 * 6, 540 + 96 - paradeSpan},
		{"justo al borde de la vuelta", 585, 0, 585},
		{"dos vueltas completas (4725 ticks) y queda igual", 40, 4725, 40},
	}
	for _, tt := range tests {
		if got := paradeX(tt.x0, tt.tick); got != tt.want {
			t.Errorf("%s: paradeX(%d,%d) = %d, quiero %d", tt.name, tt.x0, tt.tick, got, tt.want)
		}
	}
	// Siempre queda entre -44 y el final de la fila: nunca se pierde un personaje.
	for _, w := range parade {
		for tick := 0; tick < 60*120; tick += 7 {
			if x := paradeX(w.x0, tick); x < -paradeMargin || x >= paradeSpan-paradeMargin {
				t.Fatalf("%s en el tick %d: x = %d fuera de rango", w.sprite, tick, x)
			}
		}
	}
}

func TestParadeFrame(t *testing.T) {
	tests := []struct{ start, tick, want int }{
		{0, 0, 0}, {0, 8, 0}, {0, 9, 1}, {0, 27, 3}, {0, 36, 0},
		{3, 0, 3}, {3, 9, 0}, {2, 18, 0},
	}
	for _, tt := range tests {
		if got := paradeFrame(tt.start, tt.tick); got != tt.want {
			t.Errorf("paradeFrame(%d,%d) = %d, quiero %d", tt.start, tt.tick, got, tt.want)
		}
	}
}

func TestCreditsGoesBack(t *testing.T) {
	menu := &menuScene{selected: optCredits}
	tests := []struct {
		name string
		in   Input
		back bool
	}{
		{"sin entrada se queda", Input{keys: map[string]bool{}}, false},
		{"Esc vuelve", Input{keys: map[string]bool{"Escape": true}}, true},
		{"Enter vuelve", Input{keys: map[string]bool{"Enter": true}}, true},
		{"Espacio vuelve", Input{keys: map[string]bool{"Space": true}}, true},
		{"clic vuelve", Input{keys: map[string]bool{}, Click: true}, true},
		{"otra tecla no", Input{keys: map[string]bool{"Down": true}}, false},
	}
	for _, tt := range tests {
		c := newCredits(nil, menu)
		next := c.Update(&tt.in)
		if (next == Scene(menu)) != tt.back {
			t.Errorf("%s: Update devolvió %v", tt.name, next)
		}
	}
}

// spriteOwner saca la clave de una hoja del desfile: "pac_egg1_camina" -> "egg1".
func spriteOwner(sprite string) string {
	parts := strings.Split(sprite, "_")
	return parts[1]
}

// TestParadeHasWholeCast: el desfile trae a todo el elenco y a los easter eggs,
// cada uno una sola vez. Se arma desde las tablas del juego, así que si se
// agrega un personaje y se olvida aquí, el test falla.
func TestParadeHasWholeCast(t *testing.T) {
	want := map[string]string{} // clave del sprite -> nombre
	for day := 1; ; day++ {
		cfg, err := game.ConfigForDay(day)
		if err != nil {
			break
		}
		for _, d := range cfg.Doctors {
			want[spriteKey(d.ID, d.Name)] = d.Name
		}
		for _, o := range cfg.Orderlies {
			want[spriteKey(o.ID, o.Name)] = o.Name
		}
		for _, p := range cfg.Patients {
			want[spriteKey(p.ID, p.Name)] = p.Name
		}
	}
	for n := 1; ; n++ {
		egg, ok := game.EggByNumber(n)
		if !ok {
			break
		}
		want[spriteKey(egg.ID, egg.Name)] = egg.Name
	}
	if len(want) != 20 {
		t.Fatalf("el elenco tiene %d personajes, quiero 20", len(want))
	}

	seen := map[string]int{}
	for _, w := range parade {
		seen[spriteOwner(w.sprite)]++
	}
	for key, name := range want {
		if seen[key] != 1 {
			t.Errorf("%s (%s) sale %d veces en el desfile, quiero 1", name, key, seen[key])
		}
	}
	if len(seen) != len(want) {
		t.Errorf("el desfile tiene %d personajes distintos, quiero %d", len(seen), len(want))
	}
}

// TestParadeSpacing: el desfile está parejo y su largo coincide con paradeSpan.
func TestParadeSpacing(t *testing.T) {
	end := 0
	for i, w := range parade {
		if w.x0 != end {
			t.Errorf("%s: x0 = %d, quiero %d", w.sprite, w.x0, end)
		}
		if w.start < 0 || w.start >= paradeFrames {
			t.Errorf("%s: cuadro de arranque %d fuera de rango (puesto %d)", w.sprite, w.start, i)
		}
		if w.sprite == "cam_chepe_camilla" {
			end += paradeStretcher
		} else {
			end += paradeSlot
		}
	}
	if end != paradeSpan {
		t.Errorf("el desfile mide %d y paradeSpan vale %d", end, paradeSpan)
	}
}
