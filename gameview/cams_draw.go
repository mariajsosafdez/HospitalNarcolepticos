package gameview

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
	"costenos-narcolepsia/hospital"
)

// Dibujo de la pantalla de cámaras. Las capas van de atrás hacia adelante,
// en el orden de la nota de Figma "Cam / zona común" (frame 9:205).

// Rectángulos fijos de la pantalla (a 1x).
var (
	demoBoxRect  = rect{8, 200, 328, 62}
	demoNextRect = rect{210, 244, 70, 13}
	demoAutoRect = rect{286, 244, 42, 13}
	noticeRect   = rect{60, 203, 220, 46}
	noticeTopY   = 36  // arriba: debajo del rótulo REC
	noticeDemoY  = 140 // en la Demo se sube para no tapar la caja de controles
)

// dischargeRect es el botón "Dar de alta", centrado bajo la cama. Va en y = 208
// (94 del fondo): entre la cama (termina en 84 a 86) y las alfombras de las
// habitaciones 101 y 103 (empiezan en 103 y 109), o sea sobre piso liso en las
// seis habitaciones.
func dischargeRect(bed point) rect {
	sx, _ := feedToScreen(bed.x+bedW/2, 0)
	return rect{sx - 30, 208, 60, 13}
}

// bed devuelve la cama de la habitación que se está mirando.
func (c *camsScene) bed() point {
	info, _ := game.CameraByID(c.shift.CurrentCamera())
	return roomLayout(info.RoomNumber).bed
}

// La cama de las habitaciones: UNA por cuarto (ver roomBed). Sprite de 32x30.
const (
	bedW = 32
	bedH = 30
)

// placement es un paciente ya colocado en la cámara que se mira: qué sprite,
// dónde (coordenadas del fondo) y con qué efectos.
type placement struct {
	patient *game.GamePatient
	imgName string // PNG del sprite ("" = sin arte todavía: placeholder)
	frames  int
	index   int
	x, y    int // esquina de arriba a la izquierda, en coordenadas del fondo
	w, h    int
	flip    bool
	footY   float64
	zzz     bool
	alert   bool
	effectX int // esquina del efecto (zzz o !), coordenadas del fondo
	effectY int
}

// screenRect es el rectángulo del sprite en la pantalla (a x2).
func (p placement) screenRect() rect {
	sx, sy := feedToScreen(p.x, p.y)
	return rect{sx, sy, p.w * feedScale, p.h * feedScale}
}

// placements calcula dónde y cómo se dibuja cada paciente de la cámara
// actual. Lo usan el dibujo y el clic del mouse, así coinciden siempre.
func (c *camsScene) placements() []placement {
	a := c.env.Assets
	cam := c.shift.CurrentCamera()
	info, _ := game.CameraByID(cam)

	var out []placement

	for _, g := range c.shift.Patients() {
		if g.Camera() != cam {
			continue
		}
		state := g.State()
		inRoom := state == game.InBed || state == game.AwakeInBed
		if inRoom != (info.Kind == game.RoomCam) && state != game.Carried {
			continue
		}
		w := c.walkers[g.Patient().ID()]
		if w == nil {
			continue
		}

		// En camilla: se dibuja el personal empujándola, no el paciente.
		if state == game.Carried {
			m := g.Carrier()
			if m == nil {
				continue
			}
			out = append(out, c.carryPlacement(g, m, w))
			continue
		}

		spec, ok := patientSprite(state, w.walking)
		if !ok {
			continue
		}
		name := "pac_" + spriteKey(g.Patient().ID(), g.Patient().Name()) + "_" + spec.suffix
		p := placement{patient: g, frames: spec.frames, flip: w.left}
		p.index = spriteFrame(spec, c.env.Tick)
		if a.Has(name) {
			p.imgName = name
			p.w, p.h = a.FrameSize(name, spec.frames)
		} else {
			p.w, p.h = 16, 24 // placeholder: una figura simple
			if inRoom {
				p.w, p.h = bedW, bedH
			}
		}
		if inRoom {
			bed := roomLayout(info.RoomNumber).bed
			p.x, p.y = bed.x, bed.y
			p.footY = float64(p.y + p.h)
			p.flip = false
			if state == game.InBed {
				p.zzz, p.effectX, p.effectY = true, bed.x+12, bed.y-12
			} else {
				p.alert, p.effectX, p.effectY = true, bed.x+8, bed.y-16
			}
		} else {
			p.x, p.y = int(math.Round(w.x))-p.w/2, int(math.Round(w.y))-p.h
			p.footY = w.y
			if state == game.AsleepWaiting || state == game.NoBedWaiting {
				p.flip = false
				// Dormido en el piso: la fila de abajo del sprite toca el piso
				// (los pies del walker) y los sprites anchos no se salen de
				// la cámara.
				p.x = min(max(p.x, 0), feedW-p.w)
				fxW, _ := a.FrameSize("fx_zzz", 3)
				p.zzz = true
				p.effectX, p.effectY = zzzSpot(p.x, p.y, p.w, p.h, fxW)
			}
		}
		out = append(out, p)
	}
	sortPlacements(out)
	return out
}

// carryPlacement coloca al personal empujando la camilla donde estaba el
// paciente. Usa el sprite <doc|cam>_<nombre>_camilla (4 cuadros de 44x24).
func (c *camsScene) carryPlacement(g *game.GamePatient, m *game.StaffMember, w *walker) placement {
	a := c.env.Assets
	prefix := "cam_"
	if m.IsDoctor() {
		prefix = "doc_"
	}
	name := prefix + spriteKey(m.ID(), m.Name()) + "_camilla"
	p := placement{patient: g, frames: 4, flip: w.left}
	p.index = animFrame(c.env.Tick, 8, 4, 0)
	if a.Has(name) {
		p.imgName = name
		p.w, p.h = a.FrameSize(name, 4)
	} else {
		p.w, p.h = 44, 24
	}
	p.x, p.y = int(math.Round(w.x))-p.w/2, int(math.Round(w.y))-p.h
	p.footY = w.y
	if b, ok := c.demoBeat(); ok && b.Kind == game.BeatCarry {
		// En la Demo la camilla sale caminando de la cámara, por el lado más
		// cercano, al ritmo del traslado (CooldownFraction va de 0 a 1).
		left := entersFromLeft(w.x)
		p.x = int(math.Round(stretcherX(w.x, left, p.w, m.CooldownFraction()))) - p.w/2
		p.flip = left // el sprite mira a la derecha: yendo a la izquierda se refleja
	}
	return p
}

// demoBeat devuelve el beat de la Demo que se está animando (ok es false si
// no es la Demo o está esperando).
func (c *camsScene) demoBeat() (game.Beat, bool) {
	if c.demo == nil {
		return game.Beat{}, false
	}
	return c.demo.Beat()
}

// stagedStaff pinta al personal que entra caminando, espera junto al paciente
// o se va a descansar (beats Arrive, NoBed y Leave de la Demo). El que se
// lleva la camilla ya sale en carryPlacement.
func (c *camsScene) drawStagedStaff(feed *ebiten.Image, placed []placement) {
	b, ok := c.demoBeat()
	if !ok || (b.Kind != game.BeatArrive && b.Kind != game.BeatNoBed && b.Kind != game.BeatLeave) {
		return
	}
	m := c.shift.FindStaff(b.Staff)
	if m == nil {
		return
	}
	for _, target := range placed {
		if target.patient.Patient().ID() != b.Patient {
			continue
		}
		e := c.env
		prefix := "cam_"
		if m.IsDoctor() {
			prefix = "doc_"
		}
		base := prefix + spriteKey(m.ID(), m.Name())
		left := entersFromLeft(float64(target.x + target.w/2))
		stand := standX(target.x, target.w, left, 16)
		x, walking := stagedX(b.Kind, c.demo.BeatProgress(), stand, offscreenX(left, 16))
		name, frames, index := base+"_quieto", 2, animFrame(e.Tick, 30, 2, 0)
		if walking {
			name, frames, index = base+"_camina", 4, animFrame(e.Tick, 9, 4, 0)
		}
		if !e.Assets.Has(name) {
			return
		}
		w, h := e.Assets.FrameSize(name, frames)
		// Mira hacia donde camina; de pie, hacia el paciente. Los sprites miran a la derecha.
		facingRight := left
		if b.Kind == game.BeatLeave {
			facingRight = !left
		}
		sx := int(math.Round(x)) - w/2
		sy := int(math.Round(target.footY)) + 1 - h // un pixel delante del paciente
		e.DrawSprite(feed, e.Assets.Frame(name, frames, index), sx, sy, !facingRight)
		return
	}
}

// Draw pinta toda la pantalla de cámaras.
func (c *camsScene) Draw(dst *ebiten.Image) {
	e := c.env
	tick := c.shift.Clock().Tick()
	light := lightFrame(tick)
	cam := c.shift.CurrentCamera()
	camInfo, _ := game.CameraByID(cam)

	// 1. Fondo de la pantalla y marco de la cámara.
	e.FillRect(dst, rect{0, 0, screenW, screenH}, '1', 1)
	e.FillRect(dst, rect{3, 19, 338, 248}, '0', 1)

	// 2. Barra superior.
	c.drawTopBar(dst)

	// 3 y 4. La cámara: fondo, camas y pacientes a 168x123, luego a x2.
	hasBackground := c.drawFeed(light, cam, camInfo)
	e.drawImage(dst, c.feed, feedScreenX, feedScreenY, feedScale, false, 1)
	if !hasBackground {
		e.Text(dst, "Fondo pendiente", feedScreenX+168, 70, textStyle{fontPixel, 8, '2', alignCenter})
	}
	if camInfo.Kind == game.RoomCam {
		// El número de la habitación va en la placa de la puerta: centrado en
		// la placa (9 px de ancho) y 2 px de pantalla por encima de su borde.
		plate := roomLayout(camInfo.RoomNumber).plate
		sx, sy := feedToScreen(plate.x, plate.y)
		e.Text(dst, fmt.Sprintf("%d", camInfo.RoomNumber), sx+feedScale*9/2, sy-2, textStyle{fontPixel, 8, '1', alignCenter})
	}

	// 5. Encima de la cámara: tinte por hora, viñeta y grano.
	t := tintForFrame(light)
	e.FillRect(dst, feedRect, t.key, t.alpha)
	c.drawOverlay(dst, "camara_vineta", 1, 0, 0.35)
	c.drawOverlay(dst, "camara_grano", 2, animFrame(e.Tick, 6, 2, 0), 0.12)

	// 6. Rótulo de la cámara: REC, sello y lugar, con borde oscuro.
	c.drawStamp(dst, cam, camInfo)

	// Etiqueta del paciente elegido y ayuda.
	if g := c.shift.FindPatient(c.selected); g != nil {
		c.drawSelectedLabel(dst, g)
	}
	if c.selected != "" {
		// Con borde oscuro como el rótulo, y pegado a la derecha para no
		// chocar con el nombre del lugar (el más largo: "Patio de las palmeras").
		e.TextOutline(dst, "Elige a alguien del panel · Esc para cancelar", 331, 253, textStyle{fontPixel, 8, 'w', alignRight})
	}
	if c.demo == nil {
		c.drawDischargeButton(dst)
	}

	// 7 y 8. Panel de personal y mapa.
	c.drawStaffPanel(dst)
	c.drawMap(dst)

	if c.demo != nil {
		c.drawDemoBox(dst)
		c.drawDischargeButton(dst) // sobre la caja: en la Demo el alta se ve pulsarse sola
	}
	if c.notice != nil {
		c.drawNotice(dst)
	}
	if c.msg != nil {
		c.drawMessage(dst)
	}
	if c.paused {
		c.drawPause(dst)
	}
	if c.settings != nil {
		c.settings.Draw(dst)
	}
}

// drawTopBar pinta la barra de arriba: día, turno, hora y estrellas.
func (c *camsScene) drawTopBar(dst *ebiten.Image) {
	e := c.env
	e.DrawSprite(dst, e.Assets.Image("ui_barra_superior"), 0, 0, false)
	title, subtitle := fmt.Sprintf("Día %d", c.shift.Day()), "Turno de 12 a 6"
	if c.demo != nil {
		title, subtitle = "Demo", "Escenario del enunciado"
	}
	e.Text(dst, title, 8, 2, textStyle{fontPixelBold, 10, 'M', alignLeft})
	e.Text(dst, subtitle, 50, 4, textStyle{fontPixel, 8, 'k', alignLeft})
	e.Text(dst, c.shift.Clock().Now(), 240, 2, textStyle{fontVT, 12, 'w', alignCenter})
	stars := c.shift.Stars()
	for i := 0; i < 5; i++ {
		frame := 1 // vacía
		if i < stars {
			frame = 0
		}
		e.DrawSprite(dst, e.Assets.Frame("ui_estrella", 2, frame), 430+10*i, 4, false)
	}
}

// drawFeed pinta el fondo, las camas y los pacientes en c.feed (168x123).
// Devuelve false si el fondo de esta cámara todavía no está dibujado.
func (c *camsScene) drawFeed(light int, cam game.CameraID, info game.Camera) bool {
	e := c.env
	a := e.Assets
	feed := c.feed
	feed.Clear()
	// La imagen de la cámara es chica (168x123): se pinta a 1x y luego se
	// agranda junto con el resto de la pantalla.
	saved := e.Scale
	e.Scale = 1.0
	defer func() { e.Scale = saved }()

	bg := backgroundName(cam, a.Names())
	hasBackground := bg != ""
	if hasBackground {
		w, _ := a.Size(bg)
		e.DrawSprite(feed, a.Frame(bg, max(w/feedW, 1), backgroundFrame(w, light)), 0, 0, false)
	} else {
		e.FillRect(feed, rect{0, 0, feedW, feedH}, 'W', 1)
		e.FillRect(feed, rect{0, floorRect.y - 4, feedW, feedH - floorRect.y + 4}, 'k', 1)
	}

	placed := c.placements()
	if info.Kind == game.RoomCam && !roomHasBedPatient(placed) {
		bed := c.bed()
		e.DrawSprite(feed, a.Image("cama"), bed.x, bed.y, false)
	}
	for _, p := range placed {
		c.drawPlacement(feed, p)
	}
	c.drawStagedStaff(feed, placed)
	return hasBackground
}

// roomHasBedPatient dice si hay un paciente dibujado en la cama.
func roomHasBedPatient(list []placement) bool {
	for _, p := range list {
		if st := p.patient.State(); st == game.InBed || st == game.AwakeInBed {
			return true
		}
	}
	return false
}

// drawPlacement pinta un paciente y sus efectos en el lienzo de la cámara.
func (c *camsScene) drawPlacement(feed *ebiten.Image, p placement) {
	e := c.env
	a := e.Assets
	if p.imgName != "" {
		// Contorno del elegido: parpadea cada 0,4 s (24 ticks).
		if p.patient.Patient().ID() == c.selected && animFrame(e.Tick, 24, 2, 0) == 0 {
			sil := a.Silhouette(p.imgName, p.frames, p.index, 'y')
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				e.DrawSprite(feed, sil, p.x+d[0], p.y+d[1], p.flip)
			}
		}
		e.DrawSprite(feed, a.Frame(p.imgName, p.frames, p.index), p.x, p.y, p.flip)
	} else {
		c.drawPlaceholder(feed, p)
	}
	if p.zzz {
		e.DrawSprite(feed, a.Frame("fx_zzz", 3, animFrame(e.Tick, 30, 3, 0)), p.effectX, p.effectY, false)
	}
	if p.alert {
		e.DrawSprite(feed, a.Frame("fx_alerta", 2, animFrame(e.Tick, 18, 2, 0)), p.effectX, p.effectY, false)
	}
}

// drawPlaceholder pinta una figura simple para los personajes que todavía no
// tienen sprite (P-006 a P-010). Se cambia sola cuando aparece el PNG.
func (c *camsScene) drawPlaceholder(feed *ebiten.Image, p placement) {
	e := c.env
	body := rect{p.x + 2, p.y + 2, p.w - 4, p.h - 2}
	e.FillRect(feed, rect{body.x - 1, body.y - 1, body.w + 2, body.h + 2}, '0', 1)
	e.FillRect(feed, body, 'c', 1)
	e.FillRect(feed, rect{body.x, body.y, body.w, min(8, body.h)}, 'C', 1)
	if p.patient.State() == game.AsleepWaiting || p.patient.State() == game.NoBedWaiting {
		// Dormido: acostado de lado, más bajo y ancho.
		e.FillRect(feed, rect{p.x, p.y + p.h - 10, p.w + 4, 10}, '0', 1)
		e.FillRect(feed, rect{p.x + 1, p.y + p.h - 9, p.w + 2, 8}, 'c', 1)
	}
}

// drawOverlay dibuja una capa de la cámara (viñeta, grano) con su opacidad.
func (c *camsScene) drawOverlay(dst *ebiten.Image, name string, frames, frame int, alpha float32) {
	img := c.env.Assets.Frame(name, frames, frame)
	if img == nil {
		return
	}
	c.env.drawImage(dst, img, feedScreenX, feedScreenY, 1, false, alpha)
}

// drawStamp pinta el rótulo de la cámara con borde oscuro de 1 px.
func (c *camsScene) drawStamp(dst *ebiten.Image, cam game.CameraID, info game.Camera) {
	e := c.env
	e.DrawSprite(dst, e.Assets.Frame("ui_rec", 2, animFrame(e.Tick, 60, 2, 0)), 10, 26, false)
	e.TextOutline(dst, "REC", 19, 24, textStyle{fontVT, stampFontSize, 'w', alignLeft})
	e.TextOutline(dst, cameraStamp(cam, c.shift.Clock().Now()), 327, 24, textStyle{fontVT, stampFontSize, 'w', alignRight})
	e.TextOutline(dst, info.Name, 10, 253, textStyle{fontVT, stampFontSize, 'w', alignLeft})
}

// levelChip devuelve el texto y el color de la manilla de un nivel.
func levelChip(l hospital.NarcolepsyLevel) (string, byte) {
	switch l {
	case hospital.Mild:
		return "Leve", 'L'
	case hospital.Moderate:
		return "Moderado", 'Y'
	default:
		return "Severo", 'R'
	}
}

// drawSelectedLabel pinta la caja con el nombre y el nivel junto al paciente.
func (c *camsScene) drawSelectedLabel(dst *ebiten.Image, g *game.GamePatient) {
	e := c.env
	var sprite rect
	found := false
	for _, p := range c.placements() {
		if p.patient == g {
			sprite, found = p.screenRect(), true
		}
	}
	if !found {
		return
	}
	box := labelPosition(sprite, feedRect)
	e.FillRect(dst, box, 'W', 1)
	e.StrokeRect(dst, box, '0', 1)
	name := patientLabelName(g.Patient().ID(), g.Patient().Name())
	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	name = fitText(name, 60, measure)
	e.Text(dst, name, box.x+4, box.y+2, textStyle{fontPixel, 8, '1', alignLeft})
	text, key := levelChip(g.Patient().Level())
	chip := rect{box.x + 64, box.y + 2, 48, 10}
	e.FillRect(dst, chip, key, 1)
	e.StrokeRect(dst, chip, '0', 1)
	e.Text(dst, text, chip.x+chip.w/2, chip.y, textStyle{fontPixel, 8, '0', alignCenter})
}

// drawButton pinta un botón: relleno, borde 0, sombra de 1 px abajo y texto.
func (c *camsScene) drawButton(dst *ebiten.Image, r rect, label string, fill byte, hover bool) {
	c.drawButtonState(dst, r, label, fill, hover, false)
}

// drawButtonState es drawButton con la opción de dibujarlo hundido (pulsado):
// baja 1 px y pierde la sombra.
func (c *camsScene) drawButtonState(dst *ebiten.Image, r rect, label string, fill byte, hover, pressed bool) {
	e := c.env
	if pressed {
		r.y++
	} else {
		e.FillRect(dst, rect{r.x, r.y + 1, r.w, r.h}, '0', 1)
	}
	e.FillRect(dst, r, fill, 1)
	if hover {
		e.FillRect(dst, r, 'w', 0.25)
	}
	e.StrokeRect(dst, r, '0', 1)
	e.Text(dst, label, r.x+r.w/2, r.y+1, textStyle{fontPixel, 8, '0', alignCenter})
}

// ---------------------------------------------------------------------------
// Panel de personal
// ---------------------------------------------------------------------------

func (c *camsScene) drawStaffPanel(dst *ebiten.Image) {
	e := c.env
	title := "Personal"
	if c.selected != "" {
		title = "¿Quién lo atiende?"
	}
	e.Text(dst, title, 346, 20, textStyle{fontPixelBold, 10, 'M', alignLeft})

	staff := c.shift.Staff()
	compact := compactCards(len(staff))
	for i, m := range staff {
		r := staffCardRect(i, len(staff))
		state := m.State()
		if c.selected != "" && !compact {
			// Tarjeta normal: borde de 2 px por fuera.
			if state == game.StaffAvailable {
				e.FillRect(dst, staffCardHighlight(i), 'm', 1)
			}
			if i == c.cursor {
				e.FillRect(dst, staffCardHighlight(i), 'w', 1) // la marca del teclado
			}
		}
		if compact {
			c.drawCompactCard(dst, m, r, state)
			if c.selected != "" {
				// Las tarjetas compactas se tocan entre sí: el borde de 2 px va por dentro.
				if state == game.StaffAvailable {
					e.StrokeRect(dst, r, 'm', 2)
				}
				if i == c.cursor {
					e.StrokeRect(dst, r, 'w', 2)
				}
			}
		} else {
			c.drawNormalCard(dst, m, r, state)
		}
	}
}

// drawNormalCard pinta la tarjeta de 132x36 (hasta 3 personas).
func (c *camsScene) drawNormalCard(dst *ebiten.Image, m *game.StaffMember, r rect, state game.StaffState) {
	e := c.env
	a := e.Assets
	e.DrawSprite(dst, a.Image("ui_tarjeta"), r.x, r.y, false)
	c.drawHead(dst, m, r.x+4, r.y+4)

	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	e.Text(dst, fitText(displayName(m.ID(), m.Name()), 82, measure), r.x+24, r.y+2, textStyle{fontPixel, 8, '1', alignLeft})
	e.Text(dst, fitText(staffRole(m), 82, measure), r.x+24, r.y+10, textStyle{fontPixel, 8, '2', alignLeft})
	if m.IsDoctor() {
		quota := fmt.Sprintf("%d/%d", m.PatientCount(), game.DoctorQuota)
		if state == game.StaffNoQuota {
			e.Text(dst, quota, r.x+108, r.y+2, textStyle{fontPixelBold, 8, 'R', alignLeft})
		} else {
			e.Text(dst, quota, r.x+108, r.y+2, textStyle{fontPixel, 8, '1', alignLeft})
		}
	}
	e.DrawSprite(dst, a.Frame("ui_chip_estado", 4, chipFrame(state)), r.x+24, r.y+22, false)
	c.drawChipText(dst, state, r.x+26, r.y+22)
	e.FillRect(dst, rect{r.x + 86, r.y + 25, 40, 4}, 'k', 1)
	e.FillRect(dst, rect{r.x + 86, r.y + 25, int(40 * m.CooldownFraction()), 4}, 'q', 1)
}

// drawCompactCard pinta la tarjeta de 132x24 (4 o 5 personas, días 3 a 5),
// con el nombre y el rol cortos. Medidas leídas de la vista "Panel / 5
// personas" (preview/turno/panel-5-personas.png): el chip de estado (58x10)
// va a la derecha y la barra de cooldown (58x3) justo debajo.
func (c *camsScene) drawCompactCard(dst *ebiten.Image, m *game.StaffMember, r rect, state game.StaffState) {
	e := c.env
	a := e.Assets
	e.DrawSprite(dst, a.Image("ui_tarjeta_compacta"), r.x, r.y, false)
	c.drawHead(dst, m, r.x+3, r.y+6)

	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	e.Text(dst, fitText(shortName(m.ID(), m.Name()), 80, measure), r.x+22, r.y+1, textStyle{fontPixel, 8, '1', alignLeft})
	role := "Camillero"
	if m.IsDoctor() {
		role = shortRole(m.ID(), m.Doctor().Specialty())
	}
	e.Text(dst, fitText(role, 46, measure), r.x+22, r.y+12, textStyle{fontPixel, 8, '2', alignLeft})
	if m.IsDoctor() {
		quota := fmt.Sprintf("%d/%d", m.PatientCount(), game.DoctorQuota)
		if state == game.StaffNoQuota {
			e.Text(dst, quota, r.x+125, r.y+1, textStyle{fontPixelBold, 8, 'R', alignRight})
		} else {
			e.Text(dst, quota, r.x+125, r.y+1, textStyle{fontPixel, 8, '1', alignRight})
		}
	}
	e.DrawSprite(dst, a.Frame("ui_chip_estado", 4, chipFrame(state)), r.x+70, r.y+10, false)
	c.drawChipText(dst, state, r.x+72, r.y+10)
	e.FillRect(dst, rect{r.x + 70, r.y + 20, 58, 3}, 'k', 1)
	e.FillRect(dst, rect{r.x + 70, r.y + 20, int(58 * m.CooldownFraction()), 3}, 'q', 1)
}

// drawChipText escribe el estado dentro del chip ("Disponible"...).
func (c *camsScene) drawChipText(dst *ebiten.Image, state game.StaffState, x, y int) {
	key := byte('0')
	if state == game.StaffNoQuota {
		key = 'w'
	}
	c.env.Text(dst, chipText(state), x, y, textStyle{fontPixel, 8, key, alignLeft})
}

// drawHead pinta la cabeza (filas 0 a 11 del cuadro 1 del sprite quieto) de
// una persona del personal; sin arte, un círculo simple.
func (c *camsScene) drawHead(dst *ebiten.Image, m *game.StaffMember, x, y int) {
	e := c.env
	a := e.Assets
	prefix := "cam_"
	if m.IsDoctor() {
		prefix = "doc_"
	}
	name := prefix + spriteKey(m.ID(), m.Name()) + "_quieto"
	if img := a.Image(name); img != nil {
		head := img.SubImage(image.Rect(0, 0, 16, 12)).(*ebiten.Image)
		e.DrawSprite(dst, head, x, y, false)
		return
	}
	e.FillRect(dst, rect{x + 3, y, 10, 12}, '0', 1)
	e.FillRect(dst, rect{x + 4, y + 1, 8, 10}, 'C', 1)
}

// ---------------------------------------------------------------------------
// Mapa
// ---------------------------------------------------------------------------

func (c *camsScene) drawMap(dst *ebiten.Image) {
	e := c.env
	a := e.Assets
	e.DrawSprite(dst, a.Image("ui_mapa"), mapOrigin.x, mapOrigin.y, false)
	e.Text(dst, "Mapa", 348, 157, textStyle{fontPixelBold, 8, 'k', alignLeft})
	current := c.shift.CurrentCamera()
	for _, id := range mapCameras {
		r := mapButtonRect(id)
		locked := !c.shift.World().Has(id)
		e.DrawSprite(dst, a.Frame("ui_boton_cam", 3, buttonFrame(id == current, locked)), r.x, r.y, false)
		key := byte('1')
		switch {
		case id == current:
			key = '0'
		case locked:
			key = 'X'
		}
		if b, ok := c.blinks[id]; ok && id != current {
			// Parpadeo suave de una cámara donde alguien se durmió (mango) o se
			// despertó (menta). Solo en las cámaras que no se están mirando.
			if alpha := blinkAlpha(e.Tick - b.start); alpha > 0 {
				e.FillRect(dst, r, blinkKey(b.wake), alpha)
			}
		}
		if id != c.hoverLocked { // con el candado encima, el rótulo se esconde
			e.Text(dst, string(id), r.x+r.w/2, r.y, textStyle{fontVT, stampFontSize, key, alignCenter})
		}
	}
	e.Text(dst, "Monitoreo", 391, 233, textStyle{fontPixel, 8, 'k', alignLeft})

	// Candado (7x9) al pasar el mouse por una cámara bloqueada: el botón se
	// oscurece con la clave 0 al 40 % y el candado va en (x+5, y+1) del botón
	// (nota de Figma 18:176).
	if c.hoverLocked != "" {
		r := mapButtonRect(c.hoverLocked)
		e.FillRect(dst, r, '0', 0.4)
		e.DrawSprite(dst, a.Image("ui_candado"), r.x+5, r.y+1, false)
	}
}

// ---------------------------------------------------------------------------
// Avisos, Demo, pausa y reporte
// ---------------------------------------------------------------------------

// drawNotice pinta el aviso amable (por ejemplo "No hay cama libre!").
func (c *camsScene) drawNotice(dst *ebiten.Image) {
	box := noticeRect
	box.y = c.notice.y
	e := c.env
	drawBox(e, dst, box)
	e.Text(dst, c.notice.title, box.x+8, box.y+5, textStyle{fontPixelBold, 10, '2', alignLeft})
	for i, line := range c.notice.lines {
		e.Text(dst, line, box.x+8, box.y+20+11*i, textStyle{fontPixel, 8, '1', alignLeft})
	}
}

// drawDemoBox pinta la caja de la Demo con el paso y los dos botones.
func (c *camsScene) drawDemoBox(dst *ebiten.Image) {
	e := c.env
	drawBox(e, dst, demoBoxRect)
	e.DrawSprite(dst, e.Assets.Image("director_retrato"), 14, 207, false)
	info := demoParts[c.demo.Part()-1]
	e.Text(dst, info.title, 68, 206, textStyle{fontPixelBold, 10, '2', alignLeft})
	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	lines := wrapText(info.text, 258, measure)
	for i, line := range lines {
		if i >= 2 {
			break
		}
		e.Text(dst, line, 68, 220+11*i, textStyle{fontPixel, 8, '1', alignLeft})
	}
	c.drawButton(dst, demoNextRect, "Siguiente paso", 'm', c.hovered == "next")
	autoLabel, autoFill := "Auto", byte('k')
	if c.demoAuto {
		autoLabel, autoFill = "Pausa", 'm'
	}
	c.drawButton(dst, demoAutoRect, autoLabel, autoFill, c.hovered == "auto")
}

// stampFontSize es el tamaño de VT323 del rótulo y de los botones del mapa:
// 10 px a 1x, como dice la nota de Figma. Como el texto se dibuja a tamaño de
// pantalla (ver drawText), a 10 ya se ve nítido; el 13 anterior era un
// parche para el borroso de dibujarlo a 1x.
const stampFontSize = 10

// demoDischargeY es la y del botón "Dar de alta" en la Demo: la caja de la
// Demo ocupa desde y = 200, así que el botón sube para quedar entre la cama y
// la caja (y se dibuja encima de ella).
const demoDischargeY = 186

// drawDischargeButton pinta "Dar de alta" si hay un paciente despierto en la
// cama. En la Demo lo pulsa el guion: se resalta y al final se hunde.
func (c *camsScene) drawDischargeButton(dst *ebiten.Image) {
	if c.roomPatient() == nil {
		return
	}
	r := dischargeRect(c.bed())
	hover, pressed := c.hovered == "discharge", false
	if c.demo != nil {
		r.y = demoDischargeY
		hover, pressed = false, false
		if b, ok := c.demoBeat(); ok && b.Kind == game.BeatHighlight {
			hover, pressed = true, c.demo.BeatProgress() >= 0.75
		}
	}
	if hover && !pressed && c.demo != nil {
		// El aro amarillo avisa que este es el botón que la Demo va a pulsar.
		c.env.StrokeRect(dst, rect{r.x - 2, r.y - 2, r.w + 4, r.h + 5}, 'y', 1)
	}
	c.drawButtonState(dst, r, "Dar de alta", 'm', hover, pressed)
}
