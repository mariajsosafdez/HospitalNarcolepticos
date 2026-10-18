package gameview

import (
	"bytes"
	_ "embed" // para incluir las fuentes en el ejecutable
	"image/color"
	"math"

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
	quarts int // tamaño en cuartos de píxel de pantalla (ver face)
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

// face devuelve la cara pedida; size es el tamaño en píxeles de pantalla (con
// la escala fraccionaria puede tener decimales). Se redondea a cuartos de
// píxel para que, al agrandar la ventana poco a poco, no se creen caras sin fin.
func (f *Fonts) face(family fontFamily, size float64) *text.GoTextFace {
	quarts := int(math.Round(size * 4))
	key := faceKey{family, quarts}
	if face, ok := f.faces[key]; ok {
		return face
	}
	face := &text.GoTextFace{Source: f.pixel, Size: float64(quarts) / 4}
	face.SetVariation(text.MustParseTag("wght"), 400)
	// Pixelify Sans trae ligaduras ("fi", "fl") que salen como un glifo roto
	// ("la fila" se veía "la Ala"): se apagan para escribir letra por letra.
	for _, tag := range []string{"liga", "clig", "dlig"} {
		face.SetFeature(text.MustParseTag(tag), 0)
	}
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
	return int(f.lineWidth(s, family, size, 8)/8 + 0.5)
}

// screenWidth mide el ancho en píxeles de pantalla con el tamaño ya escalado.
func (f *Fonts) screenWidth(s string, family fontFamily, size float64) float64 {
	w, _ := text.Measure(screenText(s), f.face(family, size), 0)
	return w
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
// dibuja con la letra ya multiplicada por Scale (tamaño de pantalla) y en un
// píxel entero de la ventana, así cada trazo queda nítido; si se dibujara a
// 1x y se agrandara después, saldría borroso.
func (e *Env) drawText(dst *ebiten.Image, s string, family fontFamily, size, x, y int, c color.RGBA, align textAlign) {
	big := float64(size) * e.Scale
	px, py := float64(e.px(float64(x))), float64(e.px(float64(y)))
	switch align {
	case alignCenter:
		px -= e.Fonts.lineWidth(s, family, size, e.Scale) / 2
	case alignRight:
		px -= e.Fonts.lineWidth(s, family, size, e.Scale)
	}
	spec, hasSprites := digitSpecFor(family, size)
	if !hasSprites {
		e.drawFontRun(dst, s, family, big, px, py, c)
		return
	}
	// Los dígitos van como sprites y lo demás con la fuente, tanda por tanda.
	for _, run := range splitRuns(s) {
		if run.digits {
			px += e.drawDigits(dst, run.text, spec, px, y, c)
		} else {
			e.drawFontRun(dst, run.text, family, big, px, py, c)
			px += e.Fonts.screenWidth(run.text, family, big)
		}
	}
}

// drawFontRun dibuja un pedazo de texto con la fuente; (px, py) es el píxel
// de pantalla de su esquina (se redondea para que las letras no caigan a medio píxel).
func (e *Env) drawFontRun(dst *ebiten.Image, s string, family fontFamily, big, px, py float64, c color.RGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(math.Round(px), math.Round(py))
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

// px pasa una medida del juego a píxeles de pantalla, redondeada a un entero.
// Todo borde se calcula con esta función, así dos cosas que se tocan en el
// juego (el fin de una y el comienzo de otra) también se tocan en pantalla,
// sin rendijas.
func (e *Env) px(v float64) int { return int(math.Round(v * e.Scale)) }

// fillPixels rellena el rectángulo [x0,x1) x [y0,y1) en píxeles de pantalla.
func (e *Env) fillPixels(dst *ebiten.Image, x0, y0, x1, y1 int, key byte, alpha float64) {
	if x1 <= x0 || y1 <= y0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(x1-x0), float64(y1-y0))
	op.GeoM.Translate(float64(x0), float64(y0))
	op.ColorScale.ScaleWithColor(e.Assets.Color(key))
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(e.Assets.pixel, op)
}

// FillRect rellena un rectángulo con un color de la paleta y una opacidad (0 a 1).
// Los bordes se redondean a píxeles de pantalla, y un rectángulo con área
// nunca desaparece (mide al menos 1 píxel).
func (e *Env) FillRect(dst *ebiten.Image, r rect, key byte, alpha float64) {
	if r.w <= 0 || r.h <= 0 {
		return
	}
	x0, y0 := e.px(float64(r.x)), e.px(float64(r.y))
	x1, y1 := e.px(float64(r.x+r.w)), e.px(float64(r.y+r.h))
	e.fillPixels(dst, x0, y0, max(x1, x0+1), max(y1, y0+1), key, alpha)
}

// StrokeRect dibuja el borde de 1 px (o thickness) de un rectángulo. El grosor
// se pasa a píxeles de pantalla UNA vez (mínimo 1) para que los cuatro lados
// midan igual aunque la escala sea fraccionaria.
func (e *Env) StrokeRect(dst *ebiten.Image, r rect, key byte, thickness int) {
	if r.w <= 0 || r.h <= 0 {
		return
	}
	t := max(1, e.px(float64(thickness)))
	x0, y0 := e.px(float64(r.x)), e.px(float64(r.y))
	x1, y1 := max(e.px(float64(r.x+r.w)), x0+1), max(e.px(float64(r.y+r.h)), y0+1)
	e.fillPixels(dst, x0, y0, x1, y0+t, key, 1) // arriba
	e.fillPixels(dst, x0, y1-t, x1, y1, key, 1) // abajo
	e.fillPixels(dst, x0, y0, x0+t, y1, key, 1) // izquierda
	e.fillPixels(dst, x1-t, y0, x1, y1, key, 1) // derecha
}

// imageGeo da la transformación que pinta una imagen de w x h en la esquina
// (x, y) del juego, agrandada mult veces además de la escala. Calcula el
// rectángulo de destino con bordes redondeados y estira la imagen justo hasta
// ahí: con escala fraccionaria algunos píxeles del arte miden un píxel de
// pantalla más que otros, pero ninguno se emborrona ni deja rendijas.
// flip la refleja de izquierda a derecha.
func (e *Env) imageGeo(w, h int, x, y float64, mult int, flip bool) ebiten.GeoM {
	x0, y0 := e.px(x), e.px(y)
	x1, y1 := e.px(x+float64(w*mult)), e.px(y+float64(h*mult))
	x1, y1 = max(x1, x0+1), max(y1, y0+1)
	sx, sy := float64(x1-x0)/float64(w), float64(y1-y0)/float64(h)
	var g ebiten.GeoM
	if flip {
		g.Scale(-sx, sy)
		g.Translate(float64(x1), float64(y0))
	} else {
		g.Scale(sx, sy)
		g.Translate(float64(x0), float64(y0))
	}
	return g
}

// DrawSpriteScaled dibuja una imagen agrandada mult veces más (por ejemplo las
// estrellas del reporte, que van a x2).
func (e *Env) DrawSpriteScaled(dst, img *ebiten.Image, x, y, mult int) {
	e.drawImage(dst, img, float64(x), float64(y), mult, false, 1)
}

// DrawSprite dibuja una imagen en (x, y) (coordenadas de 480x270, agrandada
// Scale veces sin suavizar). flip la refleja de izquierda
// a derecha (para caminar a la izquierda).
func (e *Env) DrawSprite(dst, img *ebiten.Image, x, y int, flip bool) {
	e.drawImage(dst, img, float64(x), float64(y), 1, flip, 1)
}

// drawImage es lo que comparten los dibujos de imágenes: vecino más cercano
// (nunca suavizado) y opacidad alpha.
func (e *Env) drawImage(dst, img *ebiten.Image, x, y float64, mult int, flip bool, alpha float32) {
	if img == nil {
		return
	}
	b := img.Bounds()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM = e.imageGeo(b.Dx(), b.Dy(), x, y, mult, flip)
	op.ColorScale.ScaleAlpha(alpha)
	dst.DrawImage(img, op)
}
