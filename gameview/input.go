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

// moveTo cambia la posición del mouse y recuerda si cambió.
func (in *Input) moveTo(x, y int) {
	if x != in.MouseX || y != in.MouseY {
		in.Moved = true
	}
	in.MouseX, in.MouseY = x, y
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
func (in *Input) readReal() {
	in.reset()
	x, y := ebiten.CursorPosition()
	in.moveTo(x, y)
	in.Click = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	for name, key := range keyNames {
		if inpututil.IsKeyJustPressed(key) {
			in.keys[name] = true
		}
	}
}
