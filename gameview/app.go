// Package gameview es la VISTA del juego: abre la ventana, dibuja y lee la
// entrada con Ebitengine. Es el ÚNICO paquete del proyecto que importa
// Ebitengine; las reglas viven en el paquete game, que solo se lee y se
// manda acciones desde aquí.
package gameview

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

const (
	// Resolución lógica del juego (pixel art) y cuánto se agranda en pantalla.
	logicalWidth  = 480
	logicalHeight = 270
	windowScale   = 3

	windowTitle = "Hospital de los Costeños con Narcolepsia"
)

// background es el color "w" de la paleta (art/palette.hex): crema clara.
var background = color.RGBA{R: 0xff, G: 0xf3, B: 0xdc, A: 0xff}

// App implementa ebiten.Game. Todavía no tiene pantallas: solo avanza el
// turno y pinta el fondo.
type App struct {
	shift *game.Shift
}

// Update lo llama Ebitengine 60 veces por segundo: cada llamada es un tick
// del juego, igual que el reloj de game.Clock.
func (a *App) Update() error {
	a.shift.Tick()
	return nil
}

// Draw pinta un cuadro. La lógica no se toca aquí.
func (a *App) Draw(screen *ebiten.Image) {
	screen.Fill(background)
}

// Layout fija la resolución lógica; Ebitengine escala la ventana.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return logicalWidth, logicalHeight
}

// Run abre la ventana y bloquea hasta que se cierra. Ebitengine exige
// llamarlo desde la goroutine principal.
func Run(shift *game.Shift) error {
	ebiten.SetWindowSize(logicalWidth*windowScale, logicalHeight*windowScale)
	ebiten.SetWindowTitle(windowTitle)
	return ebiten.RunGame(&App{shift: shift})
}
