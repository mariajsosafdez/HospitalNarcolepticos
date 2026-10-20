package gameview

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Pantalla de ajustes, según la vista aprobada "Ajustes" (frame 18:340 de
// Figma, nota 18:536; copia en preview/turno/ajustes.png). Es una caja que se
// abre encima del menú o de la pausa. Cada cambio se autoguarda.

// settingsBox es la caja: (110, 26) 260x218, centrada en la pantalla.
var settingsBox = rect{110, 26, 260, 218}

// Filas de la lista.
const (
	setBlinkSleep = iota
	setBlinkWake
	setFullscreen
	setSound
	setBack
	setRows
)

// settingsLabels son los rótulos de las filas, y settingsDescs lo que dice la
// línea de abajo con la fila marcada (Textos del Director.md).
var settingsLabels = [setRows]string{"Parpadeo al dormirse", "Parpadeo al despertarse", "Pantalla completa", "Sonido", "Volver"}

var settingsDescs = [setRows]string{
	"La cámara titila cuando alguien se duerme.",
	"La cámara titila cuando alguien se despierta.",
	"Llena toda la pantalla.",
	"Música y efectos. Apagado: silencio total.",
	"",
}

const settingsHelp = "Flechas: elegir  ·  Izq/Der: cambiar  ·  Esc: volver"

// settingsRowRect es la barra de la fila i (las 4 primeras): (116, 64 + 28i) 248x15.
func settingsRowRect(i int) rect { return rect{116, 64 + 28*i, 248, 15} }

// settingsBackRect es el botón "Volver".
// Va en y = 172 (era 162 en la vista aprobada, antes de que hubiera una cuarta
// fila: la fila "Sonido" ocupa de 148 a 163).
var settingsBackRect = rect{200, 172, 80, 16}

// settingsToggle es el botón Apagado/Prendido de la fila i: (300, y + 3) 52x12.
func settingsToggle(i int) rect { r := settingsRowRect(i); return rect{300, r.y + 3, 52, 12} }

// stepRow mueve la fila marcada, dando la vuelta en los extremos.
func stepRow(sel, dir int) int { return ((sel+dir)%setRows + setRows) % setRows }

// onOff escribe el estado de una opción.
func onOff(on bool) string {
	if on {
		return "Prendido"
	}
	return "Apagado"
}

// settingsView es la caja de ajustes abierta. Cambia env.Save.Settings y lo
// autoguarda con cada cambio.
type settingsView struct {
	env     *Env
	sel     int
	hovered string
}

func newSettingsView(env *Env) *settingsView { return &settingsView{env: env} }

// toggle prende o apaga la opción de la fila y autoguarda.
func (v *settingsView) toggle(row int) {
	s := &v.env.Save.Settings
	switch row {
	case setBlinkSleep:
		s.BlinkOnSleep = !s.BlinkOnSleep
	case setBlinkWake:
		s.BlinkOnWake = !s.BlinkOnWake
	case setFullscreen:
		s.Fullscreen = !s.Fullscreen
	case setSound:
		s.SoundOff = !s.SoundOff
	}
	v.env.settingsChanged()
}

// change aplica Izquierda o Derecha a la fila marcada. Las cuatro opciones son
// de prender y apagar, así que da igual la dirección; en "Volver" no hace nada.
func (v *settingsView) change() {
	if v.sel != setBack {
		v.toggle(v.sel)
	}
}

// Update atiende flechas, Enter, Esc y mouse. Devuelve true al cerrar.
func (v *settingsView) Update(in *Input) (done bool) {
	// Mouse: marca la fila bajo el cursor cuando se mueve.
	v.hovered = ""
	hover := -1
	for i := 0; i < setBack; i++ {
		if settingsRowRect(i).contains(in.MouseX, in.MouseY) {
			hover = i
		}
	}
	if settingsBackRect.contains(in.MouseX, in.MouseY) {
		hover = setBack
	}
	if hover >= 0 {
		v.hovered = settingsLabels[hover]
		if in.Moved {
			v.sel = hover
		}
	}

	switch {
	case in.Pressed("Escape"):
		return true
	case in.Pressed("Up"):
		v.sel = stepRow(v.sel, -1)
	case in.Pressed("Down"):
		v.sel = stepRow(v.sel, 1)
	case in.Pressed("Left"):
		v.change()
	case in.Pressed("Right"):
		v.change()
	case in.Pressed("Enter"), in.Pressed("Space"):
		if v.sel == setBack {
			return true
		}
		if v.sel >= setBlinkSleep {
			v.toggle(v.sel)
		}
	case in.Click && hover >= 0:
		v.sel = hover
		switch hover {
		case setBack:
			return true
		default:
			if settingsToggle(hover).contains(in.MouseX, in.MouseY) {
				v.toggle(hover)
			}
		}
	}
	return false
}

// State sirve para el volcado de pruebas.
func (v *settingsView) State() string { return settingsLabels[v.sel] }

// Draw pinta los ajustes encima de lo que haya (oscurece el fondo primero).
func (v *settingsView) Draw(dst *ebiten.Image) {
	e := v.env
	s := e.Save.Settings
	e.FillRect(dst, rect{0, 0, screenW, screenH}, '0', 0.45)
	drawBox(e, dst, settingsBox)
	e.Text(dst, "Ajustes", settingsBox.x+12, settingsBox.y+6, textStyle{fontPixelBold, 18, '1', alignLeft})

	for i := 0; i < setBack; i++ {
		r := settingsRowRect(i)
		key := byte('1')
		if i == v.sel {
			key = '0'
			e.FillRect(dst, r, 'm', 1)
			e.DrawSprite(dst, e.Assets.Frame("ui_flecha", 2, animFrame(e.Tick, 24, 2, 0)), r.x+3, r.y+3, false)
		}
		e.Text(dst, settingsLabels[i], r.x+14, r.y+2, textStyle{fontPixel, 10, key, alignLeft})
		switch i {
		case setBlinkSleep:
			v.drawToggle(dst, i, s.BlinkOnSleep)
		case setBlinkWake:
			v.drawToggle(dst, i, s.BlinkOnWake)
		case setFullscreen:
			v.drawToggle(dst, i, s.Fullscreen)
		case setSound:
			v.drawToggle(dst, i, !s.SoundOff)
		}
	}

	back := settingsBackRect
	fill := byte('k')
	if v.sel == setBack {
		fill = 'm'
	}
	e.FillRect(dst, back, fill, 1)
	e.StrokeRect(dst, back, '0', 1)
	e.Text(dst, settingsLabels[setBack], back.x+back.w/2, back.y+3, textStyle{fontPixel, 10, '0', alignCenter})

	e.FillRect(dst, rect{settingsBox.x + 8, 200, 244, 1}, 'k', 1)
	e.Text(dst, settingsDescs[v.sel], settingsBox.x+8, 205, textStyle{fontPixel, 8, '2', alignLeft})
	e.Text(dst, settingsHelp, settingsBox.x+8, 227, textStyle{fontPixel, 8, 'x', alignLeft})
}

// drawToggle pinta el botón Apagado / Prendido.
func (v *settingsView) drawToggle(dst *ebiten.Image, row int, on bool) {
	e := v.env
	r := settingsToggle(row)
	fill := byte('x')
	if on {
		fill = 'q'
	}
	e.FillRect(dst, r, fill, 1)
	e.StrokeRect(dst, r, '0', 1)
	e.Text(dst, onOff(on), r.x+r.w/2-2, r.y+2, textStyle{fontPixel, 8, 'w', alignCenter})
}

// settingsChanged autoguarda los ajustes y aplica lo que se puede aplicar ya
// (la pantalla completa). En la puerta de pruebas no se cambia la ventana de verdad.
func (e *Env) settingsChanged() {
	e.persist()
	if !e.Opts.Agent {
		ebiten.SetFullscreen(e.Save.Settings.Fullscreen)
	}
}
