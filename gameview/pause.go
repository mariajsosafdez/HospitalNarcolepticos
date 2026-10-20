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

// blinkPatient es lo único que el parpadeo necesita saber de un paciente:
// quién es, en qué cámara está y en qué estado.
type blinkPatient struct {
	ID    string
	Cam   game.CameraID
	State game.PatientState
}

const (
	blinkPeriodTicks = 40  // un titileo cada 0,67 s
	blinkMaxAlpha    = 0.7 // "suave": nunca tapa del todo el botón
)

// blinkAlpha es la opacidad del titileo en un tick de la interfaz: sube y baja
// como un triángulo, sin parar (el parpadeo ya no se acaba solo: dura mientras
// haya un motivo sin atender). Es pura para probarla con una tabla.
func blinkAlpha(tick int) float64 {
	if tick < 0 {
		return 0
	}
	half := blinkPeriodTicks / 2
	phase := tick % blinkPeriodTicks
	if phase > half {
		phase = blinkPeriodTicks - phase
	}
	return blinkMaxAlpha * float64(phase) / float64(half)
}

// reasonKey es la clave de un motivo de parpadeo ("s" dormido, "w" despierto)
// para guardarlo en el conjunto de motivos ya vistos.
func reasonKey(wake bool, id string) string {
	if wake {
		return "w:" + id
	}
	return "s:" + id
}

// blinkReasons decide qué botones del mapa parpadean. Devuelve cámara -> true
// si el color es menta (despertó) y false si es mango (dormido). Reglas:
//   - motivo de sueño: paciente AsleepWaiting o NoBedWaiting en esa cámara;
//   - motivo de despertar: paciente AwakeInBed en esa cámara;
//   - cada ajuste (sleepOn, wakeOn) prende solo su motivo;
//   - un motivo que ya está en seen (el jugador ya miró esa cámara) no cuenta;
//   - la cámara que se está mirando nunca parpadea;
//   - si una cámara tiene los dos motivos, gana menta (despertar), por regla fija.
func blinkReasons(ps []blinkPatient, current game.CameraID, seen map[string]bool, sleepOn, wakeOn bool) map[game.CameraID]bool {
	out := make(map[game.CameraID]bool)
	for _, p := range ps {
		if p.Cam == current {
			continue
		}
		switch {
		case wakeOn && p.State == game.AwakeInBed && !seen[reasonKey(true, p.ID)]:
			out[p.Cam] = true
		case sleepOn && (p.State == game.AsleepWaiting || p.State == game.NoBedWaiting) && !seen[reasonKey(false, p.ID)]:
			if _, already := out[p.Cam]; !already {
				out[p.Cam] = false
			}
		}
	}
	return out
}

// updateSeen actualiza los motivos ya vistos: marca como vistos todos los de
// la cámara actual y olvida los que ya no existen (así un paciente que se
// vuelve a dormir sí avisa otra vez).
func updateSeen(ps []blinkPatient, current game.CameraID, seen map[string]bool) {
	live := make(map[string]bool)
	for _, p := range ps {
		var key string
		switch p.State {
		case game.AwakeInBed:
			key = reasonKey(true, p.ID)
		case game.AsleepWaiting, game.NoBedWaiting:
			key = reasonKey(false, p.ID)
		default:
			continue
		}
		live[key] = true
		if p.Cam == current {
			seen[key] = true
		}
	}
	for key := range seen {
		if !live[key] {
			delete(seen, key)
		}
	}
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

// trackBlinks recalcula cada tick qué botones del mapa parpadean (ver
// blinkReasons) y toca los sonidos del alta y del que despierta mirando.
func (c *camsScene) trackBlinks() {
	if c.seen == nil {
		c.seen = make(map[string]bool)
	}
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
		_, woke := stateEvent(prev, g.State())
		if woke && g.Camera() == c.shift.CurrentCamera() {
			c.env.Sound.Play(CueWake)
		}
	}
	c.prev = cur
	c.refreshBlinks()
}

// blinkPatients saca del turno la foto que necesita blinkReasons.
func (c *camsScene) blinkPatients() []blinkPatient {
	var ps []blinkPatient
	for _, g := range c.shift.Patients() {
		ps = append(ps, blinkPatient{g.Patient().ID(), g.Camera(), g.State()})
	}
	return ps
}

// refreshBlinks marca como vistos los motivos de la cámara actual y guarda en
// c.blinks los botones que deben parpadear ahora.
func (c *camsScene) refreshBlinks() {
	if c.seen == nil {
		c.seen = make(map[string]bool)
	}
	ps, cam := c.blinkPatients(), c.shift.CurrentCamera()
	updateSeen(ps, cam, c.seen)
	set := c.env.Save.Settings
	c.blinks = blinkReasons(ps, cam, c.seen, set.BlinkOnSleep, set.BlinkOnWake)
}
