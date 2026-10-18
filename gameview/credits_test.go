package gameview

import "testing"

func TestParadeX(t *testing.T) {
	tests := []struct {
		name     string
		x0, tick int
		want     int
	}{
		{"en el tick 0 está donde empieza", 40, 0, 40},
		{"un segundo: 16 px a la derecha", 40, 60, 56},
		{"medio segundo: 8 px", 40, 30, 48},
		{"sale por la derecha y vuelve por la izquierda", 384, 60 * 6, 384 + 96 - 524},
		{"justo al borde de la vuelta", 436, 0, 436},
	}
	for _, tt := range tests {
		if got := paradeX(tt.x0, tt.tick); got != tt.want {
			t.Errorf("%s: paradeX(%d,%d) = %d, quiero %d", tt.name, tt.x0, tt.tick, got, tt.want)
		}
	}
	// Siempre queda entre -44 y 479: nunca se pierde un personaje.
	for _, w := range parade {
		for tick := 0; tick < 60*120; tick += 7 {
			if x := paradeX(w.x0, tick); x < -paradeMargin || x >= screenW {
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
