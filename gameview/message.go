package gameview

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

// Mensaje del Director, según la vista aprobada "Director / mensaje" (frame
// 18:186 de Figma, nota 18:533; copia en preview/turno/director-mensaje.png).
// Sale arriba de la cámara para que los pacientes sigan a la vista. Mientras
// está abierto el reloj del turno se detiene.

// messageRect es la caja: (8, 34) 328x70, dentro de la cámara con 4 px de borde.
var messageRect = rect{8, 34, 328, 70}

// messageKind dice qué mensaje es (y qué pasa al cerrarlo).
type messageKind int

const (
	msgTutorial messageKind = iota // los 9 mensajes del Día 1
	msgDayIntro                    // inicio de los días 2 a 5
	msgOverwork                    // un doctor llegó a 4/4
)

// messageBox es el mensaje que se está mostrando: una o varias líneas de
// texto del Director, una por pantalla.
type messageBox struct {
	kind  messageKind
	lines []string
	index int
}

// newMessage arma un mensaje que empieza en la línea start (0 = la primera).
func newMessage(kind messageKind, lines []string, start int) *messageBox {
	return &messageBox{kind: kind, lines: lines, index: min(max(start, 0), len(lines)-1)}
}

// text es el texto de la pantalla actual.
func (m *messageBox) text() string { return m.lines[m.index] }

// counter es el "4 / 9" del tutorial; los mensajes sueltos no lo llevan.
func (m *messageBox) counter() string {
	if len(m.lines) < 2 {
		return ""
	}
	return fmt.Sprintf("%d / %d", m.index+1, len(m.lines))
}

// advance pasa a la siguiente línea. Devuelve true si ya no había más (el
// mensaje terminó y hay que cerrarlo).
func (m *messageBox) advance() bool {
	if m.index+1 >= len(m.lines) {
		return true
	}
	m.index++
	return false
}

// messageWrapWidths son los anchos de texto que se prueban, en píxeles del
// juego. El de la vista es 200 (el texto del Director hace saltos como en
// Figma); si así el texto pasa de 3 líneas (no cabe en la caja), se ensancha.
var messageWrapWidths = []int{200, 224, 248, 272, 290}

// wrapMessage parte el texto del Director en líneas de la caja: a lo sumo 3
// y lo más angostas posible. Es pura: recibe la función que mide el texto.
func wrapMessage(s string, measure func(string) int) []string {
	var lines []string
	for _, w := range messageWrapWidths {
		lines = wrapText(s, w, measure)
		if len(lines) <= 3 {
			return lines
		}
	}
	return lines
}

// drawMessage pinta la caja del Director sobre la cámara.
func (c *camsScene) drawMessage(dst *ebiten.Image) {
	e := c.env
	m := c.msg
	r := messageRect
	drawBox(e, dst, r)
	e.DrawSprite(dst, e.Assets.Image("director_retrato"), r.x+6, r.y+7, false)
	e.Text(dst, game.DirectorName, r.x+60, r.y+4, textStyle{fontPixelBold, 10, '2', alignLeft})
	if counter := m.counter(); counter != "" {
		e.Text(dst, counter, r.x+r.w-14, r.y+5, textStyle{fontPixel, 8, 'x', alignRight})
	}
	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	for i, line := range wrapMessage(m.text(), measure) {
		e.Text(dst, line, r.x+60, r.y+18+11*i, textStyle{fontPixel, 8, '1', alignLeft})
	}
	e.Text(dst, "Clic para seguir", r.x+r.w-36, r.y+56, textStyle{fontPixel, 8, '2', alignRight})
	e.DrawSprite(dst, e.Assets.Frame("ui_flecha", 2, animFrame(e.Tick, 24, 2, 0)), r.x+r.w-26, r.y+56, false)
}
