package gameview

import (
	"math"
	"testing"
)

func TestWindowToGame(t *testing.T) {
	tests := []struct {
		wx, wy int
		scale  float64
		x, y   int
	}{
		{0, 0, 1, 0, 0},
		{479, 269, 1, 479, 269},
		{1, 1, 2, 0, 0},
		{2, 3, 2, 1, 1},
		{959, 539, 2, 479, 269},
		{100, 92, 2, 50, 46}, // el bug: el cursor llegaba al doble y caía en otra opción
		{3, 5, 3, 1, 1},
		{1919, 1079, 4, 479, 269},
		{-1, -1, 2, -1, -1}, // fuera de la ventana: baja hacia abajo, no hacia 0
		{-3, 0, 3, -1, 0},
		{10, 10, 0, 10, 10}, // una escala inválida se trata como 1
		// escalas fraccionarias
		{2, 2, 1.5, 1, 1},
		{3, 3, 1.5, 2, 2},
		{719, 404, 1.5, 479, 269},
		{1919, 1079, 4, 479, 269},
		{1919, 1079, 1920.0 / 480, 479, 269},
		{1079, 607, 2.25, 479, 269},
		{1, 1, 2.25, 0, 0},
		{3, 3, 2.25, 1, 1},
	}
	for _, tt := range tests {
		x, y := windowToGame(tt.wx, tt.wy, viewport{scale: tt.scale})
		if x != tt.x || y != tt.y {
			t.Errorf("windowToGame(%d,%d,%v) = (%d,%d), want (%d,%d)", tt.wx, tt.wy, tt.scale, x, y, tt.x, tt.y)
		}
	}
}

// Todo píxel del juego, pasado a ventana y de vuelta, es el mismo píxel, en
// cualquier escala (también las fraccionarias) y con bordes.
func TestGameToWindowRoundTrip(t *testing.T) {
	for _, scale := range []float64{1, 1.5, 16.0 / 9, 2, 2.25, 8.0 / 3, 3, 4, 5.3333333} {
		v := viewport{scale, 5, 7}
		for y := 0; y < 270; y++ {
			for x := 0; x < 480; x++ {
				wx, wy := gameToWindow(x, y, v)
				gx, gy := windowToGame(wx, wy, v)
				if gx != x || gy != y {
					t.Fatalf("scale %v: (%d,%d) -> window (%d,%d) -> (%d,%d)", scale, x, y, wx, wy, gx, gy)
				}
			}
		}
	}
}

// Con bordes oscuros el cursor se mide desde la esquina del lienzo, no desde
// la de la ventana.
func TestWindowToGameWithOffset(t *testing.T) {
	tests := []struct {
		wx, wy int
		v      viewport
		x, y   int
	}{
		{30, 20, viewport{3, 30, 20}, 0, 0},                  // esquina del lienzo
		{30 + 1439, 20 + 809, viewport{3, 30, 20}, 479, 269}, // última esquina
		{29, 19, viewport{3, 30, 20}, -1, -1},                // en el borde oscuro
		{30 + 1440, 20, viewport{3, 30, 20}, 480, 0},         // pasado el lienzo
		{100, 100, viewport{2, 10, 0}, 45, 50},
		{0, 12, viewport{1280.0 / 480, 0, 12}, 0, 0},              // 1280x800: franja arriba
		{1279, 12 + 719, viewport{1280.0 / 480, 0, 12}, 479, 269}, // y la última esquina
		{600, 11, viewport{1280.0 / 480, 0, 12}, 225, -1},         // sobre la franja de arriba
		{1499, 450, viewport{1500.0 / 480, 0, 0}, 479, 144},       // 1500x900: llena el ancho
	}
	for _, tt := range tests {
		x, y := windowToGame(tt.wx, tt.wy, tt.v)
		if x != tt.x || y != tt.y {
			t.Errorf("windowToGame(%d,%d,%+v) = (%d,%d), want (%d,%d)", tt.wx, tt.wy, tt.v, x, y, tt.x, tt.y)
		}
	}
}

func TestFitScreen(t *testing.T) {
	tests := []struct {
		w, h   int
		s      float64
		cw, ch int
		ox, oy int
	}{
		{1920, 1080, 4, 1920, 1080, 0, 0},
		{960, 540, 2, 960, 540, 0, 0},
		{2560, 1440, 2560.0 / 480, 2560, 1440, 0, 0}, // llena, sin bordes
		{1366, 768, 768.0 / 270, 1365, 768, 0, 0},    // portátil común: manda el alto
		{1280, 800, 1280.0 / 480, 1280, 720, 0, 40},  // 16:10: franjas arriba y abajo
		{1920, 1200, 4, 1920, 1080, 0, 60},
		{1500, 900, 1500.0 / 480, 1500, 844, 0, 28}, // 1,67: sobran 56 px de alto
		{1501, 899, 1501.0 / 480, 1501, 844, 0, 27}, // tamaño impar
		{480, 270, 1, 480, 270, 0, 0},
		{2400, 1350, 5, 2400, 1350, 0, 0},
	}
	for _, tt := range tests {
		s, cw, ch, ox, oy := fitScreen(tt.w, tt.h)
		if math.Abs(s-tt.s) > 1e-9 || cw != tt.cw || ch != tt.ch || ox != tt.ox || oy != tt.oy {
			t.Errorf("fitScreen(%d,%d) = %v,%d,%d,%d,%d want %v,%d,%d,%d,%d", tt.w, tt.h, s, cw, ch, ox, oy, tt.s, tt.cw, tt.ch, tt.ox, tt.oy)
		}
	}
}

// El lienzo nunca se sale de la ventana y nunca queda más de un píxel de
// diferencia entre los bordes (se reparte lo que sobra).
func TestFitScreenStaysInside(t *testing.T) {
	for _, size := range [][2]int{{1920, 1080}, {1366, 768}, {1501, 899}, {1280, 800}, {2559, 1439}, {483, 271}, {500, 270}, {0, 0}} {
		_, cw, ch, ox, oy := fitScreen(size[0], size[1])
		if ox < 0 || oy < 0 || ox+cw > max(size[0], 1) || oy+ch > max(size[1], 1) {
			t.Errorf("%v: canvas %dx%d at (%d,%d) leaves the window", size, cw, ch, ox, oy)
		}
	}
}

// El cursor de ventana de cada escala cae en la opción de menú que se está
// apuntando (el bug de "apunto Nuevo turno y señala Créditos").
func TestMenuHoverAtEveryScale(t *testing.T) {
	for _, scale := range []float64{1, 1.5, 2, 2.25, 3, 4} {
		for i := range menuItems {
			m := newMenu(&Env{HasSave: true})
			var in Input
			in.reset()
			r := optionRect(i)
			wx, wy := gameToWindow(r.x+r.w/2, r.y+r.h/2, viewport{scale: scale})
			in.moveToWindow(wx, wy, viewport{scale: scale})
			m.Update(&in)
			if m.hovered != menuItems[i].label {
				t.Errorf("scale %v: over option %d hovered %q, want %q", scale, i, m.hovered, menuItems[i].label)
			}
			if m.selected != i {
				t.Errorf("scale %v: over option %d selected %d", scale, i, m.selected)
			}
		}
	}
}

// Con escala fraccionaria el rectángulo de una imagen empieza y termina en
// píxeles enteros, y dos imágenes vecinas del juego se tocan sin rendija.
func TestImageGeoWholePixels(t *testing.T) {
	for _, scale := range []float64{1, 1.5, 2.25, 16.0 / 9 * 1.6, 4} {
		e := &Env{Scale: scale}
		for _, flip := range []bool{false, true} {
			var end, start float64
			for i, x := range []float64{10, 26} { // dos sprites de 16 px pegados
				g := e.imageGeo(16, 12, x, 5, 1, flip)
				x0, _ := g.Apply(0, 0)
				x1, _ := g.Apply(16, 0)
				lo, hi := math.Min(x0, x1), math.Max(x0, x1)
				if lo != math.Trunc(lo) || hi != math.Trunc(hi) {
					t.Errorf("scale %v flip %v: edges %v..%v not whole pixels", scale, flip, lo, hi)
				}
				if i == 0 {
					end = hi
				} else {
					start = lo
				}
			}
			if end != start {
				t.Errorf("scale %v flip %v: gap between neighbours (%v vs %v)", scale, flip, end, start)
			}
		}
	}
}
