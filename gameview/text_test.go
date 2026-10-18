package gameview

import "testing"

// Sin ligaduras, "fi" y "fl" miden lo mismo que sus letras sueltas (con la
// ligadura el glifo único tenía otro ancho y salía roto).
// (se deja un margen de 10 px de pantalla por el kerning entre las dos letras).
func TestNoLigatures(t *testing.T) {
	fonts, err := LoadFonts()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range []fontFamily{fontPixel, fontPixelBold} {
		for _, pair := range []string{"fi", "fl"} {
			whole := fonts.screenWidth(pair, family, 160)
			parts := fonts.screenWidth(pair[:1], family, 160) + fonts.screenWidth(pair[1:], family, 160)
			if diff := whole - parts; diff < -10 || diff > 10 {
				t.Errorf("family %d %q: junto %d, suelto %d", family, pair, whole, parts)
			}
		}
	}
}
