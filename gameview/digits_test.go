package gameview

import (
	"reflect"
	"testing"
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

// Los dígitos miden advance por cantidad, a cualquier escala.
func TestLineWidthDigits(t *testing.T) {
	fonts, err := LoadFonts()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		s      string
		family fontFamily
		size   int
		k      int
		want   int
	}{
		{"105", fontPixel, 8, 1, 15},
		{"105", fontPixel, 8, 3, 45},
		{"58", fontPixelBold, 10, 2, 24},
		{"123", fontPixelBold, 18, 4, 132},
	}
	for _, tt := range tests {
		if got := fonts.lineWidth(tt.s, tt.family, tt.size, tt.k); got != tt.want {
			t.Errorf("lineWidth(%q, %d, %d) = %d, want %d", tt.s, tt.size, tt.k, got, tt.want)
		}
	}
	// Con letras: el ancho es el de la fuente más los dígitos.
	want := fonts.screenWidth("Día ", fontPixel, 16) + 5*2
	if got := fonts.lineWidth("Día 5", fontPixel, 8, 2); got != want {
		t.Errorf("Día 5 = %d, want %d", got, want)
	}
}
