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

// face devuelve la cara pedida; size es el tamaño en píxeles de pantalla.
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

// Width mide el ancho del texto en píxeles del juego (a 1x). Mide con la
// letra 8 veces más grande y divide, para no acumular el error de redondeo.
func (f *Fonts) Width(s string, family fontFamily, size int) int {
	return (f.screenWidth(s, family, size*8) + 4) / 8
}

// screenWidth mide el ancho en píxeles de pantalla con el tamaño ya escalado.
func (f *Fonts) screenWidth(s string, family fontFamily, size int) int {
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
// alineación) en (x, y), en coordenadas del lienzo de 480x270. El texto se
// dibuja con la letra ya multiplicada por Scale (tamaño de pantalla), así cada
// trazo cae en píxeles enteros de la ventana y se ve nítido; si se dibujara a
// 1x y se agrandara después, saldría borroso.
func (e *Env) drawText(dst *ebiten.Image, s string, family fontFamily, size, x, y int, c color.RGBA, align textAlign) {
	k := e.Scale
	big := size * k
	px, py := x*k, y*k
	switch align {
	case alignCenter:
		px -= e.Fonts.screenWidth(s, family, big) / 2
	case alignRight:
		px -= e.Fonts.screenWidth(s, family, big)
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(px), float64(py))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, screenText(s), e.Fonts.face(family, big), op)
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
// desplazada (+1,+1 píxel del juego), como en las vistas de Figma.
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
	k := float64(e.Scale)
	op.GeoM.Scale(float64(r.w)*k, float64(r.h)*k)
	op.GeoM.Translate(float64(r.x)*k, float64(r.y)*k)
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

// DrawSpriteScaled dibuja una imagen agrandada k veces más (por ejemplo las
// estrellas del reporte, que van a x2).
func (e *Env) DrawSpriteScaled(dst, img *ebiten.Image, x, y, k int) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	s := float64(e.Scale * k)
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(x*e.Scale), float64(y*e.Scale))
	dst.DrawImage(img, op)
}

// DrawSprite dibuja una imagen en (x, y) (coordenadas de 480x270, agrandada
// Scale veces sin suavizar). flip la refleja de izquierda
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
	k := float64(e.Scale)
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(x)*k, float64(y)*k)
	dst.DrawImage(img, op)
}
