package gameview

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input es lo que las pantallas saben de la entrada en UN tick. Las
// pantallas nunca llaman a Ebitengine para leer el teclado o el mouse: lo
// reciben aquí. Así la puerta de pruebas (-agent) puede inyectar clics y
// teclas de mentira y la pantalla no se entera de la diferencia.
type Input struct {
	MouseX, MouseY int  // posición del mouse en el lienzo de 480x270
	Moved          bool // el mouse cambió de sitio en este tick
	Click          bool // clic izquierdo en este tick
	keys           map[string]bool
}

// Pressed dice si la tecla (por su nombre del juego) se pulsó en este tick.
func (in *Input) Pressed(name string) bool { return in.keys[name] }

// reset limpia lo que dura un solo tick (el mouse conserva su posición).
func (in *Input) reset() {
	in.Moved = false
	in.Click = false
	in.keys = make(map[string]bool)
}

// moveTo cambia la posición del mouse (en píxeles del juego) y recuerda si
// cambió.
func (in *Input) moveTo(x, y int) {
	if x != in.MouseX || y != in.MouseY {
		in.Moved = true
	}
	in.MouseX, in.MouseY = x, y
}

// viewport dice dónde cae el lienzo de 480x270 dentro de la imagen de la
// ventana: agrandado scale veces y corrido (offX, offY) píxeles de ventana
// (los bordes oscuros de la pantalla completa, cuando la ventana no es un
// múltiplo exacto de 480x270).
type viewport struct{ scale, offX, offY int }

// moveToWindow recibe la posición del cursor tal como la da Ebitengine (en el
// espacio que devolvió Layout, la imagen entera de la ventana) y la pasa a
// píxeles del juego. Es el ÚNICO camino del cursor a las pantallas: el juego
// de verdad y la puerta de pruebas pasan por aquí.
func (in *Input) moveToWindow(wx, wy int, v viewport) {
	x, y := windowToGame(wx, wy, v)
	in.moveTo(x, y)
}

// moveToGameCell es para la puerta de pruebas: pone el cursor en el medio del
// cuadro (x, y) del juego, pero como posición de VENTANA, y la hace pasar por
// la misma conversión que el cursor real.
func (in *Input) moveToGameCell(x, y int, v viewport) {
	wx, wy := gameToWindow(x, y, v)
	in.moveToWindow(wx, wy, v)
}

// windowToGame convierte una posición del cursor (píxeles de la imagen de la
// ventana) en píxeles del juego (480x270): quita el corrimiento de los bordes
// y divide por la escala. Divide hacia abajo, también con negativos (el
// cursor puede caer en el borde oscuro o fuera de la ventana).
func windowToGame(wx, wy int, v viewport) (int, int) {
	scale := max(v.scale, 1)
	return floorDiv(wx-v.offX, scale), floorDiv(wy-v.offY, scale)
}

// gameToWindow es la inversa para la puerta de pruebas: da el píxel del medio
// del cuadro (x, y) del juego, que es donde caería el cursor de una persona.
func gameToWindow(x, y int, v viewport) (int, int) {
	scale := max(v.scale, 1)
	return x*scale + scale/2 + v.offX, y*scale + scale/2 + v.offY
}

// floorDiv divide redondeando hacia abajo (la división de Go redondea hacia 0).
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// keyNames traduce el nombre de una tecla del juego a la de Ebitengine. Es la
// única lista de teclas: la usan el juego de verdad y los guiones de -agent.
var keyNames = map[string]ebiten.Key{
	"Up": ebiten.KeyArrowUp, "Down": ebiten.KeyArrowDown,
	"Left": ebiten.KeyArrowLeft, "Right": ebiten.KeyArrowRight,
	"Enter": ebiten.KeyEnter, "Escape": ebiten.KeyEscape, "Space": ebiten.KeySpace,
	"M": ebiten.KeyM,
	"1": ebiten.Key1, "2": ebiten.Key2, "3": ebiten.Key3, "4": ebiten.Key4,
	"5": ebiten.Key5, "6": ebiten.Key6, "7": ebiten.Key7, "8": ebiten.Key8,
}

// checkKeyName falla con un mensaje claro si la tecla no existe.
func checkKeyName(name string) error {
	if _, ok := keyNames[name]; !ok {
		return fmt.Errorf("unknown key %q", name)
	}
	return nil
}

// readReal llena el Input con el teclado y el mouse de verdad.
func (in *Input) readReal(v viewport) {
	in.reset()
	x, y := ebiten.CursorPosition()
	in.moveToWindow(x, y, v)
	in.Click = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	for name, key := range keyNames {
		if inpututil.IsKeyJustPressed(key) {
			in.keys[name] = true
		}
	}
}
