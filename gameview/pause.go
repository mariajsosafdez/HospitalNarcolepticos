package gameview

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

// Pausa, según la vista aprobada "Pausa" (frame 18:263 de Figma, nota 18:535;
// copia en preview/turno/pausa.png).

// Opciones de la pausa.
const (
	pauseContinue = iota
	pauseSettings
	pauseExit
	pauseItems
)

var pauseLabels = [pauseItems]string{"Continuar", "Ajustes", "Salir al menú"}

// pauseDescs es lo que dice la línea bajo la lista con cada opción marcada.
// "Salir al menú" avisa que el turno no se guarda (decisión de la autora).
var pauseDescs = [pauseItems]string{
	"El turno sigue donde iba.",
	"Parpadeo, sonido y pantalla.",
	"Ojo: este turno no cuenta.",
}

// pauseBox es la caja: (160, 60) 160x146.
var pauseBox = rect{160, 60, 160, 146}

// pauseItemRect es la barra de la opción i: (166, 94 + 18i) 148x15.
func pauseItemRect(i int) rect { return rect{166, 94 + 18*i, 148, 15} }

// pauseItemAt dice qué opción está bajo el punto (-1 = ninguna).
func pauseItemAt(x, y int) int {
	for i := 0; i < pauseItems; i++ {
		if pauseItemRect(i).contains(x, y) {
			return i
		}
	}
	return -1
}

// stepPause mueve la opción marcada, dando la vuelta en los extremos.
func stepPause(sel, dir int) int { return ((sel+dir)%pauseItems + pauseItems) % pauseItems }

// drawPause pinta la pausa encima de la cámara oscurecida. Si los ajustes
// están abiertos, estos se pintan encima (ver Draw).
func (c *camsScene) drawPause(dst *ebiten.Image) {
	e := c.env
	e.FillRect(dst, rect{0, 0, screenW, screenH}, '0', 0.45)
	drawBox(e, dst, pauseBox)
	e.Text(dst, "Pausa", pauseBox.x+12, pauseBox.y+6, textStyle{fontPixelBold, 18, '1', alignLeft})
	for i := 0; i < pauseItems; i++ {
		r := pauseItemRect(i)
		key := byte('1')
		if i == c.pauseSel {
			key = '0'
			e.FillRect(dst, r, 'm', 1)
			e.DrawSprite(dst, e.Assets.Frame("ui_flecha", 2, animFrame(e.Tick, 24, 2, 0)), r.x+3, r.y+3, false)
		}
		e.Text(dst, pauseLabels[i], r.x+16, r.y+2, textStyle{fontPixel, 10, key, alignLeft})
	}
	e.FillRect(dst, rect{pauseBox.x + 8, 152, 144, 1}, 'k', 1)
	e.Text(dst, pauseDescs[c.pauseSel], pauseBox.x+8, 157, textStyle{fontPixel, 8, '2', alignLeft})
	e.Text(dst, fmt.Sprintf("Reloj parado en las %s", c.shift.Clock().Now()), pauseBox.x+8, 169, textStyle{fontPixel, 8, 'x', alignLeft})
	e.Text(dst, "Esc: seguir jugando", pauseBox.x+8, 191, textStyle{fontPixel, 8, 'x', alignLeft})
}

// ---------------------------------------------------------------------------
// Parpadeo de los botones del mapa (opciones de Ajustes)
// ---------------------------------------------------------------------------

// blink es un botón del mapa que está parpadeando: desde qué tick de la
// interfaz y si fue porque alguien despertó (si no, porque se durmió).
type blink struct {
	start int
	wake  bool
}

const (
	blinkTotalTicks  = 5 * game.TicksPerSecond // dura 5 s
	blinkPeriodTicks = 40                      // un titileo cada 0,67 s
	blinkMaxAlpha    = 0.7                     // "suave": nunca tapa del todo el botón
)

// blinkAlpha es la opacidad del titileo a los elapsed ticks de empezar: sube y
// baja como un triángulo y vale 0 antes de empezar y después de los 5 s. Es
// pura para probarla con una tabla.
func blinkAlpha(elapsed int) float64 {
	if elapsed < 0 || elapsed >= blinkTotalTicks {
		return 0
	}
	half := blinkPeriodTicks / 2
	phase := elapsed % blinkPeriodTicks
	if phase > half {
		phase = blinkPeriodTicks - phase
	}
	return blinkMaxAlpha * float64(phase) / float64(half)
}

// blinkKey es el color del titileo: mango al dormirse, menta al despertar.
func blinkKey(wake bool) byte {
	if wake {
		return 'A'
	}
	return 'm'
}

// stateEvent dice si el cambio de estado de un paciente es un "se durmió"
// (de pie a dormido) o un "se despertó" (de la cama a despierto en la cama).
func stateEvent(prev, cur game.PatientState) (slept, woke bool) {
	slept = cur == game.AsleepWaiting && (prev == game.Wandering || prev == game.Arriving)
	woke = prev == game.InBed && cur == game.AwakeInBed
	return slept, woke
}

// trackBlinks compara el estado de cada paciente con el del tick anterior y,
// si el ajuste está prendido, hace parpadear el botón de la cámara donde
// alguien se durmió o se despertó (solo si no es la cámara que se mira).
func (c *camsScene) trackBlinks() {
	for cam, b := range c.blinks { // los parpadeos terminados se quitan
		if c.env.Tick-b.start >= blinkTotalTicks {
			delete(c.blinks, cam)
		}
	}
	set := c.env.Save.Settings
	cur := make(map[string]game.PatientState)
	for _, g := range c.shift.Patients() {
		id := g.Patient().ID()
		cur[id] = g.State()
		prev, seen := c.prev[id]
		if !seen {
			continue
		}
		// Sonidos: el "ding" del alta (la dé el jugador o la Demo) y el pop de
		// quien despierta en la cámara que se está mirando.
		if prev == game.AwakeInBed && g.State() == game.Gone {
			c.env.Sound.Play(CueDischarge)
		}
		slept, woke := stateEvent(prev, g.State())
		if g.Camera() == c.shift.CurrentCamera() {
			if woke {
				c.env.Sound.Play(CueWake)
			}
			continue
		}
		switch {
		case slept && set.BlinkOnSleep:
			c.addBlink(g.Camera(), false)
		case woke && set.BlinkOnWake:
			c.addBlink(g.Camera(), true)
		}
	}
	c.prev = cur
}

// addBlink empieza (o reinicia) el parpadeo del botón de una cámara.
func (c *camsScene) addBlink(cam game.CameraID, wake bool) {
	if c.blinks == nil {
		c.blinks = make(map[game.CameraID]blink)
	}
	c.blinks[cam] = blink{start: c.env.Tick, wake: wake}
}
