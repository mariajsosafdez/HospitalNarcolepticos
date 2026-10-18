package gameview

import (
	"bytes"
	_ "embed" // para incluir las fuentes en el ejecutable
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Las dos fuentes (licencia SIL OFL 1.1, junto a cada una va su OFL.txt).
//
//go:embed assets/fonts/PixelifySans.ttf
var pixelifyData []byte

//go:embed assets/fonts/VT323-Regular.ttf
var vt323Data []byte

// fontFamily elige entre las tres caras que usa el juego.
type fontFamily int

const (
	fontPixel     fontFamily = iota // Pixelify Sans Regular: casi todo el texto
	fontPixelBold                   // Pixelify Sans Bold: títulos
	fontVT                          // VT323: solo el rótulo de las cámaras y la hora
)

type faceKey struct {
	family fontFamily
	size   int
}

// Fonts crea y guarda las caras de texto (una por familia y tamaño).
type Fonts struct {
	pixel *text.GoTextFaceSource
	vt    *text.GoTextFaceSource
	faces map[faceKey]*text.GoTextFace
}

// LoadFonts lee las dos fuentes embebidas.
func LoadFonts() (*Fonts, error) {
	pixel, err := text.NewGoTextFaceSource(bytes.NewReader(pixelifyData))
	if err != nil {
		return nil, err
	}
	vt, err := text.NewGoTextFaceSource(bytes.NewReader(vt323Data))
	if err != nil {
		return nil, err
	}
	return &Fonts{pixel: pixel, vt: vt, faces: make(map[faceKey]*text.GoTextFace)}, nil
}

// face devuelve la cara pedida; size es el tamaño en píxeles a 1x.
func (f *Fonts) face(family fontFamily, size int) *text.GoTextFace {
	key := faceKey{family, size}
	if face, ok := f.faces[key]; ok {
		return face
	}
	face := &text.GoTextFace{Source: f.pixel, Size: float64(size)}
	face.SetVariation(text.MustParseTag("wght"), 400)
	switch family {
	case fontPixelBold:
		// Pixelify Sans es una fuente variable: el eje "wght" va de 400 a 700.
		face.SetVariation(text.MustParseTag("wght"), 700)
	case fontVT:
		face.Source = f.vt
	}
	f.faces[key] = face
	return face
}

// Width mide el ancho del texto en píxeles a 1x.
func (f *Fonts) Width(s string, family fontFamily, size int) int {
	w, _ := text.Measure(screenText(s), f.face(family, size), 0)
	return int(w + 0.5)
}

// textAlign dice a qué punto del texto se refiere la x que se entrega.
type textAlign int

const (
	alignLeft textAlign = iota
	alignCenter
	alignRight
)

// drawText dibuja una línea de texto con su esquina de arriba (según la
// alineación) en (x, y), en coordenadas del lienzo de 480x270.
func (e *Env) drawText(dst *ebiten.Image, s string, family fontFamily, size, x, y int, c color.RGBA, align textAlign) {
	switch align {
	case alignCenter:
		x -= e.Fonts.Width(s, family, size) / 2
	case alignRight:
		x -= e.Fonts.Width(s, family, size)
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, screenText(s), e.Fonts.face(family, size), op)
}

// textStyle agrupa lo que se repite al dibujar una línea.
type textStyle struct {
	family fontFamily
	size   int
	key    byte // color de la paleta
	align  textAlign
}

// Text dibuja una línea con el color de la paleta.
func (e *Env) Text(dst *ebiten.Image, s string, x, y int, st textStyle) {
	e.drawText(dst, s, st.family, st.size, x, y, e.Assets.Color(st.key), st.align)
}

// TextShadow dibuja el texto con sombra: la misma línea en el color 0
// desplazada (+1,+1), como en las vistas de Figma.
func (e *Env) TextShadow(dst *ebiten.Image, s string, x, y int, st textStyle) {
	e.drawText(dst, s, st.family, st.size, x+1, y+1, e.Assets.Color('0'), st.align)
	e.Text(dst, s, x, y, st)
}

// TextOutline dibuja el texto con borde oscuro de 1 px (color 0): la misma
// línea desplazada en los 4 sentidos (sin diagonales) y el texto encima. Pedido de la autora
// para el rótulo de las cámaras: "hazle borde negro al texto, así destaca".
func (e *Env) TextOutline(dst *ebiten.Image, s string, x, y int, st textStyle) {
	dark := e.Assets.Color('0')
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if (dx == 0) != (dy == 0) { // solo arriba, abajo, izquierda y derecha: más limpio
				e.drawText(dst, s, st.family, st.size, x+dx, y+dy, dark, st.align)
			}
		}
	}
	e.Text(dst, s, x, y, st)
}

// FillRect rellena un rectángulo con un color de la paleta y una opacidad (0 a 1).
func (e *Env) FillRect(dst *ebiten.Image, r rect, key byte, alpha float64) {
	if r.w <= 0 || r.h <= 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(r.w), float64(r.h))
	op.GeoM.Translate(float64(r.x), float64(r.y))
	op.ColorScale.ScaleWithColor(e.Assets.Color(key))
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(e.Assets.pixel, op)
}

// StrokeRect dibuja el borde de 1 px (o thickness) de un rectángulo.
func (e *Env) StrokeRect(dst *ebiten.Image, r rect, key byte, thickness int) {
	e.FillRect(dst, rect{r.x, r.y, r.w, thickness}, key, 1)
	e.FillRect(dst, rect{r.x, r.y + r.h - thickness, r.w, thickness}, key, 1)
	e.FillRect(dst, rect{r.x, r.y, thickness, r.h}, key, 1)
	e.FillRect(dst, rect{r.x + r.w - thickness, r.y, thickness, r.h}, key, 1)
}

// DrawSprite dibuja una imagen en (x, y) a 1x. flip la refleja de izquierda
// a derecha (para caminar a la izquierda).
func (e *Env) DrawSprite(dst, img *ebiten.Image, x, y int, flip bool) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	if flip {
		w := img.Bounds().Dx()
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(w), 0)
	}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(img, op)
}
