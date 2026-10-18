package gameview

import (
	"image/png"
	"reflect"
	"testing"

	"costenos-narcolepsia/art"
)

func TestSplitRuns(t *testing.T) {
	tests := []struct {
		in   string
		want []textRun
	}{
		{"", nil},
		{"Hola", []textRun{{"Hola", false}}},
		{"105", []textRun{{"105", true}}},
		{"Día 5", []textRun{{"Día ", false}, {"5", true}}},
		{"4 / 9", []textRun{{"4", true}, {" / ", false}, {"9", true}}},
		{"de 12 posibles (80 %)", []textRun{{"de ", false}, {"12", true}, {" posibles (", false}, {"80", true}, {" %)", false}}},
	}
	for _, tt := range tests {
		if got := splitRuns(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("splitRuns(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestDigitSpecFor(t *testing.T) {
	tests := []struct {
		family fontFamily
		size   int
		sprite string
	}{
		{fontPixel, 8, "ui_digitos_8"},
		{fontPixelBold, 8, "ui_digitos_8"},
		{fontPixelBold, 10, "ui_digitos_10b"},
		{fontPixelBold, 18, "ui_digitos_18b"},
		{fontPixel, 10, ""}, // sin sprite: sigue la fuente
		{fontVT, 8, ""},
		{fontPixelBold, 12, ""},
	}
	for _, tt := range tests {
		spec, ok := digitSpecFor(tt.family, tt.size)
		if ok != (tt.sprite != "") || spec.sprite != tt.sprite {
			t.Errorf("digitSpecFor(%d, %d) = %q, %v", tt.family, tt.size, spec.sprite, ok)
		}
	}
}

// Los dígitos miden su avance cada uno (el "1" un píxel menos), a cualquier escala.
func TestLineWidthDigits(t *testing.T) {
	fonts, err := LoadFonts()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		s      string
		family fontFamily
		size   int
		k      float64
		want   float64
	}{
		{"105", fontPixel, 8, 1, 14}, // el "1" avanza 4 en vez de 5
		{"105", fontPixel, 8, 3, 42},
		{"58", fontPixelBold, 10, 2, 24},
		{"123", fontPixelBold, 18, 4, 128},
		{"11", fontPixel, 8, 1, 8},

		{"105", fontPixel, 8, 1.5, 21}, // escala fraccionaria
		{"58", fontPixelBold, 10, 2.25, 27},
	}
	for _, tt := range tests {
		if got := fonts.lineWidth(tt.s, tt.family, tt.size, tt.k); got != tt.want {
			t.Errorf("lineWidth(%q, %d, %v) = %v, want %v", tt.s, tt.size, tt.k, got, tt.want)
		}
	}
	// Con letras: el ancho es el de la fuente más los dígitos.
	want := fonts.screenWidth("Día ", fontPixel, 16) + 5*2.0
	if got := fonts.lineWidth("Día 5", fontPixel, 8, 2); got != want {
		t.Errorf("Día 5 = %v, want %v", got, want)
	}
}

// El avance de cada dígito sale de la tinta de su sprite: tinta + el mismo
// hueco que tienen los dígitos más anchos. Así "12", "101", "1/4" y "11"
// quedan con 1 px entre dígitos, y los demás dígitos no cambian.
func TestDigitAdvanceMatchesInk(t *testing.T) {
	for _, tt := range []struct {
		family fontFamily
		size   int
	}{{fontPixel, 8}, {fontPixelBold, 10}, {fontPixelBold, 18}} {
		spec, _ := digitSpecFor(tt.family, tt.size)
		f, err := art.Files.Open("png/" + spec.sprite + ".png")
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		b := img.Bounds()
		cell := b.Dx() / 10
		if cell != spec.advance {
			t.Errorf("%s: cell is %d px wide, spec.advance is %d", spec.sprite, cell, spec.advance)
		}
		ink := make([]int, 10) // ancho de la tinta: la última columna con píxeles + 1
		widest := 0
		for d := 0; d < 10; d++ {
			for y := 0; y < b.Dy(); y++ {
				for x := 0; x < cell; x++ {
					if _, _, _, a := img.At(d*cell+x, y).RGBA(); a > 0 {
						ink[d] = max(ink[d], x+1)
					}
				}
			}
			widest = max(widest, ink[d])
		}
		gap := spec.advance - widest
		for d := 0; d < 10; d++ {
			if got, want := spec.advanceOf(d), ink[d]+gap; got != want {
				t.Errorf("%s: advance of %d = %d, want %d (ink %d + gap %d)", spec.sprite, d, got, want, ink[d], gap)
			}
		}
	}
}
