package gameview

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Pixelify Sans dibuja el "5" casi igual que la "S", así que los dígitos no
// salen de la fuente sino de tres sprites (art/png/ui_digitos_*): 10 cuadros
// (0 al 9) del mismo ancho, con el espacio entre dígitos ya incluido.

// digitSpec describe uno de esos sprites para un tamaño de letra.
type digitSpec struct {
	sprite  string // nombre del PNG
	advance int    // ancho de cada dígito, con su espacio (px del juego)
	height  int    // alto del dígito
	base    int    // fila de la línea base, contada desde la y que se entrega a drawText
}

// digitSpecFor devuelve el sprite que reemplaza los dígitos de esa cara y
// tamaño; ok es false si ese tamaño sigue usando la fuente.
func digitSpecFor(family fontFamily, size int) (spec digitSpec, ok bool) {
	switch {
	case (family == fontPixel || family == fontPixelBold) && size == 8:
		return digitSpec{"ui_digitos_8", 5, 5, 7}, true
	case family == fontPixelBold && size == 10:
		return digitSpec{"ui_digitos_10b", 6, 6, 9}, true
	case family == fontPixelBold && size == 18:
		return digitSpec{"ui_digitos_18b", 11, 12, 17}, true
	}
	return digitSpec{}, false
}

// textRun es un pedazo de la línea: solo dígitos (se dibujan con sprites) o
// solo lo demás (se dibuja con la fuente).
type textRun struct {
	text   string
	digits bool
}

// isDigit dice si r es un dígito ASCII.
func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// splitRuns parte la línea en tandas seguidas de dígitos y de no dígitos.
func splitRuns(s string) []textRun {
	var runs []textRun
	start := 0
	prev := false // si el carácter anterior era dígito
	for i, r := range s {
		d := isDigit(r)
		if i > 0 && d != prev {
			runs = append(runs, textRun{s[start:i], prev})
			start = i
		}
		prev = d
	}
	if start < len(s) {
		runs = append(runs, textRun{s[start:], prev})
	}
	return runs
}

// lineWidth mide la línea en píxeles de pantalla: size está en píxeles del
// juego y k es cuántos píxeles de pantalla mide cada uno (puede tener
// decimales). Los dígitos miden advance*k cada uno; lo demás se mide con la fuente.
func (f *Fonts) lineWidth(s string, family fontFamily, size int, k float64) float64 {
	spec, hasSprites := digitSpecFor(family, size)
	if !hasSprites {
		return f.screenWidth(s, family, float64(size)*k)
	}
	total := 0.0
	for _, run := range splitRuns(s) {
		if run.digits {
			total += float64(len(run.text)*spec.advance) * k
		} else {
			total += f.screenWidth(run.text, family, float64(size)*k)
		}
	}
	return total
}

// digitMask devuelve el cuadro del dígito d (0 al 9) del sprite, con todo
// píxel no transparente pintado de blanco: así se tiñe con ColorScale igual
// que las letras (y funcionan la sombra y el borde).
func (a *Assets) digitMask(sprite string, d int) *ebiten.Image {
	key := frameKey{sprite, 10, d}
	if m, ok := a.masks[key]; ok {
		return m
	}
	raw, ok := a.raw[sprite]
	if !ok {
		return nil
	}
	b := raw.Bounds()
	r := frameRect(b.Dx(), b.Dy(), 10, d)
	out := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			if _, _, _, alpha := raw.At(r.Min.X+x, r.Min.Y+y).RGBA(); alpha > 0 {
				out.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	m := ebiten.NewImageFromImage(out)
	a.masks[key] = m
	return m
}

// drawDigits dibuja una tanda de dígitos con su esquina izquierda en el píxel
// de pantalla px; y es la y del juego que recibió drawText. Devuelve el ancho
// dibujado en píxeles de pantalla. Cada dígito se pega con vecino más cercano
// en un rectángulo de píxeles enteros (ver imageGeo).
func (e *Env) drawDigits(dst *ebiten.Image, digits string, spec digitSpec, px float64, y int, c color.RGBA) float64 {
	top := float64(y + spec.base - spec.height)
	start := px
	for _, r := range digits {
		mask := e.Assets.digitMask(spec.sprite, int(r-'0'))
		if mask == nil {
			return 0
		}
		b := mask.Bounds()
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		op.GeoM = e.imageGeo(b.Dx(), b.Dy(), px/e.Scale, top, 1, false)
		op.ColorScale.ScaleWithColor(c)
		dst.DrawImage(mask, op)
		px += float64(spec.advance) * e.Scale
	}
	return px - start
}
