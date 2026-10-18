package gameview

import "testing"

func TestWindowToGame(t *testing.T) {
	tests := []struct {
		wx, wy, scale int
		x, y          int
	}{
		{0, 0, 1, 0, 0},
		{479, 269, 1, 479, 269},
		{0, 0, 2, 0, 0},
		{1, 1, 2, 0, 0},
		{2, 3, 2, 1, 1},
		{959, 539, 2, 479, 269},
		{100, 92, 2, 50, 46}, // el bug: el cursor llegaba al doble y caía en otra opción
		{3, 5, 3, 1, 1},
		{1439, 809, 3, 479, 269},
		{7, 7, 4, 1, 1},
		{1919, 1079, 4, 479, 269},
		{-1, -1, 2, -1, -1}, // fuera de la ventana: baja hacia abajo, no hacia 0
		{-3, 0, 3, -1, 0},
		{10, 10, 0, 10, 10}, // una escala inválida se trata como 1
	}
	for _, tt := range tests {
		x, y := windowToGame(tt.wx, tt.wy, tt.scale)
		if x != tt.x || y != tt.y {
			t.Errorf("windowToGame(%d,%d,%d) = (%d,%d), want (%d,%d)", tt.wx, tt.wy, tt.scale, x, y, tt.x, tt.y)
		}
	}
}

// Todo píxel del juego, pasado a ventana y de vuelta, es el mismo píxel, en
// cualquier escala.
func TestGameToWindowRoundTrip(t *testing.T) {
	for scale := 1; scale <= 4; scale++ {
		for _, p := range [][2]int{{0, 0}, {1, 1}, {100, 92}, {479, 269}, {26, 85}} {
			wx, wy := gameToWindow(p[0], p[1], scale)
			x, y := windowToGame(wx, wy, scale)
			if x != p[0] || y != p[1] {
				t.Errorf("scale %d: (%d,%d) -> window (%d,%d) -> (%d,%d)", scale, p[0], p[1], wx, wy, x, y)
			}
		}
	}
}

func TestLayoutScale(t *testing.T) {
	tests := []struct{ w, h, want int }{
		{960, 540, 2},
		{480, 270, 1},
		{100, 100, 1},   // ventana chiquita: mínimo 1
		{1920, 1080, 4}, // pantalla completa
		{1920, 1010, 3}, // maximizada: manda el alto
		{1366, 768, 2},
		{1439, 810, 2}, // a un píxel de 3x
		{1440, 810, 3},
		{1501, 899, 3}, // tamaño impar
		{0, 0, 1},
	}
	for _, tt := range tests {
		if got := layoutScale(tt.w, tt.h); got != tt.want {
			t.Errorf("layoutScale(%d,%d) = %d, want %d", tt.w, tt.h, got, tt.want)
		}
	}
}

// El cursor de ventana de cada escala cae en la opción de menú que se está
// apuntando (el bug de "apunto Nuevo turno y señala Créditos").
func TestMenuHoverAtEveryScale(t *testing.T) {
	for scale := 1; scale <= 4; scale++ {
		for i := range menuItems {
			m := newMenu(&Env{HasSave: true})
			var in Input
			in.reset()
			r := optionRect(i)
			in.moveToWindow((r.x+r.w/2)*scale+1, (r.y+r.h/2)*scale+1, scale)
			m.Update(&in)
			if m.hovered != menuItems[i].label {
				t.Errorf("scale %d: over option %d hovered %q, want %q", scale, i, m.hovered, menuItems[i].label)
			}
			if m.selected != i {
				t.Errorf("scale %d: over option %d selected %d", scale, i, m.selected)
			}
		}
	}
}
